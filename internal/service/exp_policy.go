package service

import (
	"strings"

	"github.com/yuqie6/mirror/internal/schema"
)

// ExpPolicy 经验计算策略（可替换）
type ExpPolicy interface {
	CalcDiffExp(diffs []schema.Diff) float64
}

// DefaultExpPolicy 默认经验策略：行数主导 + 低成本修正 + clamp
type DefaultExpPolicy struct{}

func (p DefaultExpPolicy) CalcDiffExp(diffs []schema.Diff) float64 {
	if len(diffs) == 0 {
		return 1
	}

	totalLines := 0
	filesChanged := 0
	hunksChanged := 0
	seenFiles := make(map[string]struct{})

	for _, d := range diffs {
		totalLines += d.LinesAdded + d.LinesDeleted
		if d.FilePath != "" {
			if _, ok := seenFiles[d.FilePath]; !ok {
				seenFiles[d.FilePath] = struct{}{}
				filesChanged++
			}
		}
		hunksChanged += countHunks(d.DiffContent)
	}

	base := 1.0 + float64(totalLines)/10.0
	exp := base + 0.5*float64(filesChanged) + 0.2*float64(hunksChanged)

	return clamp(exp, 1, 20)
}

func countHunks(diffContent string) int {
	if diffContent == "" {
		return 0
	}
	n := 0
	for _, line := range strings.Split(diffContent, "\n") {
		if strings.HasPrefix(line, "@@") {
			n++
		}
	}
	return n
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
