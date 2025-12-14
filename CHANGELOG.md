# Changelog

本文件记录版本变更。

## v0.2.0-alpha.1 - 2025-12-15

这是首个对外可用的 Alpha 版本：Windows 托盘常驻 Agent + 本地 Web UI，本地优先、默认隐私脱敏、可离线运行（未配置 AI Key 时自动降级为规则摘要）。

### 下载与运行（推荐）

- 从 GitHub Releases 下载：`WorkMirror-v0.2.0-alpha.1-windows-x64.zip`
- 解压到你会长期保留的位置（避免直接在“下载”目录内运行）
- 在解压后的目录内双击 `workmirror.exe`
- 首次运行会在同目录自动生成：`config/`、`data/`、`logs/`
- 迁移/备份：移动整个文件夹即可（不要只移动 exe）

### 快速上手（10 分钟闭环）

1. 启动后托盘 → “打开面板”
2. 进入“设置”：
   - 配置 `diff.watch_paths`（你的 Git 项目目录；未配置则不会采集代码变更）
   - （可选）关闭 `browser.enabled` 以禁用浏览历史采集
3. 回到“会话流/仪表盘”：确认当天会话与证据链可展开

### 你会得到什么

- Windows 托盘常驻 Agent + 本地 Web UI（本地优先）。
- 自动生成“工作片段”（会话切分/聚合/摘要），支持日报与周期回顾（日报/周报/趋势）。
- 证据链展开：窗口/代码 Diff/浏览（可按配置关闭浏览采集）。
- 离线降级：未配置 AI Key 时使用规则摘要；配置后可增强语义与建议。
- 状态页与诊断导出：异常不沉默，支持导出脱敏诊断包（zip）。

### 隐私与本地优先

- 数据默认写入本地 SQLite（默认 `./data/workmirror.db`）。
- 本地 HTTP 仅监听 `127.0.0.1`，端口随机分配；端口发现文件：`./data/http_base_url.txt`。
- 启用 AI 时仅发送脱敏后的最小必要文本，用于生成摘要与建议。

### 排障（Start Here）

- 状态快照：`GET /api/status`
- 导出诊断包（脱敏 zip）：`GET /api/diagnostics/export`

### 已知限制（Alpha）

- 仅支持 Windows 10/11；不提供云同步/账号体系/团队协作。
