package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Collector CollectorConfig `mapstructure:"collector"`
	Diff      DiffConfig      `mapstructure:"diff"`
	Browser   BrowserConfig   `mapstructure:"browser"`
	Storage   StorageConfig   `mapstructure:"storage"`
	AI        AIConfig        `mapstructure:"ai"`
	Privacy   PrivacyConfig   `mapstructure:"privacy"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name     string `mapstructure:"name"`
	Version  string `mapstructure:"version"`
	LogLevel string `mapstructure:"log_level"`
	LogPath  string `mapstructure:"log_path"`
}

// CollectorConfig 采集器配置
type CollectorConfig struct {
	PollIntervalMs   int `mapstructure:"poll_interval_ms"`
	MinDurationSec   int `mapstructure:"min_duration_sec"`
	BufferSize       int `mapstructure:"buffer_size"`
	FlushBatchSize   int `mapstructure:"flush_batch_size"`
	FlushIntervalSec int `mapstructure:"flush_interval_sec"`
	SessionIdleMin   int `mapstructure:"session_idle_min"` // 会话 idle 切分阈值（分钟）
}

// StorageConfig 存储配置
type StorageConfig struct {
	DBPath string `mapstructure:"db_path"`
}

// DiffConfig Diff 采集配置
type DiffConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	WatchPaths  []string `mapstructure:"watch_paths"`
	Extensions  []string `mapstructure:"extensions"`
	BufferSize  int      `mapstructure:"buffer_size"`
	DebounceSec int      `mapstructure:"debounce_sec"`
}

// BrowserConfig 浏览器采集配置
type BrowserConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	PollIntervalSec int    `mapstructure:"poll_interval_sec"`
	HistoryPath     string `mapstructure:"history_path"`
}

// AIConfig AI 配置
type AIConfig struct {
	DeepSeek    DeepSeekConfig    `mapstructure:"deepseek"`
	SiliconFlow SiliconFlowConfig `mapstructure:"siliconflow"`
}

// DeepSeekConfig DeepSeek 配置
type DeepSeekConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

// SiliconFlowConfig 硅基流动配置
type SiliconFlowConfig struct {
	APIKey         string `mapstructure:"api_key"`
	BaseURL        string `mapstructure:"base_url"`
	EmbeddingModel string `mapstructure:"embedding_model"`
	RerankerModel  string `mapstructure:"reranker_model"`
}

// PrivacyConfig 隐私配置
type PrivacyConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	Patterns []string `mapstructure:"patterns"`
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 设置配置文件路径
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 默认查找路径
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	// 支持环境变量
	v.SetEnvPrefix("MIRROR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Warn("配置文件未找到，使用默认配置")
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	} else {
		slog.Info("加载配置文件", "path", v.ConfigFileUsed())
	}

	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 处理环境变量占位符
	cfg.AI.DeepSeek.APIKey = expandEnv(cfg.AI.DeepSeek.APIKey)
	cfg.AI.SiliconFlow.APIKey = expandEnv(cfg.AI.SiliconFlow.APIKey)

	// 处理相对路径
	cfg.Storage.DBPath = resolvePath(cfg.Storage.DBPath)
	cfg.App.LogPath = resolvePath(cfg.App.LogPath)

	return &cfg, nil
}

// Default 返回“未解析路径”的默认配置（用于首次落盘，保持相对路径可便携）。
func Default() *Config {
	v := viper.New()
	setDefaults(v)
	var cfg Config
	_ = v.Unmarshal(&cfg)
	return &cfg
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// App
	v.SetDefault("app.name", "mirror-agent")
	v.SetDefault("app.version", "0.1.0")
	v.SetDefault("app.log_level", "info")
	v.SetDefault("app.log_path", "./logs/mirror.log")

	// Collector
	v.SetDefault("collector.poll_interval_ms", 500)
	v.SetDefault("collector.min_duration_sec", 3)
	v.SetDefault("collector.buffer_size", 2048)
	v.SetDefault("collector.flush_batch_size", 100)
	v.SetDefault("collector.flush_interval_sec", 5)
	v.SetDefault("collector.session_idle_min", 6)

	// Storage
	v.SetDefault("storage.db_path", "./data/mirror.db")

	// Diff
	v.SetDefault("diff.enabled", true)
	v.SetDefault("diff.watch_paths", []string{})
	v.SetDefault("diff.extensions", []string{".go", ".py", ".js", ".ts", ".jsx", ".tsx", ".vue", ".java", ".rs", ".c", ".cpp"})
	v.SetDefault("diff.buffer_size", 512)
	v.SetDefault("diff.debounce_sec", 2)

	// AI
	v.SetDefault("ai.deepseek.base_url", "https://api.deepseek.com")
	v.SetDefault("ai.deepseek.model", "deepseek-chat")
	v.SetDefault("ai.siliconflow.base_url", "https://api.siliconflow.cn/v1")
	v.SetDefault("ai.siliconflow.embedding_model", "BAAI/bge-large-zh-v1.5")
	v.SetDefault("ai.siliconflow.reranker_model", "BAAI/bge-reranker-v2-m3")

	// Privacy
	v.SetDefault("privacy.enabled", false)
}

// expandEnv 展开环境变量占位符 ${VAR}
func expandEnv(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		envVar := s[2 : len(s)-1]
		return os.Getenv(envVar)
	}
	return s
}

// resolvePath 解析相对路径为绝对路径
func resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	// 获取可执行文件目录
	exe, err := os.Executable()
	if err != nil {
		return path
	}

	exeDir := filepath.Dir(exe)
	return filepath.Join(exeDir, path)
}

type LoggerOptions struct {
	Level     string
	Path      string
	Component string
}

// SetupLogger 初始化默认日志（stdout + 可选落盘）
// 返回的 closer 由调用方在进程退出时关闭（通常随 Core.Close 一起）。
func SetupLogger(opts LoggerOptions) (io.Closer, error) {
	var logLevel slog.Level
	switch strings.ToLower(opts.Level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var file *os.File
	var setupErr error
	writer := io.Writer(os.Stdout)
	if strings.TrimSpace(opts.Path) != "" {
		if err := os.MkdirAll(filepath.Dir(opts.Path), 0o755); err != nil {
			setupErr = fmt.Errorf("创建日志目录失败: %w", err)
		} else if f, err := os.OpenFile(opts.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err != nil {
			setupErr = fmt.Errorf("打开日志文件失败: %w", err)
		} else {
			file = f
			writer = io.MultiWriter(os.Stdout, file)
		}
	}

	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: logLevel,
		// debug 下默认带源码位置，方便定位问题；非 debug 避免噪音。
		AddSource: logLevel == slog.LevelDebug,
	})
	logger := slog.New(handler)
	if strings.TrimSpace(opts.Component) != "" {
		logger = logger.With("component", opts.Component, "pid", os.Getpid())
	}
	slog.SetDefault(logger)

	if setupErr != nil {
		// 避免在 logger 尚未稳定时递归使用 slog；stderr 提示即可，日志仍会输出到 stdout。
		_, _ = fmt.Fprintf(os.Stderr, "logger setup warning: %v\n", setupErr)
	}
	return file, setupErr
}
