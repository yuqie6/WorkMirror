package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Event 原始事件 - 记录用户的窗口活动
// 数据量级：千万级/年
type Event struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Timestamp int64     `gorm:"index" json:"timestamp"`            // Unix 时间戳 (毫秒)
	Source    string    `gorm:"size:50" json:"source"`             // 来源: window, chrome, vscode
	AppName   string    `gorm:"size:255;index" json:"app_name"`    // 应用名: Chrome.exe
	Title     string    `gorm:"type:text" json:"title"`            // 窗口标题 (已脱敏)
	Duration  int       `gorm:"default:0" json:"duration"`         // 持续时长 (秒)
	Metadata  JSONMap   `gorm:"type:text" json:"metadata"`         // 扩展字段 (git branch, url)
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (Event) TableName() string {
	return "events"
}

// NewEvent 创建新事件
func NewEvent(source, appName, title string) *Event {
	return &Event{
		Timestamp: time.Now().UnixMilli(),
		Source:    source,
		AppName:   appName,
		Title:     title,
		Duration:  0,
		Metadata:  make(JSONMap),
	}
}

// JSONMap 用于存储 JSON 格式的元数据
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid type for JSONMap")
	}

	return json.Unmarshal(bytes, j)
}
