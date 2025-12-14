package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yuqie6/WorkMirror/internal/ai"
	"github.com/yuqie6/WorkMirror/internal/repository"
	"github.com/yuqie6/WorkMirror/internal/schema"
)

// AIService AI 分析服务
type AIService struct {
	analyzer     Analyzer
	diffRepo     DiffRepository
	eventRepo    EventRepository
	summaryRepo  SummaryRepository
	skillService *SkillService
	ragService   RAGQuerier // 可选，用于查询历史记忆/索引

	lastCallAt     atomic.Int64
	lastErrorAt    atomic.Int64
	lastErrorMsg   atomic.Value // string
	degraded       atomic.Bool
	degradedReason atomic.Value // string
}

type DailySummaryOptions struct {
	Force bool
}

// NewAIService 创建 AI 服务
func NewAIService(
	analyzer Analyzer,
	diffRepo DiffRepository,
	eventRepo EventRepository,
	summaryRepo SummaryRepository,
	skillService *SkillService,
) *AIService {
	return &AIService{
		analyzer:     analyzer,
		diffRepo:     diffRepo,
		eventRepo:    eventRepo,
		summaryRepo:  summaryRepo,
		skillService: skillService,
	}
}

// SetRAGService 设置 RAG 服务（可选）
func (s *AIService) SetRAGService(ragService RAGQuerier) {
	s.ragService = ragService
}

// AnalyzePendingDiffs 分析待处理的 Diff（使用 Worker Pool）
func (s *AIService) AnalyzePendingDiffs(ctx context.Context, limit int) (int, error) {
	s.lastCallAt.Store(time.Now().UnixMilli())
	if s.analyzer == nil {
		s.degraded.Store(true)
		s.degradedReason.Store("not_configured")
		return 0, nil
	}
	diffs, err := s.diffRepo.GetPendingAIAnalysis(ctx, limit)
	if err != nil {
		s.noteError(err, "get_pending_diffs_failed")
		return 0, err
	}

	if len(diffs) == 0 {
		slog.Debug("没有待分析的 Diff")
		return 0, nil
	}

	// 获取当前技能树（传给 AI 作为上下文）
	existingSkills := s.getSkillInfoList(ctx)

	// Worker Pool 配置
	const workerCount = 3         // 并发 worker 数
	const rateLimit = time.Second // 每个请求间隔（令牌桶）

	var (
		wg       sync.WaitGroup
		analyzed int32
		limiter  = time.NewTicker(rateLimit / time.Duration(workerCount))
	)
	defer limiter.Stop()

	// 任务通道
	tasks := make(chan schema.Diff, len(diffs))
	for _, diff := range diffs {
		tasks <- diff
	}
	close(tasks)

	// 启动 workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for diff := range tasks {
				select {
				case <-ctx.Done():
					return
				case <-limiter.C:
					// 令牌桶限流
				}

				insight, err := s.analyzer.AnalyzeDiff(ctx, diff.FilePath, diff.Language, diff.DiffContent, existingSkills)
				if err != nil {
					slog.Warn("分析 Diff 失败", "worker", workerID, "file", diff.FileName, "error", err)
					continue
				}

				// 提取技能名列表
				skillNames := make([]string, 0, len(insight.Skills))
				for _, skill := range insight.Skills {
					skillNames = append(skillNames, skill.Name)
				}

				// 更新数据库（repository 层已有事务保证一致性，无需外层锁）
				if err := s.diffRepo.UpdateAIInsight(ctx, diff.ID, insight.Insight, skillNames); err != nil {
					slog.Warn("更新 Diff 解读失败", "id", diff.ID, "error", err)
					continue
				}

				// 更新技能树
				diff.AIInsight = insight.Insight
				diff.SkillsDetected = schema.JSONArray(skillNames)
				if err := s.skillService.UpdateSkillsFromDiffsWithCategory(ctx, []schema.Diff{diff}, insight.Skills); err != nil {
					slog.Warn("更新技能失败", "file", diff.FileName, "error", err)
				}

				// 索引到 RAG（如果已配置）
				if s.ragService != nil {
					if err := s.ragService.IndexDiff(ctx, &diff); err != nil {
						slog.Warn("索引 Diff 失败", "file", diff.FileName, "error", err)
					}
				}

				atomic.AddInt32(&analyzed, 1)
				slog.Info("Diff 分析完成", "worker", workerID, "file", diff.FileName)
			}
		}(i)
	}

	wg.Wait()
	return int(analyzed), nil
}

