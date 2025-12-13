package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

func DefaultConfigPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取可执行文件路径失败: %w", err)
	}
	exeDir := filepath.Dir(exe)
	return filepath.Join(exeDir, "config", "config.yaml"), nil
}

func WriteFile(path string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("cfg 不能为空")
	}
	if path == "" {
		return fmt.Errorf("path 不能为空")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	payload := map[string]any{
		"app": map[string]any{
			"name":      cfg.App.Name,
			"version":   cfg.App.Version,
			"log_level": cfg.App.LogLevel,
			"log_path":  cfg.App.LogPath,
		},
		"collector": map[string]any{
			"poll_interval_ms":   cfg.Collector.PollIntervalMs,
			"min_duration_sec":   cfg.Collector.MinDurationSec,
			"buffer_size":        cfg.Collector.BufferSize,
			"flush_batch_size":   cfg.Collector.FlushBatchSize,
			"flush_interval_sec": cfg.Collector.FlushIntervalSec,
			"session_idle_min":   cfg.Collector.SessionIdleMin,
		},
		"storage": map[string]any{
			"db_path": cfg.Storage.DBPath,
		},
		"diff": map[string]any{
			"enabled":      cfg.Diff.Enabled,
			"watch_paths":  cfg.Diff.WatchPaths,
			"extensions":   cfg.Diff.Extensions,
			"buffer_size":  cfg.Diff.BufferSize,
			"debounce_sec": cfg.Diff.DebounceSec,
		},
		"browser": map[string]any{
			"enabled":           cfg.Browser.Enabled,
			"poll_interval_sec": cfg.Browser.PollIntervalSec,
			"history_path":      cfg.Browser.HistoryPath,
		},
		"ai": map[string]any{
			"deepseek": map[string]any{
				"api_key":  cfg.AI.DeepSeek.APIKey,
				"base_url": cfg.AI.DeepSeek.BaseURL,
				"model":    cfg.AI.DeepSeek.Model,
			},
			"siliconflow": map[string]any{
				"api_key":         cfg.AI.SiliconFlow.APIKey,
				"base_url":        cfg.AI.SiliconFlow.BaseURL,
				"embedding_model": cfg.AI.SiliconFlow.EmbeddingModel,
				"reranker_model":  cfg.AI.SiliconFlow.RerankerModel,
			},
		},
		"privacy": map[string]any{
			"enabled":  cfg.Privacy.Enabled,
			"patterns": cfg.Privacy.Patterns,
		},
	}

	b, err := yaml.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}
