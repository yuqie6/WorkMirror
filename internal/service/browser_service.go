//go:build windows

package service

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yuqie6/WorkMirror/internal/collector"
	"github.com/yuqie6/WorkMirror/internal/pkg/privacy"
	"github.com/yuqie6/WorkMirror/internal/schema"
)

// BrowserService 浏览器采集服务
type BrowserService struct {
	collector   *collector.BrowserCollector
	browserRepo BrowserEventRepository
	buffer      []*schema.BrowserEvent
	bufferSize  int
	mu          sync.Mutex
	stopChan    chan struct{}
	wg          sync.WaitGroup
	running     bool
	onPersisted func(count int)
	sanitizer   *privacy.Sanitizer

	lastPersistAt atomic.Int64
	persistErrors atomic.Int64
	lastErrorAt   atomic.Int64
	lastErrorMsg  atomic.Value // string
}

// NewBrowserService 创建浏览器服务
func NewBrowserService(
	collector *collector.BrowserCollector,
	browserRepo BrowserEventRepository,
) *BrowserService {
	return &BrowserService{
		collector:   collector,
		browserRepo: browserRepo,
		buffer:      make([]*schema.BrowserEvent, 0, 100),
		bufferSize:  50,
		stopChan:    make(chan struct{}),
	}
}

// SetOnPersisted 设置持久化后的回调函数
func (s *BrowserService) SetOnPersisted(fn func(count int)) {
	s.onPersisted = fn
}

func (s *BrowserService) SetSanitizer(z *privacy.Sanitizer) {
	s.sanitizer = z
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
func (s *BrowserService) handleEvent(ctx context.Context, event *schema.BrowserEvent) {
	if s.sanitizer != nil && event != nil {
		event.Title = s.sanitizer.SanitizeBrowserTitle(event.Title)
		event.URL = s.sanitizer.SanitizeURL(event.URL)
	}

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
	s.buffer = make([]*schema.BrowserEvent, 0, 100)
	s.mu.Unlock()

	if err := s.browserRepo.BatchInsert(ctx, events); err != nil {
		s.persistErrors.Add(1)
		s.lastErrorAt.Store(time.Now().UnixMilli())
		s.lastErrorMsg.Store(err.Error())
		slog.Error("保存浏览器事件失败", "error", err)
	} else {
		s.lastPersistAt.Store(time.Now().UnixMilli())
		slog.Info("浏览器事件已保存", "count", len(events))
		if s.onPersisted != nil {
			s.onPersisted(len(events))
		}
	}
}

type BrowserServiceStats struct {
	Running       bool   `json:"running"`
	LastPersistAt int64  `json:"last_persist_at"`
	PersistErrors int64  `json:"persist_errors"`
	LastErrorAt   int64  `json:"last_error_at"`
	LastError     string `json:"last_error"`
}

func (s *BrowserService) Stats() BrowserServiceStats {
	if s == nil {
		return BrowserServiceStats{}
	}
	raw := s.lastErrorMsg.Load()
	msg, _ := raw.(string)
	return BrowserServiceStats{
		Running:       s.running,
		LastPersistAt: s.lastPersistAt.Load(),
		PersistErrors: s.persistErrors.Load(),
		LastErrorAt:   s.lastErrorAt.Load(),
		LastError:     msg,
	}
}
