//go:build windows

package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/yuqie6/WorkMirror/internal/dto"
	"github.com/yuqie6/WorkMirror/internal/service"
)

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

	dailyStats := make([]dto.DailyTrendStatDTO, 0, len(report.DailyStats))
	for _, st := range report.DailyStats {
		dailyStats = append(dailyStats, dto.DailyTrendStatDTO{
			Date:           st.Date,
			TotalDiffs:      st.TotalDiffs,
			TotalCodingMins: st.TotalCodingMins,
			SessionCount:    st.SessionCount,
		})
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
		DailyStats:      dailyStats,
	})
}

func (a *API) HandleAppStats(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	startTime := now.AddDate(0, 0, -7).UnixMilli()
	endTime := now.UnixMilli()

	if date := strings.TrimSpace(r.URL.Query().Get("date")); date != "" {
		t, err := time.ParseInLocation("2006-01-02", date, time.Local)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "日期格式错误，请使用 YYYY-MM-DD")
			return
		}
		startTime = t.UnixMilli()
		endTime = t.Add(24*time.Hour).UnixMilli() - 1
	}

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
