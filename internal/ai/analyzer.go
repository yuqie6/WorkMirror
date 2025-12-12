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

// GenerateDailySummary 生成每日总结
func (a *DiffAnalyzer) GenerateDailySummary(ctx context.Context, req *DailySummaryRequest) (*DailySummaryResult, error) {
	if !a.client.IsConfigured() {
		return nil, fmt.Errorf("DeepSeek API 未配置")
	}

	// 构建窗口使用摘要
	var windowSummary strings.Builder
	for _, w := range req.WindowEvents {
		windowSummary.WriteString(fmt.Sprintf("- %s: %d 分钟\n", w.AppName, w.Duration))
	}

	// 构建 Diff 摘要
	var diffSummary strings.Builder
	for _, d := range req.Diffs {
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

应用使用时长:
%s

代码变更:
%s

请用 JSON 格式返回（不要 markdown 代码块）:
{
  "summary": "今日总结（2-3句话概括今天做了什么，如有相关历史记忆请关联分析）",
  "highlights": "今日亮点（最有价值的学习或成果）",
  "struggles": "今日困难（遇到的问题或挑战，如果没有则写'无'）",
  "skills_gained": ["今日涉及的技能"],
  "suggestions": "明日建议（基于今天和历史工作给出建议）"
}`, historySummary.String(), req.Date, windowSummary.String(), diffSummary.String())

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

	prompt := fmt.Sprintf(`请分析以下一周的工作记录，生成周报总结：

时间范围: %s 至 %s
总编码时长: %d 分钟
总代码变更: %d 次

每日记录:
%s

请用 JSON 格式返回（不要 markdown 代码块）:
{
  "overview": "本周整体概述（3-4句话总结这周做了什么，有什么进展）",
  "achievements": ["成就1", "成就2", "成就3"],
  "patterns": "学习模式分析（发现了什么规律或趋势）",
  "suggestions": "下周建议（基于本周情况给出具体可行的建议）",
  "top_skills": ["本周最常用的技能"]
}`, req.StartDate, req.EndDate, req.TotalCoding, req.TotalDiffs, dailyDetails.String())

	messages := []Message{
		{Role: "system", Content: "你是一个个人成长助手，帮助用户回顾一周的工作和学习，提供有深度的分析和建设性的反馈。回复必须是纯 JSON。"},
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
