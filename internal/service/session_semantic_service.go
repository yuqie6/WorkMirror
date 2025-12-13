package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/ai"
	"github.com/yuqie6/mirror/internal/schema"
)

// SessionSemanticService 将低层事件/证据转为“可理解会话”（语义摘要 + 技能 + 证据链）
type SessionSemanticService struct {
	analyzer    Analyzer
	sessionRepo SessionRepository
	diffRepo    DiffRepository
	eventRepo   EventRepository
	browserRepo BrowserEventRepository
	rag         RAGQuerier // 可选
}

type SessionSemanticServiceConfig struct {
	Lookback time.Duration
	Limit    int
}

func NewSessionSemanticService(
	analyzer Analyzer,
	sessionRepo SessionRepository,
	diffRepo DiffRepository,
	eventRepo EventRepository,
	browserRepo BrowserEventRepository,
) *SessionSemanticService {
	return &SessionSemanticService{
		analyzer:    analyzer,
		sessionRepo: sessionRepo,
		diffRepo:    diffRepo,
		eventRepo:   eventRepo,
		browserRepo: browserRepo,
	}
}

func (s *SessionSemanticService) SetRAG(rag RAGQuerier) {
	s.rag = rag
}

// EnrichSessionsIncremental 增量补全最近会话的语义字段（只处理 summary 为空或证据元信息缺失的会话）
func (s *SessionSemanticService) EnrichSessionsIncremental(ctx context.Context, cfg *SessionSemanticServiceConfig) (int, error) {
	if cfg == nil {
		cfg = &SessionSemanticServiceConfig{
			Lookback: 24 * time.Hour,
			Limit:    20,
		}
	}
	if cfg.Lookback <= 0 {
		cfg.Lookback = 24 * time.Hour
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 20
	}

	end := time.Now().UnixMilli()
	start := time.Now().Add(-cfg.Lookback).UnixMilli()
	sessions, err := s.sessionRepo.GetByTimeRange(ctx, start, end)
	if err != nil {
		return 0, err
	}

	updated := 0
	for _, sess := range sessions {
		if updated >= cfg.Limit {
			break
		}
		if !shouldEnrichSession(&sess) {
			continue
		}
		if err := s.enrichOne(ctx, &sess); err != nil {
			slog.Warn("补全会话语义失败", "id", sess.ID, "error", err)
			continue
		}
		updated++
	}
	return updated, nil
}

// EnrichSessionsForDate 按日期补全会话语义字段
func (s *SessionSemanticService) EnrichSessionsForDate(ctx context.Context, date string, limit int) (int, error) {
	if limit <= 0 {
		limit = 200
	}
	sessions, err := s.sessionRepo.GetByDate(ctx, date)
	if err != nil {
		return 0, err
	}

	updated := 0
	for _, sess := range sessions {
		if updated >= limit {
			break
		}
		if !shouldEnrichSession(&sess) {
			continue
		}
		if err := s.enrichOne(ctx, &sess); err != nil {
			slog.Warn("补全会话语义失败", "id", sess.ID, "error", err)
			continue
		}
		updated++
	}
	return updated, nil
}

// GetSessionsBySkill 返回某技能相关的会话（依赖 session.metadata.skill_keys 作为证据索引）
func (s *SessionSemanticService) GetSessionsBySkill(ctx context.Context, skillKey string, lookback time.Duration, limit int) ([]schema.Session, error) {
	skillKey = strings.TrimSpace(skillKey)
	if skillKey == "" {
		return nil, nil
	}
	if lookback <= 0 {
		lookback = 30 * 24 * time.Hour
	}
	if limit <= 0 {
		limit = 20
	}

	end := time.Now().UnixMilli()
	start := time.Now().Add(-lookback).UnixMilli()
	sessions, err := s.sessionRepo.GetByTimeRange(ctx, start, end)
	if err != nil {
		return nil, err
	}

	matched := make([]schema.Session, 0, limit)
	for i := len(sessions) - 1; i >= 0; i-- { // 最近优先
		if len(matched) >= limit {
			break
		}
		sess := sessions[i]
		keys := schema.GetStringSlice(sess.Metadata, "skill_keys")
		for _, k := range keys {
			if k == skillKey {
				matched = append(matched, sess)
				break
			}
		}
	}

	sort.Slice(matched, func(i, j int) bool { return matched[i].StartTime > matched[j].StartTime })
	return matched, nil
}

func shouldEnrichSession(sess *schema.Session) bool {
	if sess == nil || sess.ID == 0 {
		return false
	}
	if strings.TrimSpace(sess.Summary) == "" {
		return true
	}
	// 证据索引缺失也需要补齐（用于 skill→session 追溯）
	if len(schema.GetStringSlice(sess.Metadata, "skill_keys")) == 0 && len(sess.SkillsInvolved) == 0 {
		return true
	}
	return false
}

