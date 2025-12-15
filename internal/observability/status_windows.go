//go:build windows

package observability

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/yuqie6/WorkMirror/internal/bootstrap"
	"github.com/yuqie6/WorkMirror/internal/collector"
	"github.com/yuqie6/WorkMirror/internal/dto"
	"github.com/yuqie6/WorkMirror/internal/pkg/config"
	"github.com/yuqie6/WorkMirror/internal/pkg/privacy"
	"github.com/yuqie6/WorkMirror/internal/schema"
)

func BuildStatus(ctx context.Context, rt *bootstrap.AgentRuntime, startedAt time.Time) (*dto.StatusDTO, error) {
	if rt == nil || rt.Cfg == nil || rt.Core == nil || rt.Core.DB == nil {
		return nil, ErrNotReady
	}

	cfg := rt.Cfg
	now := time.Now()

	start24h := now.Add(-24 * time.Hour).UnixMilli()
	end24h := now.UnixMilli()

	eventCount24h := int64(0)
	if rt.Repos.Event != nil {
		if n, err := rt.Repos.Event.CountByTimeRange(ctx, start24h, end24h); err == nil {
			eventCount24h = n
		}
	}
	diffCount24h := int64(0)
	if rt.Repos.Diff != nil {
		if n, err := rt.Repos.Diff.CountByDateRange(ctx, start24h, end24h); err == nil {
			diffCount24h = n
		}
	}
	browserCount24h := int64(0)
	if rt.Repos.Browser != nil {
		if n, err := rt.Repos.Browser.CountByDateRange(ctx, start24h, end24h); err == nil {
			browserCount24h = n
		}
	}

	windowCollectedAt := int64(0)
	windowDropped := int64(0)
	windowRunning := false
	if rt.Collectors.Window != nil {
		if s, ok := rt.Collectors.Window.(interface {
			Stats() collector.WindowCollectorStats
		}); ok {
			st := s.Stats()
			windowCollectedAt = st.LastEmitAt
			windowDropped = st.Dropped
			windowRunning = st.Running
		}
	}

	windowPersistAt := int64(0)
	windowPersistDropped := int64(0)
	if rt.Services.Tracker != nil {
		st := rt.Services.Tracker.Stats()
		windowPersistAt = st.LastPersistAt
		windowPersistDropped = st.DroppedBatches
		windowRunning = windowRunning || st.Running
	}

	diffCollectedAt := int64(0)
	diffDropped := int64(0)
	diffSkipped := int64(0)
	diffRunning := false
	diffWatchPaths := []string(nil)
	if rt.Collectors.Diff != nil {
		st := rt.Collectors.Diff.Stats()
		diffCollectedAt = st.LastEmitAt
		diffDropped = st.Dropped
		diffSkipped = st.SkippedNonGit
		diffRunning = st.Running
		diffWatchPaths = st.WatchPaths
	}
	diffPersistAt := int64(0)
	if rt.Services.Diff != nil {
		ds := rt.Services.Diff.Stats()
		diffPersistAt = ds.LastPersistAt
		diffRunning = diffRunning || ds.Running
	}

	browserCollectedAt := int64(0)
	browserDropped := int64(0)
	browserRunning := false
	browserHistoryPath := strings.TrimSpace(cfg.Browser.HistoryPath)
	if rt.Collectors.Browser != nil {
		st := rt.Collectors.Browser.Stats()
		browserCollectedAt = st.LastEmitAt
		browserDropped = st.Dropped
		browserRunning = st.Running
		if browserHistoryPath == "" {
			browserHistoryPath = st.HistoryPath
		}
	}
	browserPersistAt := int64(0)
	if rt.Services.Browser != nil {
		bs := rt.Services.Browser.Stats()
		browserPersistAt = bs.LastPersistAt
		browserRunning = browserRunning || bs.Running
	}

	sessionCount24h := int64(0)
	pendingSemantic24h := int64(0)
	if rt.Repos.Session != nil {
		if n, err := rt.Repos.Session.CountByTimeRange(ctx, start24h, end24h); err == nil {
			sessionCount24h = n
		}
		if n, err := rt.Repos.Session.CountPendingSemanticByTimeRange(ctx, start24h, end24h); err == nil {
			pendingSemantic24h = n
		}
	}

	lastSplitAt := int64(0)
	if rt.Core.Services.Sessions != nil {
		lastSplitAt = rt.Core.Services.Sessions.Stats().LastSplitAt
	}
	if lastSplitAt == 0 && rt.Repos.Session != nil {
		if last, err := rt.Repos.Session.GetLastSession(ctx); err == nil && last != nil {
			lastSplitAt = last.EndTime
		}
	}

	lastSemanticAt := int64(0)
	if rt.Core.Services.SessionSemantic != nil {
		lastSemanticAt = rt.Core.Services.SessionSemantic.Stats().LastEnrichAt
	}

	aiConfigured := rt.Core.Clients.LLM != nil && rt.Core.Clients.LLM.IsConfigured()
	aiMode := "offline"
	if aiConfigured {
		aiMode = "ai"
	}

	aiStats := dto.AIPipelineStatusDTO{
		Configured: aiConfigured,
		Mode:       aiMode,
	}
	if rt.Core.Services.AI != nil {
		st := rt.Core.Services.AI.Stats()
		aiStats.LastCallAt = st.LastCallAt
		aiStats.LastErrorAt = st.LastErrorAt
		aiStats.LastError = st.LastError
		aiStats.Degraded = st.Degraded
		aiStats.DegradedReason = st.DegradedReason
	}
	if !aiConfigured {
		aiStats.Degraded = true
		if aiStats.DegradedReason == "" {
			aiStats.DegradedReason = "not_configured"
		}
	}

	ragEnabled := rt.Services.RAG != nil

	evidence := dto.EvidenceStatusDTO{
		Sessions24h: sessionCount24h,
	}
	if rt.Repos.Session != nil {
		if sessions, err := rt.Repos.Session.GetByTimeRange(ctx, start24h, end24h); err == nil {
			refDiffIDs := make(map[int64]struct{}, 256)
			refBrowserIDs := make(map[int64]struct{}, 256)
			for _, s := range sessions {
				diffIDs := schema.GetInt64Slice(s.Metadata, "diff_ids")
				browserIDs := schema.GetInt64Slice(s.Metadata, "browser_event_ids")
				hasDiff := len(diffIDs) > 0
				hasBrowser := len(browserIDs) > 0
				if hasDiff {
					evidence.WithDiff++
				}
				if hasBrowser {
					evidence.WithBrowser++
				}
				if hasDiff && hasBrowser {
					evidence.WithDiffBrowser++
				}
				if !hasDiff && !hasBrowser {
					evidence.WeakEvidence++
				}
				for _, id := range diffIDs {
					if id > 0 {
						refDiffIDs[id] = struct{}{}
					}
				}
				for _, id := range browserIDs {
					if id > 0 {
						refBrowserIDs[id] = struct{}{}
					}
				}
			}

			if rt.Repos.Diff != nil {
				if diffs, err := rt.Repos.Diff.GetByTimeRange(ctx, start24h, end24h); err == nil {
					for _, d := range diffs {
						if d.ID <= 0 {
							continue
						}
						if _, ok := refDiffIDs[d.ID]; !ok {
							evidence.OrphanDiffs24h++
						}
					}
				}
			}
			if rt.Repos.Browser != nil {
				if evs, err := rt.Repos.Browser.GetByTimeRange(ctx, start24h, end24h); err == nil {
					for _, e := range evs {
						if e.ID <= 0 {
							continue
						}
						if _, ok := refBrowserIDs[e.ID]; !ok {
							evidence.OrphanBrowser24h++
						}
					}
				}
			}
		}
	}

	logPath := strings.TrimSpace(cfg.App.LogPath)
	recentErr := ReadRecentErrors(logPath, privacy.New(cfg.Privacy.Enabled, cfg.Privacy.Patterns), 20)

	cfgPath, _ := config.DefaultConfigPath()

	return &dto.StatusDTO{
		App: dto.AppStatusDTO{
			Name:       cfg.App.Name,
			Version:    cfg.App.Version,
			StartedAt:  startedAt.Format(time.RFC3339),
			UptimeSec:  int64(now.Sub(startedAt).Seconds()),
			SafeMode:   rt.Core.DB.SafeMode,
			ConfigPath: cfgPath,
		},
		Storage: dto.StorageStatusDTO{
			DBPath:         cfg.Storage.DBPath,
			SchemaVersion:  rt.Core.DB.SchemaVersion,
			SafeModeReason: strings.TrimSpace(rt.Core.DB.MigrationError),
		},
		Privacy: dto.PrivacyStatusDTO{
			Enabled:      cfg.Privacy.Enabled,
			PatternCount: len(cfg.Privacy.Patterns),
		},
		Collectors: dto.CollectorsStatusDTO{
			Window: dto.CollectorStatusDTO{
				Enabled:         true,
				Running:         windowRunning,
				LastCollectedAt: windowCollectedAt,
				LastPersistedAt: windowPersistAt,
				Count24h:        eventCount24h,
				DroppedEvents:   windowDropped,
				DroppedBatches:  windowPersistDropped,
			},
			Diff: dto.CollectorStatusDTO{
				Enabled:         cfg.Diff.Enabled && len(cfg.Diff.WatchPaths) > 0,
				Running:         diffRunning,
				LastCollectedAt: diffCollectedAt,
				LastPersistedAt: diffPersistAt,
				Count24h:        diffCount24h,
				DroppedEvents:   diffDropped,
				Skipped:         diffSkipped,
				WatchPaths:      diffWatchPaths,
				EffectivePaths:  len(cfg.Diff.WatchPaths),
			},
			Browser: dto.CollectorStatusDTO{
				Enabled:          cfg.Browser.Enabled,
				Running:          browserRunning,
				LastCollectedAt:  browserCollectedAt,
				LastPersistedAt:  browserPersistAt,
				Count24h:         browserCount24h,
				DroppedEvents:    browserDropped,
				HistoryPath:      browserHistoryPath,
				SanitizedEnabled: cfg.Privacy.Enabled,
			},
		},
		Pipeline: dto.PipelineStatusDTO{
			Sessions: dto.SessionPipelineStatusDTO{
				LastSplitAt:        lastSplitAt,
				Sessions24h:        sessionCount24h,
				PendingSemantic24h: pendingSemantic24h,
				LastSemanticAt:     lastSemanticAt,
			},
			AI:  aiStats,
			RAG: dto.RAGPipelineStatusDTO{Enabled: ragEnabled},
		},
		Evidence:     evidence,
		RecentErrors: recentErr,
	}, nil
}

var (
	reLogTime  = regexp.MustCompile(`\btime=([^ ]+)`)
	reLogLevel = regexp.MustCompile(`\blevel=([^ ]+)`)
	reLogMsg   = regexp.MustCompile(`\bmsg=([^\n]+)$`)
)

func parseLogLine(line string) dto.RecentErrorDTO {
	e := dto.RecentErrorDTO{Raw: line, Message: line}
	if m := reLogTime.FindStringSubmatch(line); len(m) == 2 {
		e.Time = m[1]
	}
	if m := reLogLevel.FindStringSubmatch(line); len(m) == 2 {
		e.Level = strings.Trim(m[1], "\"")
	}
	if m := reLogMsg.FindStringSubmatch(line); len(m) == 2 {
		msg := strings.TrimSpace(m[1])
		msg = strings.TrimPrefix(msg, "\"")
		msg = strings.TrimSuffix(msg, "\"")
		e.Message = msg
	}
	return e
}
