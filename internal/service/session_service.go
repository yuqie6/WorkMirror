package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/yuqie6/WorkMirror/internal/repository"
	"github.com/yuqie6/WorkMirror/internal/schema"
)

// SessionService 基于事件流切分会话（工程规则优先）
type SessionService struct {
	eventRepo       EventRepository
	diffRepo        DiffRepository
	browserRepo     BrowserEventRepository
	sessionRepo     SessionRepository
	sessionDiffRepo SessionDiffRepository
	cfg             *SessionServiceConfig

	lastSplitAt  atomic.Int64
	splitErrors  atomic.Int64
	lastErrorAt  atomic.Int64
	lastErrorMsg atomic.Value // string
}

// SessionServiceConfig 会话服务配置
type SessionServiceConfig struct {
	IdleGapMinutes    int // 空闲间隔分钟数，超过则切分会话
	MinSessionMinutes int // 会话最小时长（分钟），低于此值的会话将被过滤
}

// NewSessionService 创建会话服务
func NewSessionService(
	eventRepo EventRepository,
	diffRepo DiffRepository,
	browserRepo BrowserEventRepository,
	sessionRepo SessionRepository,
	sessionDiffRepo SessionDiffRepository,
	cfg *SessionServiceConfig,
) *SessionService {
	if cfg == nil {
		cfg = &SessionServiceConfig{
			IdleGapMinutes:    10, // 增加到10分钟，减少碎片化
			MinSessionMinutes: 2,  // 过滤掉小于2分钟的碎片会话
		}
	}
	if cfg.IdleGapMinutes <= 0 {
		cfg.IdleGapMinutes = 10
	}
	if cfg.MinSessionMinutes <= 0 {
		cfg.MinSessionMinutes = 2
	}
	return &SessionService{
		eventRepo:       eventRepo,
		diffRepo:        diffRepo,
		browserRepo:     browserRepo,
		sessionRepo:     sessionRepo,
		sessionDiffRepo: sessionDiffRepo,
		cfg:             cfg,
	}
}

// BuildSessionsIncremental 从最近一次会话结束处增量切分
func (s *SessionService) BuildSessionsIncremental(ctx context.Context) (int, error) {
	last, err := s.sessionRepo.GetLastSession(ctx)
	if err != nil {
		s.noteError(err)
		return 0, err
	}
	start := int64(0)
	if last != nil && last.EndTime > 0 {
		start = last.EndTime
	} else {
		// 冷启动：避免从 0 纪元扫全库，先回溯最近 24h
		start = time.Now().Add(-24 * time.Hour).UnixMilli()
	}
	end := time.Now().UnixMilli()
	created, buildErr := s.BuildSessionsForRange(ctx, start, end)
	if buildErr != nil {
		s.noteError(buildErr)
		return 0, buildErr
	}
	if created > 0 {
		s.lastSplitAt.Store(time.Now().UnixMilli())
	}
	return created, nil
}

// BuildSessionsForDate 按日期全量切分
func (s *SessionService) BuildSessionsForDate(ctx context.Context, date string) (int, error) {
	loc := time.Local
	t, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return 0, fmt.Errorf("解析日期失败: %w", err)
	}
	start := t.UnixMilli()
	end := t.Add(24*time.Hour).UnixMilli() - 1
	created, err := s.buildSessionsForRange(ctx, start, end, nil)
	if err != nil {
		s.noteError(err)
		return 0, err
	}
	if created > 0 {
		s.lastSplitAt.Store(time.Now().UnixMilli())
	}
	return created, nil
}

// RebuildSessionsForDate 重建某天会话：创建一个更高的切分版本，以“覆盖展示”方式清理旧碎片（不删除旧数据）
func (s *SessionService) RebuildSessionsForDate(ctx context.Context, date string) (int, error) {
	loc := time.Local
	t, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return 0, fmt.Errorf("解析日期失败: %w", err)
	}
	start := t.UnixMilli()
	end := t.Add(24*time.Hour).UnixMilli() - 1

	targetDate := strings.TrimSpace(date)
	created, err := s.buildSessionsForRange(ctx, start, end, func(d string, max int) int {
		if d == targetDate {
			if max <= 0 {
				return 1
			}
			return max + 1
		}
		if max <= 0 {
			return 1
		}
		return max
	})
	if err != nil {
		s.noteError(err)
		return 0, err
	}
	if created > 0 {
		s.lastSplitAt.Store(time.Now().UnixMilli())
	}
	return created, nil
}

