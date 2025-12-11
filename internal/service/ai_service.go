package service

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yuqie6/mirror/internal/ai"
	"github.com/yuqie6/mirror/internal/model"
	"github.com/yuqie6/mirror/internal/repository"
)

// AIService AI 分析服务
type AIService struct {
	analyzer     *ai.DiffAnalyzer
	diffRepo     *repository.DiffRepository
	eventRepo    *repository.EventRepository
	summaryRepo  *repository.SummaryRepository
	skillService *SkillService
	ragService   *RAGService // 可选，用于查询历史记忆
}

// NewAIService 创建 AI 服务
func NewAIService(
	analyzer *ai.DiffAnalyzer,
	diffRepo *repository.DiffRepository,
	eventRepo *repository.EventRepository,
	summaryRepo *repository.SummaryRepository,
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
func (s *AIService) SetRAGService(ragService *RAGService) {
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

	// Worker Pool 配置
	const workerCount = 3         // 并发 worker 数
	const rateLimit = time.Second // 每个请求间隔（令牌桶）

	var (
		wg       sync.WaitGroup
		analyzed int32
		mu       sync.Mutex
		limiter  = time.NewTicker(rateLimit / time.Duration(workerCount))
	)
	defer limiter.Stop()

	// 任务通道
	tasks := make(chan model.Diff, len(diffs))
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

				insight, err := s.analyzer.AnalyzeDiff(ctx, diff.FilePath, diff.Language, diff.DiffContent)
				if err != nil {
					slog.Warn("分析 Diff 失败", "worker", workerID, "file", diff.FileName, "error", err)
					continue
				}

				// 提取技能名列表
				skillNames := make([]string, 0, len(insight.Skills))
				for _, skill := range insight.Skills {
					skillNames = append(skillNames, skill.Name)
				}

				// 更新数据库（需要锁保护）
				mu.Lock()
				if err := s.diffRepo.UpdateAIInsight(ctx, diff.ID, insight.Insight, skillNames); err != nil {
					mu.Unlock()
					slog.Warn("更新 Diff 解读失败", "id", diff.ID, "error", err)
					continue
				}

				// 更新技能树
				diff.AIInsight = insight.Insight
				diff.SkillsDetected = model.JSONArray(skillNames)
				if err := s.skillService.UpdateSkillsFromDiffsWithCategory(ctx, []model.Diff{diff}, insight.Skills); err != nil {
					slog.Warn("更新技能失败", "file", diff.FileName, "error", err)
				}
				mu.Unlock()

				atomic.AddInt32(&analyzed, 1)
				slog.Info("Diff 分析完成", "worker", workerID, "file", diff.FileName)
			}
		}(i)
	}

	wg.Wait()
	return int(analyzed), nil
}

// GenerateDailySummary 生成每日总结
func (s *AIService) GenerateDailySummary(ctx context.Context, date string) (*model.DailySummary, error) {
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
	t, _ := time.ParseInLocation("2006-01-02", date, loc)
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

	// 添加窗口事件
	for _, stat := range appStats {
		req.WindowEvents = append(req.WindowEvents, ai.WindowEventInfo{
			AppName:  stat.AppName,
			Duration: int(stat.TotalDuration / 60), // 转换为分钟
		})
	}

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
	summary := &model.DailySummary{
		Date:         date,
		Summary:      result.Summary,
		Highlights:   result.Highlights,
		Struggles:    result.Struggles,
		SkillsGained: model.JSONArray(result.SkillsGained),
		TotalDiffs:   len(diffs),
	}

	// 计算编码时长
	for _, stat := range appStats {
		if isCodeEditorInList(stat.AppName, DefaultCodeEditors) {
			summary.TotalCoding += int(stat.TotalDuration / 60)
		}
	}

	if err := s.summaryRepo.Upsert(ctx, summary); err != nil {
		return nil, err
	}

	return summary, nil
}

// isCodeEditorInList 判断是否是代码编辑器（使用提供的列表）
func isCodeEditorInList(appName string, editors []string) bool {
	for _, editor := range editors {
		if appName == editor {
			return true
		}
	}
	return false
}

// DefaultCodeEditors 默认代码编辑器列表（配置未提供时使用）
var DefaultCodeEditors = []string{
	"Code.exe", "code.exe", "Cursor.exe", "cursor.exe",
	"Antigravity.exe", "antigravity.exe",
	"idea64.exe", "idea.exe", "goland64.exe", "pycharm64.exe", "webstorm64.exe",
	"devenv.exe", "Zed.exe", "Fleet.exe", "sublime_text.exe", "notepad++.exe",
	"vim.exe", "nvim.exe", "emacs.exe",
}

// GenerateWeeklySummary 生成周报（代理到 analyzer）
func (s *AIService) GenerateWeeklySummary(ctx context.Context, req *ai.WeeklySummaryRequest) (*ai.WeeklySummaryResult, error) {
	return s.analyzer.GenerateWeeklySummary(ctx, req)
}
