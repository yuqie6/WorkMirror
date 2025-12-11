//go:build windows

package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	kernel32                     = windows.NewLazySystemDLL("kernel32.dll")
	psapi                        = windows.NewLazySystemDLL("psapi.dll")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procGetModuleBaseNameW       = psapi.NewProc("GetModuleBaseNameW")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)

// WindowInfo 窗口信息
type WindowInfo struct {
	HWND      uintptr // 窗口句柄
	Title     string  // 窗口标题
	ProcessID uint32  // 进程 ID
	AppName   string  // 应用程序名称 (如 Chrome.exe)
}

// GetForegroundWindowInfo 获取当前前台窗口信息
func GetForegroundWindowInfo() (*WindowInfo, error) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return nil, errors.New("无法获取前台窗口")
	}

	// 获取窗口标题
	title, err := getWindowText(hwnd)
	if err != nil {
		slog.Debug("获取窗口标题失败", "error", err)
		title = ""
	}

	// 获取进程 ID
	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	// 获取进程名
	appName := getProcessName(pid)

	return &WindowInfo{
		HWND:      hwnd,
		Title:     title,
		ProcessID: pid,
		AppName:   appName,
	}, nil
}

// getWindowText 获取窗口标题
func getWindowText(hwnd uintptr) (string, error) {
	// 获取标题长度
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return "", nil
	}

	// 分配缓冲区
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), length+1)

	return syscall.UTF16ToString(buf), nil
}

// getProcessName 根据进程 ID 获取进程名
func getProcessName(pid uint32) string {
	// 打开进程
	handle, _, err := procOpenProcess.Call(
		PROCESS_QUERY_INFORMATION|PROCESS_VM_READ,
		0,
		uintptr(pid),
	)
	if handle == 0 {
		slog.Debug("打开进程失败", "pid", pid, "error", err)
		return "Unknown"
	}
	defer windows.CloseHandle(windows.Handle(handle))

	// 获取模块名称
	buf := make([]uint16, windows.MAX_PATH)
	ret, _, _ := procGetModuleBaseNameW.Call(
		handle,
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		windows.MAX_PATH,
	)
	if ret == 0 {
		return "Unknown"
	}

	return syscall.UTF16ToString(buf)
}

// String 格式化输出窗口信息
func (w *WindowInfo) String() string {
	return fmt.Sprintf("[%s] %s (PID: %d)", w.AppName, w.Title, w.ProcessID)
}

// IsSameWindow 判断是否是同一个窗口
func (w *WindowInfo) IsSameWindow(other *WindowInfo) bool {
	if other == nil {
		return false
	}
	return w.HWND == other.HWND
}

// IsSameApp 判断是否是同一个应用
func (w *WindowInfo) IsSameApp(other *WindowInfo) bool {
	if other == nil {
		return false
	}
	return strings.EqualFold(w.AppName, other.AppName)
}

// GetAppBaseName 获取应用基础名称（去除 .exe 后缀）
func (w *WindowInfo) GetAppBaseName() string {
	name := filepath.Base(w.AppName)
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// IsSystemWindow 判断是否是系统窗口（应忽略）
func (w *WindowInfo) IsSystemWindow() bool {
	ignoreApps := []string{
		"explorer.exe",
		"ShellExperienceHost.exe",
		"SearchHost.exe",
		"StartMenuExperienceHost.exe",
		"TextInputHost.exe",
		"LockApp.exe",
	}

	for _, app := range ignoreApps {
		if strings.EqualFold(w.AppName, app) {
			return true
		}
	}

	// 忽略空标题的窗口
	if strings.TrimSpace(w.Title) == "" {
		return true
	}

	return false
}
