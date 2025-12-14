package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/yuqie6/WorkMirror/internal/ai/prompts"
)

// DiffAnalyzer Diff 分析器
type DiffAnalyzer struct {
	client *DeepSeekClient
}

// NewDiffAnalyzer 创建 Diff 分析器
func NewDiffAnalyzer(client *DeepSeekClient) *DiffAnalyzer {
	return &DiffAnalyzer{client: client}
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

	prompt := prompts.DiffAnalysisUser(filePath, language, skillTreeContext.String(), diffContent)

	messages := []Message{
		{Role: "system", Content: prompts.DiffAnalysisSystem},
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

	prompt := prompts.DailySummaryUser(
		req.Date,
		windowTotal,
		len(windowEvents),
		diffCountTotal,
		linesChangedTotal,
		len(diffs),
		windowSummary.String(),
		diffSummary.String(),
		historySummary.String(),
	)

	messages := []Message{
		{Role: "system", Content: prompts.DailySummarySystem},
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

	// 计算会话丰富度，用于自适应摘要长度
	evidenceScore := len(diffs)*3 + len(browser)*2 + len(apps)
	summaryGuidance := "1 句话"
	if evidenceScore >= 20 {
		summaryGuidance = "4-6 句话（请详细描述主要工作内容、涉及的技术点和解决的问题）"
	} else if evidenceScore >= 8 {
		summaryGuidance = "2-3 句话（概括主要活动和技术点）"
	}

	appLines := make([]string, 0, len(apps))
	for _, a := range apps {
		appLines = append(appLines, fmt.Sprintf("%s: %d 分钟", a.AppName, a.Duration))
	}

	diffLines := make([]string, 0, len(diffs))
	for _, d := range diffs {
		desc := strings.TrimSpace(d.Insight)
		if desc == "" {
			desc = fmt.Sprintf("%d行变更", d.LinesChanged)
		}
		diffLines = append(diffLines, fmt.Sprintf("%s (%s): %s", d.FileName, d.Language, desc))
	}

	browserLines := make([]string, 0, len(browser))
	for _, it := range browser {
		title := strings.TrimSpace(it.Title)
		if title == "" {
			title = it.Domain
		}
		browserLines = append(browserLines, fmt.Sprintf("%s: %s", it.Domain, title))
	}

	skillsHintLines := make([]string, 0, len(req.SkillsHint))
	for _, s := range req.SkillsHint {
		s = strings.TrimSpace(s)
		if s != "" {
			skillsHintLines = append(skillsHintLines, s)
		}
	}

	memLines := make([]string, 0, len(mem))
	for _, m := range mem {
		m = strings.TrimSpace(m)
		if m != "" {
			memLines = append(memLines, m)
		}
	}

	prompt := prompts.SessionSummaryUser(prompts.SessionSummaryUserInput{
		Date:            req.Date,
		TimeRange:       req.TimeRange,
		PrimaryApp:      req.PrimaryApp,
		SummaryGuidance: summaryGuidance,
		AppLines:        appLines,
		DiffLines:       diffLines,
		BrowserLines:    browserLines,
		SkillsHintLines: skillsHintLines,
		MemoryLines:     memLines,
	})

	messages := []Message{
		{Role: "system", Content: prompts.SessionSummarySystem},
		{Role: "user", Content: prompt},
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

// GenerateWeeklySummary 生成周报
func (a *DiffAnalyzer) GenerateWeeklySummary(ctx context.Context, req *WeeklySummaryRequest) (*WeeklySummaryResult, error) {
	var dailyDetails strings.Builder
	for _, s := range req.DailySummaries {
		dailyDetails.WriteString(fmt.Sprintf("【%s】%s 亮点: %s\n", s.Date, s.Summary, s.Highlights))
	}

	periodType := strings.ToLower(strings.TrimSpace(req.PeriodType))
	prompt := prompts.WeeklySummaryUser(periodType, req.StartDate, req.EndDate, req.TotalCoding, req.TotalDiffs, dailyDetails.String())

	messages := []Message{
		{Role: "system", Content: prompts.WeeklySummarySystem(periodType)},
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
