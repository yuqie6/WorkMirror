//go:build windows

package bootstrap

import (
	"context"
	"time"

	"github.com/yuqie6/mirror/internal/collector"
	"github.com/yuqie6/mirror/internal/eventbus"
	"github.com/yuqie6/mirror/internal/pkg/privacy"
	"github.com/yuqie6/mirror/internal/service"
)

// AgentRuntime 包含 Agent 二进制需要启动的采集与后台任务
type AgentRuntime struct {
	*Core
	Hub *eventbus.Hub

	Collectors struct {
		Window  collector.Collector
		Diff    *collector.DiffCollector
		Browser *collector.BrowserCollector
	}

	Services struct {
		Tracker *service.TrackerService
		Diff    *service.DiffService
		Browser *service.BrowserService
		RAG     *service.RAGService
	}
}

// NewAgentRuntime 构建 Agent 运行时并启动采集服务
func NewAgentRuntime(ctx context.Context, cfgPath string) (*AgentRuntime, error) {
	core, err := NewCore(cfgPath)
	if err != nil {
		return nil, err
	}

	rt := &AgentRuntime{Core: core, Hub: eventbus.NewHub()}

	if core.DB != nil && core.DB.SafeMode {
		// 安全模式：允许 UI 启动与诊断导出，但不启动任何写库链路（采集/后台任务）。
		// 具体原因由 /api/status 展示，避免“沉默失败”。
		return rt, nil
	}

	sanitizer := privacy.New(core.Cfg.Privacy.Enabled, core.Cfg.Privacy.Patterns)

	// Window collector + tracker
	rt.Collectors.Window = collector.NewWindowCollector(&collector.CollectorConfig{
		PollIntervalMs: core.Cfg.Collector.PollIntervalMs,
		MinDurationSec: core.Cfg.Collector.MinDurationSec,
		MaxDurationSec: 60,
		BufferSize:     core.Cfg.Collector.BufferSize,
	})
	rt.Services.Tracker = service.NewTrackerService(rt.Collectors.Window, core.Repos.Event, &service.TrackerConfig{
		FlushBatchSize:   core.Cfg.Collector.FlushBatchSize,
		FlushIntervalSec: core.Cfg.Collector.FlushIntervalSec,
		Sanitizer:        sanitizer,
		OnWriteSuccess: func(count int) {
			rt.Hub.Publish(eventbus.Event{
				Type: "data_changed",
				Data: map[string]any{"source": "events", "count": count},
			})
		},
	})
	if err := rt.Services.Tracker.Start(ctx); err != nil {
		core.Close()
		return nil, err
	}

	// Diff collector + service (optional)
	if core.Cfg.Diff.Enabled && len(core.Cfg.Diff.WatchPaths) > 0 {
		diffCollector, err := collector.NewDiffCollector(&collector.DiffCollectorConfig{
			WatchPaths:  core.Cfg.Diff.WatchPaths,
			Extensions:  core.Cfg.Diff.Extensions,
			BufferSize:  core.Cfg.Diff.BufferSize,
			DebounceSec: core.Cfg.Diff.DebounceSec,
		})
		if err != nil {
			rt.Close()
			return nil, err
		}
		for _, path := range core.Cfg.Diff.WatchPaths {
			_ = diffCollector.AddWatchPath(path)
		}
		rt.Collectors.Diff = diffCollector
		rt.Services.Diff = service.NewDiffService(diffCollector, core.Repos.Diff)
		rt.Services.Diff.SetOnPersisted(func(count int) {
			rt.Hub.Publish(eventbus.Event{
				Type: "data_changed",
				Data: map[string]any{"source": "diffs", "count": count},
			})
		})
		if err := rt.Services.Diff.Start(ctx); err != nil {
			rt.Close()
			return nil, err
		}
	}

	// RAG (optional)
	if core.Clients.SiliconFlow != nil {
		rag, err := service.NewRAGService(core.Clients.SiliconFlow, core.Repos.Summary, core.Repos.Diff, nil)
		if err == nil {
			rt.Services.RAG = rag
			core.Services.AI.SetRAGService(rag)
			if core.Services.SessionSemantic != nil {
				core.Services.SessionSemantic.SetRAG(rag)
			}
		}
	}

	// Browser collector + service (optional)
	if core.Cfg.Browser.Enabled {
		bc, err := collector.NewBrowserCollector(&collector.BrowserCollectorConfig{
			HistoryPath:  core.Cfg.Browser.HistoryPath,
			PollInterval: time.Duration(core.Cfg.Browser.PollIntervalSec) * time.Second,
		})
		if err == nil {
			rt.Collectors.Browser = bc
			rt.Services.Browser = service.NewBrowserService(bc, core.Repos.Browser)
			rt.Services.Browser.SetSanitizer(sanitizer)
			rt.Services.Browser.SetOnPersisted(func(count int) {
				rt.Hub.Publish(eventbus.Event{
					Type: "data_changed",
					Data: map[string]any{"source": "browser_events", "count": count},
				})
			})
			_ = rt.Services.Browser.Start(ctx)
		}
	}

	// AI 定时分析（optional）
	if core.Clients.DeepSeek != nil && core.Clients.DeepSeek.IsConfigured() {
		go runPeriodic(ctx, 5*time.Minute, func() { analyzeWithRetry(ctx, core.Services.AI) })
	}

	// Session 定时切分（可离线，无需 AI）
	if core.Services.Sessions != nil {
		go runPeriodic(ctx, 5*time.Minute, func() { splitWithRetry(ctx, core.Services.Sessions) })
	}

	// Session 语义补全（用于证据链，DeepSeek 未配置时自动降级为规则摘要）
	if core.Services.SessionSemantic != nil {
		go runPeriodic(ctx, 10*time.Minute, func() { enrichWithRetry(ctx, core.Services.SessionSemantic) })
	}

	// Skill 衰减（本地规则，可离线）
	if core.Services.Skills != nil {
		go runPeriodic(ctx, 24*time.Hour, func() {
			_ = core.Services.Skills.ApplyDecayToAll(context.Background())
		})
	}

	return rt, nil
}

// Close 关闭 Agent 运行时资源
func (rt *AgentRuntime) Close() error {
	if rt == nil {
		return nil
	}
	if rt.Services.Tracker != nil {
		_ = rt.Services.Tracker.Stop()
	}
	if rt.Services.Diff != nil {
		_ = rt.Services.Diff.Stop()
	}
	if rt.Services.Browser != nil {
		_ = rt.Services.Browser.Stop()
	}
	if rt.Services.RAG != nil {
		_ = rt.Services.RAG.Close()
	}
	return rt.Core.Close()
}

// runPeriodic 定时执行函数
func runPeriodic(ctx context.Context, interval time.Duration, fn func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fn()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn()
		}
	}
}

// analyzeWithRetry 带重试的 Diff 分析
func analyzeWithRetry(ctx context.Context, aiService *service.AIService) {
	if aiService == nil {
		return
	}
	_, _ = aiService.AnalyzePendingDiffs(ctx, 10)
}

// splitWithRetry 带重试的会话切分
func splitWithRetry(ctx context.Context, sessionService *service.SessionService) {
	if sessionService == nil {
		return
	}
	_, _ = sessionService.BuildSessionsIncremental(ctx)
}

// enrichWithRetry 带重试的会话语义补全
func enrichWithRetry(ctx context.Context, svc *service.SessionSemanticService) {
	if svc == nil {
		return
	}
	_, _ = svc.EnrichSessionsIncremental(ctx, &service.SessionSemanticServiceConfig{
		Lookback: 48 * time.Hour,
		Limit:    20,
	})
}
