package uiassets

import (
	"embed"
	"io/fs"
)

// dist 目录由前端构建产物填充；用于将 UI 静态资源内置到 agent 二进制中。
//
//go:embed all:dist
var embedded embed.FS

func FS() fs.FS {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		return embedded
	}
	return sub
}
