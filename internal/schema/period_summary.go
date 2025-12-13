package schema

import "time"

// PeriodSummary 阶段汇总（周/月）
type PeriodSummary struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	Type         string    `gorm:"size:10;index;uniqueIndex:uniq_period_range"` // "week" | "month"
	StartDate    string    `gorm:"size:10;uniqueIndex:uniq_period_range"`       // YYYY-MM-DD
	EndDate      string    `gorm:"size:10;uniqueIndex:uniq_period_range"`       // YYYY-MM-DD
	Overview     string    `gorm:"type:text"`                                    // AI 生成的概述
	Achievements JSONArray `gorm:"type:text"`                                    // 成就列表
	Patterns     string    `gorm:"type:text"`                                    // 模式分析
	Suggestions  string    `gorm:"type:text"`                                    // 建议
	TopSkills    JSONArray `gorm:"type:text"`                                    // 重点技能
	TotalCoding  int       `gorm:"default:0"`                                    // 总编码时长（分钟）
	TotalDiffs   int       `gorm:"default:0"`                                    // 总 Diff 数量
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (PeriodSummary) TableName() string {
	return "period_summaries"
}
