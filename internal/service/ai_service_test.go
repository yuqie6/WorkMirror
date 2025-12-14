package service

import (
	"context"
	"testing"
	"time"

	"github.com/yuqie6/WorkMirror/internal/ai"
	"github.com/yuqie6/WorkMirror/internal/repository"
	"github.com/yuqie6/WorkMirror/internal/schema"
)

// ===== Mock Implementations =====

type fakeAnalyzer struct {
	diffInsight   *ai.DiffInsight
	dailyResult   *ai.DailySummaryResult
	weeklyResult  *ai.WeeklySummaryResult
	sessionResult *ai.SessionSummaryResult
	analyzeCalled int
	summaryCalled int
}

func (f *fakeAnalyzer) AnalyzeDiff(ctx context.Context, filePath, language, diffContent string, existingSkills []ai.SkillInfo) (*ai.DiffInsight, error) {
	f.analyzeCalled++
	if f.diffInsight != nil {
		return f.diffInsight, nil
	}
	return &ai.DiffInsight{
		Insight: "Mock insight for " + filePath,
		Skills:  []ai.SkillWithCategory{{Name: "Go", Category: "language"}},
	}, nil
}

func (f *fakeAnalyzer) GenerateDailySummary(ctx context.Context, req *ai.DailySummaryRequest) (*ai.DailySummaryResult, error) {
	f.summaryCalled++
	if f.dailyResult != nil {
		return f.dailyResult, nil
	}
	return &ai.DailySummaryResult{
		Summary:      "Mock daily summary",
		Highlights:   "Mock highlights",
		Struggles:    "Mock struggles",
		SkillsGained: []string{"Go"},
	}, nil
}

func (f *fakeAnalyzer) GenerateWeeklySummary(ctx context.Context, req *ai.WeeklySummaryRequest) (*ai.WeeklySummaryResult, error) {
	if f.weeklyResult != nil {
		return f.weeklyResult, nil
	}
	return &ai.WeeklySummaryResult{
		Overview:     "Mock weekly overview",
		Achievements: []string{"Achievement 1"},
		Patterns:     "Mock patterns",
		Suggestions:  "Mock suggestions",
	}, nil
}

func (f *fakeAnalyzer) GenerateSessionSummary(ctx context.Context, req *ai.SessionSummaryRequest) (*ai.SessionSummaryResult, error) {
	if f.sessionResult != nil {
		return f.sessionResult, nil
	}
	return &ai.SessionSummaryResult{
		Summary:        "Mock session summary",
		Category:       "coding",
		SkillsInvolved: []string{"Go"},
	}, nil
}

type fakeDiffRepoForAI struct {
	pending []schema.Diff
	updated map[int64]string
}

func (f *fakeDiffRepoForAI) Create(ctx context.Context, diff *schema.Diff) error { return nil }
func (f *fakeDiffRepoForAI) GetPendingAIAnalysis(ctx context.Context, limit int) ([]schema.Diff, error) {
	if limit > len(f.pending) {
		return f.pending, nil
	}
	return f.pending[:limit], nil
}
func (f *fakeDiffRepoForAI) UpdateAIInsight(ctx context.Context, id int64, insight string, skills []string) error {
	if f.updated == nil {
		f.updated = make(map[int64]string)
	}
	f.updated[id] = insight
	return nil
}
func (f *fakeDiffRepoForAI) GetByDate(ctx context.Context, date string) ([]schema.Diff, error) {
	return f.pending, nil
}
func (f *fakeDiffRepoForAI) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Diff, error) {
	return nil, nil
}
func (f *fakeDiffRepoForAI) GetByIDs(ctx context.Context, ids []int64) ([]schema.Diff, error) {
	return nil, nil
}
func (f *fakeDiffRepoForAI) GetLanguageStats(ctx context.Context, startTime, endTime int64) ([]repository.LanguageStat, error) {
	return nil, nil
}
func (f *fakeDiffRepoForAI) CountByDateRange(ctx context.Context, startTime, endTime int64) (int64, error) {
	return 0, nil
}
func (f *fakeDiffRepoForAI) GetRecentAnalyzed(ctx context.Context, limit int) ([]schema.Diff, error) {
	return nil, nil
}
func (f *fakeDiffRepoForAI) GetByID(ctx context.Context, id int64) (*schema.Diff, error) {
	return nil, nil
}

type fakeEventRepoForAI struct {
	stats []repository.AppStat
}

func (f fakeEventRepoForAI) BatchInsert(ctx context.Context, events []schema.Event) error { return nil }
func (f fakeEventRepoForAI) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Event, error) {
	return nil, nil
}
func (f fakeEventRepoForAI) GetByDate(ctx context.Context, date string) ([]schema.Event, error) {
	return nil, nil
}
func (f fakeEventRepoForAI) GetAppStats(ctx context.Context, startTime, endTime int64) ([]repository.AppStat, error) {
	return f.stats, nil
}
func (f fakeEventRepoForAI) Count(ctx context.Context) (int64, error) { return 0, nil }

type fakeSummaryRepo struct {
	summaries map[string]*schema.DailySummary
	upserted  []*schema.DailySummary
}

