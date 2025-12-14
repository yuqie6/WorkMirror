package buildinfo

// Version 在 Release 构建时通过 -ldflags 注入，例如：
// -X github.com/yuqie6/WorkMirror/internal/pkg/buildinfo.Version=v0.2.0-alpha.1
var Version = "v0.2.0-alpha.1"

// Commit 在 Release 构建时可选注入 git commit，例如：
// -X github.com/yuqie6/WorkMirror/internal/pkg/buildinfo.Commit=abcdef1
var Commit = "unknown"
