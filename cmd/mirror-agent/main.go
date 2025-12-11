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

	// 创建事件仓储
	eventRepo := repository.NewEventRepository(db.DB)

	// 创建采集器
	collector := handler.NewWindowCollector(&handler.CollectorConfig{
		PollIntervalMs: cfg.Collector.PollIntervalMs,
		MinDurationSec: cfg.Collector.MinDurationSec,
		BufferSize:     cfg.Collector.BufferSize,
	})

	// 创建追踪服务
	tracker := service.NewTrackerService(collector, eventRepo, &service.TrackerConfig{
		FlushBatchSize:   cfg.Collector.FlushBatchSize,
		FlushIntervalSec: cfg.Collector.FlushIntervalSec,
	})

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动追踪服务
	if err := tracker.Start(ctx); err != nil {
		slog.Error("启动追踪服务失败", "error", err)
		os.Exit(1)
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

	slog.Info("Mirror Agent 已退出")
}
