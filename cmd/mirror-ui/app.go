package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/yuqie6/mirror/internal/bootstrap"
	"github.com/yuqie6/mirror/internal/model"
	"github.com/yuqie6/mirror/internal/pkg/config"
	"github.com/yuqie6/mirror/internal/service"
)

// App struct
type App struct {
	mu   sync.RWMutex
	ctx  context.Context
	cfg  *config.Config
	core *bootstrap.Core
}

// SettingsDTO 设置页读取 DTO
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

// SaveSettingsRequestDTO 设置页保存 DTO（指针表示可选字段）
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

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.mu.Lock()
	a.ctx = ctx
	a.mu.Unlock()

	startAgentOnStartup()

	core, err := bootstrap.NewCore("")
	a.mu.Lock()
	defer a.mu.Unlock()
	if err != nil {
		// UI 启动时不 panic，改为延迟报错
		a.core = nil
		a.cfg = &config.Config{}
		return
	}
	a.core = core
	a.cfg = core.Cfg
}

func (a *App) GetSettings() (*SettingsDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Cfg == nil {
		return nil, errors.New("配置未初始化")
	}

	path, err := config.DefaultConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := a.core.Cfg
	dto := &SettingsDTO{
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
	}
	return dto, nil
}

func (a *App) SaveSettings(req SaveSettingsRequestDTO) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.core == nil || a.core.Cfg == nil {
		return errors.New("配置未初始化")
	}

	path, err := config.DefaultConfigPath()
	if err != nil {
		return err
	}

	next := *a.core.Cfg
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
		return err
	}

	newCore, err := bootstrap.NewCore(path)
	if err != nil {
		return err
	}

	oldCore := a.core
	a.core = newCore
	a.cfg = newCore.Cfg

	_ = oldCore.Close()
	return nil
}

// DailySummaryDTO 每日总结 DTO
type DailySummaryDTO struct {
	Date         string   `json:"date"`
	Summary      string   `json:"summary"`
	Highlights   string   `json:"highlights"`
	Struggles    string   `json:"struggles"`
	SkillsGained []string `json:"skills_gained"`
	TotalCoding  int      `json:"total_coding"`
	TotalDiffs   int      `json:"total_diffs"`
}

// GetTodaySummary 获取今日总结
func (a *App) GetTodaySummary() (*DailySummaryDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 添加超时防止长时间阻塞
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	if a.core == nil || a.core.Services.AI == nil {
		return nil, errors.New("AI 服务未初始化，请检查配置与数据库")
	}

	today := time.Now().Format("2006-01-02")
	summary, err := a.core.Services.AI.GenerateDailySummary(ctx, today)
	if err != nil {
		return nil, err
	}

	return &DailySummaryDTO{
		Date:         summary.Date,
		Summary:      summary.Summary,
		Highlights:   summary.Highlights,
		Struggles:    summary.Struggles,
		SkillsGained: summary.SkillsGained,
		TotalCoding:  summary.TotalCoding,
		TotalDiffs:   summary.TotalDiffs,
	}, nil
}

// SummaryIndexDTO 日报索引（用于历史侧边栏）
type SummaryIndexDTO struct {
	Date       string `json:"date"`
	HasSummary bool   `json:"has_summary"`
	Preview    string `json:"preview"` // 摘要前40字
}

