package schema

import "time"

// SessionDiff 会话与 Diff 的关联（Phase B 证据链）
type SessionDiff struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID int64     `gorm:"index;not null" json:"session_id"`
	DiffID    int64     `gorm:"index;not null" json:"diff_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (SessionDiff) TableName() string {
	return "session_diffs"
}
