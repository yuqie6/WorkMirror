• # Project Mirror PRD v0.2（产品化闭环版：Windows 托盘 Agent + 本地 Web UI；无 Wails/无 CLI）

## 0. 文档目的

把当前“能跑的个人工具”收敛为可长期交付的产品：对非作者用户也能安装即用、结果可解释、失败可诊断、可持续升级不炸数据。

### 0.1 文档关系与口径（UI）

- PRD 负责“要做什么/验收口径/优先级”，不在这里堆砌全部 UI 细节。
- Web UI 的交互与视觉规范以 `docs/frontend.md` 为准（工程师驾驶舱：Evidence First + No Silent Failures）。
- 若 PRD 与 `docs/frontend.md` 对 UI 细节冲突：以 PRD 的产品原则/验收为最终口径；实现细节以 `docs/frontend.md` 落地。

---

## 1. 一句话定位（对外）

Project Mirror：自动把你每天/每周的真实工作学习行为整理成可追溯的证据链回顾，并给出下一步建议（默认本地、可离线）。

---

## 2. 目标用户与非目标用户

### 2.1 目标用户（Phase 1）

- Windows 10/11 上的工程师/技术学习者

### 2.2 非目标用户（明确不支持）

- 需要多端同步/云账号/团队协作的人
- 主要在 macOS/Linux 工作流（暂不支持）
- 强依赖移动端记录、或希望做任务管理/OKR 的人

## 3. 产品原则（必须落到实现与 UI）

    1. Local-first & Privacy by default：数据本地；敏感项默认最小化/可脱敏
    2. 证据链优先于结论：任何结论必须可点开看到来源覆盖率
    3. 失败可见且可操作：空态/异常/降级必须解释原因并给出下一步
    4. Session 为权威路径：所有上层输出统一口径，避免多套“都能跑但不一致”
    5. 离线可用：无 AI Key 时仍能产出“规则回顾”，且 UI 明确标注能力差异

### 3.1 UI/UX 设计约束（v0.2 必须落地）

来自 `docs/frontend.md` 的强约束（写入 PRD 作为验收口径）：

- 工程师驾驶舱：高信息密度、暗色优先（Tailwind zinc 体系）、语义色统一（Healthy/Warn/Error/AI）。
- Evidence First：任何 summary/洞察/图表都必须在 1 次交互内 drill-down 到证据（Diff/Events/Browser）。
- No Silent Failures：全局可见的系统健康（侧边栏常驻 + 状态页），离线/降级必须“一键可解释”（点击健康指示器直达诊断页，展示原因与影响范围）。
- URL 同步：Session/Skill 的选中态必须能用 URL 表达（刷新可恢复，便于分享/定位）。

---

## 4. 产品形态与交付范围（写死，避免摇摆）

### 4.1 交付形态

- mirror.exe（Windows 托盘常驻）
- 内置本地 HTTP Server（随机端口，监听 127.0.0.1）
- 内置 Web UI（React/Vite SPA）：运行时优先加载同目录 frontend/dist/，否则回退内嵌资源
- 端口发现：写入 ./data/http_base_url.txt

### 4.2 明确不做（v0.2）

- Wails 桌面壳
- CLI 对外功能
- 云同步/账号体系/社交分享
- 复杂目标管理（TODO/OKR/习惯）

---

## 5. MVP 闭环（v0.2 必须达成）

行为采集 → Session（切分/聚合/摘要） → Skill 归因 → Weekly 回顾 → 证据链展开 → 系统状态可诊断

---

## 6. 核心概念与权威路径（数据口径）

### 6.1 核心对象

- Raw Evidence：WindowEvent / Diff / BrowserEvent
- Session：聚合证据的最小可回放单元（权威）
- Skill：从 Session 归因得出的能力结构
- Summary（Daily/Weekly/Period）：从 Session 聚合生成的回顾输出（不直接从 raw 统计）

### 6.2 权威路径（必须遵守）

