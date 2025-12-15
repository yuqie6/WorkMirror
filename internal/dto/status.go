package dto

type StatusDTO struct {
	App          AppStatusDTO        `json:"app"`
	Storage      StorageStatusDTO    `json:"storage"`
	Privacy      PrivacyStatusDTO    `json:"privacy"`
	Collectors   CollectorsStatusDTO `json:"collectors"`
	Pipeline     PipelineStatusDTO   `json:"pipeline"`
	Evidence     EvidenceStatusDTO   `json:"evidence"`
	RecentErrors []RecentErrorDTO    `json:"recent_errors"`
}

type AppStatusDTO struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	StartedAt  string `json:"started_at"`
	UptimeSec  int64  `json:"uptime_sec"`
	SafeMode   bool   `json:"safe_mode"`
	ConfigPath string `json:"config_path,omitempty"`
}

type StorageStatusDTO struct {
	DBPath         string `json:"db_path"`
	SchemaVersion  int    `json:"schema_version"`
	SafeModeReason string `json:"safe_mode_reason,omitempty"`
}

type PrivacyStatusDTO struct {
	Enabled      bool `json:"enabled"`
	PatternCount int  `json:"pattern_count"`
}

type CollectorsStatusDTO struct {
	Window  CollectorStatusDTO `json:"window"`
	Diff    CollectorStatusDTO `json:"diff"`
	Browser CollectorStatusDTO `json:"browser"`
}

type CollectorStatusDTO struct {
	Enabled          bool     `json:"enabled"`
	Running          bool     `json:"running"`
	LastCollectedAt  int64    `json:"last_collected_at"`
	LastPersistedAt  int64    `json:"last_persisted_at"`
	Count24h         int64    `json:"count_24h"`
	DroppedEvents    int64    `json:"dropped_events,omitempty"`
	DroppedBatches   int64    `json:"dropped_batches,omitempty"`
	Skipped          int64    `json:"skipped,omitempty"`
	WatchPaths       []string `json:"watch_paths,omitempty"`
	EffectivePaths   int      `json:"effective_paths,omitempty"`
	HistoryPath      string   `json:"history_path,omitempty"`
	SanitizedEnabled bool     `json:"sanitized_enabled,omitempty"`
}

type PipelineStatusDTO struct {
	Sessions SessionPipelineStatusDTO `json:"sessions"`
	AI       AIPipelineStatusDTO      `json:"ai"`
	RAG      RAGPipelineStatusDTO     `json:"rag"`
}

type SessionPipelineStatusDTO struct {
	LastSplitAt        int64 `json:"last_split_at"`
	Sessions24h        int64 `json:"sessions_24h"`
	PendingSemantic24h int64 `json:"pending_semantic_24h"`
	LastSemanticAt     int64 `json:"last_semantic_at"`
}

type AIPipelineStatusDTO struct {
	Configured     bool   `json:"configured"`
	Mode           string `json:"mode"` // "ai" | "offline"
	LastCallAt     int64  `json:"last_call_at"`
	LastError      string `json:"last_error,omitempty"`
	LastErrorAt    int64  `json:"last_error_at,omitempty"`
	Degraded       bool   `json:"degraded"`
	DegradedReason string `json:"degraded_reason,omitempty"`
}

type RAGPipelineStatusDTO struct {
	Enabled     bool   `json:"enabled"`
	IndexCount  int64  `json:"index_count,omitempty"`
	LastError   string `json:"last_error,omitempty"`
	LastErrorAt int64  `json:"last_error_at,omitempty"`
}

type EvidenceStatusDTO struct {
	Sessions24h     int64 `json:"sessions_24h"`
	WithDiff        int64 `json:"with_diff"`
	WithBrowser     int64 `json:"with_browser"`
	WithDiffBrowser int64 `json:"with_diff_and_browser"`
	WeakEvidence    int64 `json:"weak_evidence"`

	OrphanDiffs24h   int64 `json:"orphan_diffs_24h,omitempty"`
	OrphanBrowser24h int64 `json:"orphan_browser_24h,omitempty"`
}

type RecentErrorDTO struct {
	Time    string `json:"time,omitempty"`
	Level   string `json:"level,omitempty"`
	Message string `json:"message"`
	Raw     string `json:"raw,omitempty"`
}
