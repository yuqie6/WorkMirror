package service

import (
	"context"
	"testing"
	"time"

	"github.com/yuqie6/WorkMirror/internal/ai"
	"github.com/yuqie6/WorkMirror/internal/repository"
	"github.com/yuqie6/WorkMirror/internal/schema"
)

// ===== Mock Implementations for SessionSemanticService =====

type fakeSessionRepoForSemantic struct {
	sessions   []schema.Session
	updated    map[int64]schema.SessionSemanticUpdate
	maxVersion int
}

func (f *fakeSessionRepoForSemantic) Create(ctx context.Context, session *schema.Session) (bool, error) {
	return true, nil
}
func (f *fakeSessionRepoForSemantic) UpdateSemantic(ctx context.Context, id int64, update schema.SessionSemanticUpdate) error {
	if f.updated == nil {
		f.updated = make(map[int64]schema.SessionSemanticUpdate)
	}
	f.updated[id] = update
	return nil
}
func (f *fakeSessionRepoForSemantic) GetByDate(ctx context.Context, date string) ([]schema.Session, error) {
	return f.sessions, nil
}
func (f *fakeSessionRepoForSemantic) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Session, error) {
	out := make([]schema.Session, 0)
	for _, s := range f.sessions {
		if s.StartTime >= startTime && s.StartTime <= endTime {
			out = append(out, s)
		}
	}
	return out, nil
}
func (f *fakeSessionRepoForSemantic) GetMaxSessionVersionByDate(ctx context.Context, date string) (int, error) {
	return f.maxVersion, nil
}
func (f *fakeSessionRepoForSemantic) GetLastSession(ctx context.Context) (*schema.Session, error) {
	if len(f.sessions) == 0 {
		return nil, nil
	}
	return &f.sessions[len(f.sessions)-1], nil
}
func (f *fakeSessionRepoForSemantic) GetByID(ctx context.Context, id int64) (*schema.Session, error) {
	for i := range f.sessions {
		if f.sessions[i].ID == id {
			return &f.sessions[i], nil
		}
	}
	return nil, nil
}

type fakeDiffRepoForSemantic struct {
	diffs []schema.Diff
}

func (f fakeDiffRepoForSemantic) Create(ctx context.Context, diff *schema.Diff) error { return nil }
func (f fakeDiffRepoForSemantic) GetPendingAIAnalysis(ctx context.Context, limit int) ([]schema.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepoForSemantic) UpdateAIInsight(ctx context.Context, id int64, insight string, skills []string) error {
	return nil
}
func (f fakeDiffRepoForSemantic) GetByDate(ctx context.Context, date string) ([]schema.Diff, error) {
	return f.diffs, nil
}
func (f fakeDiffRepoForSemantic) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Diff, error) {
	out := make([]schema.Diff, 0)
	for _, d := range f.diffs {
		if d.Timestamp >= startTime && d.Timestamp <= endTime {
			out = append(out, d)
		}
	}
	return out, nil
}
func (f fakeDiffRepoForSemantic) GetByIDs(ctx context.Context, ids []int64) ([]schema.Diff, error) {
	out := make([]schema.Diff, 0)
	idSet := make(map[int64]struct{})
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	for _, d := range f.diffs {
		if _, ok := idSet[d.ID]; ok {
			out = append(out, d)
		}
	}
	return out, nil
}
func (f fakeDiffRepoForSemantic) GetLanguageStats(ctx context.Context, startTime, endTime int64) ([]repository.LanguageStat, error) {
	return nil, nil
}
func (f fakeDiffRepoForSemantic) CountByDateRange(ctx context.Context, startTime, endTime int64) (int64, error) {
	return 0, nil
}
func (f fakeDiffRepoForSemantic) GetRecentAnalyzed(ctx context.Context, limit int) ([]schema.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepoForSemantic) GetByID(ctx context.Context, id int64) (*schema.Diff, error) {
	return nil, nil
}

type fakeEventRepoForSemantic struct {
	stats []repository.AppStat
}

func (f fakeEventRepoForSemantic) BatchInsert(ctx context.Context, events []schema.Event) error {
	return nil
}
func (f fakeEventRepoForSemantic) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Event, error) {
	return nil, nil
}
func (f fakeEventRepoForSemantic) GetByDate(ctx context.Context, date string) ([]schema.Event, error) {
	return nil, nil
}
func (f fakeEventRepoForSemantic) GetAppStats(ctx context.Context, startTime, endTime int64) ([]repository.AppStat, error) {
	return f.stats, nil
}
func (f fakeEventRepoForSemantic) Count(ctx context.Context) (int64, error) { return 0, nil }

type fakeBrowserRepoForSemantic struct {
	events []schema.BrowserEvent
}

func (f fakeBrowserRepoForSemantic) BatchInsert(ctx context.Context, events []*schema.BrowserEvent) error {
	return nil
}
func (f fakeBrowserRepoForSemantic) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.BrowserEvent, error) {
	return f.events, nil
}
func (f fakeBrowserRepoForSemantic) GetByIDs(ctx context.Context, ids []int64) ([]schema.BrowserEvent, error) {
	return nil, nil
}

type fakeAnalyzerForSemantic struct {
	sessionResult *ai.SessionSummaryResult
	called        int
}

