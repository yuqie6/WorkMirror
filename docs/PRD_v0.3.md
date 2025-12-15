# 复盘镜 WorkMirror PRD v0.3（可信度与证据链增强版）

最后更新：2025-12-15  
基线版本：`v0.2.0-alpha.1`（见 `docs/PRD.md`）

## 0. 这版 PRD 要解决什么

v0.2 已跑通闭环，但“结论可信度”仍受两类不稳定因素影响：

1) **证据归并不稳**：当 window collector 缺失/晚到/采样不足时，diff/browser 证据可能无法落到 Session，导致 Evidence First 断链。  
2) **语义来源不透明**：UI/统计依赖启发式推断 `ai/rule`，离线/降级状态不够可解释。

v0.3 的目标是：让“每一个输出都可追溯且口径稳定”，并把不确定性显式化（可观测、可修复）。

---

## 1. 版本目标（必须可验收）

### 1.1 北极星目标（对用户）

- 任何一天的 Diff/Browser 证据，不会因为窗口采集缺口而“消失”；最差也能以 `diff_only`/`orphan` 的形式被看到并可修复归并。
- UI 不再猜测：任何 Session/Report 都能明确标注 `semantic_source=ai|rule`，并能解释降级原因与影响范围。

### 1.2 验收口径（量化）

- **证据落链率**：近 24h diffs 中，≥95% 能被归并到某个 Session（或被明确标注为 `diff_only/orphan` 且可一键修复）。  
- **口径一致性**：前端不允许再通过启发式推断 `ai/rule`；必须由后端契约字段驱动。  
- **No Silent Failures**：当 `diff_only/orphan` 比例异常升高时，`/status` 必须给出明确原因候选与下一步动作（例如：窗口采集未运行、权限、poll interval 过大等）。

---

## 2. 明确不做（v0.3）

- 云同步/账号体系/多端协作
- 对外开放远程访问（仍仅 `127.0.0.1`）
- “复杂目标管理”（TODO/OKR/习惯）
- 大规模 UI 重做（遵循 `docs/frontend.md`，只做为达成验收所必需的交互调整）

---

## 3. P0 功能需求（v0.3 必须）

### 3.1 Session 证据归并策略升级（P0）

目标：保证 Evidence First 链路“不断”，并对不确定性显式标注。

#### 3.1.1 证据归并规则（最小可解释，KISS）

- 以 Session（window 锚点）为主，但允许 **diff/browser 作为二级锚点**：
  - 若存在 window events：按现有策略切分 Session，并将 diff/browser 归并到时间窗内。
  - 若无 window events 或 window 覆盖稀疏：
    - 仍应把 diff/browser 归并到“最近的可解释单元”，策略优先级：
      1) 归并到相邻 Session（时间邻近，且间隔 ≤ `attach_gap_minutes`）
      2) 若无法归并：创建 `diff_only`（或 `browser_only`）Session（显式弱证据，避免静默丢失）
- 对任何“归并不确定”的证据，必须写入可观测指标：`orphan_diffs_24h`、`orphan_browser_24h`。

#### 3.1.2 对外契约（必须）

- SessionDTO/SessionDetailDTO 必须包含：
  - `semantic_source`：`ai|rule`
  - `evidence_hint`：`diff+browser|diff|browser|window_only|diff_only|browser_only`
  - （可选）`semantic_version`：用于复现与重建

#### 3.1.3 维护动作（P0）

- 状态页提供：
  - “重建某天 sessions”（已有）
  - “修复证据归并”（新增）：对 `diff_only/orphan` 执行一次归并尝试（按当天）
  - 所有动作必须明确风险提示：不会删除原始数据；仅新增更高版本的 Session（或更新 metadata）

### 3.2 语义来源与降级显式化（P0）

目标：离线/降级与 AI 能力差异“可见、可解释、可追溯”。

- Session 语义写入时必须记录：
  - `semantic_source`（ai/rule）
  - `semantic_provider`/`semantic_model`（仅 ai 时）
  - `degraded_reason`（仅 rule 时；例如 `not_configured` / `rate_limited` / `provider_error`）
- Report（Daily/Weekly/Period）也必须带 `mode=ai|offline_rule` 与 `degraded_reason`。

### 3.3 “弱证据”机制补齐到全链路（P0）

目标：用户能一眼判断可信度，并可一跳回证据。

- Dashboard/Reports 必须展示：
  - 证据强度（强/中/弱）与覆盖率比例
  - “为什么弱”（例如：无 diff、无 browser、窗口采集缺口）
- 对“弱证据”条目必须提供一键入口：跳转到关联 Sessions/证据明细，或进入状态页查看缺口原因。

---

## 4. P1（建议，但不阻塞 v0.3）

### 4.1 导出报告（P1）

- 导出 Weekly/Period 为 Markdown/HTML（脱敏），且保留“Claim → Session refs”链接。

### 4.2 数据管理（P1）

- 按时间范围删除（强确认）
- 原始数据备份导出（强确认）

---

## 5. 成功指标（v0.3）

- 证据落链率达标（见 1.2），且 `orphan_*` 指标能在状态页持续收敛
- 用户能在 1 次交互内从任意结论 drill-down 到证据（Session/Diff/Browser）
- 离线/降级状态不会“伪装成 AI”（UI 口径清晰且可解释）

