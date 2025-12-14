package service

import (
	"context"
	"testing"
	"time"

	"github.com/yuqie6/WorkMirror/internal/repository"
	"github.com/yuqie6/WorkMirror/internal/schema"
)

// ===== Mock Implementations =====

type fakeEventRepoForSession struct {
	events []schema.Event
}

func (f fakeEventRepoForSession) BatchInsert(ctx context.Context, events []schema.Event) error {
	return nil
}
func (f fakeEventRepoForSession) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Event, error) {
	out := make([]schema.Event, 0)
	for _, e := range f.events {
		if e.Timestamp >= startTime && e.Timestamp <= endTime {
			out = append(out, e)
		}
	}
	return out, nil
}
func (f fakeEventRepoForSession) GetByDate(ctx context.Context, date string) ([]schema.Event, error) {
	return f.events, nil
}
func (f fakeEventRepoForSession) GetAppStats(ctx context.Context, startTime, endTime int64) ([]repository.AppStat, error) {
	return nil, nil
}
func (f fakeEventRepoForSession) Count(ctx context.Context) (int64, error) { return 0, nil }

type fakeDiffRepoForSession struct {
	diffs []schema.Diff
}

func (f fakeDiffRepoForSession) Create(ctx context.Context, diff *schema.Diff) error { return nil }
func (f fakeDiffRepoForSession) GetPendingAIAnalysis(ctx context.Context, limit int) ([]schema.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepoForSession) UpdateAIInsight(ctx context.Context, id int64, insight string, skills []string) error {
	return nil
}
func (f fakeDiffRepoForSession) GetByDate(ctx context.Context, date string) ([]schema.Diff, error) {
	return f.diffs, nil
}
func (f fakeDiffRepoForSession) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Diff, error) {
	out := make([]schema.Diff, 0)
	for _, d := range f.diffs {
		if d.Timestamp >= startTime && d.Timestamp <= endTime {
			out = append(out, d)
		}
	}
	return out, nil
}
func (f fakeDiffRepoForSession) GetByIDs(ctx context.Context, ids []int64) ([]schema.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepoForSession) GetLanguageStats(ctx context.Context, startTime, endTime int64) ([]repository.LanguageStat, error) {
	return nil, nil
}
func (f fakeDiffRepoForSession) CountByDateRange(ctx context.Context, startTime, endTime int64) (int64, error) {
	return 0, nil
}
func (f fakeDiffRepoForSession) GetRecentAnalyzed(ctx context.Context, limit int) ([]schema.Diff, error) {
	return nil, nil
}
func (f fakeDiffRepoForSession) GetByID(ctx context.Context, id int64) (*schema.Diff, error) {
	return nil, nil
}

type fakeBrowserRepoForSession struct{}

func (f fakeBrowserRepoForSession) BatchInsert(ctx context.Context, events []*schema.BrowserEvent) error {
	return nil
}
func (f fakeBrowserRepoForSession) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.BrowserEvent, error) {
	return nil, nil
}
func (f fakeBrowserRepoForSession) GetByIDs(ctx context.Context, ids []int64) ([]schema.BrowserEvent, error) {
	return nil, nil
}

type fakeSessionRepoForSession struct {
	sessions    []*schema.Session
	lastSession *schema.Session
	maxVersion  int
}

func (f *fakeSessionRepoForSession) Create(ctx context.Context, session *schema.Session) (bool, error) {
	f.sessions = append(f.sessions, session)
	return true, nil
}
func (f *fakeSessionRepoForSession) UpdateSemantic(ctx context.Context, id int64, update schema.SessionSemanticUpdate) error {
	return nil
}
func (f *fakeSessionRepoForSession) GetByDate(ctx context.Context, date string) ([]schema.Session, error) {
	return nil, nil
}
func (f *fakeSessionRepoForSession) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Session, error) {
	return nil, nil
}
func (f *fakeSessionRepoForSession) GetMaxSessionVersionByDate(ctx context.Context, date string) (int, error) {
	return f.maxVersion, nil
}
func (f *fakeSessionRepoForSession) GetLastSession(ctx context.Context) (*schema.Session, error) {
	return f.lastSession, nil
}
func (f *fakeSessionRepoForSession) GetByID(ctx context.Context, id int64) (*schema.Session, error) {
	return nil, nil
}

type fakeSessionDiffRepoForSession struct {
	inserted map[int64][]int64
}

func (f *fakeSessionDiffRepoForSession) BatchInsert(ctx context.Context, sessionID int64, diffIDs []int64) error {
	if f.inserted == nil {
		f.inserted = make(map[int64][]int64)
	}
	f.inserted[sessionID] = diffIDs
	return nil
}

// ===== Test Cases =====

func TestSplitSessions_SingleContinuousSession(t *testing.T) {
	now := time.Now()
	baseTs := now.Truncate(time.Hour).UnixMilli()

	events := []schema.Event{
		{Timestamp: baseTs, AppName: "code.exe", Duration: 300},             // 0min, 5min duration
		{Timestamp: baseTs + 5*60*1000, AppName: "code.exe", Duration: 300}, // 5min, 5min duration
	}

	svc := NewSessionService(
		fakeEventRepoForSession{events: events},
		fakeDiffRepoForSession{},
		fakeBrowserRepoForSession{},
		&fakeSessionRepoForSession{},
		nil,
		&SessionServiceConfig{IdleGapMinutes: 6},
	)

	sessions := svc.splitSessions(events, nil, baseTs, baseTs+30*60*1000)

	if len(sessions) != 1 {
		t.Fatalf("sessions count=%d, want 1", len(sessions))
	}
	if sessions[0].PrimaryApp != "code.exe" {
		t.Fatalf("primaryApp=%q, want code.exe", sessions[0].PrimaryApp)
	}
}