- Raw → Session（切分 + 聚合统计）
- Session → Skill attribution（exp/活跃/衰减均以 session 为输入）
- Session → Summary/Trend/Bottleneck（统一口径；任何 UI 数字可追溯到 session）

---

## 7. 用户旅程（必须可重复、可自解释）

### 7.1 首次使用（10 分钟内完成）

    1. 托盘 → 打开面板
    2. 新手引导：
       - 选择/确认要监控的 Git 项目目录（Diff watch paths）
       - 选择是否启用浏览器采集（并展示自动探测结果）
       - 选择隐私模式：默认开启脱敏（提供推荐规则模板）
       - 选择 AI 模式：离线（默认）或配置 API Key（可跳过）
    3. 引导完成后：在“状态页”看到采集健康；在“会话页”看到第一批 session（即使是规则摘要）

### 7.2 日常（每日 2 分钟）

- 首页：今日概览（投入方向/技能变化/证据覆盖率）
- 一键查看：今日 sessions → 点开证据（diff/窗口/浏览）

### 7.3 每周（10 分钟）

- 周报页：本周概览、成就/模式/建议、Top skills
- 每条结论可点开对应 sessions（证据链）；缺证据要标注“弱证据”

### 7.4 异常处理（必须无需作者介入）

- 结果为空/不可信 → 状态页解释：哪一层缺口（采集/切分/分析/配置）→ 给出可执行修复动作（配置路径、重建某日、查看最近错误）

---

## 8. 功能需求（按优先级）

> P0：v0.2 必须；P1：v0.3 建议；P2：以后

### 8.1 系统状态与可观测性（P0）

目标：消灭沉默失败，用户能自助定位问题。
页面/接口必须提供：

- 全局 Shell（UI 验收点）：
  - 侧边栏底部常驻“系统健康”指示器（Win/Diff/AI 三色点 + 心跳），点击进入状态页（见 `docs/frontend.md` 3.1）。
  - 不强制全局顶部 Banner；但健康指示器必须在任意页面可见且可点击，一次点击可看到离线/降级原因、影响范围与下一步动作。
- 采集健康：
  - Window collector：启用状态、最近采集时间、近 24h 写入数、丢弃数（如有）
  - Diff collector：启用状态、watch paths、生效目录数、近 24h diffs 数、非 Git 跳过数
  - Browser collector：启用状态、history path、近 24h events 数、脱敏启用状态
- 管道健康（权威链路分层）：
  - Session：最近一次切分时间、近 24h session 数、待 enrich 数
  - AI：是否配置、最近一次调用时间、最近错误原因、当前是否降级模式
  - RAG（如有）：是否启用、索引数量、最近错误
- 操作入口（带明确风险提示）：
  - 重建某天 sessions（P0）
  - 补全某天语义（有 AI 时）（P0）
  - 导出诊断包（日志 + 统计，脱敏）（P0）
  - 清空数据/一键重置（P1，强确认）

验收标准（P0）：

- 任一空白页面都能引导到状态页并定位至少一个可行动原因
- “离线/降级”必须在任意页面可见且可一键解释（原因、影响范围、下一步动作）

### 8.2 Session（会话）切分与摘要（P0）

- 切分规则（可配置但提供默认）：
  - idle ≥ X 分钟切分（默认 6）
  - 活动时间间隔/应用切换辅助（保持简单，避免过拟合）
- Session 内容（最小闭环）：
  - 时间范围、主要应用分布、关键 diffs、关键浏览（可为空）
  - 摘要：
    - AI 模式：LLM 摘要 + 标签
    - 离线模式：规则摘要（基于应用分布 + diff 元数据 + 浏览域名）
- Session 证据链：
  - Session → diffs（可点开 diff detail）
  - Session → window/event 分布
  - Session → browser events（脱敏后展示）
- UI 交互（v0.2 验收点，来自 `docs/frontend.md` 3.3）：
  - 会话流卡片点击后，使用右侧 Drawer/Sheet 展示详情（Diff/Timeline/Apps/Browser）。
  - Diff 详情只读高亮展示；标题/路径等信息不污染全局样式。
  - Session 详情的打开状态应与 URL 同步（`/sessions/:id`），刷新后仍能定位。

