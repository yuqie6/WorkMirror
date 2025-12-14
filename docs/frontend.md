### 一、 设计核心概念：The "Engineer's Cockpit" (工程师驾驶舱)

- **视觉隐喻：** IDE（VS Code）、CI/CD 面板、服务器监控台。
- **配色策略：**
  - **底色：** 深色模式优先（Zinc/Slate 900），符合开发者习惯。
  - **状态色（关键）：** 既然 P0 是“诊断”，状态色必须明确。
    - 🟢 Emerald: 正常/强证据/AI 生成。
    - 🟡 Amber: 规则降级/弱证据/警告。
    - 🔴 Rose: 采集失败/错误/无数据。
  - **数据色：** 靛蓝（Indigo）或紫色（Violet）用于表示“智能/归因”。
- **字体：**
  - UI 文本：Inter 或系统默认 sans-serif。
  - **证据/Diff/路径/日志：必须用等宽字体（JetBrains Mono / Fira Code）。**

---

### 二、 布局架构 (App Shell)

由于是本地 Web UI，不需要复杂的路由过渡，追求**“信息触手可及”**。

#### 1. 侧边栏 (Sidebar Navigation)

使用收缩式侧边栏，最大化右侧内容区域。

- **顶部：** App Logo (Project Mirror)。
- **主导航：**
  - 📊 Dashboard (概览)
  - ⏱️ Sessions (会话流 - 核心)
  - 🌳 Skills (技能树)
  - 📅 Reports (周报/回顾)
- **底部（固定）：**
  - 🏥 **System Status (红绿灯指示器，P0 关键入口)**
  - ⚙️ Settings

#### 2. 全局状态栏 (Global Health Bar) - 对应 PRD 8.1

不要把状态藏在设置里。在页面顶部或侧边栏底部做一个微型组件：

- 显示：`Window: 🟢` `Diff: 🟢` `AI: 🟡 (Offline)`
- **交互：** 点击直接跳转到 `/status` 页面进行诊断。这是实现“消灭沉默失败”的关键 UI 设计。

---

### 三、 核心页面设计方案

#### 1. Dashboard (首页) - "Today at a Glance"

布局建议：**Bento Grid (便当盒布局)**

- **左上 (大卡片)：** **今日投入分布。** 饼图或进度条，显示 Code vs. Browse vs. Idle。
- **右上 (关键指标)：**
  - Session 数量 (比如 "12 Sessions")。
  - 证据覆盖率 (比如 "85% Covered" - 🟢 高亮)。
- **中段 (Action 区域)：**
  - 如果是空态/异常：显示一个醒目的警告 Banner，“Diff 采集未配置，点击修复”。
  - 如果是正常：显示“最近一个 Session”的简报。
- **底部：** **Contribution Graph (类似 GitHub)**。展示过去 30 天的活跃热力图，给用户“连续打卡”的成就感。

#### 2. Sessions (会话流) - " The Source of Truth"

这是最复杂的页面，需要展示“权威路径”。建议采用 **“Timeline Feed”** 形式。

- **列表项设计 (Card)：**
  - **Header:** `[14:00 - 15:30]` `VS Code` `Project Mirror Dev` (标题由 AI 或规则生成)。
  - **Badges (证据强度):**
    - `✨ AI Generated` (或是 `⚙️ Rule Based` 灰色标)
    - `Diff: +120/-5` (点击展开)
    - `Browser: 45 events`
  - **Summary:** 一段简短的文本摘要。
- **交互 (Drill Down)：**
  - 点击 Card，**右侧滑出 Drawer (抽屉)**，而不是跳转页面。
  - **Drawer 内容：** 完整的证据链。
    - Tab 1: Timeline (窗口切换时间轴)。
    - Tab 2: Diffs (渲染后的 git diff，代码高亮)。
    - Tab 3: Browser (访问过的 URL 列表，脱敏显示)。
- **设计理由：** 这种设计符合 PRD 的“任何结论必须可点开看到来源”。

#### 3. Reports (周报) - "The Insight"

- **布局：** 类似 notion 文档的阅读体验，单列居中布局。
- **Weak Evidence Pattern (弱证据模式)：**
  - 如果某条结论证据不足（例如 AI 瞎编的），UI 上必须显示一个**虚线边框**或**黄色三角图标**。
  - Hover 提示：“缺少 Diff 数据，仅基于窗口标题推断”。

#### 4. Status & Diagnostics (诊断页) - P0 核心

- **风格：** 必须像“服务器仪表盘”。
- **模块划分：**
  - **Collectors (卡片组):** 每个采集器一个卡片（Window, Diff, Browser）。显示：`Last Heartbeat`, `Events/24h`, `Error Rate`。
  - **Pipeline:** 显示 Session 切分任务的状态。
  - **Actions (危险区):** 使用红色边框或红色按钮。
    - `Rebuild Sessions (YYYY-MM-DD)`
    - `Export Debug Pack`
- **反馈：** 点击“重建”后，需要有一个实时的 Log 窗口或 Toast，告诉用户后台在干活，不能让用户觉得“卡死了”。

---

### 四、 组件库映射 (Shadcn UI)

为了快速落地，直接使用以下 Shadcn 组件对应你的需求：

| PRD 需求      | 推荐 UI 组件                 | 备注                               |
| :------------ | :--------------------------- | :--------------------------------- |
| **整体布局**  | `Sidebar` / `Layout`         | 刚出的 Shadcn Block 这种布局很现成 |
| **证据详情**  | `Sheet` (Side Drawer)        | 保持上下文，不要全屏跳转           |
| **诊断信息**  | `Alert`, `Badge`             | 区分 Info/Warning/Destructive      |
| **技能树**    | `Accordion` 或 `Tree`        | 简单的折叠面板即可满足 v0.2        |
| **数据卡片**  | `Card`                       | 加上 `HoverCard` 做解释说明        |
| **时间选择**  | `Calendar`, `DatePicker`     | 用于选择重建 Session 的日期        |
| **操作反馈**  | `Toast`                      | 成功/失败的非阻塞通知              |
| **设置表单**  | `Form`, `Switch`             | 尤其是隐私设置的开关               |
| **Diff 展示** | (第三方库) `react-diff-view` | 嵌入在 Card 或 Sheet 中            |

---

### 五、 针对 "AI vs. 规则" 的视觉区分策略

PRD 中提到**“无 AI 时仍能产出规则回顾，且 UI 明确标注”**。这一点在设计上非常重要，可以通过以下方式强化：

1.  **Icon 区分：**
    - AI 生成的内容：使用 ✨ (Sparkles) 图标，颜色用 Indigo-500。
    - 规则生成的内容：使用 ⚙️ (Gear) 或 📊 (Chart) 图标，颜色用 Slate-500。
2.  **背景微光：**
    - AI 生成的卡片可以有一个极其微弱的紫色渐变背景或边框发光。
    - 规则生成的卡片保持扁平灰底。

---
