//go:build windows

package bootstrap

import (
	"context"
	"time"

	"github.com/yuqie6/mirror/internal/eventbus"
	"github.com/yuqie6/mirror/internal/handler"
	"github.com/yuqie6/mirror/internal/service"
)

// AgentRuntime 包含 Agent 二进制需要启动的采集与后台任务
type AgentRuntime struct {
	*Core
	Hub *eventbus.Hub

	Collectors struct {
		Window  handler.Collector
		Diff    *handler.DiffCollector
		Browser *handler.BrowserCollector
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

	// Window collector + tracker
	rt.Collectors.Window = handler.NewWindowCollector(&handler.CollectorConfig{
		PollIntervalMs: core.Cfg.Collector.PollIntervalMs,
		MinDurationSec: core.Cfg.Collector.MinDurationSec,
		MaxDurationSec: 60,
		BufferSize:     core.Cfg.Collector.BufferSize,
	})
	rt.Services.Tracker = service.NewTrackerService(rt.Collectors.Window, core.Repos.Event, &service.TrackerConfig{
		FlushBatchSize:   core.Cfg.Collector.FlushBatchSize,
		FlushIntervalSec: core.Cfg.Collector.FlushIntervalSec,
		OnWriteSuccess: func(count int) {
			rt.Hub.Publish(eventbus.Event{
				Type: "events_persisted",
				Data: map[string]any{"count": count},
			})
		},
	})
	if err := rt.Services.Tracker.Start(ctx); err != nil {
		core.Close()
		return nil, err
	}

	// Diff collector + service (optional)
	if core.Cfg.Diff.Enabled && len(core.Cfg.Diff.WatchPaths) > 0 {
		diffCollector, err := handler.NewDiffCollector(&handler.DiffCollectorConfig{
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
				Type: "diffs_persisted",
				Data: map[string]any{"count": count},
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
		bc, err := handler.NewBrowserCollector(&handler.BrowserCollectorConfig{
			HistoryPath:  core.Cfg.Browser.HistoryPath,
			PollInterval: time.Duration(core.Cfg.Browser.PollIntervalSec) * time.Second,
		})
		if err == nil {
			rt.Collectors.Browser = bc
			rt.Services.Browser = service.NewBrowserService(bc, core.Repos.Browser)
			rt.Services.Browser.SetOnPersisted(func(count int) {
				rt.Hub.Publish(eventbus.Event{
					Type: "browser_events_persisted",
					Data: map[string]any{"count": count},
				})
			})
			_ = rt.Services.Browser.Start(ctx)
		}
	}

	// AI 定时分析（optional）
	if core.Clients.DeepSeek != nil && core.Clients.DeepSeek.IsConfigured() {
		go runAIAnalysisLoop(ctx, core.Services.AI, 5*time.Minute)
	}

	// Session 定时切分（可离线，无需 AI）
	if core.Services.Sessions != nil {
		go runSessionSplitLoop(ctx, core.Services.Sessions, 5*time.Minute)
	}

	// Session 语义补全（用于证据链，DeepSeek 未配置时自动降级为规则摘要）
	if core.Services.SessionSemantic != nil {
		go runSessionSemanticLoop(ctx, core.Services.SessionSemantic, 10*time.Minute)
	}

	return rt, nil
}

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

// runAIAnalysisLoop 定时运行 AI 分析（与 cmd 层解耦）
func runAIAnalysisLoop(ctx context.Context, aiService *service.AIService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	analyzeWithRetry(ctx, aiService)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			analyzeWithRetry(ctx, aiService)
		}
	}
}

func analyzeWithRetry(ctx context.Context, aiService *service.AIService) {
	if aiService == nil {
		return
	}
	_, _ = aiService.AnalyzePendingDiffs(ctx, 10)
}

// runSessionSplitLoop 定时运行会话切分（增量）
func runSessionSplitLoop(ctx context.Context, sessionService *service.SessionService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	splitWithRetry(ctx, sessionService)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			splitWithRetry(ctx, sessionService)
		}
	}
}

func splitWithRetry(ctx context.Context, sessionService *service.SessionService) {
	if sessionService == nil {
		return
	}
	_, _ = sessionService.BuildSessionsIncremental(ctx)
}

func runSessionSemanticLoop(ctx context.Context, svc *service.SessionSemanticService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	enrichWithRetry(ctx, svc)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			enrichWithRetry(ctx, svc)
		}
	}
}

func enrichWithRetry(ctx context.Context, svc *service.SessionSemanticService) {
	if svc == nil {
		return
	}
	_, _ = svc.EnrichSessionsIncremental(ctx, &service.SessionSemanticServiceConfig{
		Lookback: 48 * time.Hour,
		Limit:    20,
	})
}
