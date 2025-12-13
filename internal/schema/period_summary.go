package schema

import "time"

// PeriodSummary 阶段汇总（周/月）
type PeriodSummary struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Type         string    `gorm:"size:10;index;uniqueIndex:uniq_period_range" json:"type"` // "week" | "month"
	StartDate    string    `gorm:"size:10;uniqueIndex:uniq_period_range" json:"start_date"` // YYYY-MM-DD
	EndDate      string    `gorm:"size:10;uniqueIndex:uniq_period_range" json:"end_date"`   // YYYY-MM-DD
	Overview     string    `gorm:"type:text" json:"overview"`                               // AI 生成的概述
	Achievements JSONArray `gorm:"type:text" json:"achievements"`                           // 成就列表
	Patterns     string    `gorm:"type:text" json:"patterns"`                               // 模式分析
	Suggestions  string    `gorm:"type:text" json:"suggestions"`                            // 建议
	TopSkills    JSONArray `gorm:"type:text" json:"top_skills"`                             // 重点技能
	TotalCoding  int       `gorm:"default:0" json:"total_coding"`                           // 总编码时长（分钟）
	TotalDiffs   int       `gorm:"default:0" json:"total_diffs"`                            // 总 Diff 数量
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PeriodSummary) TableName() string {
	return "period_summaries"
}