验收标准（P0）：

- 无 AI Key：仍能生成 sessions，摘要为规则口径
- 任意 session 至少能展开看到“证据覆盖率”（例如：diff 有/无、browser 有/无）

### 8.3 Skill 树与能力变化（P0）

- 技能结构：层级技能（语言 → 框架 → 子领域）保持现有模型
- 归因口径（必须可解释）：
  - exp 增长来源于 session（权重由：diff 变更量、编码时长、证据强度等简单组合）
  - 衰减：基于 last_active（已有策略可沿用）
- UI 展示：
  - 技能树（可折叠，文件资源管理器式 Tree Explorer，来自 `docs/frontend.md` 4.2）
  - 选中 skill：最近活跃 sessions 列表 → 点开证据

验收标准（P0）：

- skill 的任何增减，都能追溯到 session（至少展示“贡献 sessions 列表”）

### 8.4 Daily / Weekly / Period 回顾（P0）

- Daily：今日投入方向、亮点、技能、编码时长等（允许规则降级）
- Weekly/Month：从 sessions 聚合生成：
  - Overview（投入/覆盖天数/Top skills）
  - Achievements（可追溯）
  - Patterns & Suggestions（AI 或规则降级）
- “弱证据”机制（P0）：
  - 每条结论标注证据强度（强/中/弱），弱证据必须可解释（例如：仅窗口行为、无 diff）

验收标准（P0）：

- 无 AI 仍能生成周报（规则版），且 UI 明示“规则/AI”口径差异
- 周报至少 70% 条目可点开 sessions（v0.2 可先做到“可点开关联 sessions”，强度统计在状态页显示）

### 8.5 设置与新手引导（P0）

- 引导与设置合一：可重复进入，修改后明确“是否需要重启”
- 配置项：
  - Diff watch paths（必配项，必须在引导强调）
  - Browser enabled + history path
  - Privacy：启用/规则列表 + 预览（给一段示例文本看脱敏结果）
  - AI：Key/模型/基础 URL（不回显明文）
  - DB path（高级项，默认不暴露在引导，只在设置高级区）

验收标准（P0）：

- 新用户不配置 AI 也能走完闭环并看到价值
- 配置错误（路径不存在/无权限）必须即时校验并提示

### 8.6 数据管理（P1）

- 导出：
  - 导出“回顾报告”（HTML/Markdown，脱敏）（P1）
  - 导出“原始数据备份”（P1，高风险提示）
- 删除：
  - 按时间范围删除（P1，高风险提示）
  - 一键重置（P1，高风险提示）

---

## 9. 信息架构（Web UI 页面）

### 9.1 全局 Shell（v0.2）

- 侧边栏：收缩式设计，常驻导航与健康状态（见 `docs/frontend.md` 3.1）。
- 顶部：保留快捷入口（如搜索/命令面板）但不得以“不可用占位”长期存在；未实现时应隐藏或标记为 P1。

### 9.2 路由设计（URL 同步，v0.2）

（详细交互见 `docs/frontend.md` 6）

- `/dashboard`：概览
- `/sessions`：会话流
- `/sessions/:id`：会话详情（Drawer 展示但 URL 必须同步）
- `/skills`：技能树根
- `/skills/:skillId`：选中特定节点（Branch/Leaf 统一 ID 空间）
- `/reports`：日报/周报/月报/阶段回顾 + 历史索引
- `/status`：系统诊断（P0）
- `/settings`：设置/引导合一（P0）

### 9.3 页面要点（验收摘要）

