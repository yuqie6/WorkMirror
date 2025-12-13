//go:build windows

package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/yuqie6/mirror/internal/schema"
)

// Collector 采集器接口
type Collector interface {
	// Start 启动采集
	Start(ctx context.Context) error
	// Stop 停止采集
	Stop() error
	// Events 返回事件通道
	Events() <-chan *schema.Event
}

// WindowCollector Windows 窗口采集器
type WindowCollector struct {
	pollInterval time.Duration // 轮询间隔
	minDuration  time.Duration // 最小记录时长
	maxDuration  time.Duration // 单段最大时长（用于持续窗口的心跳落库）
	eventChan    chan *schema.Event
	stopChan     chan struct{}
	lastWindow   *WindowInfo
	lastSwitchAt time.Time
	currentStart time.Time
	running      bool
	stopOnce     sync.Once  // 确保 Stop 只执行一次
	mu           sync.Mutex // 保护 running 状态
}

// CollectorConfig 采集器配置
type CollectorConfig struct {
	PollIntervalMs int // 轮询间隔（毫秒）
	MinDurationSec int // 最小记录时长（秒）
	MaxDurationSec int // 单段最大时长（秒），超过则强制落库并重新计时（0=默认60）
	BufferSize     int // 事件缓冲区大小
}

// DefaultCollectorConfig 默认配置
func DefaultCollectorConfig() *CollectorConfig {
	return &CollectorConfig{
		PollIntervalMs: 500,
		MinDurationSec: 3,
		MaxDurationSec: 60,
		BufferSize:     2048,
	}
}

// NewWindowCollector 创建窗口采集器
func NewWindowCollector(cfg *CollectorConfig) *WindowCollector {
	if cfg == nil {
		cfg = DefaultCollectorConfig()
	}
	if cfg.MaxDurationSec <= 0 {
		cfg.MaxDurationSec = 60
	}

	return &WindowCollector{
		pollInterval: time.Duration(cfg.PollIntervalMs) * time.Millisecond,
		minDuration:  time.Duration(cfg.MinDurationSec) * time.Second,
		maxDuration:  time.Duration(cfg.MaxDurationSec) * time.Second,
		eventChan:    make(chan *schema.Event, cfg.BufferSize),
		stopChan:     make(chan struct{}),
	}
}

// Start 启动采集
func (c *WindowCollector) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()

	slog.Info("窗口采集器启动", "poll_interval", c.pollInterval, "min_duration", c.minDuration)
	go c.pollLoop(ctx)
	return nil
}

// Stop 停止采集（线程安全，可重复调用）
func (c *WindowCollector) Stop() error {
	c.stopOnce.Do(func() {
		c.mu.Lock()
		if !c.running {
			c.mu.Unlock()
			return
		}
		c.running = false
		c.mu.Unlock()

		close(c.stopChan)

		// 记录最后一个窗口的时长
		if c.lastWindow != nil {
			c.emitEvent(c.lastWindow, time.Since(c.currentStart))
		}

		slog.Info("窗口采集器已停止")
	})
	return nil
}

// Events 返回事件通道
func (c *WindowCollector) Events() <-chan *schema.Event {
	return c.eventChan
}

// pollLoop 轮询循环
func (c *WindowCollector) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.Stop()
			return
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.poll()
		}
	}
}

// poll 执行一次轮询
func (c *WindowCollector) poll() {
	current, err := GetForegroundWindowInfo()
	if err != nil {
		slog.Debug("获取窗口信息失败", "error", err)
		return
	}

	// 忽略系统窗口
	if current.IsSystemWindow() {
		return
	}

	now := time.Now()

	// 首次记录
	if c.lastWindow == nil {
		c.lastWindow = current
		c.currentStart = now
		c.lastSwitchAt = now
		slog.Debug("开始追踪窗口", "window", current.String())
		return
	}

	// 窗口未变化，继续累计时长
	if current.IsSameWindow(c.lastWindow) {
		// 心跳：持续停留同一窗口时也定期落库，避免会话切分看不到“正在进行”的长事件
		duration := now.Sub(c.currentStart)
		if c.maxDuration > 0 && duration >= c.maxDuration && duration >= c.minDuration {
			c.emitEvent(c.lastWindow, duration)
			c.currentStart = now
		}
		return
	}

	// 窗口已切换，计算上一个窗口的持续时长
	duration := now.Sub(c.currentStart)

	// 只有超过最小时长的才记录
	if duration >= c.minDuration {
		c.emitEvent(c.lastWindow, duration)
	} else {
		slog.Debug("窗口停留时间过短，已忽略",
			"window", c.lastWindow.String(),
			"duration", duration,
		)
	}

	// 更新状态
	c.lastWindow = current
	c.currentStart = now
	c.lastSwitchAt = now
}

// emitEvent 发送事件
func (c *WindowCollector) emitEvent(w *WindowInfo, duration time.Duration) {
	event := &schema.Event{
		Timestamp: c.currentStart.UnixMilli(),
		Source:    "window",
		AppName:   w.AppName,
		Title:     w.Title,
		Duration:  int(duration.Seconds()),
		Metadata:  make(schema.JSONMap),
	}

	// 添加元数据
	event.Metadata["process_id"] = w.ProcessID
	event.Metadata["hwnd"] = w.HWND

	select {
	case c.eventChan <- event:
		slog.Debug("事件已发送",
			"app", w.AppName,
			"title", w.Title,
			"duration", duration,
		)
	default:
		slog.Warn("事件缓冲区已满，丢弃事件", "app", w.AppName)
	}
}
