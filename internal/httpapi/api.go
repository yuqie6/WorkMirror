//go:build windows

package httpapi

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/bootstrap"
	"github.com/yuqie6/mirror/internal/eventbus"
	"github.com/yuqie6/mirror/internal/model"
	"github.com/yuqie6/mirror/internal/pkg/config"
	"github.com/yuqie6/mirror/internal/service"
)

// ========== DTOs（与前端契约保持稳定） ==========

type DailySummaryDTO struct {
	Date         string   `json:"date"`
	Summary      string   `json:"summary"`
	Highlights   string   `json:"highlights"`
	Struggles    string   `json:"struggles"`
	SkillsGained []string `json:"skills_gained"`
	TotalCoding  int      `json:"total_coding"`
	TotalDiffs   int      `json:"total_diffs"`
}

type SummaryIndexDTO struct {
	Date       string `json:"date"`
	HasSummary bool   `json:"has_summary"`
	Preview    string `json:"preview"`
}

type PeriodSummaryDTO struct {
	Type         string   `json:"type"`
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	Overview     string   `json:"overview"`
	Achievements []string `json:"achievements"`
	Patterns     string   `json:"patterns"`
	Suggestions  string   `json:"suggestions"`
	TopSkills    []string `json:"top_skills"`
	TotalCoding  int      `json:"total_coding"`
	TotalDiffs   int      `json:"total_diffs"`
}

type PeriodSummaryIndexDTO struct {
	Type      string `json:"type"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type SkillNodeDTO struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	ParentKey  string `json:"parent_key"`
	Level      int    `json:"level"`
	Experience int    `json:"experience"`
	Progress   int    `json:"progress"`
	Status     string `json:"status"`
	LastActive int64  `json:"last_active"`
}

type SkillEvidenceDTO struct {
	Source              string `json:"source"`
	EvidenceID          int64  `json:"evidence_id"`
	Timestamp           int64  `json:"timestamp"`
	ContributionContext string `json:"contribution_context"`
	FileName            string `json:"file_name"`
}

type TrendReportDTO struct {
	Period          string             `json:"period"`
	StartDate       string             `json:"start_date"`
	EndDate         string             `json:"end_date"`
	TotalDiffs      int64              `json:"total_diffs"`
	TotalCodingMins int64              `json:"total_coding_mins"`
	AvgDiffsPerDay  float64            `json:"avg_diffs_per_day"`
	TopLanguages    []LanguageTrendDTO `json:"top_languages"`
	TopSkills       []SkillTrendDTO    `json:"top_skills"`
	Bottlenecks     []string           `json:"bottlenecks"`
}

type LanguageTrendDTO struct {
	Language   string  `json:"language"`
	DiffCount  int64   `json:"diff_count"`
	Percentage float64 `json:"percentage"`
}

type SkillTrendDTO struct {
	SkillName   string  `json:"skill_name"`
	Status      string  `json:"status"`
	DaysActive  int     `json:"days_active"`
	Changes     int     `json:"changes"`
	ExpGain     float64 `json:"exp_gain"`
	PrevExpGain float64 `json:"prev_exp_gain"`
	GrowthRate  float64 `json:"growth_rate"`
}

type AppStatsDTO struct {
	AppName       string `json:"app_name"`
	TotalDuration int    `json:"total_duration"`
	EventCount    int64  `json:"event_count"`
	IsCodeEditor  bool   `json:"is_code_editor"`
}

type DiffDetailDTO struct {
	ID           int64    `json:"id"`
	FileName     string   `json:"file_name"`
	Language     string   `json:"language"`
	DiffContent  string   `json:"diff_content"`
	Insight      string   `json:"insight"`
	Skills       []string `json:"skills"`
	LinesAdded   int      `json:"lines_added"`
	LinesDeleted int      `json:"lines_deleted"`
	Timestamp    int64    `json:"timestamp"`
}

type SettingsDTO struct {
	ConfigPath string `json:"config_path"`

	DeepSeekAPIKeySet bool   `json:"deepseek_api_key_set"`
	DeepSeekBaseURL   string `json:"deepseek_base_url"`
	DeepSeekModel     string `json:"deepseek_model"`

	SiliconFlowAPIKeySet      bool   `json:"siliconflow_api_key_set"`
	SiliconFlowBaseURL        string `json:"siliconflow_base_url"`
	SiliconFlowEmbeddingModel string `json:"siliconflow_embedding_model"`
	SiliconFlowRerankerModel  string `json:"siliconflow_reranker_model"`

	DBPath             string   `json:"db_path"`
	DiffWatchPaths     []string `json:"diff_watch_paths"`
	BrowserHistoryPath string   `json:"browser_history_path"`
}

type SaveSettingsRequestDTO struct {
	DeepSeekAPIKey  *string `json:"deepseek_api_key"`
	DeepSeekBaseURL *string `json:"deepseek_base_url"`
	DeepSeekModel   *string `json:"deepseek_model"`

	SiliconFlowAPIKey         *string `json:"siliconflow_api_key"`
	SiliconFlowBaseURL        *string `json:"siliconflow_base_url"`
	SiliconFlowEmbeddingModel *string `json:"siliconflow_embedding_model"`
	SiliconFlowRerankerModel  *string `json:"siliconflow_reranker_model"`

	DBPath             *string   `json:"db_path"`
	DiffWatchPaths     *[]string `json:"diff_watch_paths"`
	BrowserHistoryPath *string   `json:"browser_history_path"`
}

type SaveSettingsResponseDTO struct {
	RestartRequired bool `json:"restart_required"`
}

type SessionDTO struct {
	ID             int64    `json:"id"`
	Date           string   `json:"date"`
	StartTime      int64    `json:"start_time"`
	EndTime        int64    `json:"end_time"`
	TimeRange      string   `json:"time_range"`
	PrimaryApp     string   `json:"primary_app"`
	Category       string   `json:"category"`
	Summary        string   `json:"summary"`
	SkillsInvolved []string `json:"skills_involved"`
	DiffCount      int      `json:"diff_count"`
	BrowserCount   int      `json:"browser_count"`
}

type SessionAppUsageDTO struct {
	AppName       string `json:"app_name"`
	TotalDuration int    `json:"total_duration"`
}

type SessionDiffDTO struct {
	ID           int64    `json:"id"`
	FileName     string   `json:"file_name"`
	Language     string   `json:"language"`
	Insight      string   `json:"insight"`
	Skills       []string `json:"skills"`
	LinesAdded   int      `json:"lines_added"`
	LinesDeleted int      `json:"lines_deleted"`
	Timestamp    int64    `json:"timestamp"`
}

type SessionBrowserEventDTO struct {
	ID        int64  `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Domain    string `json:"domain"`
	Title     string `json:"title"`
	URL       string `json:"url"`
}