func (f *fakeSummaryRepo) GetByDate(ctx context.Context, date string) (*schema.DailySummary, error) {
	if f.summaries != nil {
		return f.summaries[date], nil
	}
	return nil, nil
}
func (f *fakeSummaryRepo) Upsert(ctx context.Context, summary *schema.DailySummary) error {
	f.upserted = append(f.upserted, summary)
	if f.summaries == nil {
		f.summaries = make(map[string]*schema.DailySummary)
	}
	f.summaries[summary.Date] = summary
	return nil
}
func (f *fakeSummaryRepo) GetRecent(ctx context.Context, limit int) ([]schema.DailySummary, error) {
	return nil, nil
}

// ===== Test Cases =====

func TestAnalyzePendingDiffs_NoPending(t *testing.T) {
	ctx := context.Background()

	analyzer := &fakeAnalyzer{}
	diffRepo := &fakeDiffRepoForAI{pending: nil}
	skillRepo := newFakeSkillRepo()
	skillSvc := NewSkillService(skillRepo, diffRepo, nil, DefaultExpPolicy{})

	svc := NewAIService(analyzer, diffRepo, fakeEventRepoForAI{}, &fakeSummaryRepo{}, skillSvc)

	count, err := svc.AnalyzePendingDiffs(ctx, 10)
	if err != nil {
		t.Fatalf("AnalyzePendingDiffs error: %v", err)
	}
	if count != 0 {
		t.Fatalf("count=%d, want 0", count)
	}
	if analyzer.analyzeCalled != 0 {
		t.Fatalf("analyzer should not be called when no pending diffs")
	}
}

func TestAnalyzePendingDiffs_ProcessesDiffs(t *testing.T) {
	ctx := context.Background()

	pending := []schema.Diff{
		{ID: 1, FilePath: "main.go", Language: "Go", DiffContent: "+new code"},
		{ID: 2, FilePath: "util.go", Language: "Go", DiffContent: "+util code"},
	}

	analyzer := &fakeAnalyzer{}
	diffRepo := &fakeDiffRepoForAI{pending: pending}
	skillRepo := newFakeSkillRepo()
	skillSvc := NewSkillService(skillRepo, diffRepo, nil, DefaultExpPolicy{})

	svc := NewAIService(analyzer, diffRepo, fakeEventRepoForAI{}, &fakeSummaryRepo{}, skillSvc)

	count, err := svc.AnalyzePendingDiffs(ctx, 10)
	if err != nil {
		t.Fatalf("AnalyzePendingDiffs error: %v", err)
	}
	if count != 2 {
		t.Fatalf("count=%d, want 2", count)
	}
	if len(diffRepo.updated) != 2 {
		t.Fatalf("updated diffs=%d, want 2", len(diffRepo.updated))
	}
}

func TestGenerateDailySummary_CacheHitForPastDate(t *testing.T) {
	ctx := context.Background()

	pastDate := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
	cached := &schema.DailySummary{
		Date:    pastDate,
		Summary: "Cached summary",
	}

	analyzer := &fakeAnalyzer{}
	summaryRepo := &fakeSummaryRepo{
		summaries: map[string]*schema.DailySummary{pastDate: cached},
	}
	skillRepo := newFakeSkillRepo()
	skillSvc := NewSkillService(skillRepo, &fakeDiffRepoForAI{}, nil, DefaultExpPolicy{})

	svc := NewAIService(analyzer, &fakeDiffRepoForAI{}, fakeEventRepoForAI{}, summaryRepo, skillSvc)

	result, err := svc.GenerateDailySummary(ctx, pastDate)
	if err != nil {
		t.Fatalf("GenerateDailySummary error: %v", err)
	}
	if result.Summary != "Cached summary" {
		t.Fatalf("summary=%q, want cached", result.Summary)
	}
	if analyzer.summaryCalled != 0 {
		t.Fatalf("analyzer should not be called for cached past date")
	}
}

func TestGenerateDailySummary_GeneratesForToday(t *testing.T) {
	ctx := context.Background()

	today := time.Now().Format("2006-01-02")

	analyzer := &fakeAnalyzer{}
	diffRepo := &fakeDiffRepoForAI{
		pending: []schema.Diff{{ID: 1, FileName: "main.go", Language: "Go"}},
	}
	summaryRepo := &fakeSummaryRepo{}
	skillRepo := newFakeSkillRepo()
	skillSvc := NewSkillService(skillRepo, diffRepo, nil, DefaultExpPolicy{})

	svc := NewAIService(analyzer, diffRepo, fakeEventRepoForAI{}, summaryRepo, skillSvc)

	result, err := svc.GenerateDailySummary(ctx, today)
	if err != nil {
		t.Fatalf("GenerateDailySummary error: %v", err)
	}
	if result.Summary != "Mock daily summary" {
		t.Fatalf("summary=%q, want mock", result.Summary)
	}
	if analyzer.summaryCalled != 1 {
		t.Fatalf("analyzer.summaryCalled=%d, want 1", analyzer.summaryCalled)
	}
	if len(summaryRepo.upserted) != 1 {
		t.Fatalf("summary not persisted")
	}
}
