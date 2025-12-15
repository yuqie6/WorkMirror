package dto

// 注意：本包用于承载“对外契约”的 DTO（与前端/HTTP API 保持稳定）。
// 不要在这里放 GORM/持久化细节；内部持久化 schema 请见 internal/schema；业务逻辑收敛在 internal/service。

type DailySummaryDTO struct {
	Date         string                   `json:"date"`
	Summary      string                   `json:"summary"`
	Highlights   string                   `json:"highlights"`
	Struggles    string                   `json:"struggles"`
	SkillsGained []string                 `json:"skills_gained"`
	TotalCoding  int                      `json:"total_coding"`
	TotalDiffs   int                      `json:"total_diffs"`
	Evidence     *DailySummaryEvidenceDTO `json:"evidence,omitempty"`
}

type SessionRefDTO struct {
	ID           int64  `json:"id"`
	Date         string `json:"date"`
	TimeRange    string `json:"time_range,omitempty"`
	Category     string `json:"category,omitempty"`
	Summary      string `json:"summary,omitempty"`
	EvidenceHint string `json:"evidence_hint,omitempty"` // e.g. "diff+browser" | "diff" | "browser" | "window_only"
}

type EvidenceBlockDTO struct {
	Sessions []SessionRefDTO `json:"sessions,omitempty"`
}

type ClaimEvidenceDTO struct {
	Claim    string          `json:"claim"`
	Sessions []SessionRefDTO `json:"sessions,omitempty"`
}

type DailySummaryEvidenceDTO struct {
	Summary    EvidenceBlockDTO   `json:"summary,omitempty"`
	Highlights []ClaimEvidenceDTO `json:"highlights,omitempty"`
	Struggles  []ClaimEvidenceDTO `json:"struggles,omitempty"`
}

type SummaryIndexDTO struct {
	Date       string `json:"date"`
	HasSummary bool   `json:"has_summary"`
	Preview    string `json:"preview"`
}

type PeriodSummaryDTO struct {
	Type         string                    `json:"type"`
	StartDate    string                    `json:"start_date"`
	EndDate      string                    `json:"end_date"`
	Overview     string                    `json:"overview"`
	Achievements []string                  `json:"achievements"`
	Patterns     string                    `json:"patterns"`
	Suggestions  string                    `json:"suggestions"`
	TopSkills    []string                  `json:"top_skills"`
	TotalCoding  int                       `json:"total_coding"`
	TotalDiffs   int                       `json:"total_diffs"`
	Evidence     *PeriodSummaryEvidenceDTO `json:"evidence,omitempty"`
}

