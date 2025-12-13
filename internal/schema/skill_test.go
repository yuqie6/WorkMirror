package schema

import (
	"math"
	"testing"
	"time"
)

func TestSkillNodeCalculateExpToNext(t *testing.T) {
	s := NewSkillNode("go", "Go", "language")
	if got := s.CalculateExpToNext(); got != BaseLevelUpCost*LevelMultiplier {
		t.Fatalf("expToNext level1 = %v, want %v", got, BaseLevelUpCost*LevelMultiplier)
	}
	s.Level = 2
	if got := s.CalculateExpToNext(); got != BaseLevelUpCost*math.Pow(LevelMultiplier, 2) {
		t.Fatalf("expToNext level2 = %v, want %v", got, BaseLevelUpCost*math.Pow(LevelMultiplier, 2))
	}
}

func TestSkillNodeAddExpLevelsUp(t *testing.T) {
	s := NewSkillNode("go", "Go", "language")
	s.AddExp(300) // 足够升两级
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

func TestSkillNodeDaysInactive(t *testing.T) {
	s := NewSkillNode("go", "Go", "language")
	if got := s.DaysInactive(); got != 0 {
		t.Fatalf("daysInactive with 0 lastActive = %d, want 0", got)
	}
	s.LastActive = time.Now().Add(-48 * time.Hour).UnixMilli()
	if got := s.DaysInactive(); got < 2 {
		t.Fatalf("daysInactive = %d, want >=2", got)
	}
}

func TestSkillNodeApplyDecayBounds(t *testing.T) {
	s := NewSkillNode("go", "Go", "language")
	s.Exp = 100
	s.LastActive = time.Now().Add(-10 * 24 * time.Hour).UnixMilli() // 10 天不活跃
	s.ApplyDecay()
	if s.Exp >= 100 {
		t.Fatalf("exp not decayed: %v", s.Exp)
	}

	s.Exp = 100
	s.LastActive = time.Now().Add(-100 * 24 * time.Hour).UnixMilli()
	s.ApplyDecay()
	if s.Exp != 50 { // 最大衰减 50%
		t.Fatalf("exp=%v, want 50", s.Exp)
	}
}
