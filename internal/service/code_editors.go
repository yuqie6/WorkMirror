package service

import "strings"

// DefaultCodeEditors 默认代码编辑器列表（单一来源）
// 注意：全部使用小写，匹配时进行大小写不敏感比较
var DefaultCodeEditors = []string{
	"code.exe", "cursor.exe", "antigravity.exe",
	"idea64.exe", "idea.exe", "goland64.exe", "goland.exe",
	"pycharm64.exe", "pycharm.exe", "webstorm64.exe", "webstorm.exe",
	"devenv.exe", "zed.exe", "fleet.exe",
	"sublime_text.exe", "notepad++.exe",
	"vim.exe", "nvim.exe", "emacs.exe",
}

// IsCodeEditor 判断是否是代码编辑器（大小写不敏感）
func IsCodeEditor(appName string) bool {
	lower := strings.ToLower(appName)
	for _, editor := range DefaultCodeEditors {
		if lower == editor {
			return true
		}
	}
	return false
}
