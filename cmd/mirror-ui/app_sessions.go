package main

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/model"
)

// SessionDTO 会话摘要 DTO（用于列表）
type SessionDTO struct {
	ID             int64    `json:"id"`
	Date           string   `json:"date"`
	StartTime      int64    `json:"start_time"`
	EndTime        int64    `json:"end_time"`
	TimeRange      string   `json:"time_range"`
	PrimaryApp     string   `json:"primary_app"`
	Category       string   `json:"category"`
	Summary        string   `json:"summary"`
	SkillsInvolved []string `json:"skills_involved"`
	DiffCount      int      `json:"diff_count"`
	BrowserCount   int      `json:"browser_count"`
}

type SessionAppUsageDTO struct {
	AppName       string `json:"app_name"`
	TotalDuration int    `json:"total_duration"` // 秒
}

type SessionDiffDTO struct {
	ID           int64    `json:"id"`
	FileName     string   `json:"file_name"`
	Language     string   `json:"language"`
	Insight      string   `json:"insight"`
	Skills       []string `json:"skills"`
	LinesAdded   int      `json:"lines_added"`
	LinesDeleted int      `json:"lines_deleted"`
	Timestamp    int64    `json:"timestamp"`
}

type SessionBrowserEventDTO struct {
	ID        int64  `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Domain    string `json:"domain"`
	Title     string `json:"title"`
	URL       string `json:"url"`
}

type SessionDetailDTO struct {
	SessionDTO
	Tags     []string                 `json:"tags"`
	RAGRefs  []map[string]any         `json:"rag_refs"`
	AppUsage []SessionAppUsageDTO     `json:"app_usage"`
	Diffs    []SessionDiffDTO         `json:"diffs"`
	Browser  []SessionBrowserEventDTO `json:"browser"`
}

type SessionBuildResultDTO struct {
	Created int `json:"created"`
}

type SessionEnrichResultDTO struct {
	Enriched int `json:"enriched"`
}

// BuildSessionsForDate 按日期切分会话（幂等，可重复调用）
func (a *App) BuildSessionsForDate(date string) (*SessionBuildResultDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Services.Sessions == nil {
		return nil, errors.New("会话服务未初始化")
	}

	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	created, err := a.core.Services.Sessions.BuildSessionsForDate(ctx, strings.TrimSpace(date))
	if err != nil {
		return nil, err
	}
	return &SessionBuildResultDTO{Created: created}, nil
}

// RebuildSessionsForDate 重建某天会话：创建一个更高的切分版本以覆盖展示，从而“清理旧碎片”
func (a *App) RebuildSessionsForDate(date string) (*SessionBuildResultDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Services.Sessions == nil {
		return nil, errors.New("会话服务未初始化")
	}

	ctx, cancel := context.WithTimeout(a.ctx, 60*time.Second)
	defer cancel()

	created, err := a.core.Services.Sessions.RebuildSessionsForDate(ctx, strings.TrimSpace(date))
	if err != nil {
		return nil, err
	}
	return &SessionBuildResultDTO{Created: created}, nil
}

// EnrichSessionsForDate 为日期内会话生成语义摘要/技能/证据索引（DeepSeek 未配置时自动降级）
func (a *App) EnrichSessionsForDate(date string) (*SessionEnrichResultDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Services.SessionSemantic == nil {
		return nil, errors.New("会话语义服务未初始化")
	}

	ctx, cancel := context.WithTimeout(a.ctx, 60*time.Second)
	defer cancel()

	enriched, err := a.core.Services.SessionSemantic.EnrichSessionsForDate(ctx, strings.TrimSpace(date), 200)
	if err != nil {
		return nil, err
	}
	return &SessionEnrichResultDTO{Enriched: enriched}, nil
}

// GetSessionsByDate 获取日期内会话列表
func (a *App) GetSessionsByDate(date string) ([]SessionDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Repos.Session == nil {
		return nil, errors.New("会话仓储未初始化")
	}

	sessions, err := a.core.Repos.Session.GetByDate(a.ctx, strings.TrimSpace(date))
	if err != nil {
		return nil, err
	}

	result := make([]SessionDTO, 0, len(sessions))
	for _, s := range sessions {
		diffIDs := toInt64Slice(s.Metadata, "diff_ids")
		browserIDs := toInt64Slice(s.Metadata, "browser_event_ids")
		timeRange := s.TimeRange
		if strings.TrimSpace(timeRange) == "" {
			timeRange = formatTimeRangeMs(s.StartTime, s.EndTime)
		}
		dto := SessionDTO{
			ID:             s.ID,
			Date:           s.Date,
			StartTime:      s.StartTime,
			EndTime:        s.EndTime,
			TimeRange:      timeRange,
			PrimaryApp:     s.PrimaryApp,
			Category:       s.Category,
			Summary:        s.Summary,
			SkillsInvolved: []string(s.SkillsInvolved),
			DiffCount:      len(diffIDs),
			BrowserCount:   len(browserIDs),
		}
		result = append(result, dto)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].StartTime < result[j].StartTime })
	return result, nil
}

// GetSessionDetail 获取会话详情（证据链展开）
func (a *App) GetSessionDetail(id int64) (*SessionDetailDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Repos.Session == nil {
		return nil, errors.New("会话仓储未初始化")
	}

	sess, err := a.core.Repos.Session.GetByID(a.ctx, id)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, errors.New("session not found")
	}

	diffIDs := toInt64Slice(sess.Metadata, "diff_ids")
	browserIDs := toInt64Slice(sess.Metadata, "browser_event_ids")

	// Diffs
	var diffs []model.Diff
	if len(diffIDs) > 0 {
		diffs, _ = a.core.Repos.Diff.GetByIDs(a.ctx, diffIDs)
	}
	diffDTOs := make([]SessionDiffDTO, 0, len(diffs))
	for _, d := range diffs {
		diffDTOs = append(diffDTOs, SessionDiffDTO{
			ID:           d.ID,
			FileName:     d.FileName,
			Language:     d.Language,
			Insight:      d.AIInsight,
			Skills:       []string(d.SkillsDetected),
			LinesAdded:   d.LinesAdded,
			LinesDeleted: d.LinesDeleted,
			Timestamp:    d.Timestamp,
		})
	}

	// Browser
	var browserEvents []model.BrowserEvent
	if len(browserIDs) > 0 {
		browserEvents, _ = a.core.Repos.Browser.GetByIDs(a.ctx, browserIDs)
	}
	browserDTOs := make([]SessionBrowserEventDTO, 0, len(browserEvents))
	for _, e := range browserEvents {
		browserDTOs = append(browserDTOs, SessionBrowserEventDTO{
			ID:        e.ID,
			Timestamp: e.Timestamp,
			Domain:    e.Domain,
			Title:     e.Title,
			URL:       e.URL,
		})
	}

	// App usage（按会话窗）
	appStats, _ := a.core.Repos.Event.GetAppStats(a.ctx, sess.StartTime, sess.EndTime)
	appUsage := make([]SessionAppUsageDTO, 0, len(appStats))
	for i, st := range appStats {
		if i >= 8 {
			break
		}
		appUsage = append(appUsage, SessionAppUsageDTO{
			AppName:       st.AppName,
			TotalDuration: st.TotalDuration,
		})
	}

	tags := toStringSlice(sess.Metadata, "tags")
	ragRefs := toMapSlice(sess.Metadata, "rag_refs")

	dto := &SessionDetailDTO{
		SessionDTO: SessionDTO{
			ID:             sess.ID,
			Date:           sess.Date,
			StartTime:      sess.StartTime,
			EndTime:        sess.EndTime,
			TimeRange:      formatTimeRangeForSession(sess),
			PrimaryApp:     sess.PrimaryApp,
			Category:       sess.Category,
			Summary:        sess.Summary,
			SkillsInvolved: []string(sess.SkillsInvolved),
			DiffCount:      len(diffIDs),
			BrowserCount:   len(browserIDs),
		},
		Tags:     tags,
		RAGRefs:  ragRefs,
		AppUsage: appUsage,
		Diffs:    diffDTOs,
		Browser:  browserDTOs,
	}
	return dto, nil
}

func formatTimeRangeForSession(sess *model.Session) string {
	if sess == nil {
		return ""
	}
	tr := strings.TrimSpace(sess.TimeRange)
	if tr != "" {
		return tr
	}
	return formatTimeRangeMs(sess.StartTime, sess.EndTime)
}

// GetSkillSessions 获取技能相关会话（用于 skill→session 追溯）
func (a *App) GetSkillSessions(skillKey string) ([]SessionDTO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.core == nil || a.core.Services.SessionSemantic == nil {
		return nil, errors.New("会话语义服务未初始化")
	}

	sessions, err := a.core.Services.SessionSemantic.GetSessionsBySkill(a.ctx, strings.TrimSpace(skillKey), 30*24*time.Hour, 10)
	if err != nil {
		return nil, err
	}

	result := make([]SessionDTO, 0, len(sessions))
	for _, s := range sessions {
		diffIDs := toInt64Slice(s.Metadata, "diff_ids")
		browserIDs := toInt64Slice(s.Metadata, "browser_event_ids")
		timeRange := s.TimeRange
		if strings.TrimSpace(timeRange) == "" {
			timeRange = formatTimeRangeMs(s.StartTime, s.EndTime)
		}
		result = append(result, SessionDTO{
			ID:             s.ID,
			Date:           s.Date,
			StartTime:      s.StartTime,
			EndTime:        s.EndTime,
			TimeRange:      timeRange,
			PrimaryApp:     s.PrimaryApp,
			Category:       s.Category,
			Summary:        s.Summary,
			SkillsInvolved: []string(s.SkillsInvolved),
			DiffCount:      len(diffIDs),
			BrowserCount:   len(browserIDs),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartTime > result[j].StartTime })
	return result, nil
}