// BuildSessionsForRange 按时间范围切分并写入 sessions 表
func (s *SessionService) BuildSessionsForRange(ctx context.Context, startTime, endTime int64) (int, error) {
	created, err := s.buildSessionsForRange(ctx, startTime, endTime, nil)
	if err != nil {
		s.noteError(err)
		return 0, err
	}
	if created > 0 {
		s.lastSplitAt.Store(time.Now().UnixMilli())
	}
	return created, nil
}

// buildSessionsForRange 内部方法：按时间范围切分会话并写入
func (s *SessionService) buildSessionsForRange(
	ctx context.Context,
	startTime, endTime int64,
	versionStrategy func(date string, maxVersion int) int,
) (int, error) {
	if startTime >= endTime {
		return 0, nil
	}

	events, err := s.eventRepo.GetByTimeRange(ctx, startTime, endTime)
	if err != nil {
		return 0, err
	}
	diffs, err := s.diffRepo.GetByTimeRange(ctx, startTime, endTime)
	if err != nil {
		return 0, err
	}
	browserEvents, err := s.browserRepo.GetByTimeRange(ctx, startTime, endTime)
	if err != nil {
		return 0, err
	}

	sessions := s.splitSessions(events, diffs, startTime, endTime)
	if len(sessions) == 0 {
		return 0, nil
	}

	// 绑定浏览器事件（不参与切分）
	s.attachBrowserEvents(sessions, browserEvents)
	for _, sess := range sessions {
		if sess == nil {
			continue
		}
		if sess.Metadata == nil {
			sess.Metadata = make(schema.JSONMap)
		}
		diffCount := len(getSessionDiffIDs(sess.Metadata))
		browserCount := len(getSessionBrowserEventIDs(sess.Metadata))
		setSessionMetaString(sess.Metadata, sessionMetaEvidenceHint, EvidenceHintFromCounts(diffCount, browserCount))
		setSessionMetaString(sess.Metadata, sessionMetaSemanticVersion, "v1")
	}

	if err := s.assignSessionVersions(ctx, sessions, versionStrategy); err != nil {
		return 0, err
	}

	created := 0
	for _, sess := range sessions {
		createdNow, err := s.sessionRepo.Create(ctx, sess)
		if err != nil {
			slog.Warn("创建会话失败", "error", err)
			continue
		}
		if !createdNow {
			// 已存在的会话不重复写入证据关联，避免重复数据。
			continue
		}
		if s.sessionDiffRepo != nil && sess.Metadata != nil {
			diffIDs := getSessionDiffIDs(sess.Metadata)
			if len(diffIDs) > 0 {
				_ = s.sessionDiffRepo.BatchInsert(ctx, sess.ID, diffIDs)
			}
		}
		created++
	}
	if created > 0 {
		slog.Info("会话切分完成", "created", created, "start", startTime, "end", endTime)
	}
	return created, nil
}

type SessionServiceStats struct {
	LastSplitAt int64  `json:"last_split_at"`
	SplitErrors int64  `json:"split_errors"`
	LastErrorAt int64  `json:"last_error_at"`
	LastError   string `json:"last_error"`
}

type EvidenceRepairResult struct {
	OrphanDiffs      int `json:"orphan_diffs"`
	OrphanBrowser    int `json:"orphan_browser"`
	AttachedDiffs    int `json:"attached_diffs"`
	AttachedBrowser  int `json:"attached_browser"`
	UpdatedSessions  int `json:"updated_sessions"`
	AttachGapMinutes int `json:"attach_gap_minutes"`
}

func (s *SessionService) Stats() SessionServiceStats {
	if s == nil {
		return SessionServiceStats{}
	}
	raw := s.lastErrorMsg.Load()
	msg, _ := raw.(string)
	return SessionServiceStats{
		LastSplitAt: s.lastSplitAt.Load(),
		SplitErrors: s.splitErrors.Load(),
		LastErrorAt: s.lastErrorAt.Load(),
		LastError:   msg,
	}
}

func (s *SessionService) noteError(err error) {
	if s == nil || err == nil {
		return
	}
	s.splitErrors.Add(1)
	s.lastErrorAt.Store(time.Now().UnixMilli())
	s.lastErrorMsg.Store(err.Error())
}