func (s *SessionSemanticService) enrichOne(ctx context.Context, sess *schema.Session) error {
	if sess == nil || sess.ID == 0 {
		return fmt.Errorf("无效会话")
	}

	meta := sess.Metadata
	if meta == nil {
		meta = make(schema.JSONMap)
	}

	diffIDs := schema.GetInt64Slice(meta, "diff_ids")
	var diffs []schema.Diff
	var err error
	if len(diffIDs) > 0 {
		diffs, err = s.diffRepo.GetByIDs(ctx, diffIDs)
	} else {
		diffs, err = s.diffRepo.GetByTimeRange(ctx, sess.StartTime, sess.EndTime)
	}
	if err != nil {
		return err
	}
	if len(diffIDs) == 0 && len(diffs) > 0 {
		diffIDs = make([]int64, 0, len(diffs))
		for _, d := range diffs {
			diffIDs = append(diffIDs, d.ID)
		}
	}
	schema.SetInt64Slice(meta, "diff_ids", diffIDs)

	// 应用使用统计
	appStats, err := s.eventRepo.GetAppStats(ctx, sess.StartTime, sess.EndTime)
	if err != nil {
		return err
	}
	topApps := WindowEventInfosFromAppStats(appStats, DefaultTopAppsLimit)

	// 浏览事件（优先用索引 ID，否则按时间窗补全）
	browserIDs := schema.GetInt64Slice(meta, "browser_event_ids")
	var browserEvents []schema.BrowserEvent
	if len(browserIDs) > 0 {
		browserEvents, err = s.browserRepo.GetByIDs(ctx, browserIDs)
	} else {
		browserEvents, err = s.browserRepo.GetByTimeRange(ctx, sess.StartTime, sess.EndTime)
	}
	if err != nil {
		return err
	}
	if len(browserIDs) == 0 && len(browserEvents) > 0 {
		browserIDs = make([]int64, 0, len(browserEvents))
		for _, e := range browserEvents {
			browserIDs = append(browserIDs, e.ID)
		}
	}
	schema.SetInt64Slice(meta, "browser_event_ids", browserIDs)

	// 技能聚合：从已分析 Diff 归因（避免凭空推断）
	skillNameToKey := make(map[string]string)
	skillKeyToName := make(map[string]string)
	for _, d := range diffs {
		for _, name := range d.SkillsDetected {
			n := strings.TrimSpace(name)
			if n == "" {
				continue
			}
			key := normalizeKey(n)
			if key == "" {
				continue
			}
			skillNameToKey[strings.ToLower(n)] = key
			if _, ok := skillKeyToName[key]; !ok {
				skillKeyToName[key] = n
			}
		}
	}

	skillKeys := make([]string, 0, len(skillKeyToName))
	for k := range skillKeyToName {
		skillKeys = append(skillKeys, k)
	}
	sort.Strings(skillKeys)

	skillNames := make([]string, 0, len(skillKeys))
	for _, k := range skillKeys {
		skillNames = append(skillNames, skillKeyToName[k])
	}

	meta["skill_keys"] = skillKeys

	// 浏览信息截断，用于 prompt/展示
	browserInfos := make([]ai.BrowserInfo, 0, 10)
	domainCount := make(map[string]int)
	for _, e := range browserEvents {
		if e.Domain != "" {
			domainCount[e.Domain]++
		}
		if len(browserInfos) < 10 {
			browserInfos = append(browserInfos, ai.BrowserInfo{
				Domain: e.Domain,
				Title:  truncateRunes(strings.TrimSpace(e.Title), 80),
				URL:    truncateRunes(strings.TrimSpace(e.URL), 200),
			})
		}
	}
	topDomains := topKeysByCount(domainCount, 6)
	if len(topDomains) > 0 {
		meta["top_domains"] = topDomains
	}

	diffInfos := make([]ai.DiffInfo, 0, len(diffs))
	for _, d := range diffs {
		diffInfos = append(diffInfos, ai.DiffInfo{
			FileName:     d.FileName,
			Language:     d.Language,
			Insight:      truncateRunes(strings.TrimSpace(d.AIInsight), 160),
			DiffContent:  "",
			LinesChanged: d.LinesAdded + d.LinesDeleted,
		})
	}

	// RAG 记忆（可选）：把引用也写进 metadata，便于 UI 展示“用到了哪些历史记忆”
	memories := []string{}
	if s.rag != nil && (len(diffs) > 0 || len(skillNames) > 0) {
		query := ""
		for _, d := range diffs {
			if strings.TrimSpace(d.AIInsight) != "" {
				query = "会话 " + d.AIInsight
				break
			}
		}
		if query == "" && len(skillNames) > 0 {
			query = "会话 " + skillNames[0]
		}
		if query != "" {
			if results, err := s.rag.Query(ctx, query, 3); err == nil && len(results) > 0 {
				refs := make([]map[string]any, 0, len(results))
				for _, r := range results {
					content := truncateRunes(strings.TrimSpace(r.Content), 180)
					if content != "" {
						memories = append(memories, content)
					}
					refs = append(refs, map[string]any{
						"type":       r.Type,
						"date":       r.Date,
						"similarity": r.Similarity,
						"content":    content,
					})
				}
				meta["rag_refs"] = refs
			}
		}
	}

	// 生成语义摘要（有 AI 就用 AI；否则规则降级）
	summary := strings.TrimSpace(sess.Summary)
	category := strings.TrimSpace(sess.Category)
	skillsInvolved := []string(nil)

	if len(skillNames) > 0 {
		skillsInvolved = append([]string(nil), skillNames...)
	}

	if s.analyzer != nil && strings.TrimSpace(summary) == "" {
		req := &ai.SessionSummaryRequest{
			SessionID:  sess.ID,
			Date:       sess.Date,
			TimeRange:  sess.TimeRange,
			PrimaryApp: sess.PrimaryApp,
			AppUsage:   topApps,
			Diffs:      diffInfos,
			Browser:    browserInfos,
			SkillsHint: skillNames,
			Memories:   memories,
		}
		if res, err := s.analyzer.GenerateSessionSummary(ctx, req); err == nil && res != nil {
			summary = strings.TrimSpace(res.Summary)
			if strings.TrimSpace(res.Category) != "" {
				category = strings.TrimSpace(res.Category)
			}
			if len(res.SkillsInvolved) > 0 {
				skillsInvolved = uniqueNonEmpty(res.SkillsInvolved, 8)
				keys := make([]string, 0, len(skillsInvolved))
				for _, n := range skillsInvolved {
					k := normalizeKey(n)
					if k != "" {
						keys = append(keys, k)
					}
				}
				meta["skill_keys"] = uniqueNonEmpty(keys, 16)
			}
			if len(res.Tags) > 0 {
				meta["tags"] = uniqueNonEmpty(res.Tags, 6)
			}
		}
	}

	if summary == "" {
		summary = fallbackSessionSummary(sess, diffs, topDomains, skillNames)
	}
	if category == "" {
		category = fallbackSessionCategory(diffs, browserEvents)
	}

	update := schema.SessionSemanticUpdate{
		TimeRange:      sess.TimeRange,
		Category:       category,
		Summary:        summary,
		SkillsInvolved: skillsInvolved,
		Metadata:       meta,
	}
	return s.sessionRepo.UpdateSemantic(ctx, sess.ID, update)
}

