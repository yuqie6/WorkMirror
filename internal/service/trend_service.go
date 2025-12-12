package service

import (
	"context"
	"time"

	"github.com/yuqie6/mirror/internal/repository"
)

// TrendService 趋势分析服务
type TrendService struct {
	skillRepo *repository.SkillRepository
	diffRepo  *repository.DiffRepository
	eventRepo *repository.EventRepository
}

// NewTrendService 创建趋势服务
func NewTrendService(
	skillRepo *repository.SkillRepository,
	diffRepo *repository.DiffRepository,
	eventRepo *repository.EventRepository,
) *TrendService {
	return &TrendService{
		skillRepo: skillRepo,
		diffRepo:  diffRepo,
		eventRepo: eventRepo,
	}
}

// TrendPeriod 趋势周期
type TrendPeriod string

const (
	TrendPeriod7Days  TrendPeriod = "7d"
	TrendPeriod30Days TrendPeriod = "30d"
)

// SkillTrend 技能趋势
type SkillTrend struct {
	SkillKey   string  `json:"skill_key"`
	SkillName  string  `json:"skill_name"`
	Category   string  `json:"category"`
	Changes    int     `json:"changes"`     // 变更次数
	GrowthRate float64 `json:"growth_rate"` // 增长率 (相比上期)
	Status     string  `json:"status"`      // growing, stable, declining
	DaysActive int     `json:"days_active"` // 活跃天数
}

// LanguageTrend 语言趋势
type LanguageTrend struct {
	Language     string  `json:"language"`
	DiffCount    int64   `json:"diff_count"`
	LinesAdded   int64   `json:"lines_added"`
	LinesDeleted int64   `json:"lines_deleted"`
	Percentage   float64 `json:"percentage"`
}

// TrendReport 趋势报告
type TrendReport struct {
	Period          TrendPeriod     `json:"period"`
	StartDate       string          `json:"start_date"`
	EndDate         string          `json:"end_date"`
	TopSkills       []SkillTrend    `json:"top_skills"`
	TopLanguages    []LanguageTrend `json:"top_languages"`
	TotalDiffs      int64           `json:"total_diffs"`
	TotalCodingMins int64           `json:"total_coding_mins"`
	AvgDiffsPerDay  float64         `json:"avg_diffs_per_day"`
	Bottlenecks     []string        `json:"bottlenecks"`
}

// GetTrendReport 获取趋势报告
func (s *TrendService) GetTrendReport(ctx context.Context, period TrendPeriod) (*TrendReport, error) {
	days := 7
	if period == TrendPeriod30Days {
		days = 30
	}

	now := time.Now()
	endTime := now.UnixMilli()
	startTime := now.AddDate(0, 0, -days).UnixMilli()

	// 获取语言统计
	langStats, err := s.diffRepo.GetLanguageStats(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 计算总量
	var totalDiffs int64
	for _, stat := range langStats {
		totalDiffs += stat.DiffCount
	}

	// 构建语言趋势
	topLanguages := make([]LanguageTrend, 0, len(langStats))
	for _, stat := range langStats {
		percentage := float64(0)
		if totalDiffs > 0 {
			percentage = float64(stat.DiffCount) / float64(totalDiffs) * 100
		}
		topLanguages = append(topLanguages, LanguageTrend{
			Language:     stat.Language,
			DiffCount:    stat.DiffCount,
			LinesAdded:   stat.LinesAdded,
			LinesDeleted: stat.LinesDeleted,
			Percentage:   percentage,
		})
	}

	// 获取指定时间窗内活跃的技能（真正的趋势TopSkills）
	skills, err := s.skillRepo.GetActiveSkillsInPeriod(ctx, startTime, endTime, 10)
	if err != nil {
		return nil, err
	}

	// 构建技能趋势
	topSkills := make([]SkillTrend, 0, len(skills))
	for _, skill := range skills {
		status := "stable"
		daysInactive := skill.DaysInactive()
		if daysInactive == 0 {
			status = "growing"
		} else if daysInactive > 7 {
			status = "declining"
		}

		topSkills = append(topSkills, SkillTrend{
			SkillKey:   skill.Key,
			SkillName:  skill.Name,
			Category:   skill.Category,
			Status:     status,
			DaysActive: days - daysInactive,
		})
	}

	// 获取应用统计计算编码时长
	appStats, err := s.eventRepo.GetAppStats(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var totalCodingMins int64
	for _, stat := range appStats {
		if IsCodeEditor(stat.AppName) {
			totalCodingMins += int64(stat.TotalDuration / 60)
		}
	}

	// 检测瓶颈
	bottlenecks := s.detectBottlenecks(topSkills)

	return &TrendReport{
		Period:          period,
		StartDate:       now.AddDate(0, 0, -days).Format("2006-01-02"),
		EndDate:         now.Format("2006-01-02"),
		TopSkills:       topSkills,
		TopLanguages:    topLanguages,
		TotalDiffs:      totalDiffs,
		TotalCodingMins: totalCodingMins,
		AvgDiffsPerDay:  float64(totalDiffs) / float64(days),
		Bottlenecks:     bottlenecks,
	}, nil
}

// detectBottlenecks 检测技能瓶颈
func (s *TrendService) detectBottlenecks(skills []SkillTrend) []string {
	bottlenecks := []string{}

	for _, skill := range skills {
		if skill.Status == "declining" {
			bottlenecks = append(bottlenecks, skill.SkillName+" 技能正在衰退，考虑复习")
		}
	}

	if len(skills) == 0 {
		bottlenecks = append(bottlenecks, "没有检测到技能活动，开始编码吧！")
	}

	return bottlenecks
}
