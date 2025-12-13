package service

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/yuqie6/mirror/internal/schema"
	"github.com/yuqie6/mirror/internal/repository"
)

// TrendService 趋势分析服务
type TrendService struct {
	skillRepo    SkillRepository
	activityRepo SkillActivityRepository
	diffRepo     DiffRepository
	eventRepo    EventRepository
}

// NewTrendService 创建趋势服务
func NewTrendService(
	skillRepo SkillRepository,
	activityRepo SkillActivityRepository,
	diffRepo DiffRepository,
	eventRepo EventRepository,
) *TrendService {
	return &TrendService{
		skillRepo:    skillRepo,
		activityRepo: activityRepo,
		diffRepo:     diffRepo,
		eventRepo:    eventRepo,
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
	SkillKey    string  `json:"skill_key"`
	SkillName   string  `json:"skill_name"`
	Category    string  `json:"category"`
	Changes     int     `json:"changes"`       // 活动次数（skill_activities events）
	ExpGain     float64 `json:"exp_gain"`      // 当前期经验增量
	PrevExpGain float64 `json:"prev_exp_gain"` // 上一期经验增量
	GrowthRate  float64 `json:"growth_rate"`   // 增长率（相对上期经验增量）
	Status      string  `json:"status"`        // growing, stable, declining
	DaysActive  int     `json:"days_active"`   // 活跃天数（按 skill_activities 统计）
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
	prevEndTime := startTime - 1
	prevStartTime := now.AddDate(0, 0, -2*days).UnixMilli()

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

	// 技能趋势：基于 skill_activities 的经验增量（更贴近“能力变化”且可追溯证据）
	currentStats := make(map[string]repository.SkillActivityStat)
	prevStats := make(map[string]repository.SkillActivityStat)
	if s.activityRepo != nil {
		cur, err := s.activityRepo.GetStatsByTimeRange(ctx, startTime, endTime)
		if err != nil {
			return nil, err
		}
		for _, st := range cur {
			currentStats[st.SkillKey] = st
		}

		prev, err := s.activityRepo.GetStatsByTimeRange(ctx, prevStartTime, prevEndTime)
		if err != nil {
			return nil, err
		}
		for _, st := range prev {
			prevStats[st.SkillKey] = st
		}
	}
	// 兼容历史数据：如果 skill_activities 为空但已经存在已分析的 Diff，则回填近 2 个周期的活动记录
	if s.activityRepo != nil && len(currentStats) == 0 && len(prevStats) == 0 {
		_, _ = BackfillSkillActivitiesFromDiffs(ctx, s.diffRepo, s.activityRepo, DefaultExpPolicy{}, prevStartTime, endTime, 500)

		cur, err := s.activityRepo.GetStatsByTimeRange(ctx, startTime, endTime)
		if err != nil {
			return nil, err
		}
		for _, st := range cur {
			currentStats[st.SkillKey] = st
		}
		prev, err := s.activityRepo.GetStatsByTimeRange(ctx, prevStartTime, prevEndTime)
		if err != nil {
			return nil, err
		}
		for _, st := range prev {
			prevStats[st.SkillKey] = st
		}
	}

	allSkills, _ := s.skillRepo.GetAll(ctx)
	skillByKey := make(map[string]schema.SkillNode, len(allSkills))
	for _, sk := range allSkills {
		skillByKey[sk.Key] = sk
	}

	type keyRank struct {
		key    string
		expSum float64
	}
	ranks := make([]keyRank, 0, len(currentStats))
	for k, st := range currentStats {
		ranks = append(ranks, keyRank{key: k, expSum: st.ExpSum})
	}
	sort.Slice(ranks, func(i, j int) bool {
		if ranks[i].expSum != ranks[j].expSum {
			return ranks[i].expSum > ranks[j].expSum
		}
		return ranks[i].key < ranks[j].key
	})

	topKeys := make([]string, 0, 10)
	for i := 0; i < len(ranks) && len(topKeys) < 10; i++ {
		topKeys = append(topKeys, ranks[i].key)
	}
	if len(topKeys) < 10 {
		declines := make([]keyRank, 0)
		for k, st := range prevStats {
			if _, ok := currentStats[k]; ok {
				continue
			}
			declines = append(declines, keyRank{key: k, expSum: st.ExpSum})
		}
		sort.Slice(declines, func(i, j int) bool {
			if declines[i].expSum != declines[j].expSum {
				return declines[i].expSum > declines[j].expSum
			}
			return declines[i].key < declines[j].key
		})
		for i := 0; i < len(declines) && len(topKeys) < 10; i++ {
			topKeys = append(topKeys, declines[i].key)
		}
	}

	// 构建技能趋势
	topSkills := make([]SkillTrend, 0, len(topKeys))
	for _, key := range topKeys {
		cur := currentStats[key]
		prev := prevStats[key]
		curExp := cur.ExpSum
		prevExp := prev.ExpSum

		name := key
		category := "other"
		daysInactive := 0
		if sk, ok := skillByKey[key]; ok {
			if sk.Name != "" {
				name = sk.Name
			}
			if sk.Category != "" {
				category = sk.Category
			}
			daysInactive = sk.DaysInactive()
		}

		growthRate := calcGrowthRateFloat(curExp, prevExp)
		status := classifyTrendStatusByExp(curExp, growthRate, daysInactive)

		daysActive := cur.DaysActive
		if daysActive <= 0 {
			daysActive = prev.DaysActive
		}

		topSkills = append(topSkills, SkillTrend{
			SkillKey:    key,
			SkillName:   name,
			Category:    category,
			Changes:     int(cur.EventCount),
			ExpGain:     curExp,
			PrevExpGain: prevExp,
			GrowthRate:  growthRate,
			Status:      status,
			DaysActive:  daysActive,
		})
	}

	sort.Slice(topSkills, func(i, j int) bool {
		if topSkills[i].ExpGain != topSkills[j].ExpGain {
			return topSkills[i].ExpGain > topSkills[j].ExpGain
		}
		return topSkills[i].SkillName < topSkills[j].SkillName
	})

	// 获取应用统计计算编码时长
	appStats, err := s.eventRepo.GetAppStats(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	totalCodingMins := SumCodingMinutesFromAppStats(appStats)

	// 检测瓶颈
	bottlenecks := s.detectBottlenecks(topSkills, totalCodingMins)

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
func (s *TrendService) detectBottlenecks(skills []SkillTrend, totalCodingMins int64) []string {
	bottlenecks := []string{}

	if len(skills) == 0 {
		if totalCodingMins > 0 {
			return []string{"检测到编码时长，但没有技能经验记录；请先运行一次“AI 分析 Diff”以建立技能证据链"}
		}
		return []string{"没有检测到技能活动，开始编码吧！"}
	}

	// 高投入低增长：经验增量显著高于均值，但相对上期增长率不高（或为负）
	avgExp := float64(0)
	for _, sk := range skills {
		avgExp += sk.ExpGain
	}
	avgExp = avgExp / float64(len(skills))

	for _, skill := range skills {
		if skill.ExpGain > 0 && avgExp > 0 && skill.ExpGain >= avgExp*1.5 && skill.GrowthRate <= 0 {
			bottlenecks = append(bottlenecks, skill.SkillName+" 仍在高投入，但相对上期增长放缓，建议复盘方法或补齐知识点")
			continue
		}
		if skill.Status == "declining" {
			bottlenecks = append(bottlenecks, skill.SkillName+" 技能活动下降，考虑安排一次刻意练习/复习")
		}
	}

	// 全局瓶颈：编码时长显著，但经验增量很低
	if totalCodingMins >= 120 && len(bottlenecks) == 0 {
		if skills[0].ExpGain <= 0 {
			bottlenecks = append(bottlenecks, "编码时长较高，但技能增长证据不足；建议检查 Diff 采集/AI 分析是否正常")
		}
	}

	return bottlenecks
}

func calcGrowthRateFloat(current, prev float64) float64 {
	if current <= 0 && prev <= 0 {
		return 0
	}
	denom := prev
	if denom < 1 {
		denom = 1
	}
	return (current - prev) / denom
}

func classifyTrendStatusByExp(expGain float64, growthRate float64, daysInactive int) string {
	if daysInactive > 7 {
		return "declining"
	}
	if expGain <= 0 {
		return "stable"
	}
	// 使用轻微阈值避免噪声（相对变化 >= 20%）
	if growthRate >= 0.2 {
		return "growing"
	}
	if growthRate <= -0.2 {
		return "declining"
	}
	// 没有明显上/下，保持 stable
	if math.Abs(growthRate) < 0.2 {
		return "stable"
	}
	return "stable"
}
