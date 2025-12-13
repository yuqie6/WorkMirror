package schema

import (
	"time"
)

// Diff 代码变更记录
// 从文件变更中推断学习内容的核心数据源
type Diff struct {
	ID             int64     `gorm:"primaryKey;autoIncrement"`
	Timestamp      int64     `gorm:"index"`           // 变更时间戳
	FilePath       string    `gorm:"size:500;index"`  // 文件路径
	FileName       string    `gorm:"size:255"`        // 文件名
	Language       string    `gorm:"size:50;index"`   // 编程语言
	DiffContent    string    `gorm:"type:text"`       // Diff 内容
	LinesAdded     int       `gorm:"default:0"`       // 添加行数
	LinesDeleted   int       `gorm:"default:0"`       // 删除行数
	AIInsight      string    `gorm:"type:text"`       // AI 解读
	SkillsDetected JSONArray `gorm:"type:text"`       // 检测到的技能
	ProjectPath    string    `gorm:"size:500;index"`  // 项目根目录
	IsGitRepo      bool      `gorm:"default:false"`   // 是否是 Git 仓库
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (Diff) TableName() string {
	return "diffs"
}

// DailySummary 每日总结
type DailySummary struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	Date         string    `gorm:"size:10;uniqueIndex"` // YYYY-MM-DD
	Summary      string    `gorm:"type:text"`           // AI 生成的总结
	Highlights   string    `gorm:"type:text"`           // 今日亮点
	Struggles    string    `gorm:"type:text"`           // 今日困难
	SkillsGained JSONArray `gorm:"type:text"`           // 获得的技能
	TotalCoding  int       `gorm:"default:0"`           // 编码时长 (分钟)
	TotalDiffs   int       `gorm:"default:0"`           // Diff 数量
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (DailySummary) TableName() string {
	return "daily_summaries"
}

// BrowserEvent 浏览器事件 (Phase 2.2)
type BrowserEvent struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Timestamp int64     `gorm:"index"`
	URL       string    `gorm:"size:2000"`
	Title     string    `gorm:"size:500"`
	Domain    string    `gorm:"size:255;index"`
	Duration  int       `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (BrowserEvent) TableName() string {
	return "browser_events"
}