func fallbackSessionCategory(diffs []schema.Diff, browser []schema.BrowserEvent) string {
	hasDiff := len(diffs) > 0
	hasBrowser := len(browser) > 0
	switch {
	case hasDiff && hasBrowser:
		return "exploration"
	case hasDiff:
		return "technical"
	case hasBrowser:
		return "learning"
	default:
		return "other"
	}
}

func fallbackSessionSummary(sess *schema.Session, diffs []schema.Diff, topDomains []string, skills []string) string {
	parts := []string{}
	if len(skills) > 0 {
		parts = append(parts, "围绕 "+strings.Join(skills[:minInt(3, len(skills))], "、"))
	}
	if len(diffs) > 0 {
		langs := map[string]struct{}{}
		for _, d := range diffs {
			if strings.TrimSpace(d.Language) != "" {
				langs[strings.TrimSpace(d.Language)] = struct{}{}
			}
		}
		langList := make([]string, 0, len(langs))
		for l := range langs {
			langList = append(langList, l)
		}
		sort.Strings(langList)
		if len(langList) > 0 {
			parts = append(parts, "进行了 "+strings.Join(langList[:minInt(2, len(langList))], "/")+" 代码变更")
		} else {
			parts = append(parts, "进行了代码变更")
		}
	}
	if sess != nil && strings.TrimSpace(sess.PrimaryApp) != "" {
		parts = append(parts, "主要在 "+sess.PrimaryApp)
	}
	if len(topDomains) > 0 {
		parts = append(parts, "查阅 "+strings.Join(topDomains[:minInt(2, len(topDomains))], "、"))
	}
	if len(parts) == 0 {
		return "完成了一段工作/学习会话"
	}
	return strings.Join(parts, "，") + "。"
}

func topKeysByCount(m map[string]int, limit int) []string {
	type kv struct {
		k string
		v int
	}
	items := make([]kv, 0, len(m))
	for k, v := range m {
		if strings.TrimSpace(k) == "" {
			continue
		}
		items = append(items, kv{k: k, v: v})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].v == items[j].v {
			return items[i].k < items[j].k
		}
		return items[i].v > items[j].v
	})
	if limit <= 0 || limit > len(items) {
		limit = len(items)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, items[i].k)
	}
	return out
}

func uniqueNonEmpty(in []string, limit int) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		v := strings.TrimSpace(s)
		if v == "" {
			continue
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
