package service

import (
	"testing"

	"github.com/yuqie6/WorkMirror/internal/schema"
)

func TestDefaultExpPolicy_EmptyDiffs(t *testing.T) {
	p := DefaultExpPolicy{}
	exp := p.CalcDiffExp(nil)
	if exp != 1 {
		t.Fatalf("empty diffs should return 1, got %v", exp)
	}
}

func TestDefaultExpPolicy_SingleDiff(t *testing.T) {
	p := DefaultExpPolicy{}
	diffs := []schema.Diff{
		{FilePath: "main.go", LinesAdded: 10, LinesDeleted: 5},
	}
	exp := p.CalcDiffExp(diffs)
	// base = 1 + 15/10 = 2.5
	// + 0.5*1 file + 0.2*0 hunks = 3.0
	if exp < 2.5 || exp > 5 {
		t.Fatalf("exp=%v, expected between 2.5 and 5", exp)
	}
}

func TestDefaultExpPolicy_MultipleDiffs(t *testing.T) {
	p := DefaultExpPolicy{}
	diffs := []schema.Diff{
		{FilePath: "main.go", LinesAdded: 20, LinesDeleted: 10, DiffContent: "@@ -1,5 +1,10 @@\n+code"},
		{FilePath: "util.go", LinesAdded: 5, LinesDeleted: 0, DiffContent: "@@ -1,2 +1,3 @@\n+more"},
	}
	exp := p.CalcDiffExp(diffs)
	// More lines and multiple files should give higher exp
	if exp < 4 {
		t.Fatalf("multiple diffs should give higher exp, got %v", exp)
	}
}

func TestDefaultExpPolicy_Clamped(t *testing.T) {
	p := DefaultExpPolicy{}
	// Large diff should be clamped to max 20
	diffs := []schema.Diff{
		{FilePath: "main.go", LinesAdded: 500, LinesDeleted: 500},
	}
	exp := p.CalcDiffExp(diffs)
	if exp > 20 {
		t.Fatalf("exp should be clamped to 20, got %v", exp)
	}
}

func TestCountHunks(t *testing.T) {
	cases := []struct {
		content string
		want    int
	}{
		{"", 0},
		{"@@ -1,5 +1,10 @@\n+code", 1},
		{"@@ -1,5 +1,10 @@\n+code\n@@ -20,3 +25,5 @@\n+more", 2},
		{"regular line\nno hunks here", 0},
	}

	for _, tc := range cases {
		got := countHunks(tc.content)
		if got != tc.want {
			t.Errorf("countHunks(%q) = %d, want %d", tc.content[:minInt(20, len(tc.content))], got, tc.want)
		}
	}
}

func TestClamp(t *testing.T) {
	cases := []struct {
		v, min, max, want float64
	}{
		{5, 0, 10, 5},   // within range
		{-5, 0, 10, 0},  // below min
		{15, 0, 10, 10}, // above max
		{0, 0, 10, 0},   // at min
		{10, 0, 10, 10}, // at max
	}

	for _, tc := range cases {
		got := clamp(tc.v, tc.min, tc.max)
		if got != tc.want {
			t.Errorf("clamp(%v, %v, %v) = %v, want %v", tc.v, tc.min, tc.max, got, tc.want)
		}
	}
}
