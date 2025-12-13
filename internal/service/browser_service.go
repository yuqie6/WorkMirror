//go:build windows

package service

import (
	"context"
	"log/slog"
	"sync"

	"github.com/yuqie6/mirror/internal/handler"
	"github.com/yuqie6/mirror/internal/model"
)

// BrowserService 浏览器采集服务
type BrowserService struct {
	collector   *handler.BrowserCollector
	browserRepo BrowserEventRepository
	buffer      []*model.BrowserEvent
	bufferSize  int
	mu          sync.Mutex
	stopChan    chan struct{}
	wg          sync.WaitGroup
	running     bool
	onPersisted func(count int)
}

// NewBrowserService 创建浏览器服务
func NewBrowserService(
	collector *handler.BrowserCollector,
	browserRepo BrowserEventRepository,
) *BrowserService {
	return &BrowserService{
		collector:   collector,
		browserRepo: browserRepo,
		buffer:      make([]*model.BrowserEvent, 0, 100),
		bufferSize:  50,
		stopChan:    make(chan struct{}),
	}
}

func (s *BrowserService) SetOnPersisted(fn func(count int)) {
	s.onPersisted = fn
}

// Start 启动服务
func (s *BrowserService) Start(ctx context.Context) error {
	if s.running {
		return nil
	}

	s.running = true
	slog.Info("浏览器服务启动")

	if err := s.collector.Start(ctx); err != nil {
		return err
	}

	s.wg.Add(1)
	go s.processLoop(ctx)

	return nil
}

// Stop 停止服务
func (s *BrowserService) Stop() error {
	if !s.running {
		return nil
	}

	slog.Info("正在停止浏览器服务...")

	s.collector.Stop()
	close(s.stopChan)
	s.wg.Wait()

	// 刷新剩余数据
	s.flush(context.Background())

	s.running = false
	slog.Info("浏览器服务已停止")
	return nil
}

// processLoop 处理循环
func (s *BrowserService) processLoop(ctx context.Context) {
	defer s.wg.Done()

	events := s.collector.Events()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			s.handleEvent(ctx, event)
		}
	}
}

// handleEvent 处理事件
func (s *BrowserService) handleEvent(ctx context.Context, event *model.BrowserEvent) {
	s.mu.Lock()
	s.buffer = append(s.buffer, event)
	shouldFlush := len(s.buffer) >= s.bufferSize
	s.mu.Unlock()

	if shouldFlush {
		s.flush(ctx)
	}
}

// flush 刷新缓冲区
func (s *BrowserService) flush(ctx context.Context) {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return
	}

	events := s.buffer
	s.buffer = make([]*model.BrowserEvent, 0, 100)
	s.mu.Unlock()

	if err := s.browserRepo.BatchInsert(ctx, events); err != nil {
		slog.Error("保存浏览器事件失败", "error", err)
	} else {
		slog.Info("浏览器事件已保存", "count", len(events))
		if s.onPersisted != nil {
			s.onPersisted(len(events))
		}
	}
}