- Dashboard：今日/本周卡片 + 证据覆盖率 + 快捷入口；热力图使用真实按日数据（避免假数据）。
- Sessions：列表（按日/周筛选、证据覆盖率标记）+ 详情（摘要、应用分布、diff、浏览、事件时间线）。
- Skills：树 + 搜索；详情包含趋势、贡献 sessions、证据。
- Trends：7d vs 30d（技能活跃/投入）+ “缺数据”解释。（可合并到 Dashboard，但入口与语义必须清晰）
- Reports：Daily/Weekly/Period 历史索引与查看，且结论必须可追溯到 sessions。
- Status：健康、队列、错误、诊断动作（按日期执行，带风险提示）。
- Settings：基础（采集开关/路径）/隐私/AI/高级（DB），并覆盖“首次使用引导”。

### 9.4 UI 技术栈与组件约束（强约束）

来自 `docs/frontend.md` 8.1/8.2：

- React 18 + Vite；Tailwind CSS；Lucide Icons；Recharts 图表。
- 复杂交互优先复用 shadcn/radix 原语（例如：技能树用 Collapsible/Accordion，会话详情用 Sheet，诊断卡片用 Card+Alert 等）。

---

## 10. API（本地 HTTP）规范（v0.2）

原则：所有 API 错误返回统一结构 { "error": "..." }，并附带可选 code/hint（建议 v0.2 加上）。

### 10.1 必备新增/强化接口（建议）

- GET /api/status：汇总健康状态（采集/管道/AI/最近错误/覆盖率）
- POST /api/maintenance/sessions/rebuild：按日期重建 sessions
- POST /api/maintenance/sessions/enrich：按日期补全语义
- GET /api/diagnostics/export：导出诊断包（文件流或生成到 data/ 下并返回路径）

SSE（已有）：

- GET /api/events：
  - data_changed（source, count）
  - settings_updated
  - pipeline_status_changed（让 UI 状态页实时刷新）

### 10.2 UI 统计口径（新增，v0.2）

- GET /api/trends：
  - 必须包含 `daily_stats[]`（自然日维度）：用于 Dashboard 热力图与趋势概览，避免前端估算/假数据。

---

## 11. 数据与隐私（v0.2 要求）

### 11.1 默认隐私策略（P0）

- 默认开启脱敏（至少对 URL query/fragment、账号信息、token 类模式做处理）
- UI 明确告知：采集了什么、存在哪里、怎么关掉、怎么删除

### 11.2 安全边界

- Server 只监听 127.0.0.1
- 不提供远程访问/鉴权（v0.2 仅本地）
- 配置文件权限：仅当前用户可读写（已有 0600 思路，保持）

---

## 12. 可靠性与升级策略（产品化门槛）

### 12.1 数据库迁移（P0 最小化方案）

- 增加 schema_version（或等价机制），避免仅靠 AutoMigrate
- 启动时：
  - 检测版本差异 → 执行向前迁移 → 失败则进入“只读/安全模式”并在状态页提示

### 12.2 崩溃与恢复（P0）

- 后台任务失败不得阻塞 UI
- 所有后台错误进入“最近错误列表”（状态页可见）
- 写库失败要有重试策略与丢弃计数（可观测）

---

## 13. 成功指标（v0.2 可验证）

- 新用户 10 分钟完成闭环：配置 watch path → 看到 session → 能点开证据
- 连续使用 14 天：每天都有 session（允许规则摘要），每周至少生成 1 次周报
- 证据覆盖率：周报条目中 ≥70% 可跳转到关联 sessions（先做“可跳转”，再做“强证据比例”）

---

## 14. 风险与对策

- 风险：用户无数据 → 对策：状态页 + 引导强制配置 diff watch paths + 空态行动按钮
- 风险：隐私担忧 → 对策：默认脱敏 + 可预览 + 一键删除/导出说明
- 风险：AI 不稳定/成本 → 对策：规则降级 + UI 明示口径差异
- 风险：升级破坏数据 → 对策：schema_version + 备份/诊断包

---

## 15. 版本规划

- v0.2（产品化闭环）：状态页/诊断动作、Session 权威、规则降级一致、引导完善、周报可追溯
- v0.3（价值增强）：瓶颈分析、覆盖率提升（browser/diff）、导出报告、通知
- v0.4（长期演进）：更完善迁移/备份/恢复、性能与长期数据治理
