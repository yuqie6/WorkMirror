package service

import (
	"math"
	"testing"
	"time"
)

func TestCalcSkillExpToNext(t *testing.T) {
	if got := calcSkillExpToNext(1); got != skillBaseLevelUpCost*skillLevelMultiplier {
		t.Fatalf("expToNext level1 = %v, want %v", got, skillBaseLevelUpCost*skillLevelMultiplier)
	}
	if got := calcSkillExpToNext(2); got != skillBaseLevelUpCost*math.Pow(skillLevelMultiplier, 2) {
		t.Fatalf("expToNext level2 = %v, want %v", got, skillBaseLevelUpCost*math.Pow(skillLevelMultiplier, 2))
	}
}

func TestAddSkillExpLevelsUp(t *testing.T) {
	s := NewSkillNode("go", "Go", "language")
	AddSkillExp(s, 300) // 足够升两级
	if s.Level <= 1 {
		t.Fatalf("level = %d, want > 1", s.Level)
	}
	if s.Exp < 0 || s.Exp >= s.ExpToNext {
		t.Fatalf("exp=%v expToNext=%v not in range", s.Exp, s.ExpToNext)
	}
	if s.LastActive == 0 {
		t.Fatalf("LastActive not set")
	}
}

func TestSkillDaysInactive(t *testing.T) {
	s := NewSkillNode("go", "Go", "language")
	if got := SkillDaysInactive(s); got != 0 {
		t.Fatalf("daysInactive with 0 lastActive = %d, want 0", got)
	}
	s.LastActive = time.Now().Add(-48 * time.Hour).UnixMilli()
	if got := SkillDaysInactive(s); got < 2 {
		t.Fatalf("daysInactive = %d, want >=2", got)
	}
}

func TestApplySkillDecayBounds(t *testing.T) {
	s := NewSkillNode("go", "Go", "language")
	s.Exp = 100
	s.LastActive = time.Now().Add(-10 * 24 * time.Hour).UnixMilli() // 10 天不活跃
	ApplySkillDecay(s)
	if s.Exp >= 100 {
		t.Fatalf("exp not decayed: %v", s.Exp)
	}

	s.Exp = 100
	s.LastActive = time.Now().Add(-100 * 24 * time.Hour).UnixMilli()
	ApplySkillDecay(s)
	if s.Exp != 50 { // 最大衰减 50%
		t.Fatalf("exp=%v, want 50", s.Exp)
	}
}
