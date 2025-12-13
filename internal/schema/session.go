package schema

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Session AI 生成的智能会话
// 数据量级：千级/年
type Session struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Date           string    `gorm:"size:10;index" json:"date"`          // YYYY-MM-DD（冗余字段，便于按天查询）
	StartTime      int64     `gorm:"index" json:"start_time"`            // Unix 时间戳（毫秒）
	EndTime        int64     `gorm:"index" json:"end_time"`              // Unix 时间戳（毫秒）
	PrimaryApp     string    `gorm:"size:255;index" json:"primary_app"`  // 主应用（时长最大）
	SessionVersion int       `gorm:"default:1" json:"session_version"`   // 切分规则版本号
	TimeRange      string    `gorm:"size:20" json:"time_range"`          // 兼容旧字段："14:00-16:00"
	Category       string    `gorm:"size:50" json:"category"`            // 兼容旧字段：Coding, Reading, Meeting
	Summary        string    `gorm:"type:text" json:"summary"`           // AI 生成的该时段行为总结
	SkillsInvolved JSONArray `gorm:"type:text" json:"skills_involved"`   // 涉及技能 ["Go", "Redis"]
	EmbeddingID    string    `gorm:"size:100;index" json:"embedding_id"` // 向量存储 ID
	Metadata       JSONMap   `gorm:"type:text" json:"metadata"`          // 结构化上下文与证据关联（JSON）
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// SessionSemanticUpdate 会话语义字段更新（用于部分字段更新）
type SessionSemanticUpdate struct {
	TimeRange      string
	Category       string
	Summary        string
	SkillsInvolved []string
	Metadata       JSONMap
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
