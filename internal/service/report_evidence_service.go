package service

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yuqie6/WorkMirror/internal/schema"
)

// Report evidence mapping (Claim -> Sessions)
// 目标：让“每条结论”能一跳回到 Session（Evidence First），同时保持实现轻量可解释。

type EvidenceSessionRef struct {
	ID           int64
	Date         string
	TimeRange    string
	Category     string
	Summary      string
	EvidenceHint string // diff+browser | diff | browser | window_only
	EndTime      int64  // 用于排序（不对外暴露）
}

type ClaimEvidence struct {
	Claim    string
	Sessions []EvidenceSessionRef
}

type DailySummaryEvidence struct {
	Summary    []EvidenceSessionRef
	Highlights []ClaimEvidence
	Struggles  []ClaimEvidence
}

type PeriodSummaryEvidence struct {
	Overview     []EvidenceSessionRef
	Achievements []ClaimEvidence
	Patterns     []EvidenceSessionRef
	Suggestions  []ClaimEvidence
}

type sessionDoc struct {
	ref          EvidenceSessionRef
	searchText   string
	skillsLower  map[string]struct{}
	diffCount    int
	browserCount int
}

var reClaimToken = regexp.MustCompile(`[\p{Han}]{2,}|[A-Za-z][A-Za-z0-9+.#_-]{1,}`)

func parseTextItems(s string) []string {
	raw := strings.TrimSpace(s)
	if raw == "" || raw == "无" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case '\n', '\r', ';', '；', '。':
			return true
		default:
			return false
		}
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" || v == "无" {
			continue
		}
		out = append(out, v)
	}
	return out
}

func tokenizeClaim(s string) []string {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return nil
	}
	matches := reClaimToken.FindAllString(raw, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		t := strings.TrimSpace(m)
		if t == "" {
			continue
		}
		// ASCII token lowercasing；中文不受影响。
		tLower := strings.ToLower(t)
		if _, ok := seen[tLower]; ok {
			continue
		}
		seen[tLower] = struct{}{}
		out = append(out, tLower)
	}
	return out
}

func EvidenceHintFromCounts(diffCount, browserCount int) string {
	if diffCount > 0 && browserCount > 0 {
		return "diff+browser"
	}
	if diffCount > 0 {
		return "diff"
	}
	if browserCount > 0 {
		return "browser"
	}
	return "window_only"
}

