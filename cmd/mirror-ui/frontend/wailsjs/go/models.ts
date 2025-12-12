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