// RepairEvidenceForDate 尝试把“未归并证据”挂回到最邻近的 Session（不创建新 session、不删除旧数据）。
// v0.3 P0：先解决“Evidence First 断链”与 orphan 指标长期不收敛问题。
func (s *SessionService) RepairEvidenceForDate(ctx context.Context, date string, attachGapMinutes, limit int) (EvidenceRepairResult, error) {
	if s == nil || s.sessionRepo == nil {
		return EvidenceRepairResult{}, nil
	}
	if attachGapMinutes <= 0 {
		attachGapMinutes = 10
	}
	if limit <= 0 {
		limit = 500
	}

	start, end, err := repository.DayRange(date)
	if err != nil {
		return EvidenceRepairResult{}, err
	}

	sessions, err := s.sessionRepo.GetByTimeRange(ctx, start, end)
	if err != nil {
		return EvidenceRepairResult{}, err
	}
	if len(sessions) == 0 {
		return EvidenceRepairResult{AttachGapMinutes: attachGapMinutes}, nil
	}

	type sessRef struct {
		id    int64
		start int64
		end   int64
		meta  schema.JSONMap
	}
	refs := make([]sessRef, 0, len(sessions))
	refDiffIDs := make(map[int64]struct{}, 512)
	refBrowserIDs := make(map[int64]struct{}, 512)
	for _, sess := range sessions {
		meta := sess.Metadata
		refs = append(refs, sessRef{id: sess.ID, start: sess.StartTime, end: sess.EndTime, meta: meta})
		for _, id := range getSessionDiffIDs(meta) {
			if id > 0 {
				refDiffIDs[id] = struct{}{}
			}
		}
		for _, id := range getSessionBrowserEventIDs(meta) {
			if id > 0 {
				refBrowserIDs[id] = struct{}{}
			}
		}
	}

	findBestSession := func(ts int64) (sessRef, int64, bool) {
		var best sessRef
		bestGap := int64(0)
		found := false
		for _, r := range refs {
			if r.id == 0 {
				continue
			}
			gap := int64(0)
			switch {
			case ts < r.start:
				gap = r.start - ts
			case ts > r.end:
				gap = ts - r.end
			default:
				gap = 0
			}
			if !found || gap < bestGap {
				best = r
				bestGap = gap
				found = true
				if bestGap == 0 {
					break
				}
			}
		}
		return best, bestGap, found
	}

	gapMs := int64(attachGapMinutes) * 60 * 1000
	updatedMeta := make(map[int64]schema.JSONMap, 16)
	addedDiffIDs := make(map[int64][]int64, 16)

	result := EvidenceRepairResult{AttachGapMinutes: attachGapMinutes}

	if s.diffRepo != nil {
		diffs, err := s.diffRepo.GetByTimeRange(ctx, start, end)
		if err != nil {
			return EvidenceRepairResult{}, err
		}
		for _, d := range diffs {
			if d.ID <= 0 {
				continue
			}
			if _, ok := refDiffIDs[d.ID]; ok {
				continue
			}
			result.OrphanDiffs++
			if limit > 0 && result.AttachedDiffs >= limit {
				continue
			}
			best, gap, ok := findBestSession(d.Timestamp)
			if !ok || gap > gapMs {
				continue
			}
			meta := updatedMeta[best.id]
			if meta == nil {
				meta = best.meta
				if meta == nil {
					meta = make(schema.JSONMap)
				}
			}
			ids := append(getSessionDiffIDs(meta), d.ID)
			setSessionDiffIDs(meta, ids)
			setSessionMetaString(meta, sessionMetaEvidenceHint, EvidenceHintFromCounts(len(getSessionDiffIDs(meta)), len(getSessionBrowserEventIDs(meta))))
			setSessionMetaString(meta, sessionMetaSemanticVersion, "v1")
			updatedMeta[best.id] = meta
			addedDiffIDs[best.id] = append(addedDiffIDs[best.id], d.ID)
			refDiffIDs[d.ID] = struct{}{}
			result.AttachedDiffs++
		}
	}

	if s.browserRepo != nil {
		events, err := s.browserRepo.GetByTimeRange(ctx, start, end)
		if err != nil {
			return EvidenceRepairResult{}, err
		}
		for _, e := range events {
			if e.ID <= 0 {
				continue
			}
			if _, ok := refBrowserIDs[e.ID]; ok {
				continue
			}
			result.OrphanBrowser++
			if limit > 0 && (result.AttachedDiffs+result.AttachedBrowser) >= limit {
				continue
			}
			best, gap, ok := findBestSession(e.Timestamp)
			if !ok || gap > gapMs {
				continue
			}
			meta := updatedMeta[best.id]
			if meta == nil {
				meta = best.meta
				if meta == nil {
					meta = make(schema.JSONMap)
				}
			}
			ids := append(getSessionBrowserEventIDs(meta), e.ID)
			setSessionBrowserEventIDs(meta, ids)
			setSessionMetaString(meta, sessionMetaEvidenceHint, EvidenceHintFromCounts(len(getSessionDiffIDs(meta)), len(getSessionBrowserEventIDs(meta))))
			setSessionMetaString(meta, sessionMetaSemanticVersion, "v1")
			updatedMeta[best.id] = meta
			refBrowserIDs[e.ID] = struct{}{}
			result.AttachedBrowser++
		}
	}

	for sessionID, meta := range updatedMeta {
		if sessionID == 0 || meta == nil {
			continue
		}
		if err := s.sessionRepo.UpdateSemantic(ctx, sessionID, schema.SessionSemanticUpdate{Metadata: meta}); err != nil {
			return EvidenceRepairResult{}, err
		}
		if s.sessionDiffRepo != nil {
			if ids := addedDiffIDs[sessionID]; len(ids) > 0 {
				_ = s.sessionDiffRepo.BatchInsert(ctx, sessionID, ids)
			}
		}
		result.UpdatedSessions++
	}

	return result, nil
}

