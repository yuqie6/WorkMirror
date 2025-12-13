package schema

import (
	"time"
)

// SkillNode 技能树节点
// 数据量级：百级
type SkillNode struct {
	Key        string    `gorm:"primaryKey;size:100"`   // 唯一标识: go, gin, react
	Name       string    `gorm:"size:100"`              // 显示名: Go, Gin, React
	Category   string    `gorm:"size:50;index"`         // 分类: language, framework, database, devops, tool, concept, other
	ParentKey  string    `gorm:"size:100;index"`        // 父技能 Key（AI 决定），如 gin → go
	Level      int       `gorm:"default:1"`             // 当前等级: 1-99
	Exp        float64   `gorm:"default:0"`             // 当前经验值
	ExpToNext  float64   `gorm:"default:100"`           // 升级所需经验
	LastActive int64     `gorm:"index"`                 // 最后一次获得经验的时间
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (SkillNode) TableName() string {
	return "skill_nodes"
}
