// Session types (基于后端 dto/httpapi.go)

export interface SessionDTO {
  id: number;
  date: string;
  start_time: number;
  end_time: number;
  time_range: string;
  primary_app: string;
  session_version: number;
  category: string;
  summary: string;
  skills_involved: string[];
  diff_count: number;
  browser_count: number;

  semantic_source: 'ai' | 'rule' | string;
  semantic_version?: string;
  evidence_hint: string;
  degraded_reason?: string;
}

export interface SessionAppUsageDTO {
  app_name: string;
  total_duration: number;
}

export interface SessionDiffDTO {
  id: number;
  file_name: string;
  language: string;
  insight: string;
  skills: string[];
  lines_added: number;
  lines_deleted: number;
  timestamp: number;
}

export interface SessionBrowserEventDTO {
  id: number;
  timestamp: number;
  domain: string;
  title: string;
  url: string;
  duration: number;
}

export interface SessionWindowEventDTO {
  timestamp: number;
  app_name: string;
  title: string;
  duration: number;
}

export interface SessionDetailDTO extends SessionDTO {
  tags: string[];
  rag_refs: Record<string, unknown>[];
  app_usage: SessionAppUsageDTO[];
  diffs: SessionDiffDTO[];
  browser: SessionBrowserEventDTO[];
}

// 前端使用的 Session 接口（保持向后兼容）
export interface ISession {
  id: number;
  date: string;
  title: string;
  summary: string;
  duration: string;
  type: 'ai' | 'rule';
  tags: string[];
  evidenceStrength: 'strong' | 'medium' | 'weak';
}

// 从后端 DTO 转换为前端 ISession
export function toISession(dto: SessionDTO): ISession {
  const semanticSource = (dto.semantic_source || 'rule') as 'ai' | 'rule' | string;

  // 证据强度：基于 diff 和 browser 同时存在
  let evidenceStrength: 'strong' | 'medium' | 'weak' = 'weak';
  if (dto.diff_count > 0 && dto.browser_count > 0) {
    evidenceStrength = 'strong';
  } else if (dto.diff_count > 0 || dto.browser_count > 0) {
    evidenceStrength = 'medium';
  }

  return {
    id: dto.id,
    date: dto.date,
    title: dto.category || dto.primary_app || '未分类会话',
    summary: dto.summary || `${dto.primary_app} 相关活动`,
    duration: dto.time_range,
    type: semanticSource === 'ai' ? 'ai' : 'rule',
    tags: dto.skills_involved || [],
    evidenceStrength,
  };
}
