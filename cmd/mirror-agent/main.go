//go:build windows

package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/windows"

	"github.com/yuqie6/mirror/internal/bootstrap"
	"github.com/yuqie6/mirror/internal/handler"
	"github.com/yuqie6/mirror/internal/httpapi"
	"github.com/yuqie6/mirror/internal/pkg/config"
)

func main() {
	// 单实例：避免 UI 多次启动导致重复拉起多个 Agent
	// 使用 Local\ 将范围限制在当前会话，避免多用户/多会话间互相影响。
	mutex, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr(`Local\MirrorAgentSingletonMutex`))
	if err == nil {
		if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
			_ = windows.CloseHandle(mutex)
			return
		}
		defer windows.CloseHandle(mutex)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfgPath, cfgErr := config.DefaultConfigPath()
	if cfgErr == nil {
		if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
			_ = config.WriteFile(cfgPath, config.Default())
		}
	}

	rt, err := bootstrap.NewAgentRuntime(ctx, cfgPath)
	if err != nil {
		slog.Error("启动 Agent 失败", "error", err)
		os.Exit(1)
	}
	defer rt.Close()

	slog.Info("Mirror Agent 启动中...", "name", rt.Cfg.App.Name, "version", rt.Cfg.App.Version)
	slog.Info("Mirror Agent 已启动")

	uiServer, err := httpapi.Start(ctx, rt, httpapi.Options{ListenAddr: "127.0.0.1:0"})
	if err != nil {
		slog.Error("启动本地 UI/API 失败", "error", err)
	}

	// ========== 系统托盘 ==========
	quitChan := make(chan struct{})

	tray := handler.NewTrayHandler(&handler.TrayConfig{
		AppName: rt.Cfg.App.Name,
		OnOpen: func() {
			slog.Info("打开 UI 面板")
			if uiServer != nil {
				handler.OpenUI(uiServer.BaseURL() + "/")
			}
		},
		OnQuit: func() {
			slog.Info("从托盘退出")
			close(quitChan)
		},
	})

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			slog.Info("收到系统退出信号")
			tray.Quit()
		case <-quitChan:
			// 从托盘菜单退出
		}
	}()

	// 运行托盘（阻塞）
	tray.Run()

	slog.Info("正在关闭...")

	cancel()
	if uiServer != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = uiServer.Shutdown(shutdownCtx)
		shutdownCancel()
	}
	slog.Info("Mirror Agent 已退出")
}
