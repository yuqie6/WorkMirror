package service

import (
	"context"
	"sort"
	"strings"

	"github.com/yuqie6/mirror/internal/schema"
	"github.com/yuqie6/mirror/internal/repository"
)

// BackfillSkillActivitiesFromDiffs 仅回填 skill_activities（不更新 skill_nodes，避免历史双计数）。
//
// 用途：兼容旧数据（只有 diffs.skills_detected，没有 skill_activities）。
func BackfillSkillActivitiesFromDiffs(
	ctx context.Context,
	diffRepo DiffRepository,
	activityRepo SkillActivityRepository,
	expPolicy ExpPolicy,
	startTime int64,
	endTime int64,
	limit int,
) (int64, error) {
	if diffRepo == nil || activityRepo == nil {
		return 0, nil
	}
	if expPolicy == nil {
		expPolicy = DefaultExpPolicy{}
	}
	if limit <= 0 {
		limit = 500
	}

	diffs, err := diffRepo.GetByTimeRange(ctx, startTime, endTime)
	if err != nil || len(diffs) == 0 {
		return 0, err
	}

	sort.Slice(diffs, func(i, j int) bool { return diffs[i].Timestamp > diffs[j].Timestamp })
	if len(diffs) > limit {
		diffs = diffs[:limit]
	}

	type pending struct {
		key repository.SkillActivityKey
		act schema.SkillActivity
	}
	pendingList := make([]pending, 0, len(diffs)*3)

	for _, d := range diffs {
		if d.ID <= 0 || d.Timestamp <= 0 || len(d.SkillsDetected) == 0 {
			continue
		}

		uniqKeys := make(map[string]struct{}, len(d.SkillsDetected))
		for _, name := range d.SkillsDetected {
			k := normalizeKey(strings.TrimSpace(name))
			if k == "" {
				continue
			}
			uniqKeys[k] = struct{}{}
		}
		if len(uniqKeys) == 0 {
			continue
		}

		baseExp := expPolicy.CalcDiffExp([]schema.Diff{d})
		perSkillExp := baseExp / float64(len(uniqKeys))

		for skillKey := range uniqKeys {
			k := repository.SkillActivityKey{Source: "diff", EvidenceID: d.ID, SkillKey: skillKey}
			pendingList = append(pendingList, pending{
				key: k,
				act: schema.SkillActivity{
					SkillKey:   skillKey,
					Source:     "diff",
					EvidenceID: d.ID,
					Exp:        perSkillExp,
					Timestamp:  d.Timestamp,
				},
			})
		}
	}

	if len(pendingList) == 0 {
		return 0, nil
	}

	keys := make([]repository.SkillActivityKey, 0, len(pendingList))
	for _, p := range pendingList {
		keys = append(keys, p.key)
	}

	existing, err := activityRepo.ListExistingKeys(ctx, keys)
	if err != nil {
		return 0, err
	}

	activities := make([]schema.SkillActivity, 0, len(pendingList))
	for _, p := range pendingList {
		if _, ok := existing[p.key]; ok {
			continue
		}
		activities = append(activities, p.act)
	}

	if len(activities) == 0 {
		return 0, nil
	}

	return activityRepo.BatchInsert(ctx, activities)
}
