package config

import (
	"fmt"
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
	Storage   StorageConfig   `mapstructure:"storage"`
	AI        AIConfig        `mapstructure:"ai"`
	Privacy   PrivacyConfig   `mapstructure:"privacy"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name     string `mapstructure:"name"`
	Version  string `mapstructure:"version"`
	LogLevel string `mapstructure:"log_level"`
}

// CollectorConfig 采集器配置
type CollectorConfig struct {
	PollIntervalMs   int `mapstructure:"poll_interval_ms"`
	MinDurationSec   int `mapstructure:"min_duration_sec"`
	BufferSize       int `mapstructure:"buffer_size"`
	FlushBatchSize   int `mapstructure:"flush_batch_size"`
	FlushIntervalSec int `mapstructure:"flush_interval_sec"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	DBPath string `mapstructure:"db_path"`
}

// AIConfig AI 配置
type AIConfig struct {
	DeepSeek DeepSeekConfig `mapstructure:"deepseek"`
	Jina     JinaConfig     `mapstructure:"jina"`
}

// DeepSeekConfig DeepSeek 配置
type DeepSeekConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

// JinaConfig Jina 配置
type JinaConfig struct {
	APIKey         string `mapstructure:"api_key"`
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
	cfg.AI.Jina.APIKey = expandEnv(cfg.AI.Jina.APIKey)

	// 处理相对路径
	cfg.Storage.DBPath = resolvePath(cfg.Storage.DBPath)

	return &cfg, nil
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// App
	v.SetDefault("app.name", "mirror-agent")
	v.SetDefault("app.version", "0.1.0")
	v.SetDefault("app.log_level", "info")

	// Collector
	v.SetDefault("collector.poll_interval_ms", 500)
	v.SetDefault("collector.min_duration_sec", 3)
	v.SetDefault("collector.buffer_size", 2048)
	v.SetDefault("collector.flush_batch_size", 100)
	v.SetDefault("collector.flush_interval_sec", 5)

	// Storage
	v.SetDefault("storage.db_path", "./data/mirror.db")

	// AI
	v.SetDefault("ai.deepseek.base_url", "https://api.deepseek.com")
	v.SetDefault("ai.deepseek.model", "deepseek-chat")
	v.SetDefault("ai.jina.embedding_model", "jina-embeddings-v3")
	v.SetDefault("ai.jina.reranker_model", "jina-reranker-v2-base-multilingual")

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

// SetupLogger 根据配置设置日志级别
func SetupLogger(level string) {
	var logLevel slog.Level
	switch strings.ToLower(level) {
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

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	slog.SetDefault(slog.New(handler))
}
