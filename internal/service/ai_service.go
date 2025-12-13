package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yuqie6/mirror/internal/ai"
	"github.com/yuqie6/mirror/internal/schema"
)

// AIService AI 分析服务
type AIService struct {
	analyzer     Analyzer
	diffRepo     DiffRepository
	eventRepo    EventRepository
	summaryRepo  SummaryRepository
	skillService *SkillService
	ragService   RAGQuerier // 可选，用于查询历史记忆/索引
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
	diffs, err := s.diffRepo.GetPendingAIAnalysis(ctx, limit)
	if err != nil {
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

// GenerateDailySummary 生成每日总结
func (s *AIService) GenerateDailySummary(ctx context.Context, date string) (*schema.DailySummary, error) {
	// 尝试获取缓存
	cached, err := s.summaryRepo.GetByDate(ctx, date)
	if err != nil {
		slog.Warn("查询缓存总结失败", "date", date, "error", err)
	}

	today := time.Now().Format("2006-01-02")

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

	// 生成总结
	result, err := s.analyzer.GenerateDailySummary(ctx, req)
	if err != nil {
		// 如果生成失败但有缓存，返回缓存
		if cached != nil {
			slog.Error("生成总结失败，降级返回缓存", "error", err)
			return cached, nil
		}
		return nil, err
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

// GenerateWeeklySummary 生成周报（代理到 analyzer）
func (s *AIService) GenerateWeeklySummary(ctx context.Context, req *ai.WeeklySummaryRequest) (*ai.WeeklySummaryResult, error) {
	return s.analyzer.GenerateWeeklySummary(ctx, req)
}

// GeneratePeriodSummary 生成阶段汇总（周/月）
func (s *AIService) GeneratePeriodSummary(ctx context.Context, periodType, startDate, endDate string, summaries []schema.DailySummary) (*ai.WeeklySummaryResult, error) {
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

	return s.analyzer.GenerateWeeklySummary(ctx, req)
}
