//go:build windows

package collector

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/yuqie6/mirror/internal/schema"
)

// DiffCollector 文件 Diff 采集器
type DiffCollector struct {
	watcher     *fsnotify.Watcher
	watchPaths  []string
	extensions  map[string]bool
	eventChan   chan *schema.Diff
	stopChan    chan struct{}
	running     bool
	mu          sync.Mutex
	stopOnce    sync.Once
	debounceMap map[string]time.Time // 防抖：file -> lastSave
	debounceDur time.Duration
}

// DiffCollectorConfig 配置
type DiffCollectorConfig struct {
	WatchPaths  []string // 监控的目录列表
	Extensions  []string // 监控的文件扩展名
	BufferSize  int      // 事件缓冲区大小
	DebounceSec int      // 防抖时间（秒）
}

// DefaultDiffCollectorConfig 默认配置
func DefaultDiffCollectorConfig() *DiffCollectorConfig {
	return &DiffCollectorConfig{
		WatchPaths:  []string{},
		Extensions:  []string{".go", ".py", ".js", ".ts", ".jsx", ".tsx", ".vue", ".java", ".rs", ".c", ".cpp"},
		BufferSize:  512,
		DebounceSec: 2,
	}
}

// NewDiffCollector 创建 Diff 采集器
func NewDiffCollector(cfg *DiffCollectorConfig) (*DiffCollector, error) {
	if cfg == nil {
		cfg = DefaultDiffCollectorConfig()
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监控器失败: %w", err)
	}

	// 构建扩展名 map
	extMap := make(map[string]bool)
	for _, ext := range cfg.Extensions {
		extMap[strings.ToLower(ext)] = true
	}

	return &DiffCollector{
		watcher:     watcher,
		watchPaths:  cfg.WatchPaths,
		extensions:  extMap,
		eventChan:   make(chan *schema.Diff, cfg.BufferSize),
		stopChan:    make(chan struct{}),
		debounceMap: make(map[string]time.Time),
		debounceDur: time.Duration(cfg.DebounceSec) * time.Second,
	}, nil
}

// AddWatchPath 添加监控路径
func (c *DiffCollector) AddWatchPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	// 递归添加所有子目录
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// 跳过隐藏目录和常见的忽略目录
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") ||
				base == "node_modules" ||
				base == "vendor" ||
				base == "__pycache__" ||
				base == "dist" ||
				base == "build" {
				return filepath.SkipDir
			}

			if err := c.watcher.Add(path); err != nil {
				slog.Warn("添加监控目录失败", "path", path, "error", err)
			} else {
				slog.Debug("添加监控目录", "path", path)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历目录失败: %w", err)
	}

	c.watchPaths = append(c.watchPaths, absPath)
	slog.Info("添加 Diff 监控路径", "path", absPath)
	return nil
}

// Start 启动采集
func (c *DiffCollector) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()
	slog.Info("Diff 采集器启动", "watch_paths", c.watchPaths)

	go c.watchLoop(ctx)
	return nil
}

// Stop 停止采集
func (c *DiffCollector) Stop() error {
	c.stopOnce.Do(func() {
		c.mu.Lock()
		if !c.running {
			c.mu.Unlock()
			return
		}
		c.running = false
		c.mu.Unlock()

		close(c.stopChan)
		_ = c.watcher.Close()
		slog.Info("Diff 采集器已停止")
	})
	return nil
}

// Events 返回事件通道
func (c *DiffCollector) Events() <-chan *schema.Diff {
	return c.eventChan
}

// watchLoop 监控循环
func (c *DiffCollector) watchLoop(ctx context.Context) {
	defer close(c.eventChan)
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		case event, ok := <-c.watcher.Events:
			if !ok {
				return
			}
			c.handleFsEvent(ctx, event)
		case err, ok := <-c.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("文件监控错误", "error", err)
		}
	}
}

