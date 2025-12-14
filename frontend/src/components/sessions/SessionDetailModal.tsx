import React, { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import { GetSessionDetail, GetDiffDetail, GetSessionEvents } from '../../api/app';
import type { SessionDTO } from '../../types/session';

interface SessionAppUsageDTO {
    app_name: string;
    total_duration: number;
}

interface SessionDiffDTO {
    id: number;
    file_name: string;
    language: string;
    insight: string;
    skills: string[];
    lines_added: number;
    lines_deleted: number;
    timestamp: number;
}

interface SessionBrowserEventDTO {
    id: number;
    timestamp: number;
    domain: string;
    title: string;
    url: string;
}

interface SessionWindowEventDTO {
    timestamp: number;
    app_name: string;
    title: string;
    duration: number;
}

export interface SessionDetailDTO extends SessionDTO {
    tags: string[];
    rag_refs: Array<Record<string, any>>;
    app_usage: SessionAppUsageDTO[];
    diffs: SessionDiffDTO[];
    browser: SessionBrowserEventDTO[];
}

interface DiffDetailDTO {
    id: number;
    file_name: string;
    language: string;
    diff_content: string;
    insight: string;
    skills: string[];
    lines_added: number;
    lines_deleted: number;
    timestamp: number;
}

const formatTs = (ms: number) => {
    try {
        return new Date(ms).toLocaleString();
    } catch {
        return String(ms);
    }
};

const categoryLabel = (cat?: string) => {
    switch ((cat || '').toLowerCase()) {
        case 'technical': return '技术';
        case 'learning': return '学习';
        case 'exploration': return '探索';
        default: return cat || '其他';
    }
};

const SessionDetailModal: React.FC<{ sessionId: number; onClose: () => void; }> = ({ sessionId, onClose }) => {
    const [data, setData] = useState<SessionDetailDTO | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [diffDetail, setDiffDetail] = useState<DiffDetailDTO | null>(null);
    const [diffLoading, setDiffLoading] = useState(false);
    const [windowEvents, setWindowEvents] = useState<SessionWindowEventDTO[]>([]);
    const [windowEventsLoading, setWindowEventsLoading] = useState(false);
    const [windowEventsError, setWindowEventsError] = useState<string | null>(null);

    useEffect(() => {
        if (!sessionId) return;
        let cancelled = false;
        setLoading(true);
        setError(null);
        setData(null);
        setWindowEvents([]);
        setWindowEventsError(null);
        void (async () => {
            try {
                const [detail, events] = await Promise.all([
                    GetSessionDetail(sessionId),
                    (async () => {
                        setWindowEventsLoading(true);
                        try {
                            return await GetSessionEvents(sessionId, 300, 0);
                        } finally {
                            setWindowEventsLoading(false);
                        }
                    })(),
                ]);
                if (!cancelled) {
                    setData(detail);
                    setWindowEvents((events || []) as SessionWindowEventDTO[]);
                }
            } catch (e: any) {
                if (!cancelled) {
                    setError(e?.message || '获取会话详情失败');
                    setWindowEventsError(e?.message || '获取窗口证据失败');
                }
            } finally {
                if (!cancelled) setLoading(false);
            }
        })();
        return () => { cancelled = true; };
    }, [sessionId]);

    const loadDiffDetail = async (id: number) => {
        setDiffLoading(true);
        try {
            const res = await GetDiffDetail(id);
            setDiffDetail(res);
        } catch (e) {
            console.error('Failed to load diff detail', e);
        } finally {
            setDiffLoading(false);
        }
    };

    const body = (
        <div className="fixed inset-0 z-[9999] flex items-center justify-center">
            <div className="absolute inset-0 bg-black/40" onClick={onClose} />
            <div className="relative w-[min(1100px,92vw)] max-h-[90vh] overflow-hidden rounded-2xl bg-white shadow-2xl border border-gray-100 animate-fade-in">
                <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
                    <div className="space-y-1">
                        <div className="flex items-center gap-3">
                            <h2 className="text-lg font-semibold text-gray-900">会话详情</h2>
                            {data?.time_range && <span className="pill">{data.time_range}</span>}
                            {data?.category && <span className="pill">{categoryLabel(data.category)}</span>}
                        </div>
                        <div className="text-xs text-gray-400">
                            {data?.date || ''} · {data?.primary_app || ''}
                        </div>
                    </div>
                    <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-xl transition-colors text-gray-400 hover:text-gray-700">
                        <span className="text-lg">×</span>
                    </button>
                </div>

                <div className="p-6 overflow-y-auto max-h-[calc(90vh-64px)] space-y-5">
                    {loading && (
                        <div className="flex items-center gap-3 text-sm text-gray-500">
                            <div className="w-5 h-5 border-2 border-gray-200 border-t-accent-gold rounded-full animate-spin" />
                            加载中...
                        </div>
                    )}
                    {error && <div className="text-sm text-red-500">{error}</div>}
                    {data && (
                        <>
                            <div className="card">
                                <h3 className="text-sm font-semibold text-gray-900 mb-2 flex items-center gap-1.5"><svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" /></svg>摘要</h3>
                                <p className="text-gray-600 leading-relaxed">{data.summary || '（暂无摘要）'}</p>
                                {data.skills_involved?.length > 0 && (
                                    <div className="mt-3 flex flex-wrap gap-2">
                                        {data.skills_involved.map((s, i) => <span key={i} className="pill">{s}</span>)}
                                    </div>
                                )}
                                {data.tags?.length > 0 && (
                                    <div className="mt-3 flex flex-wrap gap-2">
                                        {data.tags.map((t, i) => <span key={i} className="pill">{t}</span>)}
                                    </div>
                                )}
                            </div>

                            {data.app_usage?.length > 0 && (
                                <div className="card">
                                    <h3 className="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-1.5"><svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M3 8.25V18a2.25 2.25 0 002.25 2.25h13.5A2.25 2.25 0 0021 18V8.25m-18 0V6a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 6v2.25m-18 0h18M5.25 6h.008v.008H5.25V6zM7.5 6h.008v.008H7.5V6zm2.25 0h.008v.008H9.75V6z" /></svg>应用使用</h3>
                                    <div className="space-y-2">
                                        {data.app_usage.map((a, i) => (
                                            <div key={i} className="flex items-center justify-between text-sm">
                                                <span className="text-gray-700 truncate max-w-[70%]">{a.app_name}</span>
                                                <span className="text-gray-400">{Math.round((a.total_duration || 0) / 60)}m</span>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}

                            <div className="card">
                                <div className="flex items-center justify-between mb-3">
                                    <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-1.5">
                                        <svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                                            <path strokeLinecap="round" strokeLinejoin="round" d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25m18 0A2.25 2.25 0 0018.75 3H5.25A2.25 2.25 0 003 5.25m18 0V12a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 12V5.25" />
                                        </svg>
                                        窗口证据
                                    </h3>
                                    <span className="text-xs text-gray-400">{windowEvents?.length || 0}</span>
                                </div>
                                {windowEventsLoading && (
                                    <div className="text-sm text-gray-400">加载窗口证据中...</div>
                                )}
                                {!windowEventsLoading && windowEventsError && (
                                    <div className="text-sm text-red-500">{windowEventsError}</div>
                                )}
                                {!windowEventsLoading && !windowEventsError && (windowEvents?.length ? (
                                    <div className="space-y-2">
                                        {windowEvents.slice(0, 30).map((e, i) => (
                                            <div key={`${e.timestamp}_${i}`} className="p-3 rounded-xl border border-gray-100">
                                                <div className="flex items-center justify-between gap-3">
                                                    <div className="min-w-0">
                                                        <div className="text-sm font-medium text-gray-900 truncate">{e.app_name}</div>
                                                        <div className="text-xs text-gray-400 line-clamp-2">{e.title || '（空标题）'}</div>
                                                    </div>
                                                    <div className="text-right flex-shrink-0">
                                                        <div className="text-xs text-gray-500">{Math.max(0, Math.round((e.duration || 0)))}s</div>
                                                        <div className="text-[11px] text-gray-300">{formatTs(e.timestamp)}</div>
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                        {windowEvents.length > 30 && (
                                            <div className="text-xs text-gray-400">仅展示前 30 条窗口事件</div>
                                        )}
                                    </div>
                                ) : (
                                    <div className="text-sm text-gray-400">暂无窗口证据</div>
                                ))}
                            </div>

                            <div className="grid grid-cols-12 gap-5">
                                <div className="col-span-7 card">
                                    <div className="flex items-center justify-between mb-3">
                                        <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-1.5"><svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m5.231 13.888L15 9.75l-1.519 4.138M7.5 11.25l-1.5 3 1.5 3m7.5-6l1.5-3-1.5-3M8.25 4.5h6.75c.621 0 1.125.504 1.125 1.125v.375c0 .621-.504 1.125-1.125 1.125H8.25a1.125 1.125 0 01-1.125-1.125v-.375c0-.621.504-1.125 1.125-1.125z" /></svg>Diffs</h3>
                                        <span className="text-xs text-gray-400">{data.diffs?.length || 0}</span>
                                    </div>
                                    {data.diffs?.length ? (
                                        <div className="space-y-2">
                                            {data.diffs.map((d) => (
                                                <button key={d.id} className="w-full text-left p-3 rounded-xl border border-gray-100 hover:border-amber-200 hover:bg-amber-50/40 transition" onClick={() => loadDiffDetail(d.id)}>
                                                    <div className="flex items-center justify-between gap-3">
                                                        <div className="min-w-0">
                                                            <div className="text-sm font-medium text-gray-900 truncate">{d.file_name}</div>
                                                            <div className="text-xs text-gray-400 truncate">{d.language}{d.insight ? ` · ${d.insight}` : ''}</div>
                                                        </div>
                                                        <div className="text-xs text-gray-400 whitespace-nowrap">
                                                            +{d.lines_added}/-{d.lines_deleted}
                                                        </div>
                                                    </div>
                                                </button>
                                            ))}
                                            {diffLoading && <div className="text-xs text-gray-400">加载 Diff 详情中...</div>}
                                        </div>
                                    ) : (
                                        <div className="text-sm text-gray-400">暂无 Diff 证据</div>
                                    )}
                                </div>

                                <div className="col-span-5 card">
                                    <div className="flex items-center justify-between mb-3">
                                        <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-1.5"><svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418" /></svg>浏览</h3>
                                        <span className="text-xs text-gray-400">{data.browser?.length || 0}</span>
                                    </div>
                                    {data.browser?.length ? (
                                        <div className="space-y-2">
                                            {data.browser.slice(0, 20).map((b) => (
                                                <div key={b.id} className="p-3 rounded-xl border border-gray-100">
                                                    <div className="text-sm font-medium text-gray-900 truncate">{b.domain}</div>
                                                    <div className="text-xs text-gray-400 line-clamp-2">{b.title || b.url}</div>
                                                    <div className="text-[11px] text-gray-300 mt-1">{formatTs(b.timestamp)}</div>
                                                </div>
                                            ))}
                                        </div>
                                    ) : (
                                        <div className="text-sm text-gray-400">暂无浏览证据</div>
                                    )}
                                </div>
                            </div>

                            {data.rag_refs?.length > 0 && (
                                <div className="card">
                                    <h3 className="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-1.5"><svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23-.693L5 14.5m14.8.8l1.402 1.402c1.232 1.232.65 3.318-1.067 3.611A48.309 48.309 0 0112 21c-2.773 0-5.491-.235-8.135-.687-1.718-.293-2.3-2.379-1.067-3.61L5 14.5" /></svg>使用到的历史记忆（RAG）</h3>
                                    <div className="space-y-2">
                                        {data.rag_refs.map((r, i) => (
                                            <div key={i} className="p-3 rounded-xl border border-gray-100">
                                                <div className="text-xs text-gray-400">{r.type || 'memory'} · {r.date || ''}</div>
                                                <div className="text-sm text-gray-700 mt-1 line-clamp-3">{r.content || ''}</div>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </>
                    )}
                </div>

                {/* Diff Detail Overlay */}
                {diffDetail && createPortal(
                    <div className="fixed inset-0 z-[10000] flex items-center justify-center">
                        <div className="absolute inset-0 bg-black/50" onClick={() => setDiffDetail(null)} />
                        <div className="relative w-[min(1000px,92vw)] max-h-[85vh] overflow-hidden rounded-2xl bg-white shadow-2xl border border-gray-100">
                            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
                                <div className="min-w-0">
                                    <div className="flex items-center gap-3">
                                        <span className="text-xs font-medium text-amber-700 bg-amber-100 px-2.5 py-1 rounded-lg uppercase tracking-wide">{diffDetail.language}</span>
                                        <span className="truncate font-medium">{diffDetail.file_name}</span>
                                    </div>
                                    <div className="text-xs text-gray-400 mt-1 line-clamp-1">{diffDetail.insight || 'AI 正在分析...'}</div>
                                </div>
                                <button onClick={() => setDiffDetail(null)} className="p-2 hover:bg-gray-100 rounded-xl transition-colors text-gray-400 hover:text-gray-700">
                                    <span className="text-lg">×</span>
                                </button>
                            </div>
                            <div className="p-6 overflow-y-auto max-h-[calc(85vh-64px)]">
                                <pre className="text-xs bg-gray-50 border border-gray-100 rounded-xl p-4 overflow-auto">
                                    <code>{diffDetail.diff_content}</code>
                                </pre>
                            </div>
                        </div>
                    </div>,
                    document.body
                )}
            </div>
        </div>
    );

    return createPortal(body, document.body);
};

export default SessionDetailModal;
