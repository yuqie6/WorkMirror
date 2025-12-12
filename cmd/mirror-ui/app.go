package main

import (
	"context"
	"errors"
	"time"

	"github.com/yuqie6/mirror/internal/bootstrap"
	"github.com/yuqie6/mirror/internal/pkg/config"
	"github.com/yuqie6/mirror/internal/service"
)

// App struct
type App struct {
	ctx  context.Context
	cfg  *config.Config
	core *bootstrap.Core
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	core, err := bootstrap.NewCore("")
	if err != nil {
		// UI 启动时不 panic，改为延迟报错
		a.core = nil
		a.cfg = &config.Config{}
		return
	}
	a.core = core
	a.cfg = core.Cfg
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
