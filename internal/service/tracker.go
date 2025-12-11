package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/danielsclee/mirror/internal/handler"
	"github.com/danielsclee/mirror/internal/model"
	"github.com/danielsclee/mirror/internal/repository"
)

// TrackerService 追踪服务 - 负责接收事件并批量写入数据库
type TrackerService struct {
	collector      handler.Collector
	eventRepo      *repository.EventRepository
	buffer         []model.Event
	bufferMu       sync.Mutex
	flushBatchSize int
	flushInterval  time.Duration
	stopChan       chan struct{}
	wg             sync.WaitGroup
	running        bool
}

// TrackerConfig 追踪服务配置
type TrackerConfig struct {
	FlushBatchSize   int // 批量写入阈值
	FlushIntervalSec int // 强制刷新间隔（秒）
}

// DefaultTrackerConfig 默认配置
func DefaultTrackerConfig() *TrackerConfig {
	return &TrackerConfig{
		FlushBatchSize:   100,
		FlushIntervalSec: 5,
	}
}

// NewTrackerService 创建追踪服务
func NewTrackerService(
	collector handler.Collector,
	eventRepo *repository.EventRepository,
	cfg *TrackerConfig,
) *TrackerService {
	if cfg == nil {
		cfg = DefaultTrackerConfig()
	}

	return &TrackerService{
		collector:      collector,
		eventRepo:      eventRepo,
		buffer:         make([]model.Event, 0, cfg.FlushBatchSize),
		flushBatchSize: cfg.FlushBatchSize,
		flushInterval:  time.Duration(cfg.FlushIntervalSec) * time.Second,
		stopChan:       make(chan struct{}),
	}
}

// Start 启动追踪服务
func (t *TrackerService) Start(ctx context.Context) error {
	if t.running {
		return nil
	}

	t.running = true
	slog.Info("追踪服务启动",
		"flush_batch_size", t.flushBatchSize,
		"flush_interval", t.flushInterval,
	)

	// 启动采集器
	if err := t.collector.Start(ctx); err != nil {
		return err
	}

	// 启动事件处理循环
	t.wg.Add(1)
	go t.processLoop(ctx)

	return nil
}

// Stop 停止追踪服务
func (t *TrackerService) Stop() error {
	if !t.running {
		return nil
	}

	slog.Info("正在停止追踪服务...")

	// 停止采集器
	t.collector.Stop()

	// 发送停止信号
	close(t.stopChan)

	// 等待处理循环结束
	t.wg.Wait()

	// 最终刷新
	t.flush(context.Background())

	t.running = false
	slog.Info("追踪服务已停止")

	return nil
}

// processLoop 事件处理循环
func (t *TrackerService) processLoop(ctx context.Context) {
	defer t.wg.Done()

	ticker := time.NewTicker(t.flushInterval)
	defer ticker.Stop()

	events := t.collector.Events()

	for {
		select {
		case <-ctx.Done():
			return

		case <-t.stopChan:
			return

		case event, ok := <-events:
			if !ok {
				return
			}
			t.handleEvent(event)

		case <-ticker.C:
			// 定时刷新
			t.flush(ctx)
		}
	}
}

// handleEvent 处理单个事件
func (t *TrackerService) handleEvent(event *model.Event) {
	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()

	t.buffer = append(t.buffer, *event)

	// 检查是否达到批量写入阈值
	if len(t.buffer) >= t.flushBatchSize {
		t.flushLocked(context.Background())
	}
}

// flush 刷新缓冲区（带锁）
func (t *TrackerService) flush(ctx context.Context) {
	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()
	t.flushLocked(ctx)
}

// flushLocked 刷新缓冲区（无锁版本，调用者必须持有锁）
func (t *TrackerService) flushLocked(ctx context.Context) {
	if len(t.buffer) == 0 {
		return
	}

	// 复制缓冲区
	events := make([]model.Event, len(t.buffer))
	copy(events, t.buffer)

	// 清空缓冲区
	t.buffer = t.buffer[:0]

	// 异步写入数据库
	go func() {
		if err := t.eventRepo.BatchInsert(ctx, events); err != nil {
			slog.Error("批量写入事件失败", "count", len(events), "error", err)
			return
		}
		slog.Debug("批量写入事件成功", "count", len(events))
	}()
}

// Stats 返回服务统计信息
func (t *TrackerService) Stats() TrackerStats {
	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()

	return TrackerStats{
		BufferSize: len(t.buffer),
		Running:    t.running,
	}
}

// TrackerStats 追踪服务统计
type TrackerStats struct {
	BufferSize int  `json:"buffer_size"`
	Running    bool `json:"running"`
}
