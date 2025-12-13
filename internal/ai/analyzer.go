package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// DiffAnalyzer Diff 分析器
type DiffAnalyzer struct {
	client *DeepSeekClient
}

// NewDiffAnalyzer 创建 Diff 分析器
func NewDiffAnalyzer(client *DeepSeekClient) *DiffAnalyzer {
	return &DiffAnalyzer{client: client}
}

// SkillWithCategory 带分类的技能（AI 返回）
type SkillWithCategory struct {
	Name     string `json:"name"`             // 技能名称（标准名称如 Go, React）
	Category string `json:"category"`         // 分类: language/framework/database/devops/tool/concept/other
	Parent   string `json:"parent,omitempty"` // 父技能名（AI 决定），如 Gin → Go
}

// SkillInfo 简化的技能信息（传给 AI 作为上下文）
type SkillInfo struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Parent   string `json:"parent,omitempty"`
}

// DiffInsight Diff 解读结果
type DiffInsight struct {
	Insight    string              `json:"insight"`    // AI 解读
	Skills     []SkillWithCategory `json:"skills"`     // 涉及技能（带分类和层级）
	Difficulty float64             `json:"difficulty"` // 难度 0-1
	Category   string              `json:"category"`   // 代码变更分类: learning, refactoring, bugfix, feature
}

