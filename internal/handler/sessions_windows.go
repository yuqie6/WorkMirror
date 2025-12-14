//go:build windows

package handler

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/yuqie6/WorkMirror/internal/dto"
	"github.com/yuqie6/WorkMirror/internal/eventbus"
	"github.com/yuqie6/WorkMirror/internal/schema"
	"github.com/yuqie6/WorkMirror/internal/service"
)

func (a *API) HandleBuildSessionsForDate(w http.ResponseWriter, r *http.Request) {
	if !a.requireWritableDB(w) {
		return
	}
	var req dto.DateRequestDTO
	if err := readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	date := strings.TrimSpace(req.Date)
	if date == "" {
		WriteError(w, http.StatusBadRequest, "date 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Sessions == nil {
		WriteError(w, http.StatusBadRequest, "会话服务未初始化")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	created, err := a.rt.Core.Services.Sessions.BuildSessionsForDate(ctx, date)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if a.hub != nil {
		a.hub.Publish(eventbus.Event{Type: "pipeline_status_changed"})
	}
	WriteJSON(w, http.StatusOK, &dto.SessionBuildResultDTO{Created: created})
}

func (a *API) HandleRebuildSessionsForDate(w http.ResponseWriter, r *http.Request) {
	if !a.requireWritableDB(w) {
		return
	}
	var req dto.DateRequestDTO
	if err := readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	date := strings.TrimSpace(req.Date)
	if date == "" {
		WriteError(w, http.StatusBadRequest, "date 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Sessions == nil {
		WriteError(w, http.StatusBadRequest, "会话服务未初始化")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	created, err := a.rt.Core.Services.Sessions.RebuildSessionsForDate(ctx, date)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 重建后自动触发语义丰富，确保 skill_keys 等字段被填充，恢复技能树证据链
	enriched := 0
	if created > 0 && a.rt.Core.Services.SessionSemantic != nil {
		enriched, _ = a.rt.Core.Services.SessionSemantic.EnrichSessionsForDate(ctx, date, 200)
	}

	if a.hub != nil {
		a.hub.Publish(eventbus.Event{Type: "pipeline_status_changed"})
	}
	WriteJSON(w, http.StatusOK, &dto.SessionBuildResultDTO{Created: created, Enriched: enriched})
}

func (a *API) HandleEnrichSessionsForDate(w http.ResponseWriter, r *http.Request) {
	if !a.requireWritableDB(w) {
		return
	}
	var req dto.DateRequestDTO
	if err := readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	date := strings.TrimSpace(req.Date)
	if date == "" {
		WriteError(w, http.StatusBadRequest, "date 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.SessionSemantic == nil {
		WriteError(w, http.StatusBadRequest, "会话语义服务未初始化")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	enriched, err := a.rt.Core.Services.SessionSemantic.EnrichSessionsForDate(ctx, date, 200)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if a.hub != nil {
		a.hub.Publish(eventbus.Event{Type: "pipeline_status_changed"})
	}
	WriteJSON(w, http.StatusOK, &dto.SessionEnrichResultDTO{Enriched: enriched})
}

func (a *API) HandleSessionsByDate(w http.ResponseWriter, r *http.Request) {
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		WriteError(w, http.StatusBadRequest, "date 不能为空")
		return
	}
	if a.rt == nil || a.rt.Repos.Session == nil {
		WriteError(w, http.StatusBadRequest, "会话仓储未初始化")
		return
	}
	sessions, err := a.rt.Repos.Session.GetByDate(r.Context(), date)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]dto.SessionDTO, 0, len(sessions))
	for _, s := range sessions {
		diffIDs := schema.GetInt64Slice(s.Metadata, "diff_ids")
		browserIDs := schema.GetInt64Slice(s.Metadata, "browser_event_ids")
		timeRange := strings.TrimSpace(s.TimeRange)
		if timeRange == "" {
			timeRange = service.FormatTimeRangeMs(s.StartTime, s.EndTime)
		}
		result = append(result, dto.SessionDTO{
			ID:             s.ID,
			Date:           s.Date,
			StartTime:      s.StartTime,
			EndTime:        s.EndTime,
			TimeRange:      timeRange,
			PrimaryApp:     s.PrimaryApp,
			Category:       s.Category,
			Summary:        s.Summary,
			SkillsInvolved: []string(s.SkillsInvolved),
			DiffCount:      len(diffIDs),
			BrowserCount:   len(browserIDs),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartTime < result[j].StartTime })
	WriteJSON(w, http.StatusOK, result)
}

func (a *API) HandleSessionDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := parseInt64Param(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "id 无效")
		return
	}
	if a.rt == nil || a.rt.Repos.Session == nil {
		WriteError(w, http.StatusBadRequest, "会话仓储未初始化")
		return
	}
	sess, err := a.rt.Repos.Session.GetByID(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sess == nil {
		WriteError(w, http.StatusNotFound, "session not found")
		return
	}

	diffIDs := schema.GetInt64Slice(sess.Metadata, "diff_ids")
	browserIDs := schema.GetInt64Slice(sess.Metadata, "browser_event_ids")

	var diffs []schema.Diff
	if len(diffIDs) > 0 {
		diffs, _ = a.rt.Repos.Diff.GetByIDs(r.Context(), diffIDs)
	}
	diffDTOs := make([]dto.SessionDiffDTO, 0, len(diffs))
	for _, d := range diffs {
		diffDTOs = append(diffDTOs, dto.SessionDiffDTO{
			ID:           d.ID,
			FileName:     d.FileName,
			Language:     d.Language,
			Insight:      d.AIInsight,
			Skills:       []string(d.SkillsDetected),
			LinesAdded:   d.LinesAdded,
			LinesDeleted: d.LinesDeleted,
			Timestamp:    d.Timestamp,
		})
	}

	var browserEvents []schema.BrowserEvent
	if len(browserIDs) > 0 && a.rt.Repos.Browser != nil {
		browserEvents, _ = a.rt.Repos.Browser.GetByIDs(r.Context(), browserIDs)
	}
	browserDTOs := make([]dto.SessionBrowserEventDTO, 0, len(browserEvents))
	for _, e := range browserEvents {
		browserDTOs = append(browserDTOs, dto.SessionBrowserEventDTO{
			ID:        e.ID,
			Timestamp: e.Timestamp,
			Domain:    e.Domain,
			Title:     e.Title,
			URL:       e.URL,
		})
	}

	appStats, _ := a.rt.Repos.Event.GetAppStats(r.Context(), sess.StartTime, sess.EndTime)
	appUsage := make([]dto.SessionAppUsageDTO, 0, len(appStats))
	for _, st := range service.TopAppStats(appStats, service.DefaultTopAppsLimit) {
		appUsage = append(appUsage, dto.SessionAppUsageDTO{
			AppName:       st.AppName,
			TotalDuration: st.TotalDuration,
		})
	}

	timeRange := strings.TrimSpace(sess.TimeRange)
	if timeRange == "" {
		timeRange = service.FormatTimeRangeMs(sess.StartTime, sess.EndTime)
	}

	resp := &dto.SessionDetailDTO{
		SessionDTO: dto.SessionDTO{
			ID:             sess.ID,
			Date:           sess.Date,
			StartTime:      sess.StartTime,
			EndTime:        sess.EndTime,
			TimeRange:      timeRange,
			PrimaryApp:     sess.PrimaryApp,
			Category:       sess.Category,
			Summary:        sess.Summary,
			SkillsInvolved: []string(sess.SkillsInvolved),
			DiffCount:      len(diffIDs),
			BrowserCount:   len(browserIDs),
		},
		Tags:     schema.GetStringSlice(sess.Metadata, "tags"),
		RAGRefs:  schema.GetMapSlice(sess.Metadata, "rag_refs"),
		AppUsage: appUsage,
		Diffs:    diffDTOs,
		Browser:  browserDTOs,
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (a *API) HandleSessionEvents(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := parseInt64Param(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "id 无效")
		return
	}

	limit := 200
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		if n, err := strconvAtoiSafe(s); err == nil && n > 0 {
			limit = n
		}
	}
	offset := 0
	if s := strings.TrimSpace(r.URL.Query().Get("offset")); s != "" {
		if n, err := strconvAtoiSafe(s); err == nil && n >= 0 {
			offset = n
		}
	}

	if a.rt == nil || a.rt.Repos.Session == nil || a.rt.Repos.Event == nil {
		WriteError(w, http.StatusBadRequest, "仓储未初始化")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	sess, err := a.rt.Repos.Session.GetByID(ctx, id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sess == nil {
		WriteError(w, http.StatusNotFound, "session not found")
		return
	}

	events, err := a.rt.Repos.Event.GetByTimeRange(ctx, sess.StartTime, sess.EndTime)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if offset >= len(events) {
		WriteJSON(w, http.StatusOK, []dto.SessionWindowEventDTO{})
		return
	}
	events = events[offset:]
	if limit > 0 && limit < len(events) {
		events = events[:limit]
	}

	out := make([]dto.SessionWindowEventDTO, 0, len(events))
	for _, e := range events {
		out = append(out, dto.SessionWindowEventDTO{
			Timestamp: e.Timestamp,
			AppName:   e.AppName,
			Title:     e.Title,
			Duration:  e.Duration,
		})
	}
	WriteJSON(w, http.StatusOK, out)
}