// assignSessionVersions 为会话分配切分版本号
func (s *SessionService) assignSessionVersions(
	ctx context.Context,
	sessions []*schema.Session,
	versionStrategy func(date string, maxVersion int) int,
) error {
	if len(sessions) == 0 || s.sessionRepo == nil {
		return nil
	}
	if versionStrategy == nil {
		versionStrategy = func(_ string, max int) int {
			if max <= 0 {
				return 1
			}
			return max
		}
	}

	uniqueDates := make(map[string]struct{}, 2)
	for _, sess := range sessions {
		if sess == nil {
			continue
		}
		d := strings.TrimSpace(sess.Date)
		if d == "" {
			d = formatDate(sess.StartTime)
			sess.Date = d
		}
		uniqueDates[d] = struct{}{}
	}

	maxByDate := make(map[string]int, len(uniqueDates))
	for d := range uniqueDates {
		maxV, err := s.sessionRepo.GetMaxSessionVersionByDate(ctx, d)
		if err != nil {
			return err
		}
		maxByDate[d] = maxV
	}

	for _, sess := range sessions {
		if sess == nil {
			continue
		}
		d := strings.TrimSpace(sess.Date)
		if d == "" {
			d = formatDate(sess.StartTime)
			sess.Date = d
		}
		sess.SessionVersion = versionStrategy(d, maxByDate[d])
		if sess.SessionVersion <= 0 {
			sess.SessionVersion = 1
		}
	}
	return nil
}

