package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Session AI 生成的智能会话
// 数据量级：千级/年
type Session struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Date           string    `gorm:"size:10;index" json:"date"`          // YYYY-MM-DD
	TimeRange      string    `gorm:"size:20" json:"time_range"`          // "14:00-16:00"
	Category       string    `gorm:"size:50" json:"category"`            // Coding, Reading, Meeting
	Summary        string    `gorm:"type:text" json:"summary"`           // AI 生成的该时段行为总结
	SkillsInvolved JSONArray `gorm:"type:text" json:"skills_involved"`   // 涉及技能 ["Go", "Redis"]
	EmbeddingID    string    `gorm:"size:100;index" json:"embedding_id"` // 向量存储 ID
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Session) TableName() string {
	return "sessions"
}

// JSONArray 用于存储 JSON 数组
type JSONArray []string

// Value 实现 driver.Valuer 接口
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONArray, 0)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*j = make(JSONArray, 0)
		return nil
	}

	return json.Unmarshal(bytes, j)
}
