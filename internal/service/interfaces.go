package service

import (
	"context"

	"github.com/yuqie6/mirror/internal/ai"
	"github.com/yuqie6/mirror/internal/model"
	"github.com/yuqie6/mirror/internal/repository"
)

// 仓储/外部依赖的最小接口集合（ISP）

type DiffRepository interface {
	Create(ctx context.Context, diff *model.Diff) error
	GetPendingAIAnalysis(ctx context.Context, limit int) ([]model.Diff, error)
	UpdateAIInsight(ctx context.Context, id int64, insight string, skills []string) error
	GetByDate(ctx context.Context, date string) ([]model.Diff, error)
	GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]model.Diff, error)
	GetByIDs(ctx context.Context, ids []int64) ([]model.Diff, error)
	GetLanguageStats(ctx context.Context, startTime, endTime int64) ([]repository.LanguageStat, error)
	CountByDateRange(ctx context.Context, startTime, endTime int64) (int64, error)
	GetRecentAnalyzed(ctx context.Context, limit int) ([]model.Diff, error)
	GetByID(ctx context.Context, id int64) (*model.Diff, error)
}

type EventRepository interface {
	BatchInsert(ctx context.Context, events []model.Event) error
	GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]model.Event, error)
	GetByDate(ctx context.Context, date string) ([]model.Event, error)
	GetAppStats(ctx context.Context, startTime, endTime int64) ([]repository.AppStat, error)
	Count(ctx context.Context) (int64, error)
}

type BrowserEventRepository interface {
	BatchInsert(ctx context.Context, events []*model.BrowserEvent) error
	GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]model.BrowserEvent, error)
	GetByIDs(ctx context.Context, ids []int64) ([]model.BrowserEvent, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) (bool, error)
	UpdateSummaryOnly(ctx context.Context, id int64, summary string, metadata model.JSONMap) error
	UpdateSemantic(ctx context.Context, id int64, update model.SessionSemanticUpdate) error
	GetByDate(ctx context.Context, date string) ([]model.Session, error)
	GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]model.Session, error)
	GetLastSession(ctx context.Context) (*model.Session, error)
	GetByID(ctx context.Context, id int64) (*model.Session, error)
}

type SessionDiffRepository interface {
	BatchInsert(ctx context.Context, sessionID int64, diffIDs []int64) error
}

type SummaryRepository interface {
	GetByDate(ctx context.Context, date string) (*model.DailySummary, error)
	Upsert(ctx context.Context, summary *model.DailySummary) error
	GetRecent(ctx context.Context, limit int) ([]model.DailySummary, error)
}

type SkillRepository interface {
	GetAll(ctx context.Context) ([]model.SkillNode, error)
	GetByKey(ctx context.Context, key string) (*model.SkillNode, error)
	Upsert(ctx context.Context, skill *model.SkillNode) error
	UpsertBatch(ctx context.Context, skills []*model.SkillNode) error
	GetTopSkills(ctx context.Context, limit int) ([]model.SkillNode, error)
	GetActiveSkillsInPeriod(ctx context.Context, startTime, endTime int64, limit int) ([]model.SkillNode, error)
}

type Analyzer interface {
	AnalyzeDiff(ctx context.Context, filePath, language, diffContent string, existingSkills []ai.SkillInfo) (*ai.DiffInsight, error)
	GenerateDailySummary(ctx context.Context, req *ai.DailySummaryRequest) (*ai.DailySummaryResult, error)
	GenerateWeeklySummary(ctx context.Context, req *ai.WeeklySummaryRequest) (*ai.WeeklySummaryResult, error)
	GenerateSessionSummary(ctx context.Context, req *ai.SessionSummaryRequest) (*ai.SessionSummaryResult, error)
}

type RAGQuerier interface {
	Query(ctx context.Context, query string, topK int) ([]MemoryResult, error)
	IndexDiff(ctx context.Context, diff *model.Diff) error
	IndexDailySummary(ctx context.Context, summary *model.DailySummary) error
}
