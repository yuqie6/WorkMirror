//go:build windows

package handler

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/bootstrap"
	"github.com/yuqie6/mirror/internal/dto"
	"github.com/yuqie6/mirror/internal/eventbus"
	"github.com/yuqie6/mirror/internal/pkg/config"
	"github.com/yuqie6/mirror/internal/schema"
	"github.com/yuqie6/mirror/internal/service"
)

// ========== handlers ==========

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

	if a.rt.Repos.PeriodSummary != nil {
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

func (a *API) HandleSkillTree(w http.ResponseWriter, r *http.Request) {
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Skills == nil {
		WriteError(w, http.StatusBadRequest, "技能服务未初始化")
		return
	}
	skillTree, err := a.rt.Core.Services.Skills.GetSkillTree(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var result []dto.SkillNodeDTO
	for category, skills := range skillTree.Categories {
		for _, skill := range skills {
			result = append(result, dto.SkillNodeDTO{
				Key:        skill.Key,
				Name:       skill.Name,
				Category:   category,
				ParentKey:  skill.ParentKey,
				Level:      skill.Level,
				Experience: int(skill.Exp),
				Progress:   int(skill.Progress),
				Status:     skill.Trend,
				LastActive: skill.LastActive.UnixMilli(),
			})
		}
	}
	WriteJSON(w, http.StatusOK, result)
}

func (a *API) HandleSkillEvidence(w http.ResponseWriter, r *http.Request) {
	skillKey := strings.TrimSpace(r.URL.Query().Get("skill_key"))
	if skillKey == "" {
		WriteError(w, http.StatusBadRequest, "skill_key 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Skills == nil {
		WriteError(w, http.StatusBadRequest, "技能服务未初始化")
		return
	}
	evs, err := a.rt.Core.Services.Skills.GetSkillEvidence(r.Context(), skillKey, 3)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]dto.SkillEvidenceDTO, len(evs))
	for i, e := range evs {
		result[i] = dto.SkillEvidenceDTO{
			Source:              e.Source,
			EvidenceID:          e.EvidenceID,
			Timestamp:           e.Timestamp,
			ContributionContext: e.ContributionContext,
			FileName:            e.FileName,
		}
	}
	WriteJSON(w, http.StatusOK, result)
}

func (a *API) HandleSkillSessions(w http.ResponseWriter, r *http.Request) {
	skillKey := strings.TrimSpace(r.URL.Query().Get("skill_key"))
	if skillKey == "" {
		WriteError(w, http.StatusBadRequest, "skill_key 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.SessionSemantic == nil {
		WriteError(w, http.StatusBadRequest, "会话语义服务未初始化")
		return
	}

	sessions, err := a.rt.Core.Services.SessionSemantic.GetSessionsBySkill(r.Context(), skillKey, 30*24*time.Hour, 10)
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
	sort.Slice(result, func(i, j int) bool { return result[i].StartTime > result[j].StartTime })
	WriteJSON(w, http.StatusOK, result)
}

func (a *API) HandleTrends(w http.ResponseWriter, r *http.Request) {
	days := 7
	if s := strings.TrimSpace(r.URL.Query().Get("days")); s != "" {
		if n, err := strconvAtoiSafe(s); err == nil && (n == 7 || n == 30) {
			days = n
		}
	}

	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Trends == nil {
		WriteError(w, http.StatusBadRequest, "趋势服务未初始化")
		return
	}

	period := service.TrendPeriod7Days
	if days == 30 {
		period = service.TrendPeriod30Days
	}
	report, err := a.rt.Core.Services.Trends.GetTrendReport(r.Context(), period)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	languages := make([]dto.LanguageTrendDTO, len(report.TopLanguages))
	for i, l := range report.TopLanguages {
		languages[i] = dto.LanguageTrendDTO{
			Language:   l.Language,
			DiffCount:  l.DiffCount,
			Percentage: l.Percentage,
		}
	}

	skills := make([]dto.SkillTrendDTO, len(report.TopSkills))
	for i, s := range report.TopSkills {
		skills[i] = dto.SkillTrendDTO{
			SkillName:   s.SkillName,
			Status:      s.Status,
			DaysActive:  s.DaysActive,
			Changes:     s.Changes,
			ExpGain:     s.ExpGain,
			PrevExpGain: s.PrevExpGain,
			GrowthRate:  s.GrowthRate,
		}
	}

	WriteJSON(w, http.StatusOK, &dto.TrendReportDTO{
		Period:          string(report.Period),
		StartDate:       report.StartDate,
		EndDate:         report.EndDate,
		TotalDiffs:      report.TotalDiffs,
		TotalCodingMins: report.TotalCodingMins,
		AvgDiffsPerDay:  report.AvgDiffsPerDay,
		TopLanguages:    languages,
		TopSkills:       skills,
		Bottlenecks:     report.Bottlenecks,
	})
}

func (a *API) HandleAppStats(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	startTime := now.AddDate(0, 0, -7).UnixMilli()
	endTime := now.UnixMilli()

	if a.rt == nil || a.rt.Repos.Event == nil {
		WriteError(w, http.StatusBadRequest, "数据库未初始化")
		return
	}
	stats, err := a.rt.Repos.Event.GetAppStats(r.Context(), startTime, endTime)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]dto.AppStatsDTO, len(stats))
	for i, s := range stats {
		result[i] = dto.AppStatsDTO{
			AppName:       s.AppName,
			TotalDuration: s.TotalDuration,
			EventCount:    s.EventCount,
			IsCodeEditor:  service.IsCodeEditor(s.AppName),
		}
	}
	WriteJSON(w, http.StatusOK, result)
}

func (a *API) HandleDiffDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := parseInt64Param(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "id 无效")
		return
	}

	if a.rt == nil || a.rt.Repos.Diff == nil {
		WriteError(w, http.StatusBadRequest, "Diff 仓储未初始化")
		return
	}

	diff, err := a.rt.Repos.Diff.GetByID(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if diff == nil {
		WriteError(w, http.StatusNotFound, "Diff not found")
		return
	}

	var skills []string
	if len(diff.SkillsDetected) > 0 {
		skills = []string(diff.SkillsDetected)
	}

	WriteJSON(w, http.StatusOK, &dto.DiffDetailDTO{
		ID:           diff.ID,
		FileName:     diff.FileName,
		Language:     diff.Language,
		DiffContent:  diff.DiffContent,
		Insight:      diff.AIInsight,
		Skills:       skills,
		LinesAdded:   diff.LinesAdded,
		LinesDeleted: diff.LinesDeleted,
		Timestamp:    diff.Timestamp,
	})
}

func (a *API) HandleBuildSessionsForDate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date string `json:"date"`
	}
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
	WriteJSON(w, http.StatusOK, &dto.SessionBuildResultDTO{Created: created})
}

func (a *API) HandleRebuildSessionsForDate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date string `json:"date"`
	}
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
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	created, err := a.rt.Core.Services.Sessions.RebuildSessionsForDate(ctx, date)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, &dto.SessionBuildResultDTO{Created: created})
}

func (a *API) HandleEnrichSessionsForDate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date string `json:"date"`
	}
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

	a.hub.Publish(eventbus.Event{Type: "settings_updated"})
	WriteJSON(w, http.StatusOK, &dto.SaveSettingsResponseDTO{RestartRequired: true})
}

func strconvAtoiSafe(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty")
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}
