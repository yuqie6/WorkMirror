//go:build windows

package service

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yuqie6/mirror/internal/handler"
	"github.com/yuqie6/mirror/internal/model"
	"github.com/yuqie6/mirror/internal/repository"
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
	running        atomic.Bool // 并发安全的状态标识

	// 有界写入队列
	writeChan     chan []model.Event
	writerDone    chan struct{}
	writeQueueCap int
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

	// 写入队列容量：允许积压 10 个批次
	writeQueueCap := 10

	return &TrackerService{
		collector:      collector,
		eventRepo:      eventRepo,
		buffer:         make([]model.Event, 0, cfg.FlushBatchSize),
		flushBatchSize: cfg.FlushBatchSize,
		flushInterval:  time.Duration(cfg.FlushIntervalSec) * time.Second,
		stopChan:       make(chan struct{}),
		writeChan:      make(chan []model.Event, writeQueueCap),
		writerDone:     make(chan struct{}),
		writeQueueCap:  writeQueueCap,
	}
}

// Start 启动追踪服务
func (t *TrackerService) Start(ctx context.Context) error {
	if t.running.Load() {
		return nil
	}

	t.running.Store(true)
	slog.Info("追踪服务启动",
		"flush_batch_size", t.flushBatchSize,
		"flush_interval", t.flushInterval,
		"write_queue_cap", t.writeQueueCap,
	)

	// 启动采集器
	if err := t.collector.Start(ctx); err != nil {
		t.running.Store(false)
		return err
	}

	// 启动写入协程（单一 writer，避免并发写库冲突）
	go t.writerLoop(ctx)

	// 启动事件处理循环
	t.wg.Add(1)
	go t.processLoop(ctx)

	return nil
}

// Stop 停止追踪服务（等待所有事件落库）
func (t *TrackerService) Stop() error {
	if !t.running.Load() {
		return nil
	}

	slog.Info("正在停止追踪服务...")

	// 停止采集器
	t.collector.Stop()

	// 发送停止信号，等待处理循环结束
	close(t.stopChan)
	t.wg.Wait()

	// 最终刷新：将缓冲区剩余事件发送到写入队列
	t.flushToWriter()

	// 关闭写入队列，等待 writer 完成所有写入
	close(t.writeChan)
	<-t.writerDone

	t.running.Store(false)
	slog.Info("追踪服务已停止")

	return nil
}

// writerLoop 单一写入协程，消费 writeChan 同步写库
func (t *TrackerService) writerLoop(ctx context.Context) {
	defer close(t.writerDone)

	// 写库使用独立的 background ctx，避免外部 cancel 影响 Stop 时的数据落库
	writeCtx := context.Background()
	for events := range t.writeChan {
		if len(events) == 0 {
			continue
		}

		// 同步写入数据库
		if err := t.eventRepo.BatchInsert(writeCtx, events); err != nil {
			slog.Error("批量写入事件失败", "count", len(events), "error", err)
			continue
		}
		slog.Debug("批量写入事件成功", "count", len(events))
	}
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
			t.flushToWriter()
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
		t.flushLockedToWriter()
	}
}

// flushToWriter 刷新缓冲区到写入队列（带锁）
func (t *TrackerService) flushToWriter() {
	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()
	t.flushLockedToWriter()
}

// flushLockedToWriter 刷新缓冲区到写入队列（无锁版本，调用者必须持有锁）
func (t *TrackerService) flushLockedToWriter() {
	if len(t.buffer) == 0 {
		return
	}

	// 复制缓冲区
	events := make([]model.Event, len(t.buffer))
	copy(events, t.buffer)

	// 清空缓冲区
	t.buffer = t.buffer[:0]

	// 发送到写入队列（有界，满时会阻塞）
	t.writeChan <- events
}

// Stats 返回服务统计信息
func (t *TrackerService) Stats() TrackerStats {
	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()

	return TrackerStats{
		BufferSize:    len(t.buffer),
		WriteQueueLen: len(t.writeChan),
		WriteQueueCap: t.writeQueueCap,
		Running:       t.running.Load(),
	}
}

// TrackerStats 追踪服务统计
type TrackerStats struct {
	BufferSize    int  `json:"buffer_size"`
	WriteQueueLen int  `json:"write_queue_len"`
	WriteQueueCap int  `json:"write_queue_cap"`
	Running       bool `json:"running"`
}
