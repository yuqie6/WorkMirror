//go:build windows

package handler

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/yuqie6/mirror/internal/dto"
	"github.com/yuqie6/mirror/internal/eventbus"
	"github.com/yuqie6/mirror/internal/pkg/config"
)

func (a *API) HandleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.getSettings(w, r)
	case http.MethodPost:
		a.saveSettings(w, r)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *API) getSettings(w http.ResponseWriter, r *http.Request) {
	path, err := config.DefaultConfigPath()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	cfg, err := config.Load(path)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, &dto.SettingsDTO{
		ConfigPath: path,

		DeepSeekAPIKeySet: cfg.AI.DeepSeek.APIKey != "",
		DeepSeekBaseURL:   cfg.AI.DeepSeek.BaseURL,
		DeepSeekModel:     cfg.AI.DeepSeek.Model,

		SiliconFlowAPIKeySet:      cfg.AI.SiliconFlow.APIKey != "",
		SiliconFlowBaseURL:        cfg.AI.SiliconFlow.BaseURL,
		SiliconFlowEmbeddingModel: cfg.AI.SiliconFlow.EmbeddingModel,
		SiliconFlowRerankerModel:  cfg.AI.SiliconFlow.RerankerModel,

		DBPath:             cfg.Storage.DBPath,
		DiffEnabled:        cfg.Diff.Enabled,
		DiffWatchPaths:     append([]string(nil), cfg.Diff.WatchPaths...),
		BrowserEnabled:     cfg.Browser.Enabled,
		BrowserHistoryPath: cfg.Browser.HistoryPath,

		PrivacyEnabled:  cfg.Privacy.Enabled,
		PrivacyPatterns: append([]string(nil), cfg.Privacy.Patterns...),
	})
}

func (a *API) saveSettings(w http.ResponseWriter, r *http.Request) {
	var req dto.SaveSettingsRequestDTO
	if err := readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	path, err := config.DefaultConfigPath()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cur, err := config.Load(path)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	next := *cur
	if req.DeepSeekAPIKey != nil {
		next.AI.DeepSeek.APIKey = *req.DeepSeekAPIKey
	}
	if req.DeepSeekBaseURL != nil {
		next.AI.DeepSeek.BaseURL = *req.DeepSeekBaseURL
	}
	if req.DeepSeekModel != nil {
		next.AI.DeepSeek.Model = *req.DeepSeekModel
	}

	if req.SiliconFlowAPIKey != nil {
		next.AI.SiliconFlow.APIKey = *req.SiliconFlowAPIKey
	}
	if req.SiliconFlowBaseURL != nil {
		next.AI.SiliconFlow.BaseURL = *req.SiliconFlowBaseURL
	}
	if req.SiliconFlowEmbeddingModel != nil {
		next.AI.SiliconFlow.EmbeddingModel = *req.SiliconFlowEmbeddingModel
	}
	if req.SiliconFlowRerankerModel != nil {
		next.AI.SiliconFlow.RerankerModel = *req.SiliconFlowRerankerModel
	}

	if req.DBPath != nil {
		next.Storage.DBPath = *req.DBPath
	}
	if req.DiffEnabled != nil {
		next.Diff.Enabled = *req.DiffEnabled
	}
	if req.DiffWatchPaths != nil {
		paths := make([]string, 0, len(*req.DiffWatchPaths))
		for _, p := range *req.DiffWatchPaths {
			v := strings.TrimSpace(p)
			if v == "" {
				continue
			}
			if err := validateDirExists(v); err != nil {
				WriteAPIError(w, http.StatusBadRequest, APIError{
					Error: "diff watch path 无效: " + v,
					Code:  "invalid_diff_watch_path",
					Hint:  err.Error(),
				})
				return
			}
			paths = append(paths, v)
		}
		next.Diff.WatchPaths = paths
	}
	if req.BrowserEnabled != nil {
		next.Browser.Enabled = *req.BrowserEnabled
	}
	if req.BrowserHistoryPath != nil {
		p := strings.TrimSpace(*req.BrowserHistoryPath)
		if p != "" {
			if err := validateFileExists(p); err != nil {
				WriteAPIError(w, http.StatusBadRequest, APIError{
					Error: "browser history path 无效",
					Code:  "invalid_browser_history_path",
					Hint:  err.Error(),
				})
				return
			}
		}
		next.Browser.HistoryPath = p
	}
	if req.PrivacyEnabled != nil {
		next.Privacy.Enabled = *req.PrivacyEnabled
	}
	if req.PrivacyPatterns != nil {
		next.Privacy.Patterns = append([]string(nil), (*req.PrivacyPatterns)...)
	}

	if err := config.WriteFile(path, &next); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if a.hub != nil {
		a.hub.Publish(eventbus.Event{Type: "settings_updated"})
		a.hub.Publish(eventbus.Event{Type: "pipeline_status_changed"})
	}
	WriteJSON(w, http.StatusOK, &dto.SaveSettingsResponseDTO{RestartRequired: true})
}

func validateDirExists(p string) error {
	info, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("路径不存在或不可访问: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("不是目录")
	}
	return nil
}

func validateFileExists(p string) error {
	info, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("文件不存在或不可访问: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("是目录，不是文件")
	}
	return nil
}
