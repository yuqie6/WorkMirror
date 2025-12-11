package main

import (
	"context"
	"time"

	"github.com/yuqie6/mirror/internal/ai"
	"github.com/yuqie6/mirror/internal/pkg/config"
	"github.com/yuqie6/mirror/internal/repository"
	"github.com/yuqie6/mirror/internal/service"
)

// App struct
type App struct {
	ctx          context.Context
	cfg          *config.Config
	db           *repository.Database
	aiService    *service.AIService
	skillService *service.SkillService
	trendService *service.TrendService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		cfg, _ = config.Load("../config/config.yaml")
	}
	a.cfg = cfg

	// 初始化数据库
	db, err := repository.NewDatabase(cfg.Storage.DBPath)
	if err != nil {
		return
	}
	a.db = db

	// 初始化服务
	diffRepo := repository.NewDiffRepository(db.DB)
	eventRepo := repository.NewEventRepository(db.DB)
	summaryRepo := repository.NewSummaryRepository(db.DB)
	skillRepo := repository.NewSkillRepository(db.DB)

	deepseek := ai.NewDeepSeekClient(&ai.DeepSeekConfig{
		APIKey:  cfg.AI.DeepSeek.APIKey,
		BaseURL: cfg.AI.DeepSeek.BaseURL,
		Model:   cfg.AI.DeepSeek.Model,
	})
	analyzer := ai.NewDiffAnalyzer(deepseek)

	a.skillService = service.NewSkillService(skillRepo, diffRepo)
	a.aiService = service.NewAIService(analyzer, diffRepo, eventRepo, summaryRepo, a.skillService)
	a.trendService = service.NewTrendService(skillRepo, diffRepo, eventRepo)
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
	today := time.Now().Format("2006-01-02")
	summary, err := a.aiService.GenerateDailySummary(a.ctx, today)
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
	Level      int    `json:"level"`
	Experience int    `json:"experience"`
	Progress   int    `json:"progress"`
	Status     string `json:"status"`
}

// GetSkillTree 获取技能树
func (a *App) GetSkillTree() ([]SkillNodeDTO, error) {
	skillTree, err := a.skillService.GetSkillTree(a.ctx)
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
				Level:      skill.Level,
				Experience: int(skill.Exp),
				Progress:   int(skill.Progress),
				Status:     skill.Trend,
			})
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
	period := service.TrendPeriod7Days
	if days == 30 {
		period = service.TrendPeriod30Days
	}

	report, err := a.trendService.GetTrendReport(a.ctx, period)
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

	eventRepo := repository.NewEventRepository(a.db.DB)
	stats, err := eventRepo.GetAppStats(a.ctx, startTime, endTime)
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
