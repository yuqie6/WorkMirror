# Project Mirror

智能个人行为量化与成长归因系统 - 本地优先的自动化个人成长 Agent

## 功能特性

- **无感采集**: 24/7 记录操作系统的行为流（窗口、代码提交、浏览器行为）
- **语义理解**: 利用 LLM 将日志转化为语义化的"工作/学习会话"
- **能力建模**: 自动将行为归因到技能树，可视化能力增长
- **隐私安全**: 数据 100% 本地存储

## 技术栈

- **后端**: Go 1.25.4
- **存储**: SQLite (WAL 模式)
- **AI**: DeepSeek (LLM) + SiliconFlow (Embedding/Reranker)

## 快速开始

```bash
# 构建
go build -o mirror.exe ./cmd/mirror-agent/

# 运行
./mirror.exe
```

## 构建与分发（Windows）

推荐“便携分发”（一个 `mirror.exe` + 同目录 `config/` 与 `data/`）。Agent 首次启动会自动生成 `config/config.yaml` 与 `data/` 目录。

在 PowerShell 执行：

```powershell
go build -trimpath -ldflags "-H=windowsgui -s -w" -o .\mirror.exe .\cmd\mirror-agent\
```

说明：Agent 启动后托盘菜单“打开面板”会以 app-mode 窗口打开本地 UI（由 Agent 内置并服务）。

## 前端开发（UI）

前端源码位于 `frontend/`，开发调试建议：

```powershell
# 启动 agent 后，设置 VITE_API_TARGET 指向 agent 的本地地址（端口为自动分配）
$env:VITE_API_TARGET="http://127.0.0.1:<port>"
pnpm dev
```

发布时将前端构建产物（`dist/`）写入 `internal/uiassets/dist/` 后重新编译 agent，即可将 UI 静态资源内置到单个二进制中。

## 项目结构

```
├── cmd/mirror-agent/    # 主程序入口
├── config/              # 配置文件
├── internal/
│   ├── handler/         # 采集层 (Win32 API)
│   ├── bootstrap/       # 依赖组装/运行时构建
│   ├── service/         # 业务逻辑层
│   ├── repository/      # 数据访问层
│   ├── model/           # 数据模型
│   └── pkg/             # 内部工具包
└── data/                # 运行时数据
```

## License

MIT
