package repository

import (
	"context"
	"testing"

	"github.com/yuqie6/mirror/internal/schema"
	"github.com/yuqie6/mirror/internal/testutil"
)

func TestSkillRepositoryUpsertAndGet(t *testing.T) {
	db := testutil.OpenTestDB(t)
	repo := NewSkillRepository(db)
	ctx := context.Background()

	skill := &schema.SkillNode{Key: "go", Name: "Go", Category: "language", Level: 1, ExpToNext: 100}
	skill.Exp = 10

	if err := repo.Upsert(ctx, skill); err != nil {
		t.Fatalf("Upsert error: %v", err)
	}

	got, err := repo.GetByKey(ctx, "go")
	if err != nil {
		t.Fatalf("GetByKey error: %v", err)
	}
	if got == nil || got.Name != "Go" || got.Exp != 10 {
		t.Fatalf("got=%+v, want name Go exp 10", got)
	}
}

func TestSkillRepositoryUpsertBatchUpdatesExisting(t *testing.T) {
	db := testutil.OpenTestDB(t)
	repo := NewSkillRepository(db)
	ctx := context.Background()

	skill := &schema.SkillNode{Key: "go", Name: "Go", Category: "language", Level: 1, ExpToNext: 100}
	if err := repo.Upsert(ctx, skill); err != nil {
		t.Fatalf("Upsert error: %v", err)
	}

	updated := &schema.SkillNode{Key: "go", Name: "Go", Category: "language", Level: 1, ExpToNext: 100}
	updated.Exp = 42
	newSkill := &schema.SkillNode{Key: "react", Name: "React", Category: "framework", Level: 1, ExpToNext: 100}
	if err := repo.UpsertBatch(ctx, []*schema.SkillNode{updated, newSkill}); err != nil {
		t.Fatalf("UpsertBatch error: %v", err)
	}

	got, _ := repo.GetByKey(ctx, "go")
	if got.Exp != 42 {
		t.Fatalf("exp=%v, want 42", got.Exp)
	}
}
