import React, { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
// @ts-ignore
import { GetSessionDetail, GetDiffDetail } from '../../../wailsjs/go/main/App';

export interface SessionDTO {
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
}

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

    useEffect(() => {
        if (!sessionId) return;
        let cancelled = false;
        setLoading(true);
        setError(null);
        setData(null);
        void (async () => {
            try {
                const res = await GetSessionDetail(sessionId);
                if (!cancelled) setData(res);
            } catch (e: any) {
                if (!cancelled) setError(e?.message || '获取会话详情失败');
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
                                <h3 className="text-sm font-semibold text-gray-900 mb-2">📝 摘要</h3>
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
                                    <h3 className="text-sm font-semibold text-gray-900 mb-3">🪟 应用使用</h3>
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

                            <div className="grid grid-cols-12 gap-5">
                                <div className="col-span-7 card">
                                    <div className="flex items-center justify-between mb-3">
                                        <h3 className="text-sm font-semibold text-gray-900">🧾 Diffs</h3>
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
                                        <h3 className="text-sm font-semibold text-gray-900">🌐 浏览</h3>
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
                                    <h3 className="text-sm font-semibold text-gray-900 mb-3">🧠 使用到的历史记忆（RAG）</h3>
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

