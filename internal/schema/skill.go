package schema

import (
	"time"
)

// SkillNode 技能树节点
// 数据量级：百级
type SkillNode struct {
	Key        string    `gorm:"primaryKey;size:100" json:"key"`   // 唯一标识: go, gin, react
	Name       string    `gorm:"size:100" json:"name"`             // 显示名: Go, Gin, React
	Category   string    `gorm:"size:50;index" json:"category"`    // 分类: language, framework, database, devops, tool, concept, other
	ParentKey  string    `gorm:"size:100;index" json:"parent_key"` // 父技能 Key（AI 决定），如 gin → go
	Level      int       `gorm:"default:1" json:"level"`           // 当前等级: 1-99
	Exp        float64   `gorm:"default:0" json:"exp"`             // 当前经验值
	ExpToNext  float64   `gorm:"default:100" json:"exp_to_next"`   // 升级所需经验
	LastActive int64     `gorm:"index" json:"last_active"`         // 最后一次获得经验的时间
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (SkillNode) TableName() string {
	return "skill_nodes"
}

// BaseLevelUpCost 基础升级经验
const BaseLevelUpCost = 100.0

// LevelMultiplier 等级经验倍率
const LevelMultiplier = 1.5

// CalculateExpToNext 计算升级所需经验
// 公式: BaseCost × (1.5)^CurrentLevel
func (s *SkillNode) CalculateExpToNext() float64 {
	multiplier := 1.0
	for i := 0; i < s.Level; i++ {
		multiplier *= LevelMultiplier
	}
	return BaseLevelUpCost * multiplier
}

// AddExp 增加经验值，自动处理升级
func (s *SkillNode) AddExp(exp float64) {
	s.Exp += exp
	s.LastActive = time.Now().UnixMilli()

	// 检查是否升级
	for s.Exp >= s.ExpToNext && s.Level < 99 {
		s.Exp -= s.ExpToNext
		s.Level++
		s.ExpToNext = s.CalculateExpToNext()
	}
}

// DaysInactive 返回不活跃天数
func (s *SkillNode) DaysInactive() int {
	if s.LastActive == 0 {
		return 0
	}
	lastActiveTime := time.UnixMilli(s.LastActive)
	return int(time.Since(lastActiveTime).Hours() / 24)
}

// ApplyDecay 应用遗忘衰减
// 若超过 7 天不活跃，每日扣除 2% 当前经验值
func (s *SkillNode) ApplyDecay() {
	daysInactive := s.DaysInactive()
	if daysInactive > 7 {
		decayDays := daysInactive - 7
		decayRate := 0.02 * float64(decayDays)
		if decayRate > 0.5 { // 最多扣除 50%
			decayRate = 0.5
		}
		s.Exp *= (1 - decayRate)
	}
}

// NewSkillNode 创建新的技能节点
func NewSkillNode(key, name, category string) *SkillNode {
	node := &SkillNode{
		Key:       key,
		Name:      name,
		Category:  category,
		Level:     1,
		Exp:       0,
		ExpToNext: BaseLevelUpCost,
	}
	return node
}