type PeriodSummaryEvidenceDTO struct {
	Overview     EvidenceBlockDTO   `json:"overview,omitempty"`
	Achievements []ClaimEvidenceDTO `json:"achievements,omitempty"`
	Patterns     EvidenceBlockDTO   `json:"patterns,omitempty"`
	Suggestions  []ClaimEvidenceDTO `json:"suggestions,omitempty"`
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

type DailyTrendStatDTO struct {
	Date            string `json:"date"`
	TotalDiffs      int64  `json:"total_diffs"`
	TotalCodingMins int64  `json:"total_coding_mins"`
	SessionCount    int64  `json:"session_count"`
}

type TrendReportDTO struct {
	Period          string              `json:"period"`
	StartDate       string              `json:"start_date"`
	EndDate         string              `json:"end_date"`
	TotalDiffs      int64               `json:"total_diffs"`
	TotalCodingMins int64               `json:"total_coding_mins"`
	AvgDiffsPerDay  float64             `json:"avg_diffs_per_day"`
	TopLanguages    []LanguageTrendDTO  `json:"top_languages"`
	TopSkills       []SkillTrendDTO     `json:"top_skills"`
	Bottlenecks     []string            `json:"bottlenecks"`
	DailyStats      []DailyTrendStatDTO `json:"daily_stats,omitempty"`
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

	Language string `json:"language"` // AI Prompt 语言偏好：zh/en

	AI AISettingsDTO `json:"ai"`

	DBPath             string   `json:"db_path"`
	DiffEnabled        bool     `json:"diff_enabled"`
	DiffWatchPaths     []string `json:"diff_watch_paths"`
	BrowserEnabled     bool     `json:"browser_enabled"`
	BrowserHistoryPath string   `json:"browser_history_path"`

	PrivacyEnabled  bool     `json:"privacy_enabled"`
	PrivacyPatterns []string `json:"privacy_patterns"`
}

type AISettingsDTO struct {
	Provider string `json:"provider"`

	Default   AIProviderSettingsDTO `json:"default"`
	OpenAI    AIProviderSettingsDTO `json:"openai"`
	Anthropic AIProviderSettingsDTO `json:"anthropic"`
	Google    AIProviderSettingsDTO `json:"google"`
	Zhipu     AIProviderSettingsDTO `json:"zhipu"`

	SiliconFlow SiliconFlowSettingsDTO `json:"siliconflow"`
}

type AIProviderSettingsDTO struct {
	Enabled       bool   `json:"enabled,omitempty"` // 仅 default 用；其他 provider 忽略
	APIKeySet     bool   `json:"api_key_set"`
	APIKey        string `json:"api_key,omitempty"` // 永不返回（占位，避免前端误用）
	BaseURL       string `json:"base_url"`
	Model         string `json:"model"`
	APIKeyLocked  bool   `json:"api_key_locked,omitempty"`  // 仅 default 用；为 true 时禁止修改 api_key
	BaseURLLocked bool   `json:"base_url_locked,omitempty"` // 仅 default 用；为 true 时禁止修改 base_url
	ModelLocked   bool   `json:"model_locked,omitempty"`    // 仅 default 用；为 true 时禁止修改 model
}

type SiliconFlowSettingsDTO struct {
	APIKeySet      bool   `json:"api_key_set"`
	BaseURL        string `json:"base_url"`
	EmbeddingModel string `json:"embedding_model"`
	RerankerModel  string `json:"reranker_model"`
}

type SaveSettingsRequestDTO struct {
	Language *string `json:"language"` // AI Prompt 语言偏好：zh/en

	AI *AISettingsPatchDTO `json:"ai"`

	DBPath             *string   `json:"db_path"`
	DiffEnabled        *bool     `json:"diff_enabled"`
	DiffWatchPaths     *[]string `json:"diff_watch_paths"`
	BrowserEnabled     *bool     `json:"browser_enabled"`
	BrowserHistoryPath *string   `json:"browser_history_path"`

	PrivacyEnabled  *bool     `json:"privacy_enabled"`
	PrivacyPatterns *[]string `json:"privacy_patterns"`
}

type AISettingsPatchDTO struct {
	Provider *string `json:"provider"`

	Default   *AIProviderPatchDTO `json:"default"`
	OpenAI    *AIProviderPatchDTO `json:"openai"`
	Anthropic *AIProviderPatchDTO `json:"anthropic"`
	Google    *AIProviderPatchDTO `json:"google"`
	Zhipu     *AIProviderPatchDTO `json:"zhipu"`

	SiliconFlow *SiliconFlowPatchDTO `json:"siliconflow"`
}

type AIProviderPatchDTO struct {
	Enabled *bool   `json:"enabled"`
	APIKey  *string `json:"api_key"`
	BaseURL *string `json:"base_url"`
	Model   *string `json:"model"`
}

type SiliconFlowPatchDTO struct {
	APIKey         *string `json:"api_key"`
	BaseURL        *string `json:"base_url"`
	EmbeddingModel *string `json:"embedding_model"`
	RerankerModel  *string `json:"reranker_model"`
}

type SaveSettingsResponseDTO struct {
	RestartRequired bool `json:"restart_required"`
}

type DateRequestDTO struct {
	Date string `json:"date"`
}

type RepairEvidenceRequestDTO struct {
	Date             string `json:"date"`
	AttachGapMinutes int    `json:"attach_gap_minutes,omitempty"`
	Limit            int    `json:"limit,omitempty"`
}

type RepairEvidenceResultDTO struct {
	OrphanDiffs      int `json:"orphan_diffs"`
	OrphanBrowser    int `json:"orphan_browser"`
	AttachedDiffs    int `json:"attached_diffs"`
	AttachedBrowser  int `json:"attached_browser"`
	UpdatedSessions  int `json:"updated_sessions"`
	AttachGapMinutes int `json:"attach_gap_minutes"`
}

type SessionDTO struct {
	ID             int64    `json:"id"`
	Date           string   `json:"date"`
	StartTime      int64    `json:"start_time"`
	EndTime        int64    `json:"end_time"`
	TimeRange      string   `json:"time_range"`
	PrimaryApp     string   `json:"primary_app"`
	SessionVersion int      `json:"session_version"`
	Category       string   `json:"category"`
	Summary        string   `json:"summary"`
	SkillsInvolved []string `json:"skills_involved"`
	DiffCount      int      `json:"diff_count"`
	BrowserCount   int      `json:"browser_count"`

	SemanticSource  string `json:"semantic_source"`            // ai | rule
	SemanticVersion string `json:"semantic_version,omitempty"` // e.g. "v1"
	EvidenceHint    string `json:"evidence_hint"`              // diff+browser | diff | browser | window_only | diff_only | browser_only
	DegradedReason  string `json:"degraded_reason,omitempty"`  // only meaningful when semantic_source=rule
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
	Duration  int    `json:"duration"`
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
