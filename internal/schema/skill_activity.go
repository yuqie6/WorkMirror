package schema

import "time"

// SkillActivity 记录一次“技能经验贡献”（用于趋势/证据链）
// 设计目标：最小必要字段，且可幂等（同一证据不会重复记账）。
type SkillActivity struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SkillKey   string    `gorm:"size:100;index;not null;uniqueIndex:uniq_skill_activity,priority:3" json:"skill_key"`
	Source     string    `gorm:"size:32;index;not null;uniqueIndex:uniq_skill_activity,priority:1" json:"source"` // diff/session/browser/other
	EvidenceID int64     `gorm:"index;not null;uniqueIndex:uniq_skill_activity,priority:2" json:"evidence_id"`
	Exp        float64   `gorm:"not null" json:"exp"`
	Timestamp  int64     `gorm:"index;not null" json:"timestamp"` // Unix ms
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (SkillActivity) TableName() string {
	return "skill_activities"
}
