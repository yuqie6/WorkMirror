package handler

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows/registry"
)

// TrayHandler 系统托盘处理器
type TrayHandler struct {
	onOpen  func()
	onQuit  func()
	appName string
}

// TrayConfig 托盘配置
type TrayConfig struct {
	AppName string
	OnOpen  func()
	OnQuit  func()
}

// NewTrayHandler 创建托盘处理器
func NewTrayHandler(cfg *TrayConfig) *TrayHandler {
	return &TrayHandler{
		appName: cfg.AppName,
		onOpen:  cfg.OnOpen,
		onQuit:  cfg.OnQuit,
	}
}

// Run 启动托盘（阻塞）
func (t *TrayHandler) Run() {
	systray.Run(t.onReady, t.onExit)
}

// Quit 退出托盘
func (t *TrayHandler) Quit() {
	systray.Quit()
}

func (t *TrayHandler) onReady() {
	// 设置标题和提示（不设置图标，避免格式问题）
	systray.SetTitle(t.appName)
	systray.SetTooltip(t.appName + " - 个人成长量化系统")

	// 菜单项
	mOpen := systray.AddMenuItem("打开面板", "打开 Mirror UI")
	systray.AddSeparator()
	mAutoStart := systray.AddMenuItemCheckbox("开机自启动", "设置开机自动启动", isAutoStartEnabled())
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "退出 Mirror")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				if t.onOpen != nil {
					t.onOpen()
				}
			case <-mAutoStart.ClickedCh:
				if mAutoStart.Checked() {
					if err := disableAutoStart(); err != nil {
						slog.Warn("禁用开机自启动失败", "error", err)
					} else {
						mAutoStart.Uncheck()
					}
				} else {
					if err := enableAutoStart(); err != nil {
						slog.Warn("启用开机自启动失败", "error", err)
					} else {
						mAutoStart.Check()
					}
				}
			case <-mQuit.ClickedCh:
				if t.onQuit != nil {
					t.onQuit()
				}
				systray.Quit()
				return
			}
		}
	}()
}

func (t *TrayHandler) onExit() {
	// 清理资源
}

// ========== 开机自启动 ==========

const autoStartKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const appRegistryName = "MirrorAgent"

// isAutoStartEnabled 检查是否已设置开机自启
func isAutoStartEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, autoStartKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(appRegistryName)
	return err == nil
}

// enableAutoStart 启用开机自启
func enableAutoStart() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, autoStartKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetStringValue(appRegistryName, exePath)
}

// disableAutoStart 禁用开机自启
func disableAutoStart() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, autoStartKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.DeleteValue(appRegistryName)
}

// ========== 打开 UI ==========

// OpenUI 打开 UI 程序
func OpenUI() {
	exePath, _ := os.Executable()
	dir := filepath.Dir(exePath)

	// 可能的 UI 路径列表
	possiblePaths := []string{
		filepath.Join(dir, "mirror-ui.exe"),                                     // 同目录
		filepath.Join(dir, "mirror-ui", "mirror-ui.exe"),                        // 子目录
		filepath.Join(dir, "..", "mirror-ui", "build", "bin", "mirror-ui.exe"),  // 开发时相对路径
		filepath.Join(dir, "cmd", "mirror-ui", "build", "bin", "mirror-ui.exe"), // 从项目根目录运行
	}

	for _, uiPath := range possiblePaths {
		if _, err := os.Stat(uiPath); err == nil {
			if err := exec.Command(uiPath).Start(); err != nil {
				slog.Error("启动 UI 失败", "path", uiPath, "error", err)
			}
			return
		}
	}

	// 没找到 UI 程序
	slog.Warn("未找到 mirror-ui.exe", "searched_paths", possiblePaths)
}

// _getDefaultIcon 返回默认图标（保留以供将来使用）
func _getDefaultIcon() []byte {
	// 简单的 16x16 ICO 图标（蓝色方块）
	// 实际项目应该嵌入真正的图标文件
	return []byte{
		0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10,
		0x00, 0x00, 0x01, 0x00, 0x20, 0x00, 0x68, 0x04,
		0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00,
		0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x20, 0x00,
		0x00, 0x00, 0x01, 0x00, 0x20, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}
