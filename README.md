中文 | [English](README_EN.md)

# 复盘镜（WorkMirror）

<p align="center">
  <img src="assets/icons/app-icon.svg" width="128" height="128" alt="WorkMirror Logo">
</p>

每天自动记录你的真实工作，一键生成可回溯的日报与周报。

## 它能做什么？

复盘镜会在后台默默记录你用了哪些软件、改了哪些代码、看了哪些网页，然后自动帮你整理成工作记录。每条记录都可以点开查看是基于哪些原始数据生成的。

- **自动生成日报/周报**：把一天的工作自动拆分成多个片段，生成摘要
- **记录可追溯**：每条总结都能展开看到来源（代码修改、网页、应用使用）
- **技能成长追踪**：根据你的工作内容自动归纳技能点和成长趋势
- **离线可用**：不配置 AI 也能使用，会用自动规则生成总结
- **隐私优先**：数据存在本地，网址和标题可以自动脱敏，只监听本机

## 不适合的场景

- 需要云同步、账号体系、团队协作
- 主要在 macOS/Linux 上工作（目前只支持 Windows 10/11）
- 想用它管理待办、OKR、习惯打卡

## 快速开始

### 下载

推荐从 [Releases](https://github.com/yuqie6/WorkMirror/releases) 下载最新版本。

### 安装与运行

1. 解压 zip 包到一个固定位置（比如 `D:\WorkMirror\`）
2. 双击 `workmirror.exe` 运行
3. 程序会最小化到系统托盘，右键托盘图标 → 「打开面板」即可使用

**提示**：首次启动会自动创建 `config/`、`data/`、`logs/` 文件夹。迁移时请移动整个目录，不要只移动 exe 文件。

### 首次配置

打开界面后，建议先去「设置」页面配置**代码监控目录**（你写代码的文件夹路径）。不配置的话，系统无法记录你的代码修改。

## 遇到问题？

1. **打开「运行状态」页**：查看各项后台服务是否正常运行
2. **某天记录缺失**：在「运行状态」页选择日期，点击「生成工作记录」或「重新整理记录」
3. **需要技术支持**：点击「导出诊断包」，将生成的 zip 文件发给我们

## 配置说明

配置文件在 `config/config.yaml`，通常不需要手动编辑。你可以在界面的「设置」页完成所有配置。

### 常用配置项

**代码监控目录**：设置你写代码的文件夹，支持添加多个
```yaml
diff:
  enabled: true
  watch_paths:
    - "D:\\Projects\\MyProject"
    - "D:\\Work\\AnotherRepo"
```

**AI 服务**：默认使用内置免费服务，也可以配置自己的 API Key
```yaml
ai:
  provider: default  # 可选: openai, anthropic, google, zhipu
```

**隐私保护**：默认开启，会自动隐藏网址中的敏感参数
```yaml
privacy:
  enabled: true
```

## 技术信息

- **后端**: Go 1.25.4
- **存储**: SQLite (WAL 模式)
- **AI**: 内置免费 LLM / OpenAI 兼容 / Anthropic / Google / 智谱

## 开发者

从源码构建、前端开发、打包脚本等信息请查看 `docs/development.md`。

## 许可证

MIT