// handleFsEvent 处理文件系统事件
func (c *DiffCollector) handleFsEvent(ctx context.Context, event fsnotify.Event) {
	// 只处理写入事件
	if !event.Has(fsnotify.Write) {
		return
	}

	filePath := event.Name
	ext := strings.ToLower(filepath.Ext(filePath))

	// 检查是否是监控的文件类型
	if !c.extensions[ext] {
		return
	}

	// 防抖检查
	c.mu.Lock()
	lastSave, exists := c.debounceMap[filePath]
	now := time.Now()
	if exists && now.Sub(lastSave) < c.debounceDur {
		c.mu.Unlock()
		return
	}
	c.debounceMap[filePath] = now
	c.mu.Unlock()

	// 获取 Diff
	diff, err := c.captureDiff(ctx, filePath)
	if err != nil {
		slog.Debug("获取 Diff 失败", "file", filePath, "error", err)
		return
	}

	if diff == nil || diff.DiffContent == "" {
		return
	}

	// 发送事件
	select {
	case c.eventChan <- diff:
		slog.Debug("Diff 事件已发送",
			"file", diff.FileName,
			"lines_added", diff.LinesAdded,
			"lines_deleted", diff.LinesDeleted,
		)
	default:
		slog.Warn("Diff 缓冲区已满，丢弃事件", "file", filePath)
	}
}

// captureDiff 捕获文件 Diff
func (c *DiffCollector) captureDiff(ctx context.Context, filePath string) (*schema.Diff, error) {
	// 检查是否在 Git 仓库中
	projectPath, isGit := c.findGitRoot(filePath)

	var diffContent string
	var linesAdded, linesDeleted int

	if isGit {
		// 使用 git diff
		content, added, deleted, err := c.gitDiff(ctx, projectPath, filePath)
		if err != nil {
			return nil, err
		}
		diffContent = content
		linesAdded = added
		linesDeleted = deleted
	} else {
		// 非 Git 仓库，暂时跳过（或实现简单的文件快照对比）
		slog.Debug("非 Git 仓库，跳过 Diff", "file", filePath)
		return nil, nil
	}

	if diffContent == "" {
		return nil, nil
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	language := GetLanguageFromExt(ext)

	return &schema.Diff{
		Timestamp:    time.Now().UnixMilli(),
		FilePath:     filePath,
		FileName:     filepath.Base(filePath),
		Language:     language,
		DiffContent:  diffContent,
		LinesAdded:   linesAdded,
		LinesDeleted: linesDeleted,
		ProjectPath:  projectPath,
		IsGitRepo:    isGit,
	}, nil
}

// findGitRoot 查找 Git 仓库根目录
func (c *DiffCollector) findGitRoot(filePath string) (string, bool) {
	dir := filepath.Dir(filePath)

	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return dir, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", false
}

// gitDiff 使用 git diff 获取差异（带超时）
func (c *DiffCollector) gitDiff(ctx context.Context, repoRoot, filePath string) (string, int, int, error) {
	// 统一在仓库根目录执行 git
	dir := repoRoot
	if dir == "" {
		dir = filepath.Dir(filePath)
	}

	// 超时保护
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 检查文件是否被 Git 跟踪
	relPath := filePath
	if repoRoot != "" {
		if p, err := filepath.Rel(repoRoot, filePath); err == nil {
			relPath = p
		}
	}

	checkCmd := exec.CommandContext(timeoutCtx, "git", "ls-files", "--", relPath)
	checkCmd.Dir = dir
	checkOutput, _ := checkCmd.Output()

	isTracked := strings.TrimSpace(string(checkOutput)) != ""

	if !isTracked {
		// 新文件：读取整个文件内容作为 diff
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", 0, 0, fmt.Errorf("读取新文件失败: %w", err)
		}

		lines := strings.Split(string(content), "\n")
		diffContent := fmt.Sprintf("--- /dev/null\n+++ b/%s\n@@ -0,0 +1,%d @@\n",
			filepath.Base(filePath), len(lines))
		for _, line := range lines {
			diffContent += "+" + line + "\n"
		}

		slog.Debug("检测到新文件", "file", filePath, "lines", len(lines))
		return diffContent, len(lines), 0, nil
	}

	// 获取未暂存的改动
	cmd := exec.CommandContext(timeoutCtx, "git", "diff", "--", relPath)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		// 尝试获取已暂存的改动
		cmd = exec.Command("git", "diff", "--cached", "--", filePath)
		cmd.Dir = dir
		output, err = cmd.Output()
		if err != nil {
			return "", 0, 0, fmt.Errorf("执行 git diff 失败: %w", err)
		}
	}

	diffContent := string(output)
	if diffContent == "" {
		return "", 0, 0, nil
	}

	// 统计行数
	linesAdded, linesDeleted := c.countDiffLines(diffContent)

	return diffContent, linesAdded, linesDeleted, nil
}

// countDiffLines 统计 Diff 的增删行数
func (c *DiffCollector) countDiffLines(diff string) (added, deleted int) {
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			deleted++
		}
	}
	return
}