type SessionDetailDTO struct {
	SessionDTO
	Tags     []string                 `json:"tags"`
	RAGRefs  []map[string]any         `json:"rag_refs"`
	AppUsage []SessionAppUsageDTO     `json:"app_usage"`
	Diffs    []SessionDiffDTO         `json:"diffs"`
	Browser  []SessionBrowserEventDTO `json:"browser"`
}

type SessionBuildResultDTO struct {
	Created int `json:"created"`
}

type SessionEnrichResultDTO struct {
	Enriched int `json:"enriched"`
}

// ========== routes ==========

func (a *apiServer) registerJSONRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/summary/today", a.wrapGET(a.getTodaySummary))
	mux.HandleFunc("/api/summary/daily", a.wrapGET(a.getDailySummary))
	mux.HandleFunc("/api/summary/index", a.wrapGET(a.listSummaryIndex))
	mux.HandleFunc("/api/summary/period", a.wrapGET(a.getPeriodSummary))
	mux.HandleFunc("/api/summary/period/index", a.wrapGET(a.listPeriodSummaryIndex))

	mux.HandleFunc("/api/skills/tree", a.wrapGET(a.getSkillTree))
	mux.HandleFunc("/api/skills/evidence", a.wrapGET(a.getSkillEvidence))
	mux.HandleFunc("/api/skills/sessions", a.wrapGET(a.getSkillSessions))

	mux.HandleFunc("/api/trends", a.wrapGET(a.getTrends))
	mux.HandleFunc("/api/app-stats", a.wrapGET(a.getAppStats))

	mux.HandleFunc("/api/diffs/detail", a.wrapGET(a.getDiffDetail))

	mux.HandleFunc("/api/sessions/by-date", a.wrapGET(a.getSessionsByDate))
	mux.HandleFunc("/api/sessions/detail", a.wrapGET(a.getSessionDetail))
	mux.HandleFunc("/api/sessions/build", a.wrapPOST(a.buildSessionsForDate))
	mux.HandleFunc("/api/sessions/rebuild", a.wrapPOST(a.rebuildSessionsForDate))
	mux.HandleFunc("/api/sessions/enrich", a.wrapPOST(a.enrichSessionsForDate))

	mux.HandleFunc("/api/settings", a.wrapAny(a.settings))
}

