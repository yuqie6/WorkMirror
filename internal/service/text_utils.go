package service

// truncateRunes 按 rune 数量截断字符串
// 正确处理 Unicode 字符，超过 max 长度时添加省略号
func truncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
