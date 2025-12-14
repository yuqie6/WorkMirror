//go:build windows

package collector

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	_ "github.com/glebarez/sqlite"
	"github.com/yuqie6/WorkMirror/internal/schema"
)

// BrowserCollector 浏览器历史采集器
type BrowserCollector struct {
	historyPath   string
	tempPath      string
	pollInterval  time.Duration
	lastVisitTime int64
	eventChan     chan *schema.BrowserEvent
	stopChan      chan struct{}
	running       bool

	lastEmitAt atomic.Int64
	dropped    atomic.Int64
}

// BrowserCollectorConfig 配置
type BrowserCollectorConfig struct {
	HistoryPath  string        // Chrome History 文件路径（可选，自动检测）
	PollInterval time.Duration // 轮询间隔
}

// NewBrowserCollector 创建浏览器采集器
func NewBrowserCollector(cfg *BrowserCollectorConfig) (*BrowserCollector, error) {
	if cfg == nil {
		cfg = &BrowserCollectorConfig{}
	}

	historyPath := cfg.HistoryPath
	if historyPath == "" {
		// 自动检测 Chrome History 路径
		historyPath = getChromeHistoryPath()
	}

	if historyPath == "" {
		return nil, fmt.Errorf("未找到 chrome history 文件")
	}

	// 检查文件是否存在
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("chrome history 文件不存在: %s", historyPath)
	}

	pollInterval := cfg.PollInterval
	if pollInterval == 0 {
		pollInterval = 30 * time.Second // 默认 30 秒轮询一次
	}

	// 临时文件路径
	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, "mirror_chrome_history.db")

	return &BrowserCollector{
		historyPath:  historyPath,
		tempPath:     tempPath,
		pollInterval: pollInterval,
		eventChan:    make(chan *schema.BrowserEvent, 256),
		stopChan:     make(chan struct{}),
	}, nil
}

// getChromeHistoryPath 获取 Chrome History 文件路径
func getChromeHistoryPath() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return ""
	}

	// 检查常见的 Chrome 路径
	paths := []string{
		filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "History"),
		filepath.Join(localAppData, "Google", "Chrome Beta", "User Data", "Default", "History"),
		filepath.Join(localAppData, "Chromium", "User Data", "Default", "History"),
		filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "History"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			slog.Debug("发现浏览器历史文件", "path", path)
			return path
		}
	}

	return ""
}

// Start 启动采集
func (c *BrowserCollector) Start(ctx context.Context) error {
	if c.running {
		return nil
	}

	c.running = true
	slog.Info("浏览器历史采集器启动", "history_path", c.historyPath)

	go c.pollLoop(ctx)
	return nil
}

// Stop 停止采集
func (c *BrowserCollector) Stop() error {
	if !c.running {
		return nil
	}

	close(c.stopChan)
	c.running = false

	// 清理临时文件
	os.Remove(c.tempPath)

	slog.Info("浏览器历史采集器已停止")
	return nil
}

// Events 返回事件通道
func (c *BrowserCollector) Events() <-chan *schema.BrowserEvent {
	return c.eventChan
}

// pollLoop 轮询循环
func (c *BrowserCollector) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	// 首次获取设置基准时间
	c.lastVisitTime = time.Now().Add(-1*time.Hour).UnixMicro() * 10 // Chrome 使用 WebKit 时间戳

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.collectHistory()
		}
	}
}

// collectHistory 采集历史记录
func (c *BrowserCollector) collectHistory() {
	// 复制 History 文件（避免锁定问题）
	if err := copyFile(c.historyPath, c.tempPath); err != nil {
		slog.Debug("复制 History 文件失败", "error", err)
		return
	}
	defer os.Remove(c.tempPath)

	// 打开临时数据库
	db, err := sql.Open("sqlite", c.tempPath)
	if err != nil {
		slog.Debug("打开 History 数据库失败", "error", err)
		return
	}
	defer db.Close()

	// 查询最新的访问记录
	// Chrome 使用 WebKit 时间戳：微秒 from 1601-01-01
	query := `
		SELECT urls.url, urls.title, visits.visit_time
		FROM visits
		JOIN urls ON visits.url = urls.id
		WHERE visits.visit_time > ?
		ORDER BY visits.visit_time ASC
		LIMIT 100
	`

	rows, err := db.Query(query, c.lastVisitTime)
	if err != nil {
		slog.Debug("查询历史记录失败", "error", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var urlStr, title string
		var visitTime int64

		if err := rows.Scan(&urlStr, &title, &visitTime); err != nil {
			continue
		}

		// 更新最后访问时间
		if visitTime > c.lastVisitTime {
			c.lastVisitTime = visitTime
		}

		// 解析 URL 获取域名
		domain := extractDomain(urlStr)
		if domain == "" {
			continue
		}

		// 过滤内部页面
		if shouldSkipURL(urlStr) {
			continue
		}

		// 转换 WebKit 时间戳为 Unix 时间戳
		// WebKit: 微秒 from 1601-01-01
		// Unix: 毫秒 from 1970-01-01
		unixMilli := (visitTime - 11644473600000000) / 1000

		event := &schema.BrowserEvent{
			Timestamp: unixMilli,
			URL:       urlStr,
			Title:     title,
			Domain:    domain,
		}

		select {
		case c.eventChan <- event:
			c.lastEmitAt.Store(unixMilli)
			count++
		default:
			c.dropped.Add(1)
			slog.Warn("浏览器事件缓冲区已满")
		}
	}

	if count > 0 {
		slog.Debug("采集到浏览器历史", "count", count)
	}
}

type BrowserCollectorStats struct {
	Running     bool   `json:"running"`
	HistoryPath string `json:"history_path"`
	LastEmitAt  int64  `json:"last_emit_at"`
	Dropped     int64  `json:"dropped"`
}

func (c *BrowserCollector) Stats() BrowserCollectorStats {
	if c == nil {
		return BrowserCollectorStats{}
	}
	return BrowserCollectorStats{
		Running:     c.running,
		HistoryPath: c.historyPath,
		LastEmitAt:  c.lastEmitAt.Load(),
		Dropped:     c.dropped.Load(),
	}
}

// extractDomain 从 URL 提取域名
func extractDomain(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Host
}

// shouldSkipURL 判断是否应该跳过该 URL
func shouldSkipURL(urlStr string) bool {
	skipPrefixes := []string{
		"chrome://",
		"chrome-extension://",
		"edge://",
		"about:",
		"file://",
		"data:",
	}

	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(urlStr, prefix) {
			return true
		}
	}

	return false
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
