package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/danielsclee/mirror/internal/ai"
	"github.com/danielsclee/mirror/internal/model"
	"github.com/danielsclee/mirror/internal/repository"
)

// AIService AI 分析服务
type AIService struct {
	analyzer     *ai.DiffAnalyzer
	diffRepo     *repository.DiffRepository
	eventRepo    *repository.EventRepository
	summaryRepo  *repository.SummaryRepository
	skillService *SkillService
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

// AnalyzePendingDiffs 分析待处理的 Diff
func (s *AIService) AnalyzePendingDiffs(ctx context.Context, limit int) (int, error) {
	diffs, err := s.diffRepo.GetPendingAIAnalysis(ctx, limit)
	if err != nil {
		return 0, err
	}

	if len(diffs) == 0 {
		slog.Debug("没有待分析的 Diff")
		return 0, nil
	}

	analyzed := 0
	for _, diff := range diffs {
		insight, err := s.analyzer.AnalyzeDiff(ctx, diff.FilePath, diff.Language, diff.DiffContent)
		if err != nil {
			slog.Warn("分析 Diff 失败", "file", diff.FileName, "error", err)
			continue
		}

		// 更新数据库
		if err := s.diffRepo.UpdateAIInsight(ctx, diff.ID, insight.Insight, insight.Skills); err != nil {
			slog.Warn("更新 Diff 解读失败", "id", diff.ID, "error", err)
			continue
		}

		// 更新技能树
		diff.AIInsight = insight.Insight
		diff.SkillsDetected = model.JSONArray(insight.Skills)
		if err := s.skillService.UpdateSkillsFromDiffs(ctx, []model.Diff{diff}); err != nil {
			slog.Warn("更新技能失败", "file", diff.FileName, "error", err)
		}

		slog.Info("Diff 分析完成",
			"file", diff.FileName,
			"insight", insight.Insight,
			"skills", insight.Skills,
		)
		analyzed++

		// 避免 API 限流
		time.Sleep(500 * time.Millisecond)
	}

	return analyzed, nil
}

// GenerateDailySummary 生成每日总结
func (s *AIService) GenerateDailySummary(ctx context.Context, date string) (*model.DailySummary, error) {
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

	// 构建请求
	req := &ai.DailySummaryRequest{
		Date: date,
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
			LinesChanged: diff.LinesAdded + diff.LinesDeleted,
		})
	}

	// 生成总结
	result, err := s.analyzer.GenerateDailySummary(ctx, req)
	if err != nil {
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
		if isCodeEditor(stat.AppName) {
			summary.TotalCoding += int(stat.TotalDuration / 60)
		}
	}

	if err := s.summaryRepo.Upsert(ctx, summary); err != nil {
		return nil, err
	}

	return summary, nil
}

// isCodeEditor 判断是否是代码编辑器
func isCodeEditor(appName string) bool {
	editors := []string{
		"Code.exe", "code.exe", // VS Code
		"Antigravity.exe",        // Antigravity IDE
		"devenv.exe",             // Visual Studio
		"idea64.exe", "idea.exe", // IntelliJ
		"goland64.exe", "goland.exe", // GoLand
		"pycharm64.exe", "pycharm.exe", // PyCharm
		"webstorm64.exe", "webstorm.exe", // WebStorm
		"vim.exe", "nvim.exe", // Vim
		"sublime_text.exe", // Sublime
		"notepad++.exe",    // Notepad++
	}

	for _, editor := range editors {
		if appName == editor {
			return true
		}
	}
	return false
}

// GenerateWeeklySummary 生成周报（代理到 analyzer）
func (s *AIService) GenerateWeeklySummary(ctx context.Context, req *ai.WeeklySummaryRequest) (*ai.WeeklySummaryResult, error) {
	return s.analyzer.GenerateWeeklySummary(ctx, req)
}
