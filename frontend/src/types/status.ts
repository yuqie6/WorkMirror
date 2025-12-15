// System Status types (基于后端 dto/status.go)

export interface AppStatusDTO {
    name: string;
    version: string;
    started_at: string;
    uptime_sec: number;
    safe_mode: boolean;
    config_path?: string;
}

export interface StorageStatusDTO {
    db_path: string;
    schema_version: number;
    safe_mode_reason?: string;
}

export interface PrivacyStatusDTO {
    enabled: boolean;
    pattern_count: number;
}

export interface CollectorStatusDTO {
    enabled: boolean;
    running: boolean;
    last_collected_at: number;
    last_persisted_at: number;
    count_24h: number;
    dropped_events?: number;
    dropped_batches?: number;
    skipped?: number;
    watch_paths?: string[];
    effective_paths?: number;
    history_path?: string;
    sanitized_enabled?: boolean;
}

export interface CollectorsStatusDTO {
    window: CollectorStatusDTO;
    diff: CollectorStatusDTO;
    browser: CollectorStatusDTO;
}

export interface SessionPipelineStatusDTO {
    last_split_at: number;
    sessions_24h: number;
    pending_semantic_24h: number;
    last_semantic_at: number;
}

export interface AIPipelineStatusDTO {
    configured: boolean;
    mode: 'ai' | 'offline';
    last_call_at: number;
    last_error?: string;
    last_error_at?: number;
    degraded: boolean;
    degraded_reason?: string;
}

export interface RAGPipelineStatusDTO {
    enabled: boolean;
    index_count?: number;
    last_error?: string;
    last_error_at?: number;
}

export interface PipelineStatusDTO {
    sessions: SessionPipelineStatusDTO;
    ai: AIPipelineStatusDTO;
    rag: RAGPipelineStatusDTO;
}

export interface EvidenceStatusDTO {
    sessions_24h: number;
    with_diff: number;
    with_browser: number;
    with_diff_and_browser: number;
    weak_evidence: number;
    orphan_diffs_24h?: number;
    orphan_browser_24h?: number;
}

export interface RecentErrorDTO {
    time?: string;
    level?: string;
    message: string;
    raw?: string;
}

export interface StatusDTO {
    app: AppStatusDTO;
    storage: StorageStatusDTO;
    privacy: PrivacyStatusDTO;
    collectors: CollectorsStatusDTO;
    pipeline: PipelineStatusDTO;
    evidence: EvidenceStatusDTO;
    recent_errors: RecentErrorDTO[];
}

// 系统健康指示器类型
export type HealthStatus = 'healthy' | 'warning' | 'error' | 'offline';

export interface SystemHealthIndicator {
    window: HealthStatus;
    diff: HealthStatus;
    ai: HealthStatus;
    lastHeartbeat: string;
}

// 从 StatusDTO 提取健康指示器
export function extractHealthIndicator(status: StatusDTO): SystemHealthIndicator {
    const now = Date.now();

    const getCollectorHealth = (c: CollectorStatusDTO): HealthStatus => {
        if (!c.enabled) return 'offline';
        if (!c.running) return 'error';
        const lastActive = Math.max(c.last_collected_at, c.last_persisted_at);
        const staleMs = now - lastActive;
        if (staleMs > 5 * 60 * 1000) return 'warning'; // 5分钟未活跃
        return 'healthy';
    };

    const getAIHealth = (p: AIPipelineStatusDTO): HealthStatus => {
        if (!p.configured) return 'offline';
        if (p.degraded) return 'warning';
        if (p.mode === 'offline') return 'warning';
        return 'healthy';
    };

    const windowHealth = getCollectorHealth(status.collectors.window);
    const diffHealth = getCollectorHealth(status.collectors.diff);
    const aiHealth = getAIHealth(status.pipeline.ai);

    // 计算最近心跳
    const lastActive = Math.max(
        status.collectors.window.last_collected_at,
        status.collectors.diff.last_collected_at
    );
    const secAgo = Math.floor((now - lastActive) / 1000);
    const lastHeartbeat = secAgo < 60 ? `${secAgo}s` : `${Math.floor(secAgo / 60)}m`;

    return { window: windowHealth, diff: diffHealth, ai: aiHealth, lastHeartbeat };
}
