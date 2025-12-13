import React, { useEffect, useMemo, useState } from 'react';
import { BuildSessionsForDate, EnrichSessionsForDate, GetSessionsByDate, RebuildSessionsForDate } from '../../api/app';
import SessionDetailModal, { SessionDTO } from '../sessions/SessionDetailModal';

export interface DailySummary {
    date: string;
    summary: string;
    highlights: string;
    struggles: string;
    skills_gained: string[];
    total_coding: number;
    total_diffs: number;
}

export interface PeriodSummary {
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
}

export interface AppStat {
    app_name: string;
    total_duration: number;
    event_count: number;
    is_code_editor: boolean;
}

export interface SkillNode {
    key: string;
    name: string;
    category: string;
}

export interface SummaryIndex {
    date: string;
    has_summary: boolean;
    preview: string;
}

export interface PeriodSummaryIndex {
    type: 'week' | 'month';
    start_date: string;
    end_date: string;
}

export interface SummaryViewProps {
    summary: DailySummary | null;
    periodSummary?: PeriodSummary | null;
    loading: boolean;
    error: string | null;
    onGenerate: () => void;
    onGeneratePeriod?: (type: 'week' | 'month') => void;
    skills?: SkillNode[];
    appStats?: AppStat[];
    summaryIndex?: SummaryIndex[];
    weekSummaryIndex?: PeriodSummaryIndex[];
    monthSummaryIndex?: PeriodSummaryIndex[];
    selectedDate?: string | null;
    onSelectDate?: (date: string) => void;
    onReloadIndex?: () => void;
    onSelectPeriod?: (type: 'week' | 'month', startDate: string) => void;
    onReloadPeriodIndex?: (type: 'week' | 'month') => void;
}

const sessionCategoryLabel = (cat: string): string => {
    switch ((cat || '').toLowerCase()) {
        case 'technical': return '技术';
        case 'learning': return '学习';
        case 'exploration': return '探索';
        case 'other': return '其他';
        default: return cat || '其他';
    }
};

const StatCard: React.FC<{ value: string | number; label: string; }> = ({ value, label }) => (
    <div className="stat-card">
        <div className="flex items-center gap-2 text-gray-400">
            <span className="text-xs font-medium uppercase tracking-wider">{label}</span>
        </div>
        <div className="text-4xl font-bold text-gray-900 tracking-tight">{value}</div>
    </div>
);

const AlertCard: React.FC<{ alerts: { title: string; subtitle: string }[]; total: number; }> = ({ alerts, total }) => (
    <div className="card-dark">
        <div className="flex items-center justify-between mb-4">
            <span className="text-sm font-medium text-white">系统提醒</span>
            <span className="text-2xl font-bold text-white">{alerts.length}/{total}</span>
        </div>
        <div className="space-y-3">
            {alerts.map((alert, i) => (
                <div key={i} className="flex items-start gap-3">
                    <div className="w-6 h-6 rounded-full bg-white/10 flex items-center justify-center mt-0.5">
                        <svg className="w-3 h-3" viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="12" r="4" /></svg>
                    </div>
                    <div>
                        <p className="text-sm font-medium text-white">{alert.title}</p>
                        <p className="text-xs text-gray-400">{alert.subtitle}</p>
                    </div>
                </div>
            ))}
        </div>
    </div>
);

