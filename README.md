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

## 打包安装包（Windows）

前置：Go 1.25.x、Node 16+、Wails v2、Inno Setup 6（提供 `ISCC.exe`）。

在 PowerShell 执行：

```powershell
.\scripts\package-windows.ps1
```

产物输出到 `build/installer/windows/out/`（可用 `-Version` / `-OutDir` 覆盖）。

说明：安装后以 UI 为主入口，UI 启动时会自动拉起后台 Agent（托盘常驻，避免重复启动已做单实例保护）。

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
