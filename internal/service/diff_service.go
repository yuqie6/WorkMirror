//go:build windows

package service

import (
	"context"
	"log/slog"
	"sync"

	"github.com/yuqie6/mirror/internal/collector"
	"github.com/yuqie6/mirror/internal/schema"
)

// DiffService Diff 处理服务
type DiffService struct {
	collector   *collector.DiffCollector
	diffRepo    DiffRepository
	stopChan    chan struct{}
	wg          sync.WaitGroup
	running     bool
	onPersisted func(count int)
}

// NewDiffService 创建 Diff 服务
func NewDiffService(
	collector *collector.DiffCollector,
	diffRepo DiffRepository,
) *DiffService {
	return &DiffService{
		collector: collector,
		diffRepo:  diffRepo,
		stopChan:  make(chan struct{}),
	}
}

func (s *DiffService) SetOnPersisted(fn func(count int)) {
	s.onPersisted = fn
}

// Start 启动服务
func (s *DiffService) Start(ctx context.Context) error {
	if s.running {
		return nil
	}

	s.running = true
	slog.Info("Diff 服务启动")

	// 启动采集器
	if err := s.collector.Start(ctx); err != nil {
		return err
	}

	// 启动处理循环
	s.wg.Add(1)
	go s.processLoop(ctx)

	return nil
}

// Stop 停止服务
func (s *DiffService) Stop() error {
	if !s.running {
		return nil
	}

	slog.Info("正在停止 Diff 服务...")

	s.collector.Stop()
	close(s.stopChan)
	s.wg.Wait()

	s.running = false
	slog.Info("Diff 服务已停止")
	return nil
}

// processLoop 处理循环
func (s *DiffService) processLoop(ctx context.Context) {
	defer s.wg.Done()

	events := s.collector.Events()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case diff, ok := <-events:
			if !ok {
				return
			}
			s.handleDiff(ctx, diff)
		}
	}
}

// handleDiff 处理单个 Diff
func (s *DiffService) handleDiff(ctx context.Context, diff *schema.Diff) {
	// 保存到数据库
	if err := s.diffRepo.Create(ctx, diff); err != nil {
		slog.Error("保存 Diff 失败", "file", diff.FileName, "error", err)
		return
	}

	slog.Info("Diff 已记录",
		"file", diff.FileName,
		"language", diff.Language,
		"lines_added", diff.LinesAdded,
		"lines_deleted", diff.LinesDeleted,
	)
	if s.onPersisted != nil {
		s.onPersisted(1)
	}
}

// AddWatchPath 添加监控路径
func (s *DiffService) AddWatchPath(path string) error {
	return s.collector.AddWatchPath(path)
}
