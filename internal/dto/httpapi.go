package dto

// 注意：本包用于承载“对外契约”的 DTO（与前端/HTTP API 保持稳定）。
// 不要在这里放 GORM/持久化细节；内部持久化 schema 请见 internal/schema；业务逻辑收敛在 internal/service。

type DailySummaryDTO struct {
	Date         string   `json:"date"`
	Summary      string   `json:"summary"`
	Highlights   string   `json:"highlights"`
	Struggles    string   `json:"struggles"`
	SkillsGained []string `json:"skills_gained"`
	TotalCoding  int      `json:"total_coding"`
	TotalDiffs   int      `json:"total_diffs"`
}

type SummaryIndexDTO struct {
	Date       string `json:"date"`
	HasSummary bool   `json:"has_summary"`
	Preview    string `json:"preview"`
}

type PeriodSummaryDTO struct {
	Type         string   `json:"type"`
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	Overview     string   `json:"overview"`
	Achievements []string `json:"achievements"`
	Patterns     string   `json:"patterns"`
	Suggestions  string   `json:"suggestions"`
	TopSkills    []string `json:"top_skills"`
	TotalCoding  int      `json:"total_coding"`
	TotalDiffs   int      `json:"total_diffs"`
}

type PeriodSummaryIndexDTO struct {
	Type      string `json:"type"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type SkillNodeDTO struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	ParentKey  string `json:"parent_key"`
	Level      int    `json:"level"`
	Experience int    `json:"experience"`
	Progress   int    `json:"progress"`
	Status     string `json:"status"`
	LastActive int64  `json:"last_active"`
}

type SkillEvidenceDTO struct {
	Source              string `json:"source"`
	EvidenceID          int64  `json:"evidence_id"`
	Timestamp           int64  `json:"timestamp"`
	ContributionContext string `json:"contribution_context"`
	FileName            string `json:"file_name"`
}

type TrendReportDTO struct {
	Period          string             `json:"period"`
	StartDate       string             `json:"start_date"`
	EndDate         string             `json:"end_date"`
	TotalDiffs      int64              `json:"total_diffs"`
	TotalCodingMins int64              `json:"total_coding_mins"`
	AvgDiffsPerDay  float64            `json:"avg_diffs_per_day"`
	TopLanguages    []LanguageTrendDTO `json:"top_languages"`
	TopSkills       []SkillTrendDTO    `json:"top_skills"`
	Bottlenecks     []string           `json:"bottlenecks"`
}

type LanguageTrendDTO struct {
	Language   string  `json:"language"`
	DiffCount  int64   `json:"diff_count"`
	Percentage float64 `json:"percentage"`
}

type SkillTrendDTO struct {
	SkillName   string  `json:"skill_name"`
	Status      string  `json:"status"`
	DaysActive  int     `json:"days_active"`
	Changes     int     `json:"changes"`
	ExpGain     float64 `json:"exp_gain"`
	PrevExpGain float64 `json:"prev_exp_gain"`
	GrowthRate  float64 `json:"growth_rate"`
}

type AppStatsDTO struct {
	AppName       string `json:"app_name"`
	TotalDuration int    `json:"total_duration"`
	EventCount    int64  `json:"event_count"`
	IsCodeEditor  bool   `json:"is_code_editor"`
}

type DiffDetailDTO struct {
	ID           int64    `json:"id"`
	FileName     string   `json:"file_name"`
	Language     string   `json:"language"`
	DiffContent  string   `json:"diff_content"`
	Insight      string   `json:"insight"`
	Skills       []string `json:"skills"`
	LinesAdded   int      `json:"lines_added"`
	LinesDeleted int      `json:"lines_deleted"`
	Timestamp    int64    `json:"timestamp"`
}

type SettingsDTO struct {
	ConfigPath string `json:"config_path"`

	DeepSeekAPIKeySet bool   `json:"deepseek_api_key_set"`
	DeepSeekBaseURL   string `json:"deepseek_base_url"`
	DeepSeekModel     string `json:"deepseek_model"`

	SiliconFlowAPIKeySet      bool   `json:"siliconflow_api_key_set"`
	SiliconFlowBaseURL        string `json:"siliconflow_base_url"`
	SiliconFlowEmbeddingModel string `json:"siliconflow_embedding_model"`
	SiliconFlowRerankerModel  string `json:"siliconflow_reranker_model"`

	DBPath             string   `json:"db_path"`
	DiffEnabled        bool     `json:"diff_enabled"`
	DiffWatchPaths     []string `json:"diff_watch_paths"`
	BrowserEnabled     bool     `json:"browser_enabled"`
	BrowserHistoryPath string   `json:"browser_history_path"`

	PrivacyEnabled  bool     `json:"privacy_enabled"`
	PrivacyPatterns []string `json:"privacy_patterns"`
}

type SaveSettingsRequestDTO struct {
	DeepSeekAPIKey  *string `json:"deepseek_api_key"`
	DeepSeekBaseURL *string `json:"deepseek_base_url"`
	DeepSeekModel   *string `json:"deepseek_model"`

	SiliconFlowAPIKey         *string `json:"siliconflow_api_key"`
	SiliconFlowBaseURL        *string `json:"siliconflow_base_url"`
	SiliconFlowEmbeddingModel *string `json:"siliconflow_embedding_model"`
	SiliconFlowRerankerModel  *string `json:"siliconflow_reranker_model"`

	DBPath             *string   `json:"db_path"`
	DiffEnabled        *bool     `json:"diff_enabled"`
	DiffWatchPaths     *[]string `json:"diff_watch_paths"`
	BrowserEnabled     *bool     `json:"browser_enabled"`
	BrowserHistoryPath *string   `json:"browser_history_path"`

	PrivacyEnabled  *bool     `json:"privacy_enabled"`
	PrivacyPatterns *[]string `json:"privacy_patterns"`
}

type SaveSettingsResponseDTO struct {
	RestartRequired bool `json:"restart_required"`
}

type DateRequestDTO struct {
	Date string `json:"date"`
}

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
	TotalDuration int    `json:"total_duration"`
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

type SessionWindowEventDTO struct {
	Timestamp int64  `json:"timestamp"`
	AppName   string `json:"app_name"`
	Title     string `json:"title"`
	Duration  int    `json:"duration"`
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
	Created  int `json:"created"`
	Enriched int `json:"enriched,omitempty"` // 语义丰富的会话数量（重建时自动触发）
}

type SessionEnrichResultDTO struct {
	Enriched int `json:"enriched"`
}
