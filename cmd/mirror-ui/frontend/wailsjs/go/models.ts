export namespace main {
	
	export class AppStatsDTO {
	    app_name: string;
	    total_duration: number;
	    event_count: number;
	
	    static createFrom(source: any = {}) {
	        return new AppStatsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.app_name = source["app_name"];
	        this.total_duration = source["total_duration"];
	        this.event_count = source["event_count"];
	    }
	}
	export class DailySummaryDTO {
	    date: string;
	    summary: string;
	    highlights: string;
	    struggles: string;
	    skills_gained: string[];
	    total_coding: number;
	    total_diffs: number;
	
	    static createFrom(source: any = {}) {
	        return new DailySummaryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.summary = source["summary"];
	        this.highlights = source["highlights"];
	        this.struggles = source["struggles"];
	        this.skills_gained = source["skills_gained"];
	        this.total_coding = source["total_coding"];
	        this.total_diffs = source["total_diffs"];
	    }
	}
	export class DiffDetailDTO {
	    id: number;
	    file_name: string;
	    language: string;
	    diff_content: string;
	    insight: string;
	    skills: string[];
	    lines_added: number;
	    lines_deleted: number;
	    timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new DiffDetailDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.file_name = source["file_name"];
	        this.language = source["language"];
	        this.diff_content = source["diff_content"];
	        this.insight = source["insight"];
	        this.skills = source["skills"];
	        this.lines_added = source["lines_added"];
	        this.lines_deleted = source["lines_deleted"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class SessionAppUsageDTO {
	    app_name: string;
	    total_duration: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionAppUsageDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.app_name = source["app_name"];
	        this.total_duration = source["total_duration"];
	    }
	}
	export class SessionBrowserEventDTO {
	    id: number;
	    timestamp: number;
	    domain: string;
	    title: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionBrowserEventDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.domain = source["domain"];
	        this.title = source["title"];
	        this.url = source["url"];
	    }
	}
	export class SessionBuildResultDTO {
	    created: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionBuildResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.created = source["created"];
	    }
	}
	export class SessionDiffDTO {
	    id: number;
	    file_name: string;
	    language: string;
	    insight: string;
	    skills: string[];
	    lines_added: number;
	    lines_deleted: number;
	    timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionDiffDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.file_name = source["file_name"];
	        this.language = source["language"];
	        this.insight = source["insight"];
	        this.skills = source["skills"];
	        this.lines_added = source["lines_added"];
	        this.lines_deleted = source["lines_deleted"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class SessionDTO {
	    id: number;
	    date: string;
	    start_time: number;
	    end_time: number;
	    time_range: string;
	    primary_app: string;
	    category: string;
	    summary: string;
	    skills_involved: string[];
	    diff_count: number;
	    browser_count: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.date = source["date"];
	        this.start_time = source["start_time"];
	        this.end_time = source["end_time"];
	        this.time_range = source["time_range"];
	        this.primary_app = source["primary_app"];
	        this.category = source["category"];
	        this.summary = source["summary"];
	        this.skills_involved = source["skills_involved"];
	        this.diff_count = source["diff_count"];
	        this.browser_count = source["browser_count"];
	    }
	}
	export class SessionDetailDTO {
	    id: number;
	    date: string;
	    start_time: number;
	    end_time: number;
	    time_range: string;
	    primary_app: string;
	    category: string;
	    summary: string;
	    skills_involved: string[];
	    diff_count: number;
	    browser_count: number;
	    tags: string[];
	    rag_refs: Array<{[key: string]: any}>;
	    app_usage: SessionAppUsageDTO[];
	    diffs: SessionDiffDTO[];
	    browser: SessionBrowserEventDTO[];
	
	    static createFrom(source: any = {}) {
	        return new SessionDetailDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.date = source["date"];
	        this.start_time = source["start_time"];
	        this.end_time = source["end_time"];
	        this.time_range = source["time_range"];
	        this.primary_app = source["primary_app"];
	        this.category = source["category"];
	        this.summary = source["summary"];
	        this.skills_involved = source["skills_involved"];
	        this.diff_count = source["diff_count"];
	        this.browser_count = source["browser_count"];
	        this.tags = source["tags"];
	        this.rag_refs = source["rag_refs"];
	        this.app_usage = (source["app_usage"] || []).map((e: any) => SessionAppUsageDTO.createFrom(e));
	        this.diffs = (source["diffs"] || []).map((e: any) => SessionDiffDTO.createFrom(e));
	        this.browser = (source["browser"] || []).map((e: any) => SessionBrowserEventDTO.createFrom(e));
	    }
	}
	export class SessionEnrichResultDTO {
	    enriched: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionEnrichResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enriched = source["enriched"];
	    }
	}
	export class LanguageTrendDTO {
	    language: string;
	    diff_count: number;
	    percentage: number;
	
	    static createFrom(source: any = {}) {
	        return new LanguageTrendDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.diff_count = source["diff_count"];
	        this.percentage = source["percentage"];
	    }
	}
	export class PeriodSummaryDTO {
	    type: string;
	    start_date: string;
	    end_date: string;
	    overview: string;
	    achievements: string[];
	    patterns: string;
	    suggestions: string;
	    top_skills: string[];
	    total_coding: number;
	    total_diffs: number;
	
	    static createFrom(source: any = {}) {
	        return new PeriodSummaryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.start_date = source["start_date"];
	        this.end_date = source["end_date"];
	        this.overview = source["overview"];
	        this.achievements = source["achievements"];
	        this.patterns = source["patterns"];
	        this.suggestions = source["suggestions"];
	        this.top_skills = source["top_skills"];
	        this.total_coding = source["total_coding"];
	        this.total_diffs = source["total_diffs"];
	    }
	}
	export class PeriodSummaryIndexDTO {
	    type: string;
	    start_date: string;
	    end_date: string;
	
	    static createFrom(source: any = {}) {
	        return new PeriodSummaryIndexDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.start_date = source["start_date"];
	        this.end_date = source["end_date"];
	    }
	}
	export class SaveSettingsRequestDTO {
	    deepseek_api_key?: string;
	    deepseek_base_url?: string;
	    deepseek_model?: string;
	    siliconflow_api_key?: string;
	    siliconflow_base_url?: string;
	    siliconflow_embedding_model?: string;
	    siliconflow_reranker_model?: string;
	    db_path?: string;
	    diff_watch_paths?: string[];
	    browser_history_path?: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveSettingsRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deepseek_api_key = source["deepseek_api_key"];
	        this.deepseek_base_url = source["deepseek_base_url"];
	        this.deepseek_model = source["deepseek_model"];
	        this.siliconflow_api_key = source["siliconflow_api_key"];
	        this.siliconflow_base_url = source["siliconflow_base_url"];
	        this.siliconflow_embedding_model = source["siliconflow_embedding_model"];
	        this.siliconflow_reranker_model = source["siliconflow_reranker_model"];
	        this.db_path = source["db_path"];
	        this.diff_watch_paths = source["diff_watch_paths"];
	        this.browser_history_path = source["browser_history_path"];
	    }
	}
	export class SettingsDTO {
	    config_path: string;
	    deepseek_api_key_set: boolean;
	    deepseek_base_url: string;
	    deepseek_model: string;
	    siliconflow_api_key_set: boolean;
	    siliconflow_base_url: string;
	    siliconflow_embedding_model: string;
	    siliconflow_reranker_model: string;
	    db_path: string;
	    diff_watch_paths: string[];
	    browser_history_path: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config_path = source["config_path"];
	        this.deepseek_api_key_set = source["deepseek_api_key_set"];
	        this.deepseek_base_url = source["deepseek_base_url"];
	        this.deepseek_model = source["deepseek_model"];
	        this.siliconflow_api_key_set = source["siliconflow_api_key_set"];
	        this.siliconflow_base_url = source["siliconflow_base_url"];
	        this.siliconflow_embedding_model = source["siliconflow_embedding_model"];
	        this.siliconflow_reranker_model = source["siliconflow_reranker_model"];
	        this.db_path = source["db_path"];
	        this.diff_watch_paths = source["diff_watch_paths"];
	        this.browser_history_path = source["browser_history_path"];
	    }
	}
	export class SkillEvidenceDTO {
	    source: string;
	    evidence_id: number;
	    timestamp: number;
	    contribution_context: string;
	    file_name: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillEvidenceDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.evidence_id = source["evidence_id"];
	        this.timestamp = source["timestamp"];
	        this.contribution_context = source["contribution_context"];
	        this.file_name = source["file_name"];
	    }
	}
	export class SkillNodeDTO {
	    key: string;
	    name: string;
	    category: string;
	    parent_key: string;
	    level: number;
	    experience: number;
	    progress: number;
	    status: string;
	    last_active: number;
	
	    static createFrom(source: any = {}) {
	        return new SkillNodeDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.category = source["category"];
	        this.parent_key = source["parent_key"];
	        this.level = source["level"];
	        this.experience = source["experience"];
	        this.progress = source["progress"];
	        this.status = source["status"];
	        this.last_active = source["last_active"];
	    }
	}
	export class SkillTrendDTO {
	    skill_name: string;
	    status: string;
	    days_active: number;
	
	    static createFrom(source: any = {}) {
	        return new SkillTrendDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skill_name = source["skill_name"];
	        this.status = source["status"];
	        this.days_active = source["days_active"];
	    }
	}
	export class SummaryIndexDTO {
	    date: string;
	    has_summary: boolean;
	    preview: string;
	
	    static createFrom(source: any = {}) {
	        return new SummaryIndexDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.has_summary = source["has_summary"];
	        this.preview = source["preview"];
	    }
	}
	export class TrendReportDTO {
	    period: string;
	    start_date: string;
	    end_date: string;
	    total_diffs: number;
	    total_coding_mins: number;
	    avg_diffs_per_day: number;
	    top_languages: LanguageTrendDTO[];
	    top_skills: SkillTrendDTO[];
	    bottlenecks: string[];
	
	    static createFrom(source: any = {}) {
	        return new TrendReportDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.period = source["period"];
	        this.start_date = source["start_date"];
	        this.end_date = source["end_date"];
	        this.total_diffs = source["total_diffs"];
	        this.total_coding_mins = source["total_coding_mins"];
	        this.avg_diffs_per_day = source["avg_diffs_per_day"];
	        this.top_languages = this.convertValues(source["top_languages"], LanguageTrendDTO);
	        this.top_skills = this.convertValues(source["top_skills"], SkillTrendDTO);
	        this.bottlenecks = source["bottlenecks"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}
