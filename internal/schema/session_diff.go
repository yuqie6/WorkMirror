package schema

import "time"

// SessionDiff 会话与 Diff 的关联（Phase B 证据链）
type SessionDiff struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	SessionID int64     `gorm:"index;not null"`
	DiffID    int64     `gorm:"index;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (SessionDiff) TableName() string {
	return "session_diffs"
}