// splitSessions 根据空闲间隔切分会话
func (s *SessionService) splitSessions(events []schema.Event, diffs []schema.Diff, startTime, endTime int64) []*schema.Session {
	idleMs := int64(s.cfg.IdleGapMinutes) * 60 * 1000

	// 确保按时间排序
	sort.Slice(events, func(i, j int) bool { return events[i].Timestamp < events[j].Timestamp })
	sort.Slice(diffs, func(i, j int) bool { return diffs[i].Timestamp < diffs[j].Timestamp })

	var sessions []*schema.Session

	var currentStart int64
	var lastActivityEnd int64
	appDurations := map[string]int{}
	diffIDs := make([]int64, 0, 8)
	windowSeconds := 0

	openSession := func(start int64) {
		currentStart = start
		lastActivityEnd = start
		appDurations = map[string]int{}
		diffIDs = diffIDs[:0]
		windowSeconds = 0
	}

	closeSession := func(end int64) {
		if currentStart == 0 || end <= currentStart {
			return
		}
		// 不产生纯 diff 会话：window 事件是会话锚点，否则容易因事件晚到导致碎片化/重复。
		if windowSeconds <= 0 {
			currentStart = 0
			return
		}
		primaryApp := ""
		maxDur := 0
		for app, dur := range appDurations {
			if dur > maxDur {
				maxDur = dur
				primaryApp = app
			}
		}

		meta := make(schema.JSONMap)
		setSessionDiffIDs(meta, diffIDs)

		sessions = append(sessions, &schema.Session{
			Date:       formatDate(currentStart),
			StartTime:  currentStart,
			EndTime:    end,
			PrimaryApp: primaryApp,
			TimeRange:  FormatTimeRangeMs(currentStart, end),
			Metadata:   meta,
		})
		currentStart = 0
	}

	// 没有 window events 时不切分：避免仅靠 diff 产生“碎片会话”，且 window 事件可能是晚到数据。
	if len(events) == 0 {
		return nil
	}

	openSession(events[0].Timestamp)
	diffIdx := 0

	for _, ev := range events {
		evStart := ev.Timestamp
		evEnd := ev.Timestamp + int64(ev.Duration)*1000

		// 先处理落在当前窗口开始前的 diffs（可能在 idle gap 内）
		for diffIdx < len(diffs) && diffs[diffIdx].Timestamp < evStart {
			dt := diffs[diffIdx].Timestamp
			if dt-lastActivityEnd >= idleMs {
				closeSession(lastActivityEnd)
				openSession(dt)
			} else if currentStart == 0 {
				openSession(dt)
			}
			diffIDs = append(diffIDs, diffs[diffIdx].ID)
			if dt > lastActivityEnd {
				lastActivityEnd = dt
			}
			diffIdx++
		}

		// idle hard boundary（不产生 idle session）
		if evStart-lastActivityEnd >= idleMs {
			closeSession(lastActivityEnd)
			openSession(evStart)
		} else if currentStart == 0 {
			openSession(evStart)
		}

		// 累计 app 时长
		if ev.AppName != "" && ev.Duration > 0 {
			appDurations[ev.AppName] += ev.Duration
			windowSeconds += ev.Duration
		}

		if evEnd > lastActivityEnd {
			lastActivityEnd = evEnd
		}
	}

	// 处理剩余 diffs（在最后窗口之后）
	for diffIdx < len(diffs) {
		dt := diffs[diffIdx].Timestamp
		if dt-lastActivityEnd >= idleMs {
			closeSession(lastActivityEnd)
			openSession(dt)
		} else if currentStart == 0 {
			openSession(dt)
		}
		diffIDs = append(diffIDs, diffs[diffIdx].ID)
		if dt > lastActivityEnd {
			lastActivityEnd = dt
		}
		diffIdx++
	}

	closeSession(lastActivityEnd)

	// 过滤空洞会话（无窗口且无 diff）并过滤碎片会话（时长过短）
	filtered := sessions[:0]
	minDurationMs := int64(s.cfg.MinSessionMinutes) * 60 * 1000
	for _, sess := range sessions {
		if sess == nil {
			continue
		}
		hasDiffs := sess.Metadata != nil && len(getSessionDiffIDs(sess.Metadata)) > 0
		if sess.PrimaryApp == "" && !hasDiffs {
			continue
		}
		// 过滤时长过短的碎片会话（避免1-2分钟的无意义会话）
		duration := sess.EndTime - sess.StartTime
		if duration < minDurationMs {
			continue
		}
		filtered = append(filtered, sess)
	}
	return filtered
}

// attachBrowserEvents 将浏览器事件绑定到对应的会话
func (s *SessionService) attachBrowserEvents(sessions []*schema.Session, events []schema.BrowserEvent) {
	if len(sessions) == 0 || len(events) == 0 {
		return
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Timestamp < events[j].Timestamp })

	sessIdx := 0
	for _, be := range events {
		for sessIdx < len(sessions) && be.Timestamp > sessions[sessIdx].EndTime {
			sessIdx++
		}
		if sessIdx >= len(sessions) {
			break
		}
		sess := sessions[sessIdx]
		if be.Timestamp < sess.StartTime || be.Timestamp > sess.EndTime {
			continue
		}
		if sess.Metadata == nil {
			sess.Metadata = make(schema.JSONMap)
		}
		raw := getSessionBrowserEventIDs(sess.Metadata)
		raw = append(raw, be.ID)
		setSessionBrowserEventIDs(sess.Metadata, raw)
	}
}

// formatDate 将时间戳格式化为日期字符串
func formatDate(ts int64) string {
	return time.UnixMilli(ts).Format("2006-01-02")
}
