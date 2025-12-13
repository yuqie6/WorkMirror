package service

import (
	"context"
	"testing"
	"time"

	"github.com/yuqie6/mirror/internal/ai"
	"github.com/yuqie6/mirror/internal/model"
	"github.com/yuqie6/mirror/internal/repository"
)

type fakeSkillRepo struct {
	items       map[string]*model.SkillNode
	upserted    []*model.SkillNode
	upsertBatch []*model.SkillNode
}

func newFakeSkillRepo(skills ...*model.SkillNode) *fakeSkillRepo {
	m := make(map[string]*model.SkillNode)
	for _, s := range skills {
		copy := *s
		m[s.Key] = &copy
	}
	return &fakeSkillRepo{items: m}
}

func (r *fakeSkillRepo) GetAll(ctx context.Context) ([]model.SkillNode, error) {
	out := make([]model.SkillNode, 0, len(r.items))
	for _, s := range r.items {
		out = append(out, *s)
	}
	return out, nil
}
func (r *fakeSkillRepo) GetByKey(ctx context.Context, key string) (*model.SkillNode, error) {
	if s, ok := r.items[key]; ok {
		copy := *s
		return &copy, nil
	}
	return nil, nil
}
func (r *fakeSkillRepo) Upsert(ctx context.Context, skill *model.SkillNode) error {
	copy := *skill
	r.items[skill.Key] = &copy
	r.upserted = append(r.upserted, &copy)
	return nil
}
func (r *fakeSkillRepo) UpsertBatch(ctx context.Context, skills []*model.SkillNode) error {
	r.upsertBatch = append(r.upsertBatch, skills...)
	for _, s := range skills {
		copy := *s
		r.items[s.Key] = &copy
	}
	return nil
}
func (r *fakeSkillRepo) GetTopSkills(ctx context.Context, limit int) ([]model.SkillNode, error) {
	return nil, nil
}
func (r *fakeSkillRepo) GetActiveSkillsInPeriod(ctx context.Context, startTime, endTime int64, limit int) ([]model.SkillNode, error) {
	return nil, nil
}

type fakeDiffRepo struct{}

func (f fakeDiffRepo) Create(ctx context.Context, diff *model.Diff) error { return nil }
func (f fakeDiffRepo) GetPendingAIAnalysis(ctx context.Context, limit int) ([]model.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepo) UpdateAIInsight(ctx context.Context, id int64, insight string, skills []string) error {
	return nil
}
func (f fakeDiffRepo) GetByDate(ctx context.Context, date string) ([]model.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepo) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]model.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepo) GetByIDs(ctx context.Context, ids []int64) ([]model.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepo) GetLanguageStats(ctx context.Context, startTime, endTime int64) ([]repository.LanguageStat, error) {
	return nil, nil
}
func (f fakeDiffRepo) CountByDateRange(ctx context.Context, startTime, endTime int64) (int64, error) {
	return 0, nil
}
func (f fakeDiffRepo) GetRecentAnalyzed(ctx context.Context, limit int) ([]model.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepo) GetByID(ctx context.Context, id int64) (*model.Diff, error) {
	return nil, nil
}

func TestNormalizeKey(t *testing.T) {
	cases := map[string]string{
		"React.js":       "reactjs",
		"Next.js App":    "nextjs-app",
		"C++":            "cpp",
		"C#/.NET":        "csharpdotnet",
		"TypeScript":     "typescript",
		"foo bar baz":    "foo-bar-baz",
		"hello_world.js": "helloworld-js",
	}
	for in, want := range cases {
		if got := normalizeKey(in); got != want {
			t.Fatalf("normalizeKey(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestUpdateSkillsFromDiffsWithCategory(t *testing.T) {
	ctx := context.Background()
	existingParent := model.NewSkillNode("go", "Go", "language")
	existingChild := model.NewSkillNode("reactjs", "ReactJS", "other")
	repo := newFakeSkillRepo(existingParent, existingChild)

	svc := NewSkillService(repo, fakeDiffRepo{}, DefaultExpPolicy{})

	diffs := []model.Diff{
		{LinesAdded: 5, LinesDeleted: 5},
	}
	aiSkills := []ai.SkillWithCategory{
		{Name: "React.js", Category: "framework", Parent: "Go"},
		{Name: "TypeScript", Category: "language"},
	}

	if err := svc.UpdateSkillsFromDiffsWithCategory(ctx, diffs, aiSkills); err != nil {
		t.Fatalf("UpdateSkillsFromDiffsWithCategory error: %v", err)
	}

	if len(repo.upsertBatch) != 2 {
		t.Fatalf("upsertBatch count=%d, want 2", len(repo.upsertBatch))
	}

	updatedReact := repo.items["reactjs"]
	if updatedReact == nil || updatedReact.Category != "framework" {
		t.Fatalf("react category=%q, want framework", updatedReact.Category)
	}
	if updatedReact.ParentKey != "go" {
		t.Fatalf("react parentKey=%q, want go (existing key)", updatedReact.ParentKey)
	}

	ts := repo.items["typescript"]
	if ts == nil || ts.Category != "language" {
		t.Fatalf("typescript category=%q, want language", ts.Category)
	}
	if ts.Exp <= 0 {
		t.Fatalf("typescript exp not added")
	}
}

func TestUpdateSkillsFromDiffsWithCategory_EmptySkillsNoop(t *testing.T) {
	ctx := context.Background()
	repo := newFakeSkillRepo()
	svc := NewSkillService(repo, fakeDiffRepo{}, DefaultExpPolicy{})
	if err := svc.UpdateSkillsFromDiffsWithCategory(ctx, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.upsertBatch) != 0 {
		t.Fatalf("upsertBatch should be empty")
	}
}

func TestApplyDecayToAll(t *testing.T) {
	ctx := context.Background()
	oldSkill := model.NewSkillNode("go", "Go", "language")
	oldSkill.Exp = 100
	oldSkill.LastActive = time.Now().Add(-10 * 24 * time.Hour).UnixMilli()
	recentSkill := model.NewSkillNode("react", "React", "framework")
	recentSkill.Exp = 50
	recentSkill.LastActive = time.Now().Add(-2 * 24 * time.Hour).UnixMilli()

	repo := newFakeSkillRepo(oldSkill, recentSkill)
	svc := NewSkillService(repo, fakeDiffRepo{}, DefaultExpPolicy{})

	if err := svc.ApplyDecayToAll(ctx); err != nil {
		t.Fatalf("ApplyDecayToAll error: %v", err)
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("upserted=%d, want 1", len(repo.upserted))
	}
	if repo.items["go"].Exp >= 100 {
		t.Fatalf("old skill exp not decayed")
	}
}
