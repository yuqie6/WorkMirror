package service

import "strings"

// SkillCategory 技能类别
type SkillCategory string

const (
	CategoryLanguage  SkillCategory = "language"  // 编程语言
	CategoryFramework SkillCategory = "framework" // 框架
	CategoryDatabase  SkillCategory = "database"  // 数据库
	CategoryDevOps    SkillCategory = "devops"    // 运维工具
	CategoryTool      SkillCategory = "tool"      // 工具
	CategoryConcept   SkillCategory = "concept"   // 概念/模式
	CategoryOther     SkillCategory = "other"     // 其他
)

// SkillCategoryInfo 技能类别信息
type SkillCategoryInfo struct {
	Name        string
	DisplayName string
	Icon        string
	Priority    int // 显示优先级, 越小越靠前
}

// SkillCategories 所有类别信息（用于 UI 显示）
var SkillCategories = map[SkillCategory]SkillCategoryInfo{
	CategoryLanguage:  {Name: "language", DisplayName: "编程语言", Icon: "💻", Priority: 1},
	CategoryFramework: {Name: "framework", DisplayName: "框架", Icon: "🏗️", Priority: 2},
	CategoryDatabase:  {Name: "database", DisplayName: "数据库", Icon: "🗄️", Priority: 3},
	CategoryDevOps:    {Name: "devops", DisplayName: "DevOps", Icon: "⚙️", Priority: 4},
	CategoryTool:      {Name: "tool", DisplayName: "工具", Icon: "🔧", Priority: 5},
	CategoryConcept:   {Name: "concept", DisplayName: "概念", Icon: "💡", Priority: 6},
	CategoryOther:     {Name: "other", DisplayName: "其他", Icon: "📦", Priority: 7},
}

// GetSkillCategory 获取技能类别（备用，AI 优先决定分类）
func GetSkillCategory(skillName string) SkillCategory {
	// AI 已决定分类时不会调用这里
	// 这只是降级方案
	return CategoryOther
}

// NormalizeSkillName 标准化技能名称
func NormalizeSkillName(skillName string) string {
	s := strings.TrimSpace(skillName)
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}