// getSkillInfoList 获取简化的技能列表（传给 AI）
func (s *AIService) getSkillInfoList(ctx context.Context) []ai.SkillInfo {
	if s.skillService == nil {
		return nil
	}
	skills, err := s.skillService.GetAllSkills(ctx)
	if err != nil {
		slog.Warn("获取技能列表失败", "error", err)
		return nil
	}

	// ParentKey 是内部 Key；传给 AI 的 parent 需要是“父技能名称”以保持语义一致。
	keyToName := make(map[string]string, len(skills))
	for _, skill := range skills {
		if strings.TrimSpace(skill.Key) == "" {
			continue
		}
		keyToName[skill.Key] = skill.Name
	}

	result := make([]ai.SkillInfo, 0, len(skills))
	for _, skill := range skills {
		parentName := ""
		if strings.TrimSpace(skill.ParentKey) != "" {
			parentName = keyToName[skill.ParentKey]
		}
		result = append(result, ai.SkillInfo{
			Name:     skill.Name,
			Category: skill.Category,
			Parent:   parentName,
		})
	}
	return result
}

func (s *AIService) GenerateDailySummaryWithOptions(ctx context.Context, date string, opts DailySummaryOptions) (*schema.DailySummary, error) {
	s.lastCallAt.Store(time.Now().UnixMilli())
	// 尝试获取缓存
	cached, err := s.summaryRepo.GetByDate(ctx, date)
	if err != nil {
		slog.Warn("查询缓存总结失败", "date", date, "error", err)
	}

	today := time.Now().Format("2006-01-02")

	if !opts.Force {
		// 如果是过去日期的总结，直接返回缓存
		if date != today && cached != nil {
			slog.Info("返回历史总结缓存", "date", date)
			return cached, nil
		}

		// 如果是今日总结，且最近 5 分钟内生成过，直接返回
		if date == today && cached != nil && time.Since(cached.UpdatedAt) < 5*time.Minute {
			slog.Info("返回最近生成的今日总结", "date", date)
			return cached, nil
		}
	}

	// 获取当日 Diff
	diffs, err := s.diffRepo.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	// 获取当日事件统计
	loc := time.Local
	t, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return nil, fmt.Errorf("无效日期格式: %w", err)
	}
	startTime := t.UnixMilli()
	endTime := t.Add(24*time.Hour).UnixMilli() - 1

	appStats, err := s.eventRepo.GetAppStats(ctx, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 尝试从 RAG 获取相关历史记忆
	var historyMemories []string
	if s.ragService != nil {
		// 构建查询文本
		var queryParts []string
		for _, diff := range diffs {
			if diff.AIInsight != "" {
				queryParts = append(queryParts, diff.AIInsight)
			} else if diff.Language != "" {
				queryParts = append(queryParts, diff.Language)
			}
		}
		if len(queryParts) > 0 {
			query := "编程 " + queryParts[0]
			if results, err := s.ragService.Query(ctx, query, 3); err == nil {
				for _, r := range results {
					historyMemories = append(historyMemories, r.Content)
				}
				slog.Debug("获取历史记忆", "count", len(historyMemories))
			}
		}
	}

	// 构建请求
	req := &ai.DailySummaryRequest{
		Date:            date,
		HistoryMemories: historyMemories,
	}

	// 添加窗口事件（分钟）
	req.WindowEvents = WindowEventInfosFromAppStats(appStats, 0)

	// 添加 Diff 信息
	for _, diff := range diffs {
		req.Diffs = append(req.Diffs, ai.DiffInfo{
			FileName:     diff.FileName,
			Language:     diff.Language,
			Insight:      diff.AIInsight,
			DiffContent:  diff.DiffContent,
			LinesChanged: diff.LinesAdded + diff.LinesDeleted,
		})
	}

	// 离线模式：不触发任何 AI 调用，直接走规则总结（保证“无 Key 也可用”且不产生噪音错误日志）。
	if s.analyzer == nil {
		s.degraded.Store(true)
		s.degradedReason.Store("not_configured")
		summary := buildRuleBasedDailySummary(date, diffs, appStats)
		if upsertErr := s.summaryRepo.Upsert(ctx, summary); upsertErr != nil {
			s.noteError(upsertErr, "upsert_daily_summary_failed")
			return nil, upsertErr
		}
		return summary, nil
	}

	// 生成总结
	result, err := s.analyzer.GenerateDailySummary(ctx, req)
	if err != nil {
		s.noteError(err, "generate_daily_summary_failed")
		// 如果生成失败但有缓存，返回缓存
		if cached != nil {
			slog.Error("生成总结失败，降级返回缓存", "error", err)
			return cached, nil
		}
		// 离线/降级：生成一个纯规则总结，保证产品可用性（Local-first）
		slog.Warn("AI 总结不可用，使用规则总结降级", "date", date, "error", err)
		summary := buildRuleBasedDailySummary(date, diffs, appStats)
		if upsertErr := s.summaryRepo.Upsert(ctx, summary); upsertErr != nil {
			return nil, upsertErr
		}
		return summary, nil
	}

	// 保存到数据库
	summary := &schema.DailySummary{
		Date:         date,
		Summary:      result.Summary,
		Highlights:   result.Highlights,
		Struggles:    result.Struggles,
		SkillsGained: schema.JSONArray(result.SkillsGained),
		TotalDiffs:   len(diffs),
	}

	// 计算编码时长
	for _, stat := range appStats {
		if IsCodeEditor(stat.AppName) {
			summary.TotalCoding += int(stat.TotalDuration / 60)
		}
	}

	if err := s.summaryRepo.Upsert(ctx, summary); err != nil {
		s.noteError(err, "upsert_daily_summary_failed")
		return nil, err
	}

	// 索引到 RAG（如果已配置）
	if s.ragService != nil {
		if err := s.ragService.IndexDailySummary(ctx, summary); err != nil {
			slog.Warn("索引每日总结失败", "date", date, "error", err)
		}
	}

	return summary, nil
}