// AnalyzeDiff 分析单个 Diff（传入当前技能树作为上下文）
func (a *DiffAnalyzer) AnalyzeDiff(ctx context.Context, filePath, language, diffContent string, existingSkills []SkillInfo) (*DiffInsight, error) {
	if !a.client.IsConfigured() {
		return nil, fmt.Errorf("DeepSeek API 未配置")
	}

	// 限制 diff 长度
	if len(diffContent) > 3000 {
		diffContent = diffContent[:3000] + "\n... (truncated)"
	}

	// 构建技能树上下文
	var skillTreeContext strings.Builder
	if len(existingSkills) > 0 {
		skillTreeContext.WriteString("\n当前已有技能树：\n")
		for _, s := range existingSkills {
			if s.Parent == "" {
				skillTreeContext.WriteString(fmt.Sprintf("- %s (%s)\n", s.Name, s.Category))
			} else {
				skillTreeContext.WriteString(fmt.Sprintf("  - %s → %s (%s)\n", s.Name, s.Parent, s.Category))
			}
		}
		skillTreeContext.WriteString("\n")
	}

	prompt := fmt.Sprintf(`分析以下代码变更，推断开发者学习或实践了什么。

文件: %s
语言: %s
%s
Diff:
%s

请用 JSON 格式返回（不要 markdown 代码块）:
{
  "insight": "一句话描述这次修改学到了什么或做了什么（中文）",
  "skills": [
    {"name": "技能名", "category": "分类", "parent": "父技能名（可选）"}
  ],
  "difficulty": 0.3,
  "category": "learning"
}

技能层级规则：
1. 如果技能已存在于技能树中，使用**完全相同的名称**
2. 编程语言是顶级技能（parent 留空）
3. 框架/库归属到对应语言（如 Gin → Go, React → JavaScript）
4. category 可选值: language/framework/database/devops/tool/concept/other
5. 变更分类: learning/refactoring/bugfix/feature`, filePath, language, skillTreeContext.String(), diffContent)

	messages := []Message{
		{Role: "system", Content: "你是一个代码分析助手，擅长从代码变更中推断开发者的学习和成长。你能看到用户当前的技能树，请合理判断技能归属。回复必须是纯 JSON，不要 markdown。"},
		{Role: "user", Content: prompt},
	}

	response, err := a.client.ChatWithOptions(ctx, messages, 0.2, 500)
	if err != nil {
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	// 清理响应（移除可能的 markdown 代码块）
	response = cleanJSONResponse(response)

	var insight DiffInsight
	if err := json.Unmarshal([]byte(response), &insight); err != nil {
		slog.Warn("解析 AI 响应失败，使用原始文本", "response", response, "error", err)
		// 降级处理：直接使用响应文本
		insight = DiffInsight{
			Insight:    response,
			Skills:     []SkillWithCategory{{Name: language, Category: "language"}},
			Difficulty: 0.3,
			Category:   "unknown",
		}
	}

	return &insight, nil
}

// DailySummaryRequest 每日总结请求
type DailySummaryRequest struct {
	Date            string            // 日期
	WindowEvents    []WindowEventInfo // 窗口事件摘要
	Diffs           []DiffInfo        // Diff 摘要
	HistoryMemories []string          // 相关历史记忆（来自 RAG）
}

// WindowEventInfo 窗口事件信息
type WindowEventInfo struct {
	AppName  string
	Duration int // 分钟
}

// DiffInfo Diff 信息
type DiffInfo struct {
	FileName     string
	Language     string
	Insight      string // 预分析的解读（可能为空）
	DiffContent  string // 原始 diff 内容
	LinesChanged int
}

// DailySummaryResult 每日总结结果
type DailySummaryResult struct {
	Summary      string   `json:"summary"`       // 总结
	Highlights   string   `json:"highlights"`    // 亮点
	Struggles    string   `json:"struggles"`     // 困难
	SkillsGained []string `json:"skills_gained"` // 获得技能
	Suggestions  string   `json:"suggestions"`   // 建议
}

// SessionSummaryRequest 会话摘要请求
type SessionSummaryRequest struct {
	SessionID  int64             `json:"session_id"`
	Date       string            `json:"date"`
	TimeRange  string            `json:"time_range"`
	PrimaryApp string            `json:"primary_app"`
	AppUsage   []WindowEventInfo `json:"app_usage"`
	Diffs      []DiffInfo        `json:"diffs"`
	Browser    []BrowserInfo     `json:"browser"`
	SkillsHint []string          `json:"skills_hint"`
	Memories   []string          `json:"memories"`
}

type BrowserInfo struct {
	Domain string `json:"domain"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

// SessionSummaryResult 会话摘要结果
type SessionSummaryResult struct {
	Summary        string   `json:"summary"`
	Category       string   `json:"category"` // technical/learning/exploration/other
	SkillsInvolved []string `json:"skills_involved"`
	Tags           []string `json:"tags"`
}

// GenerateDailySummary 生成每日总结
func (a *DiffAnalyzer) GenerateDailySummary(ctx context.Context, req *DailySummaryRequest) (*DailySummaryResult, error) {
	if !a.client.IsConfigured() {
		return nil, fmt.Errorf("DeepSeek API 未配置")
	}

	windowTotal := 0
	for _, w := range req.WindowEvents {
		if w.Duration > 0 {
			windowTotal += w.Duration
		}
	}
	diffCountTotal := len(req.Diffs)
	linesChangedTotal := 0
	for _, d := range req.Diffs {
		if d.LinesChanged > 0 {
			linesChangedTotal += d.LinesChanged
		}
	}

	// 构建窗口使用摘要
	var windowSummary strings.Builder
	windowEvents := req.WindowEvents
	// 控制 prompt 规模：只展示前 12 个应用（通常已按时长排序）
	if len(windowEvents) > 12 {
		windowEvents = windowEvents[:12]
	}
	for _, w := range windowEvents {
		windowSummary.WriteString(fmt.Sprintf("- %s: %d 分钟\n", w.AppName, w.Duration))
	}

	// 构建 Diff 摘要
	var diffSummary strings.Builder
	diffs := req.Diffs
	// 控制 prompt 规模：只展开前 20 个 diff（其余用统计信息概括）
	if len(diffs) > 20 {
		diffs = diffs[:20]
	}
	for _, d := range diffs {
		// 优先使用预分析解读，否则使用原始 diff 内容
		description := d.Insight
		if description == "" && d.DiffContent != "" {
			// 截取前 300 字符作为描述
			content := d.DiffContent
			if len(content) > 300 {
				content = content[:300] + "..."
			}
			description = content
		}
		if description == "" {
			description = fmt.Sprintf("%d行变更", d.LinesChanged)
		}
		diffSummary.WriteString(fmt.Sprintf("- %s (%s): %s\n", d.FileName, d.Language, description))
	}

	// 构建历史记忆摘要
	var historySummary strings.Builder
	if len(req.HistoryMemories) > 0 {
		historySummary.WriteString("\n相关历史记忆（你之前的学习/工作记录，可作为参考）:\n")
		for _, mem := range req.HistoryMemories {
			historySummary.WriteString(fmt.Sprintf("- %s\n", mem))
		}
	}

	prompt := fmt.Sprintf(`根据以下行为数据，生成今日工作/学习总结。
	%s
	日期: %s

	统计概览:
	- 应用使用总时长: %d 分钟（下方仅展示 Top %d）
	- 代码变更: %d 次（共 %d 行变更；下方仅展示前 %d 条）

	应用使用时长:
	%s

	代码变更:
	%s

	请用 JSON 格式返回（不要 markdown 代码块）:
	{
	  "summary": "今日总结（请根据数据量自适应篇幅：轻量日 2-3 句；中等 5-8 句；高强度/多变更 10-16 句。尽量引用具体证据：应用名/文件名/语言/技能，避免套话。）",
	  "highlights": "今日亮点（2-6 条要点，用换行分隔；每条尽量具体。若确实没有，写'无'）",
	  "struggles": "今日困难（0-5 条要点，用换行分隔；没有就写'无'）",
	  "skills_gained": ["今日涉及的技能（按重要性排序，允许 0-12 个）"],
	  "suggestions": "明日建议（2-6 条要点，用换行分隔；优先给可执行的小动作）"
	}`, historySummary.String(), req.Date, windowTotal, len(windowEvents), diffCountTotal, linesChangedTotal, len(diffs), windowSummary.String(), diffSummary.String())

	messages := []Message{
		{Role: "system", Content: "你是一个个人成长助手，帮助用户回顾每天的工作和学习，提供有建设性的反馈。回复必须是纯 JSON。"},
		{Role: "user", Content: prompt},
	}

	response, err := a.client.ChatWithOptions(ctx, messages, 0.5, 1000)
	if err != nil {
		return nil, fmt.Errorf("生成总结失败: %w", err)
	}

	response = cleanJSONResponse(response)

	var result DailySummaryResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析总结失败: %w", err)
	}

	return &result, nil
}

// GenerateSessionSummary 生成会话语义摘要（用于证据链的可解释描述）
func (a *DiffAnalyzer) GenerateSessionSummary(ctx context.Context, req *SessionSummaryRequest) (*SessionSummaryResult, error) {
	if !a.client.IsConfigured() {
		return nil, fmt.Errorf("DeepSeek API 未配置")
	}

	// 控制输入规模
	maxApps := 8
	apps := req.AppUsage
	if len(apps) > maxApps {
		apps = apps[:maxApps]
	}
	maxDiffs := 12
	diffs := req.Diffs
	if len(diffs) > maxDiffs {
		diffs = diffs[:maxDiffs]
	}
	maxBrowser := 12
	browser := req.Browser
	if len(browser) > maxBrowser {
		browser = browser[:maxBrowser]
	}
	maxMem := 5
	mem := req.Memories
	if len(mem) > maxMem {
		mem = mem[:maxMem]
	}

	var b strings.Builder
	b.WriteString("请基于以下本地行为证据，生成一个可解释的会话摘要。\n")
	b.WriteString("要求：\n")
	b.WriteString("1) summary 用中文 1 句话（尽量具体，避免空泛）\n")
	b.WriteString("2) category 只能是 technical/learning/exploration/other\n")
	b.WriteString("3) skills_involved 最多 8 个，尽量使用用户已有技能树中的标准名称（如 Go、Redis、React）\n")
	b.WriteString("4) tags 最多 6 个，用中文短标签（如 并发、性能、数据库、文档阅读）\n")
	b.WriteString("5) 必须可追溯：summary 应对应下面的 diff/browser/app 证据，不要胡编\n\n")

	b.WriteString(fmt.Sprintf("日期: %s\n时间: %s\n主应用: %s\n\n", req.Date, req.TimeRange, req.PrimaryApp))

	if len(apps) > 0 {
		b.WriteString("应用使用:\n")
		for _, a := range apps {
			b.WriteString(fmt.Sprintf("- %s: %d 分钟\n", a.AppName, a.Duration))
		}
		b.WriteString("\n")
	}

	if len(diffs) > 0 {
		b.WriteString("代码变更:\n")
		for _, d := range diffs {
			desc := strings.TrimSpace(d.Insight)
			if desc == "" {
				desc = fmt.Sprintf("%d行变更", d.LinesChanged)
			}
			b.WriteString(fmt.Sprintf("- %s (%s): %s\n", d.FileName, d.Language, desc))
		}
		b.WriteString("\n")
	}

	if len(browser) > 0 {
		b.WriteString("浏览记录:\n")
		for _, it := range browser {
			title := strings.TrimSpace(it.Title)
			if title == "" {
				title = it.Domain
			}
			b.WriteString(fmt.Sprintf("- %s: %s\n", it.Domain, title))
		}
		b.WriteString("\n")
	}

	if len(req.SkillsHint) > 0 {
		b.WriteString("技能提示（可参考）:\n")
		for _, s := range req.SkillsHint {
			if strings.TrimSpace(s) == "" {
				continue
			}
			b.WriteString("- " + strings.TrimSpace(s) + "\n")
		}
		b.WriteString("\n")
	}

	if len(mem) > 0 {
		b.WriteString("相关历史记忆（可参考，不要编造不存在的内容）:\n")
		for _, m := range mem {
			b.WriteString("- " + strings.TrimSpace(m) + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("请用 JSON 格式返回（不要 markdown 代码块）:\n")
	b.WriteString("{\n")
	b.WriteString("  \"summary\": \"...\",\n")
	b.WriteString("  \"category\": \"technical\",\n")
	b.WriteString("  \"skills_involved\": [\"...\"],\n")
	b.WriteString("  \"tags\": [\"...\"]\n")
	b.WriteString("}\n")

	messages := []Message{
		{Role: "system", Content: "你是一个本地优先的个人成长分析助手。你必须严格基于证据生成摘要，回复必须是纯 JSON。"},
		{Role: "user", Content: b.String()},
	}

	response, err := a.client.ChatWithOptions(ctx, messages, 0.3, 600)
	if err != nil {
		return nil, fmt.Errorf("生成会话摘要失败: %w", err)
	}
	response = cleanJSONResponse(response)

	var result SessionSummaryResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析会话摘要失败: %w", err)
	}
	return &result, nil
}

// cleanJSONResponse 清理 JSON 响应（移除 markdown 代码块和额外文本）
func cleanJSONResponse(response string) string {
	response = strings.TrimSpace(response)

	// 移除 ```json ... ``` 或 ``` ... ```
	if strings.Contains(response, "```") {
		// 找 JSON 开始位置
		jsonStart := strings.Index(response, "```json")
		if jsonStart == -1 {
			jsonStart = strings.Index(response, "```")
		}
		if jsonStart != -1 {
			// 跳过 ```json\n 或 ```\n
			startIdx := strings.Index(response[jsonStart:], "\n")
			if startIdx != -1 {
				response = response[jsonStart+startIdx+1:]
			}
		}
		// 移除结尾的 ```
		if endIdx := strings.LastIndex(response, "```"); endIdx != -1 {
			response = response[:endIdx]
		}
	}

	response = strings.TrimSpace(response)

	// 尝试提取 JSON 对象（处理 AI 添加的前缀/后缀文字）
	if !strings.HasPrefix(response, "{") {
		// 找到第一个 {
		if idx := strings.Index(response, "{"); idx != -1 {
			response = response[idx:]
		}
	}
	if !strings.HasSuffix(response, "}") {
		// 找到最后一个 }
		if idx := strings.LastIndex(response, "}"); idx != -1 {
			response = response[:idx+1]
		}
	}

	return strings.TrimSpace(response)
}

