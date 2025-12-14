package service

import (
	"testing"

	"github.com/yuqie6/WorkMirror/internal/repository"
)

func TestSecondsToMinutesFloor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		seconds int
		want    int
	}{
		{seconds: -1, want: 0},
		{seconds: 0, want: 0},
		{seconds: 1, want: 0},
		{seconds: 59, want: 0},
		{seconds: 60, want: 1},
		{seconds: 61, want: 1},
		{seconds: 120, want: 2},
	}

	for _, c := range cases {
		if got := SecondsToMinutesFloor(c.seconds); got != c.want {
			t.Fatalf("SecondsToMinutesFloor(%d)=%d, want %d", c.seconds, got, c.want)
		}
	}
}

func TestTopAppStats(t *testing.T) {
	t.Parallel()

	stats := []repository.AppStat{
		{AppName: "a", TotalDuration: 10},
		{AppName: "b", TotalDuration: 9},
		{AppName: "c", TotalDuration: 8},
	}

	if got := TopAppStats(stats, 0); len(got) != 3 {
		t.Fatalf("TopAppStats(limit=0) len=%d, want 3", len(got))
	}
	if got := TopAppStats(stats, 2); len(got) != 2 {
		t.Fatalf("TopAppStats(limit=2) len=%d, want 2", len(got))
	}
	if got := TopAppStats(stats, 10); len(got) != 3 {
		t.Fatalf("TopAppStats(limit=10) len=%d, want 3", len(got))
	}
}

func TestSumCodingMinutesFromAppStats(t *testing.T) {
	t.Parallel()

	stats := []repository.AppStat{
		{AppName: "Code.exe", TotalDuration: 60},
		{AppName: "notepad.exe", TotalDuration: 60},
	}

	if got := SumCodingMinutesFromAppStats(stats); got != 1 {
		t.Fatalf("SumCodingMinutesFromAppStats=%d, want 1", got)
	}
}