// ListSummaryIndex 获取所有已生成的日报索引（只返回有数据的日期）
func (a *App) ListSummaryIndex(limit int) ([]SummaryIndexDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Repos.Summary == nil {
		return nil, errors.New("总结仓储未初始化")
	}
	if limit <= 0 {
		limit = 365 // 最多一年
	}

	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
	defer cancel()

	// 只获取已生成日报的预览（按日期倒序）
	previews, err := a.core.Repos.Summary.ListSummaryPreviews(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 只返回有数据的日期
	result := make([]SummaryIndexDTO, 0, len(previews))
	for _, p := range previews {
		result = append(result, SummaryIndexDTO{
			Date:       p.Date,
			HasSummary: true,
			Preview:    p.Preview,
		})
	}
	return result, nil
}

// GetDailySummary 获取指定日期总结（优先读取缓存，必要时生成）
func (a *App) GetDailySummary(date string) (*DailySummaryDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	if a.core == nil || a.core.Services.AI == nil {
		return nil, errors.New("AI 服务未初始化，请检查配置与数据库")
	}
	if date == "" {
		return nil, errors.New("date 不能为空")
	}

	summary, err := a.core.Services.AI.GenerateDailySummary(ctx, date)
	if err != nil {
		return nil, err
	}

	return &DailySummaryDTO{
		Date:         summary.Date,
		Summary:      summary.Summary,
		Highlights:   summary.Highlights,
		Struggles:    summary.Struggles,
		SkillsGained: summary.SkillsGained,
		TotalCoding:  summary.TotalCoding,
		TotalDiffs:   summary.TotalDiffs,
	}, nil
}

// PeriodSummaryDTO 阶段汇总 DTO
type PeriodSummaryDTO struct {
	Type         string   `json:"type"` // "week" | "month"
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

// GetPeriodSummary 生成周/月汇总（带缓存，月汇总基于周汇总）
// startDate 可选：指定起始日期，为空则使用当前周/月
func (a *App) GetPeriodSummary(periodType string, startDateStr string) (*PeriodSummaryDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Services.AI == nil {
		return nil, errors.New("AI 服务未初始化")
	}

	ctx, cancel := context.WithTimeout(a.ctx, 90*time.Second)
	defer cancel()

	// 确定时间范围
	var startDate, endDate time.Time
	now := time.Now()

	if startDateStr != "" {
		// 使用指定日期
		parsed, err := time.ParseInLocation("2006-01-02", startDateStr, now.Location())
		if err != nil {
			return nil, errors.New("日期格式错误，请使用 YYYY-MM-DD")
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
			return nil, errors.New("startDate 不能是未来日期")
		}
		// 自然周周日为结束（用于稳定缓存 key）
		endDate = startDate.AddDate(0, 0, 6)
	case "month":
		if startDateStr == "" {
			startDate = now
		}
		startDate = normalizeToMonthStart(startDate)
		if startDate.After(now) {
			return nil, errors.New("startDate 不能是未来日期")
		}
		// 自然月月末为结束（用于稳定缓存 key）
		endDate = startDate.AddDate(0, 1, -1)
	default:
		return nil, errors.New("不支持的周期类型，请使用 week 或 month")
	}

	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	// 实际数据截止日期（避免未来日期）
	dataEnd := endDate
	if dataEnd.After(now) {
		dataEnd = now
	}
	dataEndStr := dataEnd.Format("2006-01-02")

	// 检查缓存（自然周期维度 key）
	if a.core.Repos.PeriodSummary != nil {
		cached, err := a.core.Repos.PeriodSummary.GetByTypeAndRange(ctx, periodType, startStr, endStr, 365*24*time.Hour)
		if err == nil && cached != nil {
			return a.periodSummaryToDTO(cached), nil
		}
	}

	// 周/月汇总：从日报生成
	summaries, err := a.core.Repos.Summary.GetByDateRange(ctx, startStr, dataEndStr)
	if err != nil {
		return nil, err
	}

	if len(summaries) == 0 {
		return nil, errors.New("该周期内没有日报数据")
	}

	var totalCoding, totalDiffs int
	for _, s := range summaries {
		totalCoding += s.TotalCoding
		totalDiffs += s.TotalDiffs
	}

	aiResult, err := a.core.Services.AI.GeneratePeriodSummary(ctx, startStr, dataEndStr, summaries)
	if err != nil {
		return nil, err
	}

	overview := aiResult.Overview
	if dataEndStr != endStr {
		overview = fmt.Sprintf("（截至 %s）%s", dataEndStr, overview)
	}

	result := &PeriodSummaryDTO{
		Type:         periodType,
		StartDate:    startStr,
		EndDate:      endStr,
		Overview:     overview,
		Achievements: aiResult.Achievements,
		Patterns:     aiResult.Patterns,
		Suggestions:  aiResult.Suggestions,
		TopSkills:    aiResult.TopSkills,
		TotalCoding:  totalCoding,
		TotalDiffs:   totalDiffs,
	}

	// 保存到缓存
	a.savePeriodSummary(ctx, result)

	return result, nil
}

func (a *App) periodSummaryToDTO(ps *model.PeriodSummary) *PeriodSummaryDTO {
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

func (a *App) savePeriodSummary(ctx context.Context, dto *PeriodSummaryDTO) {
	if a.core.Repos.PeriodSummary == nil {
		return
	}
	_ = a.core.Repos.PeriodSummary.Upsert(ctx, &model.PeriodSummary{
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

// PeriodSummaryIndexDTO 历史汇总索引
type PeriodSummaryIndexDTO struct {
	Type      string `json:"type"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// ListPeriodSummaryIndex 获取历史周/月汇总列表
func (a *App) ListPeriodSummaryIndex(periodType string, limit int) ([]PeriodSummaryIndexDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Repos.PeriodSummary == nil {
		return nil, errors.New("仓储未初始化")
	}
	if limit <= 0 {
		limit = 20
	}

	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
	defer cancel()

	summaries, err := a.core.Repos.PeriodSummary.ListByType(ctx, periodType, limit)
	if err != nil {
		return nil, err
	}

	result := make([]PeriodSummaryIndexDTO, 0, len(summaries))
	for _, s := range summaries {
		result = append(result, PeriodSummaryIndexDTO{
			Type:      s.Type,
			StartDate: s.StartDate,
			EndDate:   s.EndDate,
		})
	}
	return result, nil
}

// SkillNodeDTO 技能节点 DTO
type SkillNodeDTO struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	ParentKey  string `json:"parent_key"` // 父技能 Key
	Level      int    `json:"level"`
	Experience int    `json:"experience"`
	Progress   int    `json:"progress"`
	Status     string `json:"status"`
	LastActive int64  `json:"last_active"` // 最后活跃时间戳
}

// SkillEvidenceDTO 技能证据 DTO
type SkillEvidenceDTO struct {
	Source              string `json:"source"`
	EvidenceID          int64  `json:"evidence_id"`
	Timestamp           int64  `json:"timestamp"`
	ContributionContext string `json:"contribution_context"`
	FileName            string `json:"file_name"`
}

// GetSkillTree 获取技能树
func (a *App) GetSkillTree() ([]SkillNodeDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Services.Skills == nil {
		return nil, errors.New("技能服务未初始化")
	}

	skillTree, err := a.core.Services.Skills.GetSkillTree(a.ctx)
	if err != nil {
		return nil, err
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
	return result, nil
}

// GetSkillEvidence 获取技能最近证据（Phase B drill-down）
func (a *App) GetSkillEvidence(skillKey string) ([]SkillEvidenceDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Services.Skills == nil {
		return nil, errors.New("技能服务未初始化")
	}
	evs, err := a.core.Services.Skills.GetSkillEvidence(a.ctx, skillKey, 3)
	if err != nil {
		return nil, err
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
	return result, nil
}

// TrendReportDTO 趋势报告 DTO
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

// LanguageTrendDTO 语言趋势 DTO
type LanguageTrendDTO struct {
	Language   string  `json:"language"`
	DiffCount  int64   `json:"diff_count"`
	Percentage float64 `json:"percentage"`
}

// SkillTrendDTO 技能趋势 DTO
type SkillTrendDTO struct {
	SkillName  string `json:"skill_name"`
	Status     string `json:"status"`
	DaysActive int    `json:"days_active"`
}

// GetTrends 获取趋势报告
func (a *App) GetTrends(days int) (*TrendReportDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Services.Trends == nil {
		return nil, errors.New("趋势服务未初始化")
	}

	period := service.TrendPeriod7Days
	if days == 30 {
		period = service.TrendPeriod30Days
	}

	report, err := a.core.Services.Trends.GetTrendReport(a.ctx, period)
	if err != nil {
		return nil, err
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
			SkillName:  s.SkillName,
			Status:     s.Status,
			DaysActive: s.DaysActive,
		}
	}

	return &TrendReportDTO{
		Period:          string(report.Period),
		StartDate:       report.StartDate,
		EndDate:         report.EndDate,
		TotalDiffs:      report.TotalDiffs,
		TotalCodingMins: report.TotalCodingMins,
		AvgDiffsPerDay:  report.AvgDiffsPerDay,
		TopLanguages:    languages,
		TopSkills:       skills,
		Bottlenecks:     report.Bottlenecks,
	}, nil
}

// AppStatsDTO 应用统计 DTO
type AppStatsDTO struct {
	AppName       string `json:"app_name"`
	TotalDuration int    `json:"total_duration"`
	EventCount    int64  `json:"event_count"`
}

// GetAppStats 获取应用统计
func (a *App) GetAppStats() ([]AppStatsDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	now := time.Now()
	startTime := now.AddDate(0, 0, -7).UnixMilli()
	endTime := now.UnixMilli()

	if a.core == nil {
		return nil, errors.New("数据库未初始化")
	}

	stats, err := a.core.Repos.Event.GetAppStats(a.ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	result := make([]AppStatsDTO, len(stats))
	for i, s := range stats {
		result[i] = AppStatsDTO{
			AppName:       s.AppName,
			TotalDuration: s.TotalDuration,
			EventCount:    s.EventCount,
		}
	}
	return result, nil
}

// DiffDetailDTO Diff 详情 DTO
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

// GetDiffDetail 获取 Diff 详情
func (a *App) GetDiffDetail(id int64) (*DiffDetailDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Repos.Diff == nil {
		return nil, errors.New("Diff 仓储未初始化")
	}

	diff, err := a.core.Repos.Diff.GetByID(a.ctx, id)
	if err != nil {
		return nil, err
	}
	if diff == nil {
		return nil, errors.New("Diff not found")
	}

	var skills []string
	if len(diff.SkillsDetected) > 0 {
		skills = []string(diff.SkillsDetected)
	}

	return &DiffDetailDTO{
		ID:           diff.ID,
		FileName:     diff.FileName,
		Language:     diff.Language,
		DiffContent:  diff.DiffContent,
		Insight:      diff.AIInsight,
		Skills:       skills,
		LinesAdded:   diff.LinesAdded,
		LinesDeleted: diff.LinesDeleted,
		Timestamp:    diff.Timestamp,
	}, nil
}