// WeeklySummaryRequest 周报请求
type WeeklySummaryRequest struct {
	PeriodType     string // week/month（为空按 week）
	StartDate      string
	EndDate        string
	DailySummaries []DailySummaryInfo
	TotalCoding    int
	TotalDiffs     int
}

// DailySummaryInfo 日报信息
type DailySummaryInfo struct {
	Date       string
	Summary    string
	Highlights string
	Skills     []string
}

// WeeklySummaryResult 周报结果
type WeeklySummaryResult struct {
	Overview     string   `json:"overview"`     // 本周整体概述
	Achievements []string `json:"achievements"` // 主要成就
	Patterns     string   `json:"patterns"`     // 学习模式分析
	Suggestions  string   `json:"suggestions"`  // 下周建议
	TopSkills    []string `json:"top_skills"`   // 本周重点技能
}

// GenerateWeeklySummary 生成周报
func (a *DiffAnalyzer) GenerateWeeklySummary(ctx context.Context, req *WeeklySummaryRequest) (*WeeklySummaryResult, error) {
	var dailyDetails strings.Builder
	for _, s := range req.DailySummaries {
		dailyDetails.WriteString(fmt.Sprintf("【%s】%s 亮点: %s\n", s.Date, s.Summary, s.Highlights))
	}

	periodType := strings.ToLower(strings.TrimSpace(req.PeriodType))
	periodLabel := "本周"
	nextLabel := "下周"
	periodScope := "一周"
	if periodType == "month" {
		periodLabel = "本月"
		nextLabel = "下月"
		periodScope = "一个月"
	}

	prompt := fmt.Sprintf(`请分析以下%s的工作记录，生成阶段汇总：

时间范围: %s 至 %s
总编码时长: %d 分钟
总代码变更: %d 次

每日记录:
%s

	请用 JSON 格式返回（不要 markdown 代码块）:
	{
	  "overview": "%s整体概述（请根据数据量自适应：轻量期 3-5 句；中等 6-10 句；高强度 10-16 句。尽量引用具体证据：哪几天在做什么、主要语言/主题变化、节奏变化。）",
	  "achievements": ["主要成就（请按重要性给 3-8 条，不要固定 3 条；每条尽量具体）"],
	  "patterns": "学习模式分析（请写成一段有观点的分析：投入/产出、节奏、语言/技能迁移、反复出现的问题）",
	  "suggestions": "%s建议（请给 3-7 条可执行建议；如果数据偏少，也要说明原因并给出补数据/改流程建议）",
	  "top_skills": ["%s重点技能（按重要性排序，允许 3-12 个；不要固定数量）"]
	}
	注意：如果这是月汇总，请不要使用“本周/下周”的措辞。`, periodScope, req.StartDate, req.EndDate, req.TotalCoding, req.TotalDiffs, dailyDetails.String(), periodLabel, nextLabel, periodLabel)

	messages := []Message{
		{Role: "system", Content: fmt.Sprintf("你是一个个人成长助手，帮助用户回顾%s的工作和学习，提供有深度的分析和建设性的反馈。回复必须是纯 JSON。", periodScope)},
		{Role: "user", Content: prompt},
	}

	response, err := a.client.ChatWithOptions(ctx, messages, 0.5, 1500)
	if err != nil {
		return nil, fmt.Errorf("生成周报失败: %w", err)
	}

	response = cleanJSONResponse(response)

	var result WeeklySummaryResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析周报失败: %w", err)
	}

	return &result, nil
}
