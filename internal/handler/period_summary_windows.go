//go:build windows

package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/bootstrap"
	"github.com/yuqie6/mirror/internal/dto"
	"github.com/yuqie6/mirror/internal/schema"
)

func normalizeToMonday(t time.Time) time.Time {
	dayStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	weekday := int(dayStart.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return dayStart.AddDate(0, 0, -(weekday - 1))
}

func normalizeToMonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func (a *API) HandlePeriodSummary(w http.ResponseWriter, r *http.Request) {
	periodType := strings.TrimSpace(r.URL.Query().Get("type"))
	startDateStr := strings.TrimSpace(r.URL.Query().Get("start_date"))
	force := strings.TrimSpace(r.URL.Query().Get("force")) == "1"
	if periodType == "" {
		WriteError(w, http.StatusBadRequest, "type 不能为空")
		return
	}

	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.AI == nil {
		WriteError(w, http.StatusBadRequest, "AI 服务未初始化")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	var startDate, endDate time.Time
	now := time.Now()

	if startDateStr != "" {
		parsed, err := time.ParseInLocation("2006-01-02", startDateStr, now.Location())
		if err != nil {
			WriteError(w, http.StatusBadRequest, "日期格式错误，请使用 YYYY-MM-DD")
			return
		}
		startDate = parsed
	}

	switch periodType {
	case "week":
		if startDateStr == "" {
			startDate = now
		}
		startDate = normalizeToMonday(startDate)
		if startDate.After(now) {
			WriteError(w, http.StatusBadRequest, "startDate 不能是未来日期")
			return
		}
		endDate = startDate.AddDate(0, 0, 6)
	case "month":
		if startDateStr == "" {
			startDate = now
		}
		startDate = normalizeToMonthStart(startDate)
		if startDate.After(now) {
			WriteError(w, http.StatusBadRequest, "startDate 不能是未来日期")
			return
		}
		endDate = startDate.AddDate(0, 1, -1)
	default:
		WriteError(w, http.StatusBadRequest, "不支持的周期类型，请使用 week 或 month")
		return
	}

	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	dataEnd := endDate
	if dataEnd.After(now) {
		dataEnd = now
	}
	dataEndStr := dataEnd.Format("2006-01-02")

	if !force && a.rt.Repos.PeriodSummary != nil {
		cached, err := a.rt.Repos.PeriodSummary.GetByTypeAndRange(ctx, periodType, startStr, endStr, 365*24*time.Hour)
		if err == nil && cached != nil {
			WriteJSON(w, http.StatusOK, periodSummaryToDTO(cached))
			return
		}
	}

	summaries, err := a.rt.Repos.Summary.GetByDateRange(ctx, startStr, dataEndStr)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(summaries) == 0 {
		WriteError(w, http.StatusBadRequest, "该周期内没有日报数据")
		return
	}

	var totalCoding, totalDiffs int
	for _, s := range summaries {
		totalCoding += s.TotalCoding
		totalDiffs += s.TotalDiffs
	}

	aiResult, err := a.rt.Core.Services.AI.GeneratePeriodSummary(ctx, periodType, startStr, dataEndStr, summaries)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	overview := normalizePeriodWording(periodType, aiResult.Overview)
	if dataEndStr != endStr {
		overview = "（截至 " + dataEndStr + "）" + overview
	}

	result := &dto.PeriodSummaryDTO{
		Type:         periodType,
		StartDate:    startStr,
		EndDate:      endStr,
		Overview:     overview,
		Achievements: normalizePeriodWordingList(periodType, aiResult.Achievements),
		Patterns:     normalizePeriodWording(periodType, aiResult.Patterns),
		Suggestions:  normalizePeriodWording(periodType, aiResult.Suggestions),
		TopSkills:    aiResult.TopSkills,
		TotalCoding:  totalCoding,
		TotalDiffs:   totalDiffs,
	}

	savePeriodSummary(ctx, a.rt, result)

	WriteJSON(w, http.StatusOK, result)
}

func normalizePeriodWording(periodType string, text string) string {
	t := strings.TrimSpace(text)
	if t == "" || periodType != "month" {
		return t
	}
	replacements := []struct {
		old string
		new string
	}{
		{"本周", "本月"},
		{"这周", "这个月"},
		{"下周", "下月"},
		{"一周", "一个月"},
	}
	for _, r := range replacements {
		if strings.HasPrefix(t, r.old) {
			return r.new + strings.TrimPrefix(t, r.old)
		}
	}
	return t
}

func normalizePeriodWordingList(periodType string, items []string) []string {
	if len(items) == 0 {
		return items
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, normalizePeriodWording(periodType, it))
	}
	return out
}

func periodSummaryToDTO(ps *schema.PeriodSummary) *dto.PeriodSummaryDTO {
	if ps == nil {
		return nil
	}
	return &dto.PeriodSummaryDTO{
		Type:         ps.Type,
		StartDate:    ps.StartDate,
		EndDate:      ps.EndDate,
		Overview:     ps.Overview,
		Achievements: []string(ps.Achievements),
		Patterns:     ps.Patterns,
		Suggestions:  ps.Suggestions,
		TopSkills:    []string(ps.TopSkills),
		TotalCoding:  ps.TotalCoding,
		TotalDiffs:   ps.TotalDiffs,
	}
}

func savePeriodSummary(ctx context.Context, rt *bootstrap.AgentRuntime, psDTO *dto.PeriodSummaryDTO) {
	if rt == nil || rt.Repos.PeriodSummary == nil || psDTO == nil {
		return
	}
	_ = rt.Repos.PeriodSummary.Upsert(ctx, &schema.PeriodSummary{
		Type:         psDTO.Type,
		StartDate:    psDTO.StartDate,
		EndDate:      psDTO.EndDate,
		Overview:     psDTO.Overview,
		Achievements: schema.JSONArray(psDTO.Achievements),
		Patterns:     psDTO.Patterns,
		Suggestions:  psDTO.Suggestions,
		TopSkills:    schema.JSONArray(psDTO.TopSkills),
		TotalCoding:  psDTO.TotalCoding,
		TotalDiffs:   psDTO.TotalDiffs,
	})
}

func (a *API) HandlePeriodSummaryIndex(w http.ResponseWriter, r *http.Request) {
	periodType := strings.TrimSpace(r.URL.Query().Get("type"))
	if periodType == "" {
		WriteError(w, http.StatusBadRequest, "type 不能为空")
		return
	}
	limit := 20
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		if n, err := strconvAtoiSafe(s); err == nil && n > 0 {
			limit = n
		}
	}

	if a.rt == nil || a.rt.Repos.PeriodSummary == nil {
		WriteError(w, http.StatusBadRequest, "仓储未初始化")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	summaries, err := a.rt.Repos.PeriodSummary.ListByType(ctx, periodType, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]dto.PeriodSummaryIndexDTO, 0, len(summaries))
	for _, s := range summaries {
		result = append(result, dto.PeriodSummaryIndexDTO{
			Type:      s.Type,
			StartDate: s.StartDate,
			EndDate:   s.EndDate,
		})
	}
	WriteJSON(w, http.StatusOK, result)
}
