package bootstrap

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yuqie6/WorkMirror/internal/ai"
	"github.com/yuqie6/WorkMirror/internal/pkg/config"
	"github.com/yuqie6/WorkMirror/internal/repository"
	"github.com/yuqie6/WorkMirror/internal/service"
)

// Core 持有跨二进制共享的核心依赖
type Core struct {
	Cfg       *config.Config
	DB        *repository.Database
	LogCloser io.Closer

	Repos struct {
		Diff          *repository.DiffRepository
		Event         *repository.EventRepository
		Summary       *repository.SummaryRepository
		Skill         *repository.SkillRepository
		SkillActivity *repository.SkillActivityRepository
		Browser       *repository.BrowserEventRepository
		Session       *repository.SessionRepository
		SessionDiff   *repository.SessionDiffRepository
		PeriodSummary *repository.PeriodSummaryRepository
	}

	Services struct {
		Skills          *service.SkillService
		AI              *service.AIService
		Trends          *service.TrendService
		Sessions        *service.SessionService
		SessionSemantic *service.SessionSemanticService
	}

	Clients struct {
		DeepSeek    *ai.DeepSeekClient
		SiliconFlow *ai.SiliconFlowClient
	}
}

// NewCore 构建核心依赖（不启动采集）
func NewCore(cfgPath string) (*Core, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}
	logCloser, _ := config.SetupLogger(config.LoggerOptions{
		Level:     cfg.App.LogLevel,
		Path:      cfg.App.LogPath,
		Component: filepath.Base(os.Args[0]),
	})

	db, err := repository.NewDatabase(cfg.Storage.DBPath)
	if err != nil {
		if logCloser != nil {
			_ = logCloser.Close()
		}
		return nil, err
	}

	c := &Core{Cfg: cfg, DB: db, LogCloser: logCloser}

	// Repos
	c.Repos.Diff = repository.NewDiffRepository(db.DB)
	c.Repos.Event = repository.NewEventRepository(db.DB)
	c.Repos.Summary = repository.NewSummaryRepository(db.DB)
	c.Repos.Skill = repository.NewSkillRepository(db.DB)
	c.Repos.SkillActivity = repository.NewSkillActivityRepository(db.DB)
	c.Repos.Browser = repository.NewBrowserEventRepository(db.DB)
	c.Repos.Session = repository.NewSessionRepository(db.DB)
	c.Repos.SessionDiff = repository.NewSessionDiffRepository(db.DB)
	c.Repos.PeriodSummary = repository.NewPeriodSummaryRepository(db.DB)

	// Clients / Analyzer
	c.Clients.DeepSeek = ai.NewDeepSeekClient(&ai.DeepSeekConfig{
		APIKey:  cfg.AI.DeepSeek.APIKey,
		BaseURL: cfg.AI.DeepSeek.BaseURL,
		Model:   cfg.AI.DeepSeek.Model,
	})
	var analyzer service.Analyzer
	if c.Clients.DeepSeek != nil && c.Clients.DeepSeek.IsConfigured() {
		analyzer = ai.NewDiffAnalyzer(c.Clients.DeepSeek)
	}

	// Services
	c.Services.Skills = service.NewSkillService(c.Repos.Skill, c.Repos.Diff, c.Repos.SkillActivity, service.DefaultExpPolicy{})
	c.Services.AI = service.NewAIService(analyzer, c.Repos.Diff, c.Repos.Event, c.Repos.Summary, c.Services.Skills)
	c.Services.Trends = service.NewTrendService(c.Repos.Skill, c.Repos.SkillActivity, c.Repos.Diff, c.Repos.Event, c.Repos.Session)
	c.Services.Sessions = service.NewSessionService(
		c.Repos.Event,
		c.Repos.Diff,
		c.Repos.Browser,
		c.Repos.Session,
		c.Repos.SessionDiff,
		&service.SessionServiceConfig{IdleGapMinutes: cfg.Collector.SessionIdleMin},
	)
	c.Services.SessionSemantic = service.NewSessionSemanticService(
		analyzer,
		c.Repos.Session,
		c.Repos.Diff,
		c.Repos.Event,
		c.Repos.Browser,
	)

	// Optional SiliconFlow client 由 Agent 侧按需启动 RAG
	if cfg.AI.SiliconFlow.APIKey != "" {
		c.Clients.SiliconFlow = ai.NewSiliconFlowClient(&ai.SiliconFlowConfig{
			APIKey:         cfg.AI.SiliconFlow.APIKey,
			BaseURL:        cfg.AI.SiliconFlow.BaseURL,
			EmbeddingModel: cfg.AI.SiliconFlow.EmbeddingModel,
			RerankerModel:  cfg.AI.SiliconFlow.RerankerModel,
		})
	}

	return c, nil
}

// Close 关闭核心依赖资源
func (c *Core) Close() error {
	if c == nil {
		return nil
	}
	var dbErr error
	if c.DB != nil {
		dbErr = c.DB.Close()
	}
	if c.LogCloser != nil {
		_ = c.LogCloser.Close()
	}
	return dbErr
}

// RequireAIConfigured 检查 AI 是否已配置
func (c *Core) RequireAIConfigured() error {
	if c.Clients.DeepSeek == nil || !c.Clients.DeepSeek.IsConfigured() {
		return fmt.Errorf("DeepSeek API 未配置")
	}
	return nil
}
