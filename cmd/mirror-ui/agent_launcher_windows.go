//go:build windows
// +build windows

package main

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

func startAgentOnStartup() {
	exePath, err := os.Executable()
	if err != nil {
		slog.Warn("获取 UI 可执行文件路径失败", "error", err)
		return
	}

	dir := filepath.Dir(exePath)
	agentPath := filepath.Join(dir, "mirror.exe")
	if _, err := os.Stat(agentPath); err != nil {
		slog.Warn("未找到 Agent 可执行文件", "path", agentPath, "error", err)
		return
	}

	if err := exec.Command(agentPath).Start(); err != nil {
		slog.Warn("拉起 Agent 失败", "path", agentPath, "error", err)
	}
}