func buildSessionDocs(ctx context.Context, diffRepo DiffRepository, sessions []schema.Session) ([]sessionDoc, error) {
	if len(sessions) == 0 {
		return nil, nil
	}

	// 批量拉取 diff 文件名，提升关键词覆盖率（不依赖 AI 语义）。
	diffIDsAll := make([]int64, 0, 256)
	diffIDsBySession := make(map[int64][]int64, len(sessions))
	for i := range sessions {
		ids := schema.GetInt64Slice(sessions[i].Metadata, "diff_ids")
		if len(ids) == 0 {
			continue
		}
		diffIDsBySession[sessions[i].ID] = ids
		diffIDsAll = append(diffIDsAll, ids...)
	}

	diffNameByID := map[int64]string{}
	if diffRepo != nil && len(diffIDsAll) > 0 {
		uniq := make(map[int64]struct{}, len(diffIDsAll))
		ids := make([]int64, 0, len(diffIDsAll))
		for _, id := range diffIDsAll {
			if id <= 0 {
				continue
			}
			if _, ok := uniq[id]; ok {
				continue
			}
			uniq[id] = struct{}{}
			ids = append(ids, id)
		}
		if len(ids) > 0 {
			diffs, err := diffRepo.GetByIDs(ctx, ids)
			if err == nil {
				for _, d := range diffs {
					if strings.TrimSpace(d.FileName) != "" {
						diffNameByID[d.ID] = d.FileName
					}
				}
			}
		}
	}

	docs := make([]sessionDoc, 0, len(sessions))
	for _, s := range sessions {
		diffCount := len(schema.GetInt64Slice(s.Metadata, "diff_ids"))
		browserCount := len(schema.GetInt64Slice(s.Metadata, "browser_event_ids"))

		timeRange := strings.TrimSpace(s.TimeRange)
		if timeRange == "" {
			timeRange = FormatTimeRangeMs(s.StartTime, s.EndTime)
		}

		ref := EvidenceSessionRef{
			ID:           s.ID,
			Date:         s.Date,
			TimeRange:    timeRange,
			Category:     strings.TrimSpace(s.Category),
			Summary:      strings.TrimSpace(s.Summary),
			EvidenceHint: EvidenceHintFromCounts(diffCount, browserCount),
			EndTime:      s.EndTime,
		}

		var b strings.Builder
		b.WriteString(ref.Category)
		b.WriteByte(' ')
		b.WriteString(strings.TrimSpace(s.PrimaryApp))
		b.WriteByte(' ')
		b.WriteString(ref.Summary)

		skillSet := make(map[string]struct{}, len(s.SkillsInvolved))
		for _, sk := range s.SkillsInvolved {
			v := strings.TrimSpace(sk)
			if v == "" {
				continue
			}
			l := strings.ToLower(v)
			skillSet[l] = struct{}{}
			b.WriteByte(' ')
			b.WriteString(v)
		}

		if ids := diffIDsBySession[s.ID]; len(ids) > 0 {
			for _, id := range ids {
				if name := strings.TrimSpace(diffNameByID[id]); name != "" {
					b.WriteByte(' ')
					b.WriteString(name)
				}
			}
		}

		docs = append(docs, sessionDoc{
			ref:          ref,
			searchText:   strings.ToLower(b.String()),
			skillsLower:  skillSet,
			diffCount:    diffCount,
			browserCount: browserCount,
		})
	}

	sort.Slice(docs, func(i, j int) bool { return docs[i].ref.EndTime > docs[j].ref.EndTime })
	return docs, nil
}

func pickTopSessionsByEvidence(docs []sessionDoc, limit int) []EvidenceSessionRef {
	if limit <= 0 || len(docs) == 0 {
		return nil
	}
	type scored struct {
		ref   EvidenceSessionRef
		score int
	}
	scores := make([]scored, 0, len(docs))
	for _, d := range docs {
		s := 0
		if d.diffCount > 0 {
			s += 2
		}
		if d.browserCount > 0 {
			s += 1
		}
		if strings.TrimSpace(d.ref.Summary) != "" {
			s += 1
		}
		scores = append(scores, scored{ref: d.ref, score: s})
	}
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].score != scores[j].score {
			return scores[i].score > scores[j].score
		}
		if scores[i].ref.EndTime != scores[j].ref.EndTime {
			return scores[i].ref.EndTime > scores[j].ref.EndTime
		}
		return scores[i].ref.ID > scores[j].ref.ID
	})
	if len(scores) > limit {
		scores = scores[:limit]
	}
	out := make([]EvidenceSessionRef, 0, len(scores))
	for _, s := range scores {
		out = append(out, s.ref)
	}
	return out
}

func linkClaimToSessions(claim string, docs []sessionDoc, topK int) ClaimEvidence {
	c := strings.TrimSpace(claim)
	if c == "" {
		return ClaimEvidence{Claim: claim}
	}
	tokens := tokenizeClaim(c)
	if len(tokens) == 0 || len(docs) == 0 {
		return ClaimEvidence{Claim: claim}
	}

	type scored struct {
		ref   EvidenceSessionRef
		score int
	}
	scoredList := make([]scored, 0, len(docs))

	for _, d := range docs {
		score := 0
		for _, t := range tokens {
			if t == "" {
				continue
			}
			if strings.Contains(d.searchText, t) {
				score += 1
			}
			if _, ok := d.skillsLower[t]; ok {
				score += 2
			}
		}
		// 轻量 tie-break：更强证据的 session 优先展示。
		if score > 0 {
			if d.diffCount > 0 {
				score += 1
			}
			if d.browserCount > 0 {
				score += 1
			}
			scoredList = append(scoredList, scored{ref: d.ref, score: score})
		}
	}

	if len(scoredList) == 0 {
		return ClaimEvidence{Claim: claim}
	}

	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].score != scoredList[j].score {
			return scoredList[i].score > scoredList[j].score
		}
		if scoredList[i].ref.EndTime != scoredList[j].ref.EndTime {
			return scoredList[i].ref.EndTime > scoredList[j].ref.EndTime
		}
		return scoredList[i].ref.ID > scoredList[j].ref.ID
	})

	if topK <= 0 {
		topK = 5
	}
	if len(scoredList) > topK {
		scoredList = scoredList[:topK]
	}
	out := make([]EvidenceSessionRef, 0, len(scoredList))
	for _, s := range scoredList {
		out = append(out, s.ref)
	}
	return ClaimEvidence{Claim: claim, Sessions: out}
}

