package service

import (
	"context"
	"testing"
	"time"

	"github.com/yuqie6/mirror/internal/model"
	"github.com/yuqie6/mirror/internal/repository"
)

type fakeDiffRepoForTrend struct {
	langStats []repository.LanguageStat
}

func (f fakeDiffRepoForTrend) Create(ctx context.Context, diff *model.Diff) error { return nil }
func (f fakeDiffRepoForTrend) GetPendingAIAnalysis(ctx context.Context, limit int) ([]model.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepoForTrend) UpdateAIInsight(ctx context.Context, id int64, insight string, skills []string) error {
	return nil
}
func (f fakeDiffRepoForTrend) GetByDate(ctx context.Context, date string) ([]model.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepoForTrend) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]model.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepoForTrend) GetByIDs(ctx context.Context, ids []int64) ([]model.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepoForTrend) GetLanguageStats(ctx context.Context, startTime, endTime int64) ([]repository.LanguageStat, error) {
	return f.langStats, nil
}
func (f fakeDiffRepoForTrend) CountByDateRange(ctx context.Context, startTime, endTime int64) (int64, error) {
	return 0, nil
}
func (f fakeDiffRepoForTrend) GetRecentAnalyzed(ctx context.Context, limit int) ([]model.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepoForTrend) GetByID(ctx context.Context, id int64) (*model.Diff, error) {
	return nil, nil
}

type fakeSkillRepoForTrend struct {
	active []model.SkillNode
}

func (f fakeSkillRepoForTrend) GetAll(ctx context.Context) ([]model.SkillNode, error) {
	return nil, nil
}
func (f fakeSkillRepoForTrend) GetByKey(ctx context.Context, key string) (*model.SkillNode, error) {
	return nil, nil
}
func (f fakeSkillRepoForTrend) Upsert(ctx context.Context, skill *model.SkillNode) error { return nil }
func (f fakeSkillRepoForTrend) UpsertBatch(ctx context.Context, skills []*model.SkillNode) error {
	return nil
}
func (f fakeSkillRepoForTrend) GetTopSkills(ctx context.Context, limit int) ([]model.SkillNode, error) {
	return nil, nil
}
func (f fakeSkillRepoForTrend) GetActiveSkillsInPeriod(ctx context.Context, startTime, endTime int64, limit int) ([]model.SkillNode, error) {
	return f.active, nil
}

type fakeEventRepoForTrend struct {
	stats []repository.AppStat
}

func (f fakeEventRepoForTrend) BatchInsert(ctx context.Context, events []model.Event) error {
	return nil
}
func (f fakeEventRepoForTrend) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]model.Event, error) {
	return nil, nil
}
func (f fakeEventRepoForTrend) GetByDate(ctx context.Context, date string) ([]model.Event, error) {
	return nil, nil
}
func (f fakeEventRepoForTrend) GetAppStats(ctx context.Context, startTime, endTime int64) ([]repository.AppStat, error) {
	return f.stats, nil
}
func (f fakeEventRepoForTrend) Count(ctx context.Context) (int64, error) { return 0, nil }

func TestGetTrendReport(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	langStats := []repository.LanguageStat{
		{Language: "go", DiffCount: 2, LinesAdded: 10, LinesDeleted: 1},
		{Language: "ts", DiffCount: 1, LinesAdded: 3, LinesDeleted: 0},
	}

	activeSkills := []model.SkillNode{
		{Key: "go", Name: "Go", Category: "language", LastActive: now.UnixMilli()},
		{Key: "react", Name: "React", Category: "framework", LastActive: now.Add(-9 * 24 * time.Hour).UnixMilli()},
	}

	appStats := []repository.AppStat{
		{AppName: "code.exe", TotalDuration: 3600},
		{AppName: "chrome.exe", TotalDuration: 7200},
	}

	svc := NewTrendService(
		fakeSkillRepoForTrend{active: activeSkills},
		fakeDiffRepoForTrend{langStats: langStats},
		fakeEventRepoForTrend{stats: appStats},
	)

	report, err := svc.GetTrendReport(ctx, TrendPeriod7Days)
	if err != nil {
		t.Fatalf("GetTrendReport error: %v", err)
	}

	if report.Period != TrendPeriod7Days {
		t.Fatalf("period=%q, want 7d", report.Period)
	}
	if report.TotalDiffs != 3 {
		t.Fatalf("totalDiffs=%d, want 3", report.TotalDiffs)
	}
	if len(report.TopLanguages) != 2 || report.TopLanguages[0].Percentage <= 0 {
		t.Fatalf("topLanguages unexpected: %+v", report.TopLanguages)
	}
	if report.TotalCodingMins != 60 { // 3600s from code.exe
		t.Fatalf("totalCodingMins=%d, want 60", report.TotalCodingMins)
	}
	if len(report.TopSkills) != 2 {
		t.Fatalf("topSkills len=%d, want 2", len(report.TopSkills))
	}
	if report.TopSkills[0].Status != "growing" {
		t.Fatalf("go status=%q, want growing", report.TopSkills[0].Status)
	}
	if report.TopSkills[1].Status != "declining" {
		t.Fatalf("react status=%q, want declining", report.TopSkills[1].Status)
	}
	if len(report.Bottlenecks) != 1 {
		t.Fatalf("bottlenecks=%v, want 1 item", report.Bottlenecks)
	}
}

func TestDetectBottlenecks(t *testing.T) {
	svc := &TrendService{}
	b := svc.detectBottlenecks([]SkillTrend{{SkillName: "Go", Status: "declining"}})
	if len(b) != 1 {
		t.Fatalf("len=%d, want 1", len(b))
	}
	b = svc.detectBottlenecks(nil)
	if len(b) != 1 || b[0] == "" {
		t.Fatalf("empty skills should return default message")
	}
}
