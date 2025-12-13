package collector

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

