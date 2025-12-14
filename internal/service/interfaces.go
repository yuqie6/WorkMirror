package service

import (
	"context"

	"github.com/yuqie6/WorkMirror/internal/ai"
	"github.com/yuqie6/WorkMirror/internal/repository"
	"github.com/yuqie6/WorkMirror/internal/schema"
)

// 仓储/外部依赖的最小接口集合（ISP）

type DiffRepository interface {
	Create(ctx context.Context, diff *schema.Diff) error
	GetPendingAIAnalysis(ctx context.Context, limit int) ([]schema.Diff, error)
	UpdateAIInsight(ctx context.Context, id int64, insight string, skills []string) error
	GetByDate(ctx context.Context, date string) ([]schema.Diff, error)
	GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Diff, error)
	GetByIDs(ctx context.Context, ids []int64) ([]schema.Diff, error)
	GetLanguageStats(ctx context.Context, startTime, endTime int64) ([]repository.LanguageStat, error)
	CountByDateRange(ctx context.Context, startTime, endTime int64) (int64, error)
	GetRecentAnalyzed(ctx context.Context, limit int) ([]schema.Diff, error)
	GetByID(ctx context.Context, id int64) (*schema.Diff, error)
}

type EventRepository interface {
	BatchInsert(ctx context.Context, events []schema.Event) error
	GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Event, error)
	GetByDate(ctx context.Context, date string) ([]schema.Event, error)
	GetAppStats(ctx context.Context, startTime, endTime int64) ([]repository.AppStat, error)
	Count(ctx context.Context) (int64, error)
}

type BrowserEventRepository interface {
	BatchInsert(ctx context.Context, events []*schema.BrowserEvent) error
	GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.BrowserEvent, error)
	GetByIDs(ctx context.Context, ids []int64) ([]schema.BrowserEvent, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *schema.Session) (bool, error)
	UpdateSemantic(ctx context.Context, id int64, update schema.SessionSemanticUpdate) error
	GetByDate(ctx context.Context, date string) ([]schema.Session, error)
	GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Session, error)
	GetMaxSessionVersionByDate(ctx context.Context, date string) (int, error)
	GetLastSession(ctx context.Context) (*schema.Session, error)
	GetByID(ctx context.Context, id int64) (*schema.Session, error)
}

type SessionDiffRepository interface {
	BatchInsert(ctx context.Context, sessionID int64, diffIDs []int64) error
}

type SummaryRepository interface {
	GetByDate(ctx context.Context, date string) (*schema.DailySummary, error)
	Upsert(ctx context.Context, summary *schema.DailySummary) error
	GetRecent(ctx context.Context, limit int) ([]schema.DailySummary, error)
}

type SkillRepository interface {
	GetAll(ctx context.Context) ([]schema.SkillNode, error)
	GetByKey(ctx context.Context, key string) (*schema.SkillNode, error)
	Upsert(ctx context.Context, skill *schema.SkillNode) error
	UpsertBatch(ctx context.Context, skills []*schema.SkillNode) error
	GetTopSkills(ctx context.Context, limit int) ([]schema.SkillNode, error)
	GetActiveSkillsInPeriod(ctx context.Context, startTime, endTime int64, limit int) ([]schema.SkillNode, error)
}

type SkillActivityRepository interface {
	BatchInsert(ctx context.Context, activities []schema.SkillActivity) (int64, error)
	ListExistingKeys(ctx context.Context, keys []repository.SkillActivityKey) (map[repository.SkillActivityKey]struct{}, error)
	GetStatsByTimeRange(ctx context.Context, startTime, endTime int64) ([]repository.SkillActivityStat, error)
}

type Analyzer interface {
	AnalyzeDiff(ctx context.Context, filePath, language, diffContent string, existingSkills []ai.SkillInfo) (*ai.DiffInsight, error)
	GenerateDailySummary(ctx context.Context, req *ai.DailySummaryRequest) (*ai.DailySummaryResult, error)
	GenerateWeeklySummary(ctx context.Context, req *ai.WeeklySummaryRequest) (*ai.WeeklySummaryResult, error)
	GenerateSessionSummary(ctx context.Context, req *ai.SessionSummaryRequest) (*ai.SessionSummaryResult, error)
}

type RAGQuerier interface {
	Query(ctx context.Context, query string, topK int) ([]MemoryResult, error)
	IndexDiff(ctx context.Context, diff *schema.Diff) error
	IndexDailySummary(ctx context.Context, summary *schema.DailySummary) error
}
