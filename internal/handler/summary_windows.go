//go:build windows

package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/dto"
)

func (a *API) HandleTodaySummary(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.AI == nil {
		WriteError(w, http.StatusBadRequest, "AI 服务未初始化，请检查配置与数据库")
		return
	}

	today := time.Now().Format("2006-01-02")
	summary, err := a.rt.Core.Services.AI.GenerateDailySummary(ctx, today)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, &dto.DailySummaryDTO{
		Date:         summary.Date,
		Summary:      summary.Summary,
		Highlights:   summary.Highlights,
		Struggles:    summary.Struggles,
		SkillsGained: summary.SkillsGained,
		TotalCoding:  summary.TotalCoding,
		TotalDiffs:   summary.TotalDiffs,
	})
}

func (a *API) HandleSummaryIndex(w http.ResponseWriter, r *http.Request) {
	limit := 365
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		if n, err := strconvAtoiSafe(s); err == nil && n > 0 {
			limit = n
		}
	}

	if a.rt == nil || a.rt.Repos.Summary == nil {
		WriteError(w, http.StatusBadRequest, "总结仓储未初始化")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	previews, err := a.rt.Repos.Summary.ListSummaryPreviews(ctx, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]dto.SummaryIndexDTO, 0, len(previews))
	for _, p := range previews {
		result = append(result, dto.SummaryIndexDTO{
			Date:       p.Date,
			HasSummary: true,
			Preview:    p.Preview,
		})
	}
	WriteJSON(w, http.StatusOK, result)
}

func (a *API) HandleDailySummary(w http.ResponseWriter, r *http.Request) {
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		WriteError(w, http.StatusBadRequest, "date 不能为空")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.AI == nil {
		WriteError(w, http.StatusBadRequest, "AI 服务未初始化，请检查配置与数据库")
		return
	}

	summary, err := a.rt.Core.Services.AI.GenerateDailySummary(ctx, date)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, &dto.DailySummaryDTO{
		Date:         summary.Date,
		Summary:      summary.Summary,
		Highlights:   summary.Highlights,
		Struggles:    summary.Struggles,
		SkillsGained: summary.SkillsGained,
		TotalCoding:  summary.TotalCoding,
		TotalDiffs:   summary.TotalDiffs,
	})
}
