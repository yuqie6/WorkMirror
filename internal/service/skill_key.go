package service

import "strings"

// normalizeKey 统一 Key 格式（稳定 slug 策略）
func normalizeKey(name string) string {
	if name == "" {
		return ""
	}
	key := strings.ToLower(strings.TrimSpace(name))

	// 常见特殊符号处理（在空格转换之前，按固定顺序替换，避免 map 遍历导致不稳定）
	orderedReplacements := []struct {
		old string
		new string
	}{
		// 更具体的后缀优先
		{"react.js", "reactjs"},
		{"vue.js", "vuejs"},
		{"next.js", "nextjs"},
		{"node.js", "nodejs"},
		// 语言/平台别名
		{"c++", "cpp"},
		{"c#", "csharp"},
		{".net", "dotnet"},
		// 通用后缀最后处理，避免误伤上面的替换结果
		{".js", "-js"},
		{".ts", "-ts"},
	}
	for _, rep := range orderedReplacements {
		key = strings.ReplaceAll(key, rep.old, rep.new)
	}

	// 空格转连字符
	key = strings.ReplaceAll(key, " ", "-")

	// 移除其他特殊字符（保留字母、数字、连字符）
	var result strings.Builder
	for _, r := range key {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
