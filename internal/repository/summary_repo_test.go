package repository

import (
	"context"
	"testing"

	"github.com/yuqie6/mirror/internal/schema"
	"github.com/yuqie6/mirror/internal/testutil"
)

func TestSummaryRepositoryUpsertAndGet(t *testing.T) {
	db := testutil.OpenTestDB(t)
	repo := NewSummaryRepository(db)
	ctx := context.Background()

	summary := &schema.DailySummary{Date: "2025-12-12", Summary: "ok"}
	if err := repo.Upsert(ctx, summary); err != nil {
		t.Fatalf("Upsert error: %v", err)
	}

	got, err := repo.GetByDate(ctx, "2025-12-12")
	if err != nil {
		t.Fatalf("GetByDate error: %v", err)
	}
	if got == nil || got.Summary != "ok" {
		t.Fatalf("got=%+v, want summary ok", got)
	}
}
