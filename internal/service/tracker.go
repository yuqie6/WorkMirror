//go:build windows

package service

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yuqie6/mirror/internal/collector"
	"github.com/yuqie6/mirror/internal/pkg/privacy"
	"github.com/yuqie6/mirror/internal/schema"
)

// TrackerService 追踪服务 - 负责接收事件并批量写入数据库
type TrackerService struct {
	collector      collector.Collector
	eventRepo      EventRepository
	buffer         []schema.Event
	bufferMu       sync.Mutex
	flushBatchSize int
	flushInterval  time.Duration
	stopChan       chan struct{}
	wg             sync.WaitGroup
	running        atomic.Bool // 并发安全的状态标识
	onWriteSuccess func(count int)
	sanitizer      *privacy.Sanitizer

	// 有界写入队列
	writeChan     chan []schema.Event
	writerDone    chan struct{}
	writeQueueCap int

	lastPersistAt  atomic.Int64
	lastErrorAt    atomic.Int64
	lastErrorMsg   atomic.Value // string
	writeErrors    atomic.Int64
	droppedBatches atomic.Int64
}

// TrackerConfig 追踪服务配置
type TrackerConfig struct {
	FlushBatchSize   int // 批量写入阈值
	FlushIntervalSec int // 强制刷新间隔（秒）
	OnWriteSuccess   func(count int)
	Sanitizer        *privacy.Sanitizer
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
	collector collector.Collector,
	eventRepo EventRepository,
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
		buffer:         make([]schema.Event, 0, cfg.FlushBatchSize),
		flushBatchSize: cfg.FlushBatchSize,
		flushInterval:  time.Duration(cfg.FlushIntervalSec) * time.Second,
		stopChan:       make(chan struct{}),
		writeChan:      make(chan []schema.Event, writeQueueCap),
		writerDone:     make(chan struct{}),
		writeQueueCap:  writeQueueCap,
		onWriteSuccess: cfg.OnWriteSuccess,
		sanitizer:      cfg.Sanitizer,
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

	// 尽量把 collector 的剩余缓冲也处理掉（Stop 场景优先“尽量不丢”而不是“保活”）
	t.drainCollectorEvents()

	// 最终刷新：将缓冲区剩余事件阻塞写入队列（Stop 场景避免丢批次）
	t.flushToWriterBlocking()

	// 关闭写入队列，等待 writer 完成所有写入
	close(t.writeChan)
	<-t.writerDone

	t.running.Store(false)
	slog.Info("追踪服务已停止")

	return nil
}

func (t *TrackerService) drainCollectorEvents() {
	if t == nil || t.collector == nil {
		return
	}
	ch := t.collector.Events()
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return
			}
			t.appendEventNoFlush(ev)
		default:
			return
		}
	}
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
			t.writeErrors.Add(1)
			t.lastErrorAt.Store(time.Now().UnixMilli())
			t.lastErrorMsg.Store(err.Error())
			slog.Error("批量写入事件失败", "count", len(events), "error", err)
			continue
		}
		t.lastPersistAt.Store(time.Now().UnixMilli())
		slog.Debug("批量写入事件成功", "count", len(events))
		if t.onWriteSuccess != nil {
			t.onWriteSuccess(len(events))
		}
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
func (t *TrackerService) handleEvent(event *schema.Event) {
	if t.sanitizer != nil && event != nil {
		event.Title = t.sanitizer.SanitizeWindowTitle(event.Title)
	}

	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()

	t.buffer = append(t.buffer, *event)

	// 检查是否达到批量写入阈值
	if len(t.buffer) >= t.flushBatchSize {
		t.flushLockedToWriter()
	}
}

func (t *TrackerService) appendEventNoFlush(event *schema.Event) {
	if t == nil || event == nil {
		return
	}
	if t.sanitizer != nil {
		event.Title = t.sanitizer.SanitizeWindowTitle(event.Title)
	}
	t.bufferMu.Lock()
	t.buffer = append(t.buffer, *event)
	t.bufferMu.Unlock()
}

// flushToWriter 刷新缓冲区到写入队列（带锁）
func (t *TrackerService) flushToWriter() {
	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()
	t.flushLockedToWriter()
}

// flushToWriterBlocking 刷新缓冲区到写入队列（阻塞版本，仅用于 Stop 阶段）
func (t *TrackerService) flushToWriterBlocking() {
	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()

	if len(t.buffer) == 0 {
		return
	}

	events := make([]schema.Event, len(t.buffer))
	copy(events, t.buffer)
	t.buffer = t.buffer[:0]

	// Stop 阶段阻塞写入，尽量不丢
	t.writeChan <- events
}

// flushLockedToWriter 刷新缓冲区到写入队列（无锁版本，调用者必须持有锁）
func (t *TrackerService) flushLockedToWriter() {
	if len(t.buffer) == 0 {
		return
	}

	// 复制缓冲区
	events := make([]schema.Event, len(t.buffer))
	copy(events, t.buffer)

	// 清空缓冲区
	t.buffer = t.buffer[:0]

	// 发送到写入队列（有界，满时丢弃以避免阻塞采集链路）
	select {
	case t.writeChan <- events:
	default:
		t.droppedBatches.Add(1)
		slog.Warn("事件写入队列已满，丢弃批次", "count", len(events), "queue_cap", t.writeQueueCap)
	}
}

// Stats 返回服务统计信息
func (t *TrackerService) Stats() TrackerStats {
	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()

	return TrackerStats{
		BufferSize:     len(t.buffer),
		WriteQueueLen:  len(t.writeChan),
		WriteQueueCap:  t.writeQueueCap,
		Running:        t.running.Load(),
		LastPersistAt:  t.lastPersistAt.Load(),
		WriteErrors:    t.writeErrors.Load(),
		DroppedBatches: t.droppedBatches.Load(),
		LastErrorAt:    t.lastErrorAt.Load(),
		LastError:      loadAtomicString(&t.lastErrorMsg),
	}
}

// TrackerStats 追踪服务统计
type TrackerStats struct {
	BufferSize     int
	WriteQueueLen  int
	WriteQueueCap  int
	Running        bool
	LastPersistAt  int64
	WriteErrors    int64
	DroppedBatches int64
	LastErrorAt    int64
	LastError      string
}

func loadAtomicString(v *atomic.Value) string {
	if v == nil {
		return ""
	}
	raw := v.Load()
	s, _ := raw.(string)
	return s
}
