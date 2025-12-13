type JSONValue = string | number | boolean | null | JSONValue[] | { [key: string]: JSONValue };

async function requestJSON<T>(url: string, init?: RequestInit): Promise<T> {
    const res = await fetch(url, {
        ...init,
        headers: {
            "Content-Type": "application/json",
            ...(init?.headers || {}),
        },
    });

    if (!res.ok) {
        const errBody = await res.json().catch(() => null) as any;
        const msg = errBody?.error || res.statusText || `HTTP ${res.status}`;
        throw new Error(msg);
    }

    return await res.json() as T;
}

export async function GetTodaySummary(): Promise<any> {
    return requestJSON("/api/summary/today");
}

export async function GetDailySummary(date: string): Promise<any> {
    return requestJSON(`/api/summary/daily?date=${encodeURIComponent(date)}`);
}

export async function ListSummaryIndex(limit: number): Promise<any> {
    const n = Number.isFinite(limit) && limit > 0 ? Math.floor(limit) : 365;
    return requestJSON(`/api/summary/index?limit=${encodeURIComponent(String(n))}`);
}

export async function GetPeriodSummary(periodType: string, startDate: string): Promise<any> {
    const qs = new URLSearchParams();
    qs.set("type", periodType);
    if (startDate) qs.set("start_date", startDate);
    return requestJSON(`/api/summary/period?${qs.toString()}`);
}

export async function ListPeriodSummaryIndex(periodType: string, limit: number): Promise<any> {
    const qs = new URLSearchParams();
    qs.set("type", periodType);
    const n = Number.isFinite(limit) && limit > 0 ? Math.floor(limit) : 20;
    qs.set("limit", String(n));
    return requestJSON(`/api/summary/period/index?${qs.toString()}`);
}

export async function GetSkillTree(): Promise<any> {
    return requestJSON("/api/skills/tree");
}

export async function GetSkillEvidence(skillKey: string): Promise<any> {
    return requestJSON(`/api/skills/evidence?skill_key=${encodeURIComponent(skillKey)}`);
}

export async function GetSkillSessions(skillKey: string): Promise<any> {
    return requestJSON(`/api/skills/sessions?skill_key=${encodeURIComponent(skillKey)}`);
}

export async function GetTrends(days: number): Promise<any> {
    const n = days === 30 ? 30 : 7;
    return requestJSON(`/api/trends?days=${encodeURIComponent(String(n))}`);
}

export async function GetAppStats(): Promise<any> {
    return requestJSON("/api/app-stats");
}

export async function GetDiffDetail(id: number): Promise<any> {
    return requestJSON(`/api/diffs/detail?id=${encodeURIComponent(String(id))}`);
}

export async function GetSessionsByDate(date: string): Promise<any> {
    return requestJSON(`/api/sessions/by-date?date=${encodeURIComponent(date)}`);
}

export async function GetSessionDetail(id: number): Promise<any> {
    return requestJSON(`/api/sessions/detail?id=${encodeURIComponent(String(id))}`);
}

export async function BuildSessionsForDate(date: string): Promise<any> {
    return requestJSON("/api/sessions/build", {
        method: "POST",
        body: JSON.stringify({ date } as any),
    });
}

export async function RebuildSessionsForDate(date: string): Promise<any> {
    return requestJSON("/api/sessions/rebuild", {
        method: "POST",
        body: JSON.stringify({ date } as any),
    });
}

export async function EnrichSessionsForDate(date: string): Promise<any> {
    return requestJSON("/api/sessions/enrich", {
        method: "POST",
        body: JSON.stringify({ date } as any),
    });
}

export async function GetSettings(): Promise<any> {
    return requestJSON("/api/settings");
}

export async function SaveSettings(req: Record<string, JSONValue>): Promise<any> {
    return requestJSON("/api/settings", {
        method: "POST",
        body: JSON.stringify(req),
    });
}