// GenerateDailySummary 生成每日总结
func (s *AIService) GenerateDailySummary(ctx context.Context, date string) (*schema.DailySummary, error) {
	return s.GenerateDailySummaryWithOptions(ctx, date, DailySummaryOptions{})
}

// GenerateWeeklySummary 生成周报（代理到 analyzer）
func (s *AIService) GenerateWeeklySummary(ctx context.Context, req *ai.WeeklySummaryRequest) (*ai.WeeklySummaryResult, error) {
	if s.analyzer == nil {
		s.degraded.Store(true)
		s.degradedReason.Store("not_configured")
		return nil, fmt.Errorf("AI 未配置")
	}
	return s.analyzer.GenerateWeeklySummary(ctx, req)
}

// GeneratePeriodSummary 生成阶段汇总（周/月）
func (s *AIService) GeneratePeriodSummary(ctx context.Context, periodType, startDate, endDate string, summaries []schema.DailySummary) (*ai.WeeklySummaryResult, error) {
	s.lastCallAt.Store(time.Now().UnixMilli())
	req := &ai.WeeklySummaryRequest{
		PeriodType: periodType,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	for _, sum := range summaries {
		req.DailySummaries = append(req.DailySummaries, ai.DailySummaryInfo{
			Date:       sum.Date,
			Summary:    sum.Summary,
			Highlights: sum.Highlights,
			Skills:     sum.SkillsGained,
		})
		req.TotalCoding += sum.TotalCoding
		req.TotalDiffs += sum.TotalDiffs
	}

	if s.analyzer == nil {
		s.degraded.Store(true)
		s.degradedReason.Store("not_configured")
		return buildRuleBasedPeriodSummary(req), nil
	}

	res, err := s.analyzer.GenerateWeeklySummary(ctx, req)
	if err == nil && res != nil {
		s.degraded.Store(false)
		return res, nil
	}
	if err != nil {
		s.noteError(err, "generate_period_summary_failed")
	}
	slog.Warn("AI 阶段汇总不可用，使用规则汇总降级", "type", periodType, "start", startDate, "end", endDate, "error", err)
	return buildRuleBasedPeriodSummary(req), nil
}

type AIServiceStats struct {
	LastCallAt     int64  `json:"last_call_at"`
	LastErrorAt    int64  `json:"last_error_at"`
	LastError      string `json:"last_error"`
	Degraded       bool   `json:"degraded"`
	DegradedReason string `json:"degraded_reason"`
}

func (s *AIService) Stats() AIServiceStats {
	if s == nil {
		return AIServiceStats{}
	}
	rawErr := s.lastErrorMsg.Load()
	lastErr, _ := rawErr.(string)
	rawReason := s.degradedReason.Load()
	reason, _ := rawReason.(string)
	return AIServiceStats{
		LastCallAt:     s.lastCallAt.Load(),
		LastErrorAt:    s.lastErrorAt.Load(),
		LastError:      lastErr,
		Degraded:       s.degraded.Load(),
		DegradedReason: reason,
	}
}

func (s *AIService) noteError(err error, reason string) {
	if s == nil || err == nil {
		return
	}
	s.lastErrorAt.Store(time.Now().UnixMilli())
	s.lastErrorMsg.Store(err.Error())
	if strings.TrimSpace(reason) != "" {
		s.degradedReason.Store(reason)
	}
	s.degraded.Store(true)
}

func buildRuleBasedDailySummary(date string, diffs []schema.Diff, appStats []repository.AppStat) *schema.DailySummary {
	// 统计编码时长（分钟）
	totalCoding := 0
	for _, stat := range appStats {
		if IsCodeEditor(stat.AppName) {
			totalCoding += SecondsToMinutesFloor(stat.TotalDuration)
		}
	}

	// 语言分布
	langCount := make(map[string]int)
	for _, d := range diffs {
		lang := strings.TrimSpace(d.Language)
		if lang == "" {
			continue
		}
		langCount[lang]++
	}
	topLangs := topKeysByCount(langCount, 2)

	// 技能（来自 diffs.skills_detected，证据链优先）
	skillCount := make(map[string]int)
	for _, d := range diffs {
		for _, sk := range d.SkillsDetected {
			name := strings.TrimSpace(sk)
			if name == "" {
				continue
			}
			skillCount[name]++
		}
	}
	topSkills := topKeysByCount(skillCount, 8)

	// 亮点：选择变更量最高的 diffs
	type diffRank struct {
		file    string
		lang    string
		insight string
		changed int
	}
	ranks := make([]diffRank, 0, len(diffs))
	for _, d := range diffs {
		ranks = append(ranks, diffRank{
			file:    strings.TrimSpace(d.FileName),
			lang:    strings.TrimSpace(d.Language),
			insight: strings.TrimSpace(d.AIInsight),
			changed: d.LinesAdded + d.LinesDeleted,
		})
	}
	sort.Slice(ranks, func(i, j int) bool {
		if ranks[i].changed != ranks[j].changed {
			return ranks[i].changed > ranks[j].changed
		}
		return ranks[i].file < ranks[j].file
	})
	highlights := make([]string, 0, 2)
	for i := 0; i < len(ranks) && len(highlights) < 2; i++ {
		if ranks[i].file == "" {
			continue
		}
		if ranks[i].insight != "" {
			highlights = append(highlights, ranks[i].file+"："+ranks[i].insight)
			continue
		}
		if ranks[i].lang != "" {
			highlights = append(highlights, ranks[i].file+"（"+ranks[i].lang+"）有较多变更")
		} else {
			highlights = append(highlights, ranks[i].file+" 有较多变更")
		}
	}

	// 总结文案（保持可解释/可追溯，不做凭空推断）
	parts := make([]string, 0, 6)
	if totalCoding > 0 {
		parts = append(parts, fmt.Sprintf("编码约 %d 分钟", totalCoding))
	}
	if len(diffs) > 0 {
		parts = append(parts, fmt.Sprintf("记录到 %d 次代码变更", len(diffs)))
	}
	if len(topLangs) > 0 {
		parts = append(parts, "主要语言："+strings.Join(topLangs, "、"))
	}
	if len(topSkills) > 0 {
		parts = append(parts, "涉及技能："+strings.Join(topSkills[:minInt(3, len(topSkills))], "、"))
	}
	summaryText := "今日暂无足够证据生成总结。"
	if len(parts) > 0 {
		summaryText = strings.Join(parts, "，") + "。"
	}

	return &schema.DailySummary{
		Date:         date,
		Summary:      summaryText,
		Highlights:   strings.Join(highlights, "\n"),
		Struggles:    "",
		SkillsGained: schema.JSONArray(topSkills),
		TotalCoding:  totalCoding,
		TotalDiffs:   len(diffs),
	}
}

func buildRuleBasedPeriodSummary(req *ai.WeeklySummaryRequest) *ai.WeeklySummaryResult {
	if req == nil {
		return &ai.WeeklySummaryResult{
			Overview:     "暂无足够数据生成阶段汇总。",
			Achievements: []string{},
			Patterns:     "",
			Suggestions:  "",
			TopSkills:    []string{},
		}
	}

	skillCount := make(map[string]int)
	activeDays := 0
	for _, d := range req.DailySummaries {
		if strings.TrimSpace(d.Summary) != "" {
			activeDays++
		}
		for _, sk := range d.Skills {
			name := strings.TrimSpace(sk)
			if name == "" {
				continue
			}
			skillCount[name]++
		}
	}
	topSkills := topKeysByCount(skillCount, 10)

	label := "本周"
	scope := "一周"
	if strings.ToLower(strings.TrimSpace(req.PeriodType)) == "month" {
		label = "本月"
		scope = "一个月"
	}

	overviewParts := []string{
		fmt.Sprintf("%s（%s ~ %s）累计编码 %d 分钟，代码变更 %d 次。", label, req.StartDate, req.EndDate, req.TotalCoding, req.TotalDiffs),
	}
	if activeDays > 0 {
		overviewParts = append(overviewParts, fmt.Sprintf("记录覆盖 %d 天（%s口径）。", activeDays, scope))
	}
	if len(topSkills) > 0 {
		overviewParts = append(overviewParts, "重点技能："+strings.Join(topSkills[:minInt(5, len(topSkills))], "、")+"。")
	}

	achievements := make([]string, 0, 6)
	if req.TotalDiffs > 0 {
		achievements = append(achievements, fmt.Sprintf("完成 %d 次代码变更，并形成可追溯的证据链。", req.TotalDiffs))
	}
	if req.TotalCoding > 0 {
		achievements = append(achievements, fmt.Sprintf("累计编码 %d 分钟，保持持续投入。", req.TotalCoding))
	}
	if len(topSkills) > 0 {
		achievements = append(achievements, "主要聚焦："+strings.Join(topSkills[:minInt(3, len(topSkills))], "、")+"。")
	}

	patterns := "当前为离线/降级汇总：建议先完善 Diff 采集与 AI 分析（或保持日报生成），以提升技能归因与趋势洞察的可信度。"
	suggestions := "建议：\n- 保持每日生成一次日报（即使离线也可），确保记录连续性\n- 对重要项目目录启用 Diff 监控，尽量覆盖主要工作流\n- 定期重建/补全会话语义（用于 Skill → Session 的证据追溯）"

	return &ai.WeeklySummaryResult{
		Overview:     strings.Join(overviewParts, ""),
		Achievements: achievements,
		Patterns:     patterns,
		Suggestions:  suggestions,
		TopSkills:    topSkills,
	}
}
