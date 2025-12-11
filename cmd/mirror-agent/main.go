package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielsclee/mirror/internal/handler"
	"github.com/danielsclee/mirror/internal/pkg/config"
	"github.com/danielsclee/mirror/internal/repository"
	"github.com/danielsclee/mirror/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	// 设置日志级别
	config.SetupLogger(cfg.App.LogLevel)

	slog.Info("Mirror Agent 启动中...",
		"name", cfg.App.Name,
		"version", cfg.App.Version,
	)

	// 初始化数据库
	db, err := repository.NewDatabase(cfg.Storage.DBPath)
	if err != nil {
		slog.Error("初始化数据库失败", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 创建仓储
	eventRepo := repository.NewEventRepository(db.DB)
	diffRepo := repository.NewDiffRepository(db.DB)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ========== 窗口采集服务 ==========
	windowCollector := handler.NewWindowCollector(&handler.CollectorConfig{
		PollIntervalMs: cfg.Collector.PollIntervalMs,
		MinDurationSec: cfg.Collector.MinDurationSec,
		BufferSize:     cfg.Collector.BufferSize,
	})

	tracker := service.NewTrackerService(windowCollector, eventRepo, &service.TrackerConfig{
		FlushBatchSize:   cfg.Collector.FlushBatchSize,
		FlushIntervalSec: cfg.Collector.FlushIntervalSec,
	})

	if err := tracker.Start(ctx); err != nil {
		slog.Error("启动窗口追踪服务失败", "error", err)
		os.Exit(1)
	}

	// ========== Diff 采集服务 ==========
	var diffService *service.DiffService
	if cfg.Diff.Enabled && len(cfg.Diff.WatchPaths) > 0 {
		diffCollector, err := handler.NewDiffCollector(&handler.DiffCollectorConfig{
			WatchPaths:  cfg.Diff.WatchPaths,
			Extensions:  cfg.Diff.Extensions,
			BufferSize:  cfg.Diff.BufferSize,
			DebounceSec: cfg.Diff.DebounceSec,
		})
		if err != nil {
			slog.Error("创建 Diff 采集器失败", "error", err)
			os.Exit(1)
		}

		// 添加监控路径
		for _, path := range cfg.Diff.WatchPaths {
			if err := diffCollector.AddWatchPath(path); err != nil {
				slog.Warn("添加监控路径失败", "path", path, "error", err)
			}
		}

		diffService = service.NewDiffService(diffCollector, diffRepo)
		if err := diffService.Start(ctx); err != nil {
			slog.Error("启动 Diff 服务失败", "error", err)
			os.Exit(1)
		}
		slog.Info("Diff 采集已启用", "watch_paths", cfg.Diff.WatchPaths)
	} else {
		slog.Info("Diff 采集未启用（需配置 watch_paths）")
	}

	slog.Info("Mirror Agent 已启动，按 Ctrl+C 退出")

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	slog.Info("收到退出信号，正在关闭...")

	// 优雅关闭
	cancel()
	tracker.Stop()
	if diffService != nil {
		diffService.Stop()
	}

	slog.Info("Mirror Agent 已退出")
}
