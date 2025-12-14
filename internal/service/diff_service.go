//go:build windows

package service

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

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

	lastPersistAt atomic.Int64
	persistErrors atomic.Int64
	lastErrorAt   atomic.Int64
	lastErrorMsg  atomic.Value // string
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

// SetOnPersisted 设置持久化后的回调函数
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
		s.persistErrors.Add(1)
		s.lastErrorAt.Store(time.Now().UnixMilli())
		s.lastErrorMsg.Store(err.Error())
		slog.Error("保存 Diff 失败", "file", diff.FileName, "error", err)
		return
	}
	s.lastPersistAt.Store(time.Now().UnixMilli())

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

type DiffServiceStats struct {
	Running       bool   `json:"running"`
	LastPersistAt int64  `json:"last_persist_at"`
	PersistErrors int64  `json:"persist_errors"`
	LastErrorAt   int64  `json:"last_error_at"`
	LastError     string `json:"last_error"`
}

func (s *DiffService) Stats() DiffServiceStats {
	if s == nil {
		return DiffServiceStats{}
	}
	raw := s.lastErrorMsg.Load()
	msg, _ := raw.(string)
	return DiffServiceStats{
		Running:       s.running,
		LastPersistAt: s.lastPersistAt.Load(),
		PersistErrors: s.persistErrors.Load(),
		LastErrorAt:   s.lastErrorAt.Load(),
		LastError:     msg,
	}
}

// AddWatchPath 添加监控路径
func (s *DiffService) AddWatchPath(path string) error {
	return s.collector.AddWatchPath(path)
}