func (f *fakeAnalyzerForSemantic) AnalyzeDiff(ctx context.Context, filePath, language, diffContent string, existingSkills []ai.SkillInfo) (*ai.DiffInsight, error) {
	return &ai.DiffInsight{}, nil
}
func (f *fakeAnalyzerForSemantic) GenerateDailySummary(ctx context.Context, req *ai.DailySummaryRequest) (*ai.DailySummaryResult, error) {
	return &ai.DailySummaryResult{}, nil
}
func (f *fakeAnalyzerForSemantic) GenerateWeeklySummary(ctx context.Context, req *ai.WeeklySummaryRequest) (*ai.WeeklySummaryResult, error) {
	return &ai.WeeklySummaryResult{}, nil
}
func (f *fakeAnalyzerForSemantic) GenerateSessionSummary(ctx context.Context, req *ai.SessionSummaryRequest) (*ai.SessionSummaryResult, error) {
	f.called++
	if f.sessionResult != nil {
		return f.sessionResult, nil
	}
	return &ai.SessionSummaryResult{
		Summary:        "Mock session summary",
		Category:       "coding",
		SkillsInvolved: []string{"Go", "React"},
		Tags:           []string{"backend"},
	}, nil
}

// ===== Test Cases =====

func TestShouldEnrichSession_EmptySummary(t *testing.T) {
	sess := &schema.Session{ID: 1, Summary: ""}
	if !shouldEnrichSession(sess) {
		t.Fatal("session with empty summary should need enrichment")
	}
}

func TestShouldEnrichSession_HasSummaryAndSkills(t *testing.T) {
	sess := &schema.Session{
		ID:             1,
		Summary:        "Already has summary",
		SkillsInvolved: []string{"Go"}, // has skills too
		Metadata: schema.JSONMap{
			"semantic_source": "ai",
			"evidence_hint":   "diff",
		},
	}
	if shouldEnrichSession(sess) {
		t.Fatal("session with summary AND skills should NOT need enrichment")
	}
}

func TestShouldEnrichSession_NilSession(t *testing.T) {
	if shouldEnrichSession(nil) {
		t.Fatal("nil session should NOT need enrichment")
	}
	if shouldEnrichSession(&schema.Session{ID: 0}) {
		t.Fatal("session with zero ID should NOT need enrichment")
	}
}

func TestEnrichSessionsForDate_EnrichesNeedingSessions(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	baseTs := now.Truncate(time.Hour).UnixMilli()

	sessions := []schema.Session{
		{ID: 1, StartTime: baseTs, EndTime: baseTs + 1000, Summary: ""}, // needs enrichment
		{ID: 2, StartTime: baseTs, EndTime: baseTs + 2000, Summary: "Already done", SkillsInvolved: []string{"Go"}, Metadata: schema.JSONMap{"semantic_source": "ai", "evidence_hint": "diff"}}, // skip - has both
	}

	diffs := []schema.Diff{
		{ID: 101, Timestamp: baseTs + 500, FileName: "main.go", Language: "Go", SkillsDetected: []string{"Go"}},
	}

	sessionRepo := &fakeSessionRepoForSemantic{sessions: sessions}
	analyzer := &fakeAnalyzerForSemantic{}

	svc := NewSessionSemanticService(
		analyzer,
		sessionRepo,
		fakeDiffRepoForSemantic{diffs: diffs},
		fakeEventRepoForSemantic{},
		fakeBrowserRepoForSemantic{},
	)

	updated, err := svc.EnrichSessionsForDate(ctx, now.Format("2006-01-02"), 10)
	if err != nil {
		t.Fatalf("EnrichSessionsForDate error: %v", err)
	}
	if updated != 1 {
		t.Fatalf("updated=%d, want 1", updated)
	}
	if analyzer.called != 1 {
		t.Fatalf("analyzer.called=%d, want 1", analyzer.called)
	}
	if _, ok := sessionRepo.updated[1]; !ok {
		t.Fatal("session 1 should have been updated")
	}
	if _, ok := sessionRepo.updated[2]; ok {
		t.Fatal("session 2 should NOT have been updated")
	}
}

func TestFallbackSessionCategory(t *testing.T) {
	cases := []struct {
		diffs   []schema.Diff
		browser []schema.BrowserEvent
		want    string
	}{
		{[]schema.Diff{{ID: 1}}, []schema.BrowserEvent{{ID: 1}}, "exploration"},
		{[]schema.Diff{{ID: 1}}, nil, "technical"},
		{nil, []schema.BrowserEvent{{ID: 1}}, "learning"},
		{nil, nil, "other"},
	}

	for _, tc := range cases {
		got := fallbackSessionCategory(tc.diffs, tc.browser)
		if got != tc.want {
			t.Errorf("fallbackSessionCategory(%d diffs, %d browser) = %q, want %q",
				len(tc.diffs), len(tc.browser), got, tc.want)
		}
	}
}

func TestFallbackSessionSummary(t *testing.T) {
	sess := &schema.Session{PrimaryApp: "code.exe"}
	diffs := []schema.Diff{{Language: "Go"}}
	domains := []string{"github.com"}
	skills := []string{"Go", "React"}

	summary := fallbackSessionSummary(sess, diffs, domains, skills)

	if summary == "" {
		t.Fatal("fallback summary should not be empty")
	}
	// 应该包含关键信息
	if !contains(summary, "Go") && !contains(summary, "React") {
		t.Errorf("summary should mention skills: %s", summary)
	}
}

func TestTopKeysByCount(t *testing.T) {
	m := map[string]int{
		"github.com":    10,
		"google.com":    5,
		"stackoverflow": 8,
		"":              100, // empty should be filtered
	}

	top := topKeysByCount(m, 2)
	if len(top) != 2 {
		t.Fatalf("len=%d, want 2", len(top))
	}
	if top[0] != "github.com" {
		t.Fatalf("top[0]=%q, want github.com", top[0])
	}
}

func TestUniqueNonEmpty(t *testing.T) {
	input := []string{"Go", " go ", "React", "", "  ", "Go", "TypeScript"}
	result := uniqueNonEmpty(input, 10)

	if len(result) != 3 {
		t.Fatalf("len=%d, want 3 (Go, React, TypeScript)", len(result))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
