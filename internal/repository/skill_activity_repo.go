package repository

import (
	"context"
	"fmt"

	"github.com/yuqie6/mirror/internal/schema"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SkillActivityKey struct {
	Source     string
	EvidenceID int64
	SkillKey   string
}

type SkillActivityStat struct {
	SkillKey    string
	ExpSum      float64
	EventCount  int64
	DaysActive  int
	LastTsMilli int64
}

type SkillActivityRepository struct {
	db *gorm.DB
}

func NewSkillActivityRepository(db *gorm.DB) *SkillActivityRepository {
	return &SkillActivityRepository{db: db}
}

func (r *SkillActivityRepository) BatchInsert(ctx context.Context, activities []schema.SkillActivity) (int64, error) {
	if len(activities) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&activities)
	if res.Error != nil {
		return 0, fmt.Errorf("写入技能活动失败: %w", res.Error)
	}
	return res.RowsAffected, nil
}

func (r *SkillActivityRepository) ListExistingKeys(ctx context.Context, keys []SkillActivityKey) (map[SkillActivityKey]struct{}, error) {
	existing := make(map[SkillActivityKey]struct{})
	if len(keys) == 0 {
		return existing, nil
	}

	q := r.db.WithContext(ctx).Model(&schema.SkillActivity{}).Select("source, evidence_id, skill_key")
	for i, k := range keys {
		if i == 0 {
			q = q.Where("(source = ? AND evidence_id = ? AND skill_key = ?)", k.Source, k.EvidenceID, k.SkillKey)
			continue
		}
		q = q.Or("(source = ? AND evidence_id = ? AND skill_key = ?)", k.Source, k.EvidenceID, k.SkillKey)
	}

	type row struct {
		Source     string
		EvidenceID int64
		SkillKey   string
	}
	var rows []row
	if err := q.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询技能活动失败: %w", err)
	}
	for _, r := range rows {
		existing[SkillActivityKey{Source: r.Source, EvidenceID: r.EvidenceID, SkillKey: r.SkillKey}] = struct{}{}
	}
	return existing, nil
}

func (r *SkillActivityRepository) GetStatsByTimeRange(ctx context.Context, startTime, endTime int64) ([]SkillActivityStat, error) {
	// timestamp 存的是 ms，需要 /1000 转秒；localtime 让“天”对齐本地时区。
	const sql = `
SELECT
  skill_key AS skill_key,
  COALESCE(SUM(exp), 0) AS exp_sum,
  COUNT(1) AS event_count,
  COUNT(DISTINCT strftime('%Y-%m-%d', timestamp/1000, 'unixepoch', 'localtime')) AS days_active,
  COALESCE(MAX(timestamp), 0) AS last_ts_milli
FROM skill_activities
WHERE timestamp >= ? AND timestamp <= ?
GROUP BY skill_key
`
	var out []SkillActivityStat
	if err := r.db.WithContext(ctx).Raw(sql, startTime, endTime).Scan(&out).Error; err != nil {
		return nil, fmt.Errorf("统计技能活动失败: %w", err)
	}
	return out, nil
}
