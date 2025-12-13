//go:build windows

package handler

import (
	"net/http"

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
		DiffWatchPaths:     append([]string(nil), cfg.Diff.WatchPaths...),
		BrowserHistoryPath: cfg.Browser.HistoryPath,
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
	if req.DiffWatchPaths != nil {
		next.Diff.WatchPaths = append([]string(nil), (*req.DiffWatchPaths)...)
	}
	if req.BrowserHistoryPath != nil {
		next.Browser.HistoryPath = *req.BrowserHistoryPath
	}

	if err := config.WriteFile(path, &next); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if a.hub != nil {
		a.hub.Publish(eventbus.Event{Type: "settings_updated"})
	}
	WriteJSON(w, http.StatusOK, &dto.SaveSettingsResponseDTO{RestartRequired: true})
}
