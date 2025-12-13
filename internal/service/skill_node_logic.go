package service

import (
	"time"

	"github.com/yuqie6/mirror/internal/schema"
)

const (
	skillBaseLevelUpCost = 100.0
	skillLevelMultiplier = 1.5
)

func NewSkillNode(key, name, category string) *schema.SkillNode {
	return &schema.SkillNode{
		Key:       key,
		Name:      name,
		Category:  category,
		Level:     1,
		Exp:       0,
		ExpToNext: skillBaseLevelUpCost,
	}
}

func AddSkillExp(skill *schema.SkillNode, exp float64) {
	if skill == nil || exp <= 0 {
		return
	}

	skill.Exp += exp
	skill.LastActive = time.Now().UnixMilli()

	for skill.Exp >= skill.ExpToNext && skill.Level < 99 {
		skill.Exp -= skill.ExpToNext
		skill.Level++
		skill.ExpToNext = calcSkillExpToNext(skill.Level)
	}
}

func SkillDaysInactive(skill *schema.SkillNode) int {
	if skill == nil || skill.LastActive == 0 {
		return 0
	}
	last := time.UnixMilli(skill.LastActive)
	return int(time.Since(last).Hours() / 24)
}

func ApplySkillDecay(skill *schema.SkillNode) {
	if skill == nil {
		return
	}
	daysInactive := SkillDaysInactive(skill)
	if daysInactive <= 7 {
		return
	}

	decayDays := daysInactive - 7
	decayRate := 0.02 * float64(decayDays)
	if decayRate > 0.5 {
		decayRate = 0.5
	}
	skill.Exp *= (1 - decayRate)
}

func calcSkillExpToNext(level int) float64 {
	if level <= 0 {
		level = 1
	}
	multiplier := 1.0
	for i := 0; i < level; i++ {
		multiplier *= skillLevelMultiplier
	}
	return skillBaseLevelUpCost * multiplier
}

