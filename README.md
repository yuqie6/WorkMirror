中文 | [English](README_EN.md)

# 复盘镜（WorkMirror）

<p align="center">
  <img src="assets/icons/app-icon.svg" width="128" height="128" alt="WorkMirror Logo">
</p>

每天自动复盘你的真实工作，一键生成可追溯证据的日报与周报（本地优先 / Windows 托盘）

复盘镜（WorkMirror）会在后台记录你电脑上的工作痕迹（窗口切换、代码变更、浏览历史），自动整理成“工作片段”，生成日报/周期回顾，并且每条结论都能展开看到来源证据（默认本地、可离线、无需手写日志）。

## 功能特性

- **自动日报/复盘**：把一天拆成工作片段并生成摘要；支持周期聚合（日报/周报/趋势）。
- **可回溯证据链**：每条结论都能一键展开到来源覆盖率与证据明细（Diff/窗口/浏览）。
- **技能树与趋势**：基于工作片段做技能归因与趋势展示（口径统一、可解释）。
- **离线可用**：不配置 AI Key 也能生成规则版摘要；配置 AI 后可选增强语义与建议。
- **隐私默认开启**：数据本地存储；URL/标题可脱敏；本地 HTTP 仅监听 `127.0.0.1`。
- **拒绝沉默失败**：内置状态页与诊断包导出，空态/异常会给出可操作原因与修复建议。

## 不适用场景（当前阶段）

- 需要云同步/账号体系/团队协作
- 主要在 macOS/Linux 工作流（当前仅支持 Windows 10/11）
- 希望它替代 TODO/OKR/习惯管理

## 技术栈

- **后端**: Go 1.25.4
- **存储**: SQLite (WAL 模式)
- **AI**: DeepSeek (LLM) + SiliconFlow (Embedding/Reranker)

## 快速开始

推荐直接下载 Releases：<https://github.com/yuqie6/WorkMirror/releases>

## 下载与运行（Windows）

推荐 zip “便携目录分发”：解压后得到一个固定目录（例如 `WorkMirror/`），在该目录内运行 `workmirror.exe`。
首次启动会自动生成 `./config/`、`./data/`、`./logs/`；迁移/备份时请移动整个目录，避免只移动 exe 导致“数据丢在下载目录”。

运行后会最小化到系统托盘：右键图标 → “打开面板”。也可读取 `.\data\http_base_url.txt` 在浏览器打开本地 UI。

## 运行与数据位置

- 运行：双击 `workmirror.exe`
- UI：托盘 → 打开面板；或读取 `.\data\http_base_url.txt` 后在浏览器打开（例如 `http://127.0.0.1:12345/`）
- 默认数据库：`.\data\workmirror.db`
- 端口发现：`.\data\http_base_url.txt`

## 配置（YAML）

配置文件默认在 `.\config\config.yaml`，模板见 `config/config.yaml.example`。

你最需要关注的是 Diff 监控目录（watch paths）：不配置时不会采集代码变更。
建议配置为 Git 项目目录（非 Git 目录会被跳过）。

```yaml
diff:
  enabled: true
  watch_paths:
    - "C:\\Users\\Dev\\Projects\\MyRepo"
```

AI Key 建议通过环境变量注入（避免写入磁盘），例如 `DEEPSEEK_API_KEY`、`SILICONFLOW_API_KEY`（详见 `config/config.yaml.example`）。
如不希望采集浏览历史，可在配置中关闭 `browser.enabled`。

## 开发者

从源码构建/前端开发/打包脚本见 `docs/development.md`。

## 排障（推荐从这里开始）

- 状态快照：`GET /api/status`
- 导出诊断包（脱敏 zip）：`GET /api/diagnostics/export`
- 会话维护：`POST /api/maintenance/sessions/rebuild`、`POST /api/maintenance/sessions/enrich`

以上接口的 base URL 来自 `.\data\http_base_url.txt`（服务仅监听 `127.0.0.1`）。

## 项目结构

```
├── cmd/workmirror-agent/    # 主程序入口
├── config/              # 配置文件
├── internal/
│   ├── collector/       # 采集器（Win32 API / Diff / Browser History）
│   ├── bootstrap/       # 依赖组装/运行时构建
│   ├── server/          # 本地 HTTP Server（JSON API + SSE + 静态 UI）
│   ├── observability/   # status/diagnostics（拒绝沉默失败）
│   ├── service/         # 业务逻辑层
│   ├── repository/      # 数据访问层
│   ├── model/           # 数据模型
│   └── pkg/             # 内部工具包
└── data/                # 运行时数据
```

## License

MIT
