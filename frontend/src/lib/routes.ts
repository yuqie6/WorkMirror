export type AppRoute =
  | { kind: 'root' }
  | { kind: 'dashboard' }
  | { kind: 'sessions'; date?: string; sessionId?: number }
  | { kind: 'skills'; skillId?: string }
  | { kind: 'reports' }
  | { kind: 'status' }
  | { kind: 'settings' }
  | { kind: 'not_found' };

function parseIntSafe(s: string | undefined): number | undefined {
  if (!s) return undefined;
  const n = Number(s);
  if (!Number.isFinite(n) || n <= 0) return undefined;
  return Math.floor(n);
}

function parseQuery(search: string): URLSearchParams {
  const s = typeof search === 'string' ? search.trim() : '';
  if (!s) return new URLSearchParams();
  return new URLSearchParams(s.startsWith('?') ? s.slice(1) : s);
}

export function parseAppRoute(pathname: string, search: string): AppRoute {
  const p = (pathname || '/').split('?')[0].trim() || '/';
  const qs = parseQuery(search);

  if (p === '/' || p === '') return { kind: 'root' };

  const parts = p.split('/').filter(Boolean);
  const head = parts[0];

  if (head === 'dashboard') return { kind: 'dashboard' };
  if (head === 'reports') return { kind: 'reports' };
  if (head === 'status') return { kind: 'status' };
  if (head === 'settings') return { kind: 'settings' };

  if (head === 'sessions') {
    const date = qs.get('date') || undefined;
    const sessionId = parseIntSafe(parts[1]);
    return { kind: 'sessions', date, sessionId };
  }

  if (head === 'skills') {
    const skillId = parts[1] ? decodeURIComponent(parts[1]) : undefined;
    return { kind: 'skills', skillId };
  }

  return { kind: 'not_found' };
}

function withQuery(path: string, query: Record<string, string | undefined>): string {
  const qs = new URLSearchParams();
  for (const [k, v] of Object.entries(query)) {
    if (typeof v === 'string' && v.trim() !== '') qs.set(k, v);
  }
  const s = qs.toString();
  return s ? `${path}?${s}` : path;
}

export function buildSessionsURL(opts?: { date?: string; sessionId?: number }): string {
  const date = opts?.date;
  const id = opts?.sessionId;
  const base = typeof id === 'number' && id > 0 ? `/sessions/${id}` : '/sessions';
  return withQuery(base, { date });
}

export function buildSkillsURL(skillId?: string): string {
  if (typeof skillId === 'string' && skillId.trim() !== '') {
    return `/skills/${encodeURIComponent(skillId)}`;
  }
  return '/skills';
}