func TestSplitSessions_IdleGapCreatesNewSession(t *testing.T) {
	now := time.Now()
	baseTs := now.Truncate(time.Hour).UnixMilli()

	// 两段事件，中间有 10 分钟间隔（超过 6 分钟阈值）
	events := []schema.Event{
		{Timestamp: baseTs, AppName: "code.exe", Duration: 60},
		{Timestamp: baseTs + 15*60*1000, AppName: "chrome.exe", Duration: 60}, // 15min later
	}

	svc := NewSessionService(
		fakeEventRepoForSession{events: events},
		fakeDiffRepoForSession{},
		fakeBrowserRepoForSession{},
		&fakeSessionRepoForSession{},
		nil,
		&SessionServiceConfig{IdleGapMinutes: 6},
	)

	sessions := svc.splitSessions(events, nil, baseTs, baseTs+30*60*1000)

	if len(sessions) != 2 {
		t.Fatalf("sessions count=%d, want 2 (idle gap should split)", len(sessions))
	}
	if sessions[0].PrimaryApp != "code.exe" {
		t.Fatalf("first session primaryApp=%q, want code.exe", sessions[0].PrimaryApp)
	}
	if sessions[1].PrimaryApp != "chrome.exe" {
		t.Fatalf("second session primaryApp=%q, want chrome.exe", sessions[1].PrimaryApp)
	}
}

func TestSplitSessions_DiffsAttachedToSession(t *testing.T) {
	now := time.Now()
	baseTs := now.Truncate(time.Hour).UnixMilli()

	events := []schema.Event{
		{Timestamp: baseTs, AppName: "code.exe", Duration: 300},
	}
	diffs := []schema.Diff{
		{ID: 101, Timestamp: baseTs + 1*60*1000}, // 1min after session start
		{ID: 102, Timestamp: baseTs + 2*60*1000}, // 2min after session start
	}

	svc := NewSessionService(
		fakeEventRepoForSession{events: events},
		fakeDiffRepoForSession{diffs: diffs},
		fakeBrowserRepoForSession{},
		&fakeSessionRepoForSession{},
		nil,
		&SessionServiceConfig{IdleGapMinutes: 6},
	)

	sessions := svc.splitSessions(events, diffs, baseTs, baseTs+30*60*1000)

	if len(sessions) != 1 {
		t.Fatalf("sessions count=%d, want 1", len(sessions))
	}

	diffIDs := getSessionDiffIDs(sessions[0].Metadata)
	if len(diffIDs) != 2 || diffIDs[0] != 101 || diffIDs[1] != 102 {
		t.Fatalf("diff_ids=%v, want [101, 102]", diffIDs)
	}
}

func TestSplitSessions_NoEventsReturnsNil(t *testing.T) {
	svc := NewSessionService(
		fakeEventRepoForSession{},
		fakeDiffRepoForSession{},
		fakeBrowserRepoForSession{},
		&fakeSessionRepoForSession{},
		nil,
		nil,
	)

	sessions := svc.splitSessions(nil, nil, 0, 1000)

	if sessions != nil {
		t.Fatalf("sessions should be nil when no events, got %v", sessions)
	}
}

func TestSplitSessions_PrimaryAppSelection(t *testing.T) {
	now := time.Now()
	baseTs := now.Truncate(time.Hour).UnixMilli()

	// code.exe 有更长的总时长
	events := []schema.Event{
		{Timestamp: baseTs, AppName: "code.exe", Duration: 600},          // 10min
		{Timestamp: baseTs + 1000, AppName: "chrome.exe", Duration: 300}, // 5min
		{Timestamp: baseTs + 2000, AppName: "code.exe", Duration: 300},   // 5min more
	}

	svc := NewSessionService(
		fakeEventRepoForSession{events: events},
		fakeDiffRepoForSession{},
		fakeBrowserRepoForSession{},
		&fakeSessionRepoForSession{},
		nil,
		&SessionServiceConfig{IdleGapMinutes: 6},
	)

	sessions := svc.splitSessions(events, nil, baseTs, baseTs+30*60*1000)

	if len(sessions) != 1 {
		t.Fatalf("sessions count=%d, want 1", len(sessions))
	}
	// code.exe total = 600+300=900s, chrome.exe = 300s
	if sessions[0].PrimaryApp != "code.exe" {
		t.Fatalf("primaryApp=%q, want code.exe (longer total duration)", sessions[0].PrimaryApp)
	}
}

func TestBuildSessionsForRange_CreatesAndPersists(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	baseTs := now.Truncate(time.Hour).UnixMilli()

	events := []schema.Event{
		{Timestamp: baseTs, AppName: "code.exe", Duration: 300},
	}

	sessionRepo := &fakeSessionRepoForSession{}
	svc := NewSessionService(
		fakeEventRepoForSession{events: events},
		fakeDiffRepoForSession{},
		fakeBrowserRepoForSession{},
		sessionRepo,
		nil,
		&SessionServiceConfig{IdleGapMinutes: 6},
	)

	created, err := svc.BuildSessionsForRange(ctx, baseTs, baseTs+30*60*1000)
	if err != nil {
		t.Fatalf("BuildSessionsForRange error: %v", err)
	}
	if created != 1 {
		t.Fatalf("created=%d, want 1", created)
	}
	if len(sessionRepo.sessions) != 1 {
		t.Fatalf("persisted sessions=%d, want 1", len(sessionRepo.sessions))
	}
}