func linkClaims(claims []string, docs []sessionDoc, topK int) []ClaimEvidence {
	if len(claims) == 0 {
		return nil
	}
	out := make([]ClaimEvidence, 0, len(claims))
	for _, c := range claims {
		ce := linkClaimToSessions(c, docs, topK)
		out = append(out, ce)
	}
	return out
}

func BuildDailySummaryEvidence(ctx context.Context, sessionRepo SessionRepository, diffRepo DiffRepository, summary *schema.DailySummary) (*DailySummaryEvidence, error) {
	if summary == nil || sessionRepo == nil {
		return nil, nil
	}

	sessions, err := sessionRepo.GetByDate(ctx, summary.Date)
	if err != nil {
		return nil, err
	}
	docs, err := buildSessionDocs(ctx, diffRepo, sessions)
	if err != nil {
		return nil, err
	}

	ev := &DailySummaryEvidence{
		// summary 是段落，不做强行逐句映射；给一个“最可能相关”的 top sessions。
		Summary:    pickTopSessionsByEvidence(docs, 5),
		Highlights: linkClaims(parseTextItems(summary.Highlights), docs, 4),
		Struggles:  linkClaims(parseTextItems(summary.Struggles), docs, 4),
	}
	return ev, nil
}

func BuildPeriodSummaryEvidence(ctx context.Context, sessionRepo SessionRepository, diffRepo DiffRepository, startDate, endDate string, achievements []string, overview, patterns, suggestions string) (*PeriodSummaryEvidence, error) {
	if sessionRepo == nil {
		return nil, nil
	}
	loc := time.Local
	startT, err := time.ParseInLocation("2006-01-02", startDate, loc)
	if err != nil {
		return nil, err
	}
	endT, err := time.ParseInLocation("2006-01-02", endDate, loc)
	if err != nil {
		return nil, err
	}
	// endDate is inclusive
	startMs := startT.UnixMilli()
	endMs := endT.Add(24*time.Hour).UnixMilli() - 1

	sessions, err := sessionRepo.GetByTimeRange(ctx, startMs, endMs)
	if err != nil {
		return nil, err
	}
	docs, err := buildSessionDocs(ctx, diffRepo, sessions)
	if err != nil {
		return nil, err
	}

	achievementClaims := make([]string, 0, len(achievements))
	for _, a := range achievements {
		v := strings.TrimSpace(a)
		if v == "" {
			continue
		}
		achievementClaims = append(achievementClaims, v)
	}

	overviewSessions := linkClaimToSessions(overview, docs, 6).Sessions
	if len(overviewSessions) == 0 {
		overviewSessions = pickTopSessionsByEvidence(docs, 6)
	}
	patternSessions := linkClaimToSessions(patterns, docs, 6).Sessions
	if len(patternSessions) == 0 {
		patternSessions = pickTopSessionsByEvidence(docs, 6)
	}

	ev := &PeriodSummaryEvidence{
		Overview:     overviewSessions,
		Achievements: linkClaims(achievementClaims, docs, 4),
		Patterns:     patternSessions,
		Suggestions:  linkClaims(parseTextItems(suggestions), docs, 4),
	}
	return ev, nil
}