func (a *apiServer) wrapGET(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		fn(w, r)
	}
}

func (a *apiServer) wrapPOST(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		fn(w, r)
	}
}

func (a *apiServer) wrapAny(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { fn(w, r) }
}

// ========== handlers ==========

func (a *apiServer) getTodaySummary(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.AI == nil {
		writeError(w, http.StatusBadRequest, "AI 服务未初始化，请检查配置与数据库")
		return
	}

	today := time.Now().Format("2006-01-02")
	summary, err := a.rt.Core.Services.AI.GenerateDailySummary(ctx, today)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &DailySummaryDTO{
		Date:         summary.Date,
		Summary:      summary.Summary,
		Highlights:   summary.Highlights,
		Struggles:    summary.Struggles,
		SkillsGained: summary.SkillsGained,
		TotalCoding:  summary.TotalCoding,
		TotalDiffs:   summary.TotalDiffs,
	})
}

func (a *apiServer) listSummaryIndex(w http.ResponseWriter, r *http.Request) {
	limit := 365
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		if n, err := strconvAtoiSafe(s); err == nil && n > 0 {
			limit = n
		}
	}

	if a.rt == nil || a.rt.Repos.Summary == nil {
		writeError(w, http.StatusBadRequest, "总结仓储未初始化")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	previews, err := a.rt.Repos.Summary.ListSummaryPreviews(ctx, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]SummaryIndexDTO, 0, len(previews))
	for _, p := range previews {
		result = append(result, SummaryIndexDTO{
			Date:       p.Date,
			HasSummary: true,
			Preview:    p.Preview,
		})
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *apiServer) getDailySummary(w http.ResponseWriter, r *http.Request) {
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		writeError(w, http.StatusBadRequest, "date 不能为空")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.AI == nil {
		writeError(w, http.StatusBadRequest, "AI 服务未初始化，请检查配置与数据库")
		return
	}

	summary, err := a.rt.Core.Services.AI.GenerateDailySummary(ctx, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &DailySummaryDTO{
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

func (a *apiServer) getPeriodSummary(w http.ResponseWriter, r *http.Request) {
	periodType := strings.TrimSpace(r.URL.Query().Get("type"))
	startDateStr := strings.TrimSpace(r.URL.Query().Get("start_date"))
	if periodType == "" {
		writeError(w, http.StatusBadRequest, "type 不能为空")
		return
	}

	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.AI == nil {
		writeError(w, http.StatusBadRequest, "AI 服务未初始化")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	var startDate, endDate time.Time
	now := time.Now()

	if startDateStr != "" {
		parsed, err := time.ParseInLocation("2006-01-02", startDateStr, now.Location())
		if err != nil {
			writeError(w, http.StatusBadRequest, "日期格式错误，请使用 YYYY-MM-DD")
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
			writeError(w, http.StatusBadRequest, "startDate 不能是未来日期")
			return
		}
		endDate = startDate.AddDate(0, 0, 6)
	case "month":
		if startDateStr == "" {
			startDate = now
		}
		startDate = normalizeToMonthStart(startDate)
		if startDate.After(now) {
			writeError(w, http.StatusBadRequest, "startDate 不能是未来日期")
			return
		}
		endDate = startDate.AddDate(0, 1, -1)
	default:
		writeError(w, http.StatusBadRequest, "不支持的周期类型，请使用 week 或 month")
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
			writeJSON(w, http.StatusOK, periodSummaryToDTO(cached))
			return
		}
	}

	summaries, err := a.rt.Repos.Summary.GetByDateRange(ctx, startStr, dataEndStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(summaries) == 0 {
		writeError(w, http.StatusBadRequest, "该周期内没有日报数据")
		return
	}

	var totalCoding, totalDiffs int
	for _, s := range summaries {
		totalCoding += s.TotalCoding
		totalDiffs += s.TotalDiffs
	}

	aiResult, err := a.rt.Core.Services.AI.GeneratePeriodSummary(ctx, periodType, startStr, dataEndStr, summaries)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	overview := normalizePeriodWording(periodType, aiResult.Overview)
	if dataEndStr != endStr {
		overview = "（截至 " + dataEndStr + "）" + overview
	}

	result := &PeriodSummaryDTO{
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

	writeJSON(w, http.StatusOK, result)
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

func periodSummaryToDTO(ps *model.PeriodSummary) *PeriodSummaryDTO {
	if ps == nil {
		return nil
	}
	return &PeriodSummaryDTO{
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

func savePeriodSummary(ctx context.Context, rt *bootstrap.AgentRuntime, dto *PeriodSummaryDTO) {
	if rt == nil || rt.Repos.PeriodSummary == nil || dto == nil {
		return
	}
	_ = rt.Repos.PeriodSummary.Upsert(ctx, &model.PeriodSummary{
		Type:         dto.Type,
		StartDate:    dto.StartDate,
		EndDate:      dto.EndDate,
		Overview:     dto.Overview,
		Achievements: model.JSONArray(dto.Achievements),
		Patterns:     dto.Patterns,
		Suggestions:  dto.Suggestions,
		TopSkills:    model.JSONArray(dto.TopSkills),
		TotalCoding:  dto.TotalCoding,
		TotalDiffs:   dto.TotalDiffs,
	})
}

func (a *apiServer) listPeriodSummaryIndex(w http.ResponseWriter, r *http.Request) {
	periodType := strings.TrimSpace(r.URL.Query().Get("type"))
	if periodType == "" {
		writeError(w, http.StatusBadRequest, "type 不能为空")
		return
	}
	limit := 20
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		if n, err := strconvAtoiSafe(s); err == nil && n > 0 {
			limit = n
		}
	}

	if a.rt == nil || a.rt.Repos.PeriodSummary == nil {
		writeError(w, http.StatusBadRequest, "仓储未初始化")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	summaries, err := a.rt.Repos.PeriodSummary.ListByType(ctx, periodType, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]PeriodSummaryIndexDTO, 0, len(summaries))
	for _, s := range summaries {
		result = append(result, PeriodSummaryIndexDTO{
			Type:      s.Type,
			StartDate: s.StartDate,
			EndDate:   s.EndDate,
		})
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *apiServer) getSkillTree(w http.ResponseWriter, r *http.Request) {
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Skills == nil {
		writeError(w, http.StatusBadRequest, "技能服务未初始化")
		return
	}
	skillTree, err := a.rt.Core.Services.Skills.GetSkillTree(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var result []SkillNodeDTO
	for category, skills := range skillTree.Categories {
		for _, skill := range skills {
			result = append(result, SkillNodeDTO{
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
	writeJSON(w, http.StatusOK, result)
}

func (a *apiServer) getSkillEvidence(w http.ResponseWriter, r *http.Request) {
	skillKey := strings.TrimSpace(r.URL.Query().Get("skill_key"))
	if skillKey == "" {
		writeError(w, http.StatusBadRequest, "skill_key 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Skills == nil {
		writeError(w, http.StatusBadRequest, "技能服务未初始化")
		return
	}
	evs, err := a.rt.Core.Services.Skills.GetSkillEvidence(r.Context(), skillKey, 3)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]SkillEvidenceDTO, len(evs))
	for i, e := range evs {
		result[i] = SkillEvidenceDTO{
			Source:              e.Source,
			EvidenceID:          e.EvidenceID,
			Timestamp:           e.Timestamp,
			ContributionContext: e.ContributionContext,
			FileName:            e.FileName,
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *apiServer) getSkillSessions(w http.ResponseWriter, r *http.Request) {
	skillKey := strings.TrimSpace(r.URL.Query().Get("skill_key"))
	if skillKey == "" {
		writeError(w, http.StatusBadRequest, "skill_key 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.SessionSemantic == nil {
		writeError(w, http.StatusBadRequest, "会话语义服务未初始化")
		return
	}

	sessions, err := a.rt.Core.Services.SessionSemantic.GetSessionsBySkill(r.Context(), skillKey, 30*24*time.Hour, 10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]SessionDTO, 0, len(sessions))
	for _, s := range sessions {
		diffIDs := toInt64Slice(s.Metadata, "diff_ids")
		browserIDs := toInt64Slice(s.Metadata, "browser_event_ids")
		timeRange := strings.TrimSpace(s.TimeRange)
		if timeRange == "" {
			timeRange = formatTimeRangeMs(s.StartTime, s.EndTime)
		}
		result = append(result, SessionDTO{
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
	writeJSON(w, http.StatusOK, result)
}

func (a *apiServer) getTrends(w http.ResponseWriter, r *http.Request) {
	days := 7
	if s := strings.TrimSpace(r.URL.Query().Get("days")); s != "" {
		if n, err := strconvAtoiSafe(s); err == nil && (n == 7 || n == 30) {
			days = n
		}
	}

	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Trends == nil {
		writeError(w, http.StatusBadRequest, "趋势服务未初始化")
		return
	}

	period := service.TrendPeriod7Days
	if days == 30 {
		period = service.TrendPeriod30Days
	}
	report, err := a.rt.Core.Services.Trends.GetTrendReport(r.Context(), period)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	languages := make([]LanguageTrendDTO, len(report.TopLanguages))
	for i, l := range report.TopLanguages {
		languages[i] = LanguageTrendDTO{
			Language:   l.Language,
			DiffCount:  l.DiffCount,
			Percentage: l.Percentage,
		}
	}

	skills := make([]SkillTrendDTO, len(report.TopSkills))
	for i, s := range report.TopSkills {
		skills[i] = SkillTrendDTO{
			SkillName:   s.SkillName,
			Status:      s.Status,
			DaysActive:  s.DaysActive,
			Changes:     s.Changes,
			ExpGain:     s.ExpGain,
			PrevExpGain: s.PrevExpGain,
			GrowthRate:  s.GrowthRate,
		}
	}

	writeJSON(w, http.StatusOK, &TrendReportDTO{
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

func (a *apiServer) getAppStats(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	startTime := now.AddDate(0, 0, -7).UnixMilli()
	endTime := now.UnixMilli()

	if a.rt == nil || a.rt.Repos.Event == nil {
		writeError(w, http.StatusBadRequest, "数据库未初始化")
		return
	}
	stats, err := a.rt.Repos.Event.GetAppStats(r.Context(), startTime, endTime)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]AppStatsDTO, len(stats))
	for i, s := range stats {
		result[i] = AppStatsDTO{
			AppName:       s.AppName,
			TotalDuration: s.TotalDuration,
			EventCount:    s.EventCount,
			IsCodeEditor:  service.IsCodeEditor(s.AppName),
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *apiServer) getDiffDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := parseInt64Param(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id 无效")
		return
	}

	if a.rt == nil || a.rt.Repos.Diff == nil {
		writeError(w, http.StatusBadRequest, "Diff 仓储未初始化")
		return
	}

	diff, err := a.rt.Repos.Diff.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if diff == nil {
		writeError(w, http.StatusNotFound, "Diff not found")
		return
	}

	var skills []string
	if len(diff.SkillsDetected) > 0 {
		skills = []string(diff.SkillsDetected)
	}

	writeJSON(w, http.StatusOK, &DiffDetailDTO{
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

func (a *apiServer) buildSessionsForDate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date string `json:"date"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	date := strings.TrimSpace(req.Date)
	if date == "" {
		writeError(w, http.StatusBadRequest, "date 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Sessions == nil {
		writeError(w, http.StatusBadRequest, "会话服务未初始化")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	created, err := a.rt.Core.Services.Sessions.BuildSessionsForDate(ctx, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, &SessionBuildResultDTO{Created: created})
}

func (a *apiServer) rebuildSessionsForDate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date string `json:"date"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	date := strings.TrimSpace(req.Date)
	if date == "" {
		writeError(w, http.StatusBadRequest, "date 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Sessions == nil {
		writeError(w, http.StatusBadRequest, "会话服务未初始化")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	created, err := a.rt.Core.Services.Sessions.RebuildSessionsForDate(ctx, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, &SessionBuildResultDTO{Created: created})
}

func (a *apiServer) enrichSessionsForDate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date string `json:"date"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	date := strings.TrimSpace(req.Date)
	if date == "" {
		writeError(w, http.StatusBadRequest, "date 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.SessionSemantic == nil {
		writeError(w, http.StatusBadRequest, "会话语义服务未初始化")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	enriched, err := a.rt.Core.Services.SessionSemantic.EnrichSessionsForDate(ctx, date, 200)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, &SessionEnrichResultDTO{Enriched: enriched})
}

func (a *apiServer) getSessionsByDate(w http.ResponseWriter, r *http.Request) {
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		writeError(w, http.StatusBadRequest, "date 不能为空")
		return
	}
	if a.rt == nil || a.rt.Repos.Session == nil {
		writeError(w, http.StatusBadRequest, "会话仓储未初始化")
		return
	}
	sessions, err := a.rt.Repos.Session.GetByDate(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]SessionDTO, 0, len(sessions))
	for _, s := range sessions {
		diffIDs := toInt64Slice(s.Metadata, "diff_ids")
		browserIDs := toInt64Slice(s.Metadata, "browser_event_ids")
		timeRange := strings.TrimSpace(s.TimeRange)
		if timeRange == "" {
			timeRange = formatTimeRangeMs(s.StartTime, s.EndTime)
		}
		result = append(result, SessionDTO{
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
	writeJSON(w, http.StatusOK, result)
}

func (a *apiServer) getSessionDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := parseInt64Param(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id 无效")
		return
	}
	if a.rt == nil || a.rt.Repos.Session == nil {
		writeError(w, http.StatusBadRequest, "会话仓储未初始化")
		return
	}
	sess, err := a.rt.Repos.Session.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sess == nil {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	diffIDs := toInt64Slice(sess.Metadata, "diff_ids")
	browserIDs := toInt64Slice(sess.Metadata, "browser_event_ids")

	var diffs []model.Diff
	if len(diffIDs) > 0 {
		diffs, _ = a.rt.Repos.Diff.GetByIDs(r.Context(), diffIDs)
	}
	diffDTOs := make([]SessionDiffDTO, 0, len(diffs))
	for _, d := range diffs {
		diffDTOs = append(diffDTOs, SessionDiffDTO{
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

	var browserEvents []model.BrowserEvent
	if len(browserIDs) > 0 && a.rt.Repos.Browser != nil {
		browserEvents, _ = a.rt.Repos.Browser.GetByIDs(r.Context(), browserIDs)
	}
	browserDTOs := make([]SessionBrowserEventDTO, 0, len(browserEvents))
	for _, e := range browserEvents {
		browserDTOs = append(browserDTOs, SessionBrowserEventDTO{
			ID:        e.ID,
			Timestamp: e.Timestamp,
			Domain:    e.Domain,
			Title:     e.Title,
			URL:       e.URL,
		})
	}

	appStats, _ := a.rt.Repos.Event.GetAppStats(r.Context(), sess.StartTime, sess.EndTime)
	appUsage := make([]SessionAppUsageDTO, 0, len(appStats))
	for i, st := range appStats {
		if i >= 8 {
			break
		}
		appUsage = append(appUsage, SessionAppUsageDTO{
			AppName:       st.AppName,
			TotalDuration: st.TotalDuration,
		})
	}

	timeRange := strings.TrimSpace(sess.TimeRange)
	if timeRange == "" {
		timeRange = formatTimeRangeMs(sess.StartTime, sess.EndTime)
	}

	dto := &SessionDetailDTO{
		SessionDTO: SessionDTO{
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
		Tags:     toStringSlice(sess.Metadata, "tags"),
		RAGRefs:  toMapSlice(sess.Metadata, "rag_refs"),
		AppUsage: appUsage,
		Diffs:    diffDTOs,
		Browser:  browserDTOs,
	}
	writeJSON(w, http.StatusOK, dto)
}

func (a *apiServer) settings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.getSettings(w, r)
	case http.MethodPost:
		a.saveSettings(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *apiServer) getSettings(w http.ResponseWriter, r *http.Request) {
	path, err := config.DefaultConfigPath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	cfg, err := config.Load(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &SettingsDTO{
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

func (a *apiServer) saveSettings(w http.ResponseWriter, r *http.Request) {
	var req SaveSettingsRequestDTO
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	path, err := config.DefaultConfigPath()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cur, err := config.Load(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
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
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	a.hub.Publish(eventbus.Event{Type: "settings_updated"})
	writeJSON(w, http.StatusOK, &SaveSettingsResponseDTO{RestartRequired: true})
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
