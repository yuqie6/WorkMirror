//go:build windows

package handler

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OpenUI 以“桌面应用窗口”优先打开本地 UI（Edge app mode），失败则回退到默认浏览器。
func OpenUI(url string) {
	u := strings.TrimSpace(url)
	if u == "" {
		return
	}

	if tryOpenEdgeApp(u) {
		return
	}

	// 默认浏览器打开
	if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start(); err != nil {
		slog.Warn("打开 UI 失败", "error", err)
	}
}

func tryOpenEdgeApp(url string) bool {
	edgePath := findEdgePath()
	if edgePath == "" {
		return false
	}

	args := []string{
		"--app=" + url,
		"--new-window",
	}

	if err := exec.Command(edgePath, args...).Start(); err != nil {
		slog.Warn("启动 Edge app-mode 失败", "path", edgePath, "error", err)
		return false
	}
	return true
}

func findEdgePath() string {
	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("LocalAppData"), "Microsoft", "Edge", "Application", "msedge.exe"),
	}
	for _, p := range candidates {
		if strings.TrimSpace(p) == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// PATH fallback
	if p, err := exec.LookPath("msedge"); err == nil {
		return p
	}
	if p, err := exec.LookPath("msedge.exe"); err == nil {
		return p
	}
	return ""
}