// 阶段汇总视图
const PeriodSummaryCard: React.FC<{ data: PeriodSummary }> = ({ data }) => (
    <div className="space-y-6 animate-slide-up">
        <header className="space-y-4">
            <div className="flex items-end justify-between">
                <div>
                    <h1 className="text-3xl font-bold text-gray-900">
                        📊 {data.type === 'week' ? '本周' : '本月'}汇总
                    </h1>
                    <p className="text-gray-500 mt-1">{data.start_date} 至 {data.end_date}</p>
                </div>
                <div className="flex items-center gap-6">
                    <StatCard value={`${Math.round(data.total_coding / 60)}h`} label="总编码" />
                    <StatCard value={data.total_diffs} label="总变更" />
                </div>
            </div>
        </header>

        <div className="grid grid-cols-12 gap-5">
            {/* 概述 */}
            <div className="col-span-8">
                <div className="card">
                    <h3 className="text-sm font-semibold text-gray-900 mb-3">📝 概述</h3>
                    <p className="text-gray-600 leading-relaxed">{data.overview}</p>
                </div>
            </div>

            {/* 成就 */}
            <div className="col-span-4">
                <div className="card-dark h-full">
                    <h3 className="text-sm font-semibold text-white mb-3">🏆 主要成就</h3>
                    <ul className="space-y-2">
                        {data.achievements?.map((item, i) => (
                            <li key={i} className="text-sm text-gray-300 flex items-start gap-2">
                                <span className="text-accent-gold">✓</span>
                                {item}
                            </li>
                        ))}
                    </ul>
                </div>
            </div>

            {/* 模式分析 */}
            <div className="col-span-6">
                <div className="card">
                    <h3 className="text-sm font-semibold text-gray-900 mb-3">🔍 模式分析</h3>
                    <p className="text-gray-600 text-sm leading-relaxed">{data.patterns}</p>
                </div>
            </div>

            {/* 建议 */}
            <div className="col-span-6">
                <div className="card">
                    <h3 className="text-sm font-semibold text-gray-900 mb-3">💡 下一步建议</h3>
                    <p className="text-gray-600 text-sm leading-relaxed">{data.suggestions}</p>
                </div>
            </div>

            {/* 重点技能 */}
            <div className="col-span-12">
                <div className="card">
                    <h3 className="text-sm font-semibold text-gray-900 mb-3">🎯 重点技能</h3>
                    <div className="flex flex-wrap gap-2">
                        {data.top_skills?.map((skill, i) => (
                            <span key={i} className="pill">{skill}</span>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    </div>
);

// 历史侧边栏
const HistorySidebar: React.FC<{
    summaryIndex: SummaryIndex[];
    selectedDate: string | null;
    onSelectDate: (date: string) => void;
    onReload: () => void;
    onGeneratePeriod?: (type: 'week' | 'month') => void;
    weekSummaryIndex?: PeriodSummaryIndex[];
    monthSummaryIndex?: PeriodSummaryIndex[];
    onSelectPeriod?: (type: 'week' | 'month', startDate: string) => void;
    onReloadPeriodIndex?: (type: 'week' | 'month') => void;
}> = ({ summaryIndex, selectedDate, onSelectDate, onReload, onGeneratePeriod, weekSummaryIndex = [], monthSummaryIndex = [], onSelectPeriod, onReloadPeriodIndex }) => {
    const groupedByMonth = useMemo(() => {
        const groups: Record<string, SummaryIndex[]> = {};
        for (const item of summaryIndex) {
            const monthKey = item.date.slice(0, 7);
            if (!groups[monthKey]) groups[monthKey] = [];
            groups[monthKey].push(item);
        }
        return Object.entries(groups).sort((a, b) => b[0].localeCompare(a[0]));
    }, [summaryIndex]);

    const latestMonth = groupedByMonth[0]?.[0];
    const [expandedMonths, setExpandedMonths] = useState<Record<string, boolean>>(() => {
        const init: Record<string, boolean> = {};
        for (const [m] of groupedByMonth) init[m] = false;
        if (selectedDate) init[selectedDate.slice(0, 7)] = true;
        else if (latestMonth) init[latestMonth] = true;
        return init;
    });

    const toggleMonth = (monthKey: string) => {
        setExpandedMonths(prev => ({ ...prev, [monthKey]: !prev[monthKey] }));
    };

    const groupedWeeks = useMemo(() => {
        const groups: Record<string, PeriodSummaryIndex[]> = {};
        for (const item of weekSummaryIndex) {
            const monthKey = item.start_date.slice(0, 7);
            const groupKey = `week:${monthKey}`;
            if (!groups[groupKey]) groups[groupKey] = [];
            groups[groupKey].push(item);
        }
        return Object.entries(groups).sort((a, b) => b[0].localeCompare(a[0]));
    }, [weekSummaryIndex]);

    const groupedMonths = useMemo(() => {
        const groups: Record<string, PeriodSummaryIndex[]> = {};
        for (const item of monthSummaryIndex) {
            const yearKey = item.start_date.slice(0, 4);
            const groupKey = `month:${yearKey}`;
            if (!groups[groupKey]) groups[groupKey] = [];
            groups[groupKey].push(item);
        }
        return Object.entries(groups).sort((a, b) => b[0].localeCompare(a[0]));
    }, [monthSummaryIndex]);

    const [expandedPeriodGroups, setExpandedPeriodGroups] = useState<Record<string, boolean>>({});
    const togglePeriodGroup = (key: string) => {
        setExpandedPeriodGroups(prev => ({ ...prev, [key]: !prev[key] }));
    };

    return (
        <aside className="card h-fit sticky top-24 w-64 flex-shrink-0">
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-semibold text-gray-900">📁 日报历史</h3>
                <button className="text-xs text-gray-500 hover:text-gray-900" onClick={onReload}>刷新</button>
            </div>

            {/* 快捷汇总按钮 */}
            {onGeneratePeriod && (
                <div className="flex gap-2 mb-4">
                    <button 
                        className="flex-1 text-xs px-2 py-1.5 rounded-lg bg-amber-50 text-amber-700 hover:bg-amber-100 transition"
                        onClick={() => onGeneratePeriod('week')}
                    >
                        📅 本周汇总
                    </button>
                    <button 
                        className="flex-1 text-xs px-2 py-1.5 rounded-lg bg-amber-50 text-amber-700 hover:bg-amber-100 transition"
                        onClick={() => onGeneratePeriod('month')}
                    >
                        📆 本月汇总
                    </button>
                </div>
            )}

            {summaryIndex.length === 0 ? (
                <div className="text-xs text-gray-400">暂无历史索引</div>
            ) : (
                <div className="space-y-1 max-h-[50vh] overflow-y-auto">
                    {groupedByMonth.map(([monthKey, items]) => {
                        const isExpanded = expandedMonths[monthKey];
                        const hasSummaryCount = items.filter(i => i.has_summary).length;
                        return (
                            <div key={monthKey}>
                                <button
                                    className="w-full flex items-center justify-between px-2 py-1.5 rounded-lg hover:bg-gray-50 transition"
                                    onClick={() => toggleMonth(monthKey)}
                                >
                                    <div className="flex items-center gap-2">
                                        <span className="text-sm">{isExpanded ? '📂' : '📁'}</span>
                                        <span className="text-sm font-medium text-gray-900">{monthKey}</span>
                                        <span className="text-xs text-gray-400">({hasSummaryCount})</span>
                                    </div>
                                    <span className="text-xs text-gray-400">{isExpanded ? '▼' : '▶'}</span>
                                </button>

                                {isExpanded && (
                                    <div className="mt-1 ml-4 space-y-0.5">
                                        {items.map((item) => {
                                            const isActive = selectedDate === item.date;
                                            return (
                                                <button
                                                    key={item.date}
                                                    className={`w-full text-left px-2 py-1.5 rounded-md text-sm transition ${isActive ? 'bg-amber-50 text-amber-900' : 'hover:bg-gray-50 text-gray-700'}`}
                                                    onClick={() => onSelectDate(item.date)}
                                                >
                                                    <div className="flex items-center gap-2">
                                                        <span>📄</span>
                                                        <span>{item.date.slice(8, 10)}日</span>
                                                    </div>
                                                    {item.preview && <div className="text-xs text-gray-400 ml-6 truncate">{item.preview}...</div>}
                                                </button>
                                            );
                                        })}
                                    </div>
                                )}
                            </div>
                        );
                    })}
                </div>
            )}

            {/* 周/月汇总历史（生成后才会出现） */}
            <div className="mt-6 space-y-6">
                <div>
                    <div className="flex items-center justify-between mb-3">
                        <h3 className="text-sm font-semibold text-gray-900">🗂️ 周汇总</h3>
                        <button
                            className="text-xs text-gray-500 hover:text-gray-900"
                            onClick={() => onReloadPeriodIndex && onReloadPeriodIndex('week')}
                        >
                            刷新
                        </button>
                    </div>
                    {groupedWeeks.length === 0 ? (
                        <div className="text-xs text-gray-400">暂无周汇总历史</div>
                    ) : (
                        <div className="space-y-2">
                            {groupedWeeks.map(([groupKey, items]) => {
                                const label = groupKey.replace('week:', '');
                                const expanded = !!expandedPeriodGroups[groupKey];
                                return (
                                    <div key={groupKey}>
                                        <button
                                            className="w-full flex items-center justify-between px-2 py-1.5 rounded-lg hover:bg-gray-50 transition"
                                            onClick={() => togglePeriodGroup(groupKey)}
                                        >
                                            <div className="flex items-center gap-2">
                                                <span className="text-sm">{expanded ? '📂' : '📁'}</span>
                                                <span className="text-sm font-medium text-gray-900">{label}</span>
                                                <span className="text-xs text-gray-400">({items.length})</span>
                                            </div>
                                            <span className="text-xs text-gray-400">{expanded ? '▼' : '▶'}</span>
                                        </button>
                                        {expanded && (
                                            <div className="mt-1 ml-4 space-y-0.5">
                                                {items.map((it) => (
                                                    <button
                                                        key={`${it.type}:${it.start_date}:${it.end_date}`}
                                                        className="w-full text-left px-2 py-1.5 rounded-md text-sm transition hover:bg-amber-50 text-gray-700"
                                                        onClick={() => onSelectPeriod && onSelectPeriod('week', it.start_date)}
                                                    >
                                                        <div className="flex items-center gap-2">
                                                            <span>📄</span>
                                                            <span className="text-xs">{it.start_date.slice(5, 10)} ~ {it.end_date.slice(5, 10)}</span>
                                                        </div>
                                                    </button>
                                                ))}
                                            </div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>

                <div>
                    <div className="flex items-center justify-between mb-3">
                        <h3 className="text-sm font-semibold text-gray-900">🗂️ 月汇总</h3>
                        <button
                            className="text-xs text-gray-500 hover:text-gray-900"
                            onClick={() => onReloadPeriodIndex && onReloadPeriodIndex('month')}
                        >
                            刷新
                        </button>
                    </div>
                    {groupedMonths.length === 0 ? (
                        <div className="text-xs text-gray-400">暂无月汇总历史</div>
                    ) : (
                        <div className="space-y-2">
                            {groupedMonths.map(([groupKey, items]) => {
                                const label = groupKey.replace('month:', '');
                                const expanded = !!expandedPeriodGroups[groupKey];
                                return (
                                    <div key={groupKey}>
                                        <button
                                            className="w-full flex items-center justify-between px-2 py-1.5 rounded-lg hover:bg-gray-50 transition"
                                            onClick={() => togglePeriodGroup(groupKey)}
                                        >
                                            <div className="flex items-center gap-2">
                                                <span className="text-sm">{expanded ? '📂' : '📁'}</span>
                                                <span className="text-sm font-medium text-gray-900">{label}</span>
                                                <span className="text-xs text-gray-400">({items.length})</span>
                                            </div>
                                            <span className="text-xs text-gray-400">{expanded ? '▼' : '▶'}</span>
                                        </button>
                                        {expanded && (
                                            <div className="mt-1 ml-4 space-y-0.5">
                                                {items.map((it) => (
                                                    <button
                                                        key={`${it.type}:${it.start_date}:${it.end_date}`}
                                                        className="w-full text-left px-2 py-1.5 rounded-md text-sm transition hover:bg-amber-50 text-gray-700"
                                                        onClick={() => onSelectPeriod && onSelectPeriod('month', it.start_date)}
                                                    >
                                                        <div className="flex items-center gap-2">
                                                            <span>📄</span>
                                                            <span>{it.start_date.slice(0, 7)}</span>
                                                        </div>
                                                    </button>
                                                ))}
                                            </div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>
            </div>
        </aside>
    );
};

const SummaryView: React.FC<SummaryViewProps> = ({
    summary, periodSummary, loading, error, onGenerate, onGeneratePeriod, skills = [], appStats = [],
    summaryIndex = [], weekSummaryIndex = [], monthSummaryIndex = [], selectedDate = null, onSelectDate, onReloadIndex, onSelectPeriod, onReloadPeriodIndex,
}) => {
    const [sessions, setSessions] = useState<SessionDTO[]>([]);
    const [sessionsLoading, setSessionsLoading] = useState(false);
    const [sessionsError, setSessionsError] = useState<string | null>(null);
    const [activeSessionId, setActiveSessionId] = useState<number | null>(null);

    const reloadSessions = async (date: string) => {
        setSessionsLoading(true);
        setSessionsError(null);
        try {
            const res = await GetSessionsByDate(date);
            setSessions(res || []);
        } catch (e: any) {
            setSessionsError(e?.message || '加载会话失败');
            setSessions([]);
        } finally {
            setSessionsLoading(false);
        }
    };

    useEffect(() => {
        if (!summary?.date) {
            setSessions([]);
            return;
        }
        void reloadSessions(summary.date);
    }, [summary?.date]);

    const buildSessions = async () => {
        if (!summary?.date) return;
        setSessionsLoading(true);
        try {
            await BuildSessionsForDate(summary.date);
            await reloadSessions(summary.date);
        } catch (e: any) {
            setSessionsError(e?.message || '切分会话失败');
        } finally {
            setSessionsLoading(false);
        }
    };

    const rebuildSessions = async () => {
        if (!summary?.date) return;
        setSessionsLoading(true);
        try {
            await RebuildSessionsForDate(summary.date);
            await reloadSessions(summary.date);
        } catch (e: any) {
            setSessionsError(e?.message || '重建会话失败');
        } finally {
            setSessionsLoading(false);
        }
    };

    const enrichSessions = async () => {
        if (!summary?.date) return;
        setSessionsLoading(true);
        try {
            await EnrichSessionsForDate(summary.date);
            await reloadSessions(summary.date);
        } catch (e: any) {
            setSessionsError(e?.message || '生成会话摘要失败');
        } finally {
            setSessionsLoading(false);
        }
    };
    const focusStats = useMemo(() => {
        if (!appStats.length) return { focusPercent: 0, codingTime: 0, totalTime: 0 };
        let codingTime = 0, totalTime = 0;
        for (const stat of appStats) {
            totalTime += stat.total_duration;
            if (stat.is_code_editor) codingTime += stat.total_duration;
        }
        return { focusPercent: totalTime > 0 ? Math.round((codingTime / totalTime) * 100) : 0, codingTime, totalTime };
    }, [appStats]);

    const skillDistribution = useMemo(() => {
        if (!skills.length) return [];
        const categoryCount: Record<string, number> = {};
        for (const skill of skills) categoryCount[skill.category || 'other'] = (categoryCount[skill.category || 'other'] || 0) + 1;
        const total = skills.length;
        const labels: Record<string, string> = { language: '编程语言', framework: '框架', database: '数据库', devops: 'DevOps', tool: '工具', concept: '概念', other: '其他' };
        return Object.entries(categoryCount).map(([cat, count]) => ({ category: cat, label: labels[cat] || cat, count, percent: Math.round((count / total) * 100) })).sort((a, b) => b.count - a.count).slice(0, 3);
    }, [skills]);

    const renderMainContent = () => {
        if (loading) {
            return (
                <div className="flex flex-col items-center justify-center min-h-[50vh] gap-6 animate-fade-in">
                    <div className="w-12 h-12 border-2 border-gray-200 border-t-accent-gold rounded-full animate-spin"></div>
                    <p className="text-gray-400 text-sm">生成中，请稍候...</p>
                </div>
            );
        }

        // 显示阶段汇总
        if (periodSummary) {
            return <PeriodSummaryCard data={periodSummary} />;
        }

        // 空状态
        if (!summary) {
            return (
                <div className="flex flex-col items-center justify-center min-h-[50vh] text-center space-y-8 animate-fade-in">
                    <h2 className="text-4xl font-bold text-gray-900">Welcome back, <span className="text-gradient">Developer</span></h2>
                    <p className="text-gray-500 text-lg max-w-md">从左侧选择日期查看历史日报，或生成今日总结。</p>
                    <button className="btn-gold" onClick={onGenerate}>✨ 生成今日总结</button>
                    {error && <div className="px-4 py-2 bg-red-50 border border-red-200 rounded-xl text-red-600 text-sm">{error}</div>}
                </div>
            );
        }

        // 日报详情
        return (
            <div className="space-y-8 animate-slide-up">
                <header className="space-y-6">
                    <div className="flex items-end justify-between">
                        <div><h1 className="text-4xl font-bold text-gray-900">Welcome back, <span className="text-gradient">Developer</span></h1><p className="text-gray-500 mt-1">{summary.date} 日报</p></div>
                        <div className="flex items-center gap-8">
                            <StatCard value={`${Math.round(summary.total_coding / 60)}h`} label="专注时间" />
                            <StatCard value={summary.total_diffs} label="代码变更" />
                            <StatCard value={summary.skills_gained?.length || 0} label="技能增长" />
                        </div>
                    </div>
                    <div className="flex items-center gap-4">
                        <span className="text-sm text-gray-500">编码专注度 {focusStats.focusPercent}%</span>
                        <div className="flex-1 h-2 bg-gray-100 rounded-full overflow-hidden"><div className="h-full bg-gradient-gold" style={{ width: `${focusStats.focusPercent}%` }} /></div>
                    </div>
                </header>

                <div className="grid grid-cols-12 gap-5">
                    <div className="col-span-8"><div className="card"><h3 className="text-sm font-semibold text-gray-900 mb-3">核心总结</h3><p className="text-gray-600 leading-relaxed">{summary.summary}</p></div></div>
                    <div className="col-span-4"><AlertCard alerts={[{ title: '高光时刻', subtitle: summary.highlights?.slice(0, 40) || '暂无' }, { title: '待改进', subtitle: summary.struggles?.slice(0, 40) || '无' }]} total={2} /></div>
                    <div className="col-span-4"><div className="card"><h3 className="text-sm font-semibold text-gray-900 mb-3">技能分布</h3><div className="space-y-2">{skillDistribution.map((item, i) => (<div key={item.category}><div className="flex justify-between text-sm"><span className="text-gray-600">{item.label}</span><span>{item.percent}%</span></div><div className="h-2 bg-gray-100 rounded-full mt-1"><div className="h-full bg-accent-gold rounded-full" style={{ width: `${item.percent}%`, opacity: 1 - i * 0.2 }} /></div></div>))}</div></div></div>
                    <div className="col-span-8"><div className="card"><h3 className="text-sm font-semibold text-gray-900 mb-3">今日习得技能</h3><div className="flex flex-wrap gap-2">{summary.skills_gained?.map((s, i) => <span key={i} className="pill">{s}</span>)}{(!summary.skills_gained?.length) && <span className="text-sm text-gray-400">暂无</span>}</div></div></div>
                </div>

                {/* 会话列表（证据链入口） */}
                <div className="card">
                    <div className="flex items-center justify-between mb-3">
                        <div className="space-y-1">
                            <h3 className="text-sm font-semibold text-gray-900">🧩 今日会话</h3>
                            <p className="text-xs text-gray-400">点击会话可展开窗口/Diff/浏览证据</p>
                        </div>
                        <div className="flex items-center gap-2">
                            <button className="text-xs px-3 py-2 rounded-lg bg-gray-50 hover:bg-gray-100 transition" onClick={buildSessions} disabled={sessionsLoading}>
                                切分
                            </button>
                            <button className="text-xs px-3 py-2 rounded-lg bg-red-50 text-red-700 hover:bg-red-100 transition" onClick={rebuildSessions} disabled={sessionsLoading}>
                                重建
                            </button>
                            <button className="text-xs px-3 py-2 rounded-lg bg-amber-50 text-amber-700 hover:bg-amber-100 transition" onClick={enrichSessions} disabled={sessionsLoading}>
                                生成摘要
                            </button>
                            <button className="text-xs px-3 py-2 rounded-lg bg-gray-50 hover:bg-gray-100 transition" onClick={() => summary?.date && reloadSessions(summary.date)} disabled={sessionsLoading}>
                                刷新
                            </button>
                        </div>
                    </div>

                    {sessionsLoading && sessions.length === 0 && (
                        <div className="text-sm text-gray-400">加载中...</div>
                    )}
                    {sessionsError && (
                        <div className="text-sm text-red-500 mb-2">{sessionsError}</div>
                    )}
                    {(!sessionsLoading && sessions.length === 0) && (
                        <div className="text-sm text-gray-400">暂无会话记录（可先点击“切分”，再点击“生成摘要”）</div>
                    )}
                    {sessions.length > 0 && (
                        <div className="space-y-2">
                            {sessions.map((s) => (
                                <button
                                    key={s.id}
                                    className="w-full text-left p-3 rounded-xl border border-gray-100 hover:border-amber-200 hover:bg-amber-50/40 transition"
                                    onClick={() => setActiveSessionId(s.id)}
                                >
                                    <div className="flex items-center justify-between gap-3">
                                        <div className="min-w-0">
                                            <div className="flex items-center gap-2">
                                                <span className="text-sm font-semibold text-gray-900">{s.time_range || '会话'}</span>
                                                {s.category && <span className="pill">{sessionCategoryLabel(s.category)}</span>}
                                                {(s.diff_count || 0) > 0 && <span className="text-xs text-gray-400">Diff {s.diff_count}</span>}
                                                {(s.browser_count || 0) > 0 && <span className="text-xs text-gray-400">Browser {s.browser_count}</span>}
                                            </div>
                                            <div className="text-xs text-gray-400 truncate">{s.primary_app || ''}</div>
                                        </div>
                                        <div className="text-sm text-gray-700 line-clamp-2 max-w-[55%]">{s.summary || '（未生成摘要）'}</div>
                                    </div>
                                </button>
                            ))}
                        </div>
                    )}
                </div>

                {activeSessionId && (
                    <SessionDetailModal sessionId={activeSessionId} onClose={() => setActiveSessionId(null)} />
                )}
            </div>
        );
    };

    return (
        <div className="flex gap-6 pb-12">
            <HistorySidebar
                summaryIndex={summaryIndex}
                selectedDate={selectedDate}
                onSelectDate={onSelectDate || (() => {})}
                onReload={onReloadIndex || (() => {})}
                onGeneratePeriod={onGeneratePeriod}
                weekSummaryIndex={weekSummaryIndex}
                monthSummaryIndex={monthSummaryIndex}
                onSelectPeriod={onSelectPeriod}
                onReloadPeriodIndex={onReloadPeriodIndex}
            />
            <div className="flex-1">{renderMainContent()}</div>
        </div>
    );
};

export default SummaryView;
