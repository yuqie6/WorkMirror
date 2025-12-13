package schema

import (
	"time"
)

// Diff 代码变更记录
// 从文件变更中推断学习内容的核心数据源
type Diff struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Timestamp      int64     `gorm:"index" json:"timestamp"`             // 变更时间戳
	FilePath       string    `gorm:"size:500;index" json:"file_path"`    // 文件路径
	FileName       string    `gorm:"size:255" json:"file_name"`          // 文件名
	Language       string    `gorm:"size:50;index" json:"language"`      // 编程语言
	DiffContent    string    `gorm:"type:text" json:"diff_content"`      // Diff 内容
	LinesAdded     int       `gorm:"default:0" json:"lines_added"`       // 添加行数
	LinesDeleted   int       `gorm:"default:0" json:"lines_deleted"`     // 删除行数
	AIInsight      string    `gorm:"type:text" json:"ai_insight"`        // AI 解读
	SkillsDetected JSONArray `gorm:"type:text" json:"skills_detected"`   // 检测到的技能
	ProjectPath    string    `gorm:"size:500;index" json:"project_path"` // 项目根目录
	IsGitRepo      bool      `gorm:"default:false" json:"is_git_repo"`   // 是否是 Git 仓库
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (Diff) TableName() string {
	return "diffs"
}

// DailySummary 每日总结
type DailySummary struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Date         string    `gorm:"size:10;uniqueIndex" json:"date"` // YYYY-MM-DD
	Summary      string    `gorm:"type:text" json:"summary"`        // AI 生成的总结
	Highlights   string    `gorm:"type:text" json:"highlights"`     // 今日亮点
	Struggles    string    `gorm:"type:text" json:"struggles"`      // 今日困难
	SkillsGained JSONArray `gorm:"type:text" json:"skills_gained"`  // 获得的技能
	TotalCoding  int       `gorm:"default:0" json:"total_coding"`   // 编码时长 (分钟)
	TotalDiffs   int       `gorm:"default:0" json:"total_diffs"`    // Diff 数量
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (DailySummary) TableName() string {
	return "daily_summaries"
}

// BrowserEvent 浏览器事件 (Phase 2.2)
type BrowserEvent struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Timestamp int64     `gorm:"index" json:"timestamp"`
	URL       string    `gorm:"size:2000" json:"url"`
	Title     string    `gorm:"size:500" json:"title"`
	Domain    string    `gorm:"size:255;index" json:"domain"`
	Duration  int       `gorm:"default:0" json:"duration"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (BrowserEvent) TableName() string {
	return "browser_events"
}

// LanguageExtensions 文件扩展名到语言的映射
var LanguageExtensions = map[string]string{
	".go":    "Go",
	".py":    "Python",
	".js":    "JavaScript",
	".ts":    "TypeScript",
	".jsx":   "React",
	".tsx":   "React",
	".vue":   "Vue",
	".java":  "Java",
	".c":     "C",
	".cpp":   "C++",
	".h":     "C/C++",
	".rs":    "Rust",
	".rb":    "Ruby",
	".php":   "PHP",
	".swift": "Swift",
	".kt":    "Kotlin",
	".scala": "Scala",
	".cs":    "C#",
	".lua":   "Lua",
	".sql":   "SQL",
	".sh":    "Shell",
	".ps1":   "PowerShell",
	".yaml":  "YAML",
	".yml":   "YAML",
	".json":  "JSON",
	".xml":   "XML",
	".html":  "HTML",
	".css":   "CSS",
	".scss":  "SCSS",
	".less":  "LESS",
	".md":    "Markdown",
}

// GetLanguageFromExt 根据扩展名获取语言
func GetLanguageFromExt(ext string) string {
	if lang, ok := LanguageExtensions[ext]; ok {
		return lang
	}
	return "Unknown"
}
