import React, { useMemo, useState } from 'react';
import { useSessionsByDate } from '../../hooks/useSessionsByDate';
import SessionList from '../sessions/SessionList';
import SessionDetailModal from '../sessions/SessionDetailModal';

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
                    <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-2">
                        <svg className="w-7 h-7 text-accent-gold" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" /></svg>
                        {data.type === 'week' ? '本周' : '本月'}汇总
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
                    <h3 className="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-1.5"><svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" /></svg>概述</h3>
                    <p className="text-gray-600 leading-relaxed">{data.overview}</p>
                </div>
            </div>

            {/* 成就 */}
            <div className="col-span-4">
                <div className="card-dark h-full">
                    <h3 className="text-sm font-semibold text-white mb-3 flex items-center gap-1.5"><svg className="w-4 h-4 text-accent-gold" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M16.5 18.75h-9m9 0a3 3 0 013 3h-15a3 3 0 013-3m9 0v-3.375c0-.621-.503-1.125-1.125-1.125h-.871M7.5 18.75v-3.375c0-.621.504-1.125 1.125-1.125h.872m5.007 0H9.497m5.007 0a7.454 7.454 0 01-.982-3.172M9.497 14.25a7.454 7.454 0 00.981-3.172M5.25 4.236c-.982.143-1.954.317-2.916.52A6.003 6.003 0 007.73 9.728M5.25 4.236V4.5c0 2.108.966 3.99 2.48 5.228M5.25 4.236V2.721C7.456 2.41 9.71 2.25 12 2.25c2.291 0 4.545.16 6.75.47v1.516M7.73 9.728a6.726 6.726 0 002.748 1.35m8.272-6.842V4.5c0 2.108-.966 3.99-2.48 5.228m2.48-5.492a46.32 46.32 0 012.916.52 6.003 6.003 0 01-5.395 4.972m0 0a6.726 6.726 0 01-2.749 1.35m0 0a6.772 6.772 0 01-3.044 0" /></svg>主要成就</h3>
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
                    <h3 className="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-1.5"><svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" /></svg>模式分析</h3>
                    <p className="text-gray-600 text-sm leading-relaxed">{data.patterns}</p>
                </div>
            </div>

            {/* 建议 */}
            <div className="col-span-6">
                <div className="card">
                    <h3 className="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-1.5"><svg className="w-4 h-4 text-accent-gold" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M12 18v-5.25m0 0a6.01 6.01 0 001.5-.189m-1.5.189a6.01 6.01 0 01-1.5-.189m3.75 7.478a12.06 12.06 0 01-4.5 0m3.75 2.383a14.406 14.406 0 01-3 0M14.25 18v-.192c0-.983.658-1.823 1.508-2.316a7.5 7.5 0 10-7.517 0c.85.493 1.509 1.333 1.509 2.316V18" /></svg>下一步建议</h3>
                    <p className="text-gray-600 text-sm leading-relaxed">{data.suggestions}</p>
                </div>
            </div>

            {/* 重点技能 */}
            <div className="col-span-12">
                <div className="card">
                    <h3 className="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-1.5"><svg className="w-4 h-4 text-accent-gold" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" /></svg>重点技能</h3>
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
                <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-1.5"><svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z" /></svg>日报历史</h3>
                <button className="text-xs text-gray-500 hover:text-gray-900" onClick={onReload}>刷新</button>
            </div>

            {/* 快捷汇总按钮 */}
            {onGeneratePeriod && (
                <div className="flex gap-2 mb-4">
                    <button 
                        className="flex-1 text-xs px-2 py-1.5 rounded-lg bg-amber-50 text-amber-700 hover:bg-amber-100 transition"
                        onClick={() => onGeneratePeriod('week')}
                    >
                                                <svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" /></svg>本周汇总
                    </button>
                    <button 
                        className="flex-1 text-xs px-2 py-1.5 rounded-lg bg-amber-50 text-amber-700 hover:bg-amber-100 transition"
                        onClick={() => onGeneratePeriod('month')}
                    >
                                                <svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5m-9-6h.008v.008H12v-.008zM12 15h.008v.008H12V15zm0 2.25h.008v.008H12v-.008zM9.75 15h.008v.008H9.75V15zm0 2.25h.008v.008H9.75v-.008zM7.5 15h.008v.008H7.5V15zm0 2.25h.008v.008H7.5v-.008zm6.75-4.5h.008v.008h-.008v-.008zm0 2.25h.008v.008h-.008V15zm0 2.25h.008v.008h-.008v-.008zm2.25-4.5h.008v.008H16.5v-.008zm0 2.25h.008v.008H16.5V15z" /></svg>本月汇总
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
                                                                                <svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M3.75 9.776c.112-.017.227-.026.344-.026h15.812c.117 0 .232.009.344.026m-16.5 0a2.25 2.25 0 00-1.883 2.542l.857 6a2.25 2.25 0 002.227 1.932H19.05a2.25 2.25 0 002.227-1.932l.857-6a2.25 2.25 0 00-1.883-2.542m-16.5 0V6A2.25 2.25 0 016 3.75h3.879a1.5 1.5 0 011.06.44l2.122 2.12a1.5 1.5 0 001.06.44H18A2.25 2.25 0 0120.25 9v.776" /></svg>
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
                                                                                                                <svg className="w-3.5 h-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" /></svg>
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
                        <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-1.5"><svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 010 3.75H5.625a1.875 1.875 0 010-3.75z" /></svg>周汇总</h3>
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
                                                                                                <svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M3.75 9.776c.112-.017.227-.026.344-.026h15.812c.117 0 .232.009.344.026m-16.5 0a2.25 2.25 0 00-1.883 2.542l.857 6a2.25 2.25 0 002.227 1.932H19.05a2.25 2.25 0 002.227-1.932l.857-6a2.25 2.25 0 00-1.883-2.542m-16.5 0V6A2.25 2.25 0 016 3.75h3.879a1.5 1.5 0 011.06.44l2.122 2.12a1.5 1.5 0 001.06.44H18A2.25 2.25 0 0120.25 9v.776" /></svg>
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
                                                                                                                        <svg className="w-3.5 h-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" /></svg>
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
                        <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-1.5"><svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 010 3.75H5.625a1.875 1.875 0 010-3.75z" /></svg>月汇总</h3>
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
                                                                                                <svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M3.75 9.776c.112-.017.227-.026.344-.026h15.812c.117 0 .232.009.344.026m-16.5 0a2.25 2.25 0 00-1.883 2.542l.857 6a2.25 2.25 0 002.227 1.932H19.05a2.25 2.25 0 002.227-1.932l.857-6a2.25 2.25 0 00-1.883-2.542m-16.5 0V6A2.25 2.25 0 016 3.75h3.879a1.5 1.5 0 011.06.44l2.122 2.12a1.5 1.5 0 001.06.44H18A2.25 2.25 0 0120.25 9v.776" /></svg>
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
                                                                                                                        <svg className="w-3.5 h-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" /></svg>
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
    const sessionsHook = useSessionsByDate(summary?.date);
    const [activeSessionId, setActiveSessionId] = useState<number | null>(null);
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
        const gained = summary?.skills_gained || [];
        if (!gained.length) return [];

        const nameToCategory = new Map<string, string>();
        for (const sk of skills) {
            const name = (sk.name || '').trim().toLowerCase();
            if (!name) continue;
            nameToCategory.set(name, sk.category || 'other');
        }

        const categoryCount: Record<string, number> = {};
        for (const rawName of gained) {
            const name = (rawName || '').trim().toLowerCase();
            if (!name) continue;
            const cat = nameToCategory.get(name) || 'other';
            categoryCount[cat] = (categoryCount[cat] || 0) + 1;
        }

        const total = Object.values(categoryCount).reduce((sum, v) => sum + v, 0);
        if (total <= 0) return [];

        const labels: Record<string, string> = { language: '编程语言', framework: '框架', database: '数据库', devops: 'DevOps', tool: '工具', concept: '概念', other: '其他' };
        return Object.entries(categoryCount)
            .map(([cat, count]) => ({ category: cat, label: labels[cat] || cat, count, percent: Math.round((count / total) * 100) }))
            .sort((a, b) => b.count - a.count)
            .slice(0, 3);
    }, [summary?.skills_gained, skills]);

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
                    <p className="text-gray-500 text-lg max-w-md">从左侧选择日期查看历史日报，或生成/刷新今日总结。</p>
                    <button className="btn-gold flex items-center gap-2" onClick={onGenerate}><svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456z" /></svg>生成/刷新今日总结</button>
                    <div className="text-xs text-gray-400">影响范围：仅重新生成总结（Summary），不会重建会话（Session）。</div>
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
                            <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-1.5"><svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M14.25 6.087c0-.355.186-.676.401-.959.221-.29.349-.634.349-1.003 0-1.036-1.007-1.875-2.25-1.875s-2.25.84-2.25 1.875c0 .369.128.713.349 1.003.215.283.401.604.401.959v0a.64.64 0 01-.657.643 48.39 48.39 0 01-4.163-.3c.186 1.613.293 3.25.315 4.907a.656.656 0 01-.658.663v0c-.355 0-.676-.186-.959-.401a1.647 1.647 0 00-1.003-.349c-1.036 0-1.875 1.007-1.875 2.25s.84 2.25 1.875 2.25c.369 0 .713-.128 1.003-.349.283-.215.604-.401.959-.401v0c.31 0 .555.26.532.57a48.039 48.039 0 01-.642 5.056c1.518.19 3.058.309 4.616.354a.64.64 0 00.657-.643v0c0-.355-.186-.676-.401-.959a1.647 1.647 0 01-.349-1.003c0-1.035 1.008-1.875 2.25-1.875 1.243 0 2.25.84 2.25 1.875 0 .369-.128.713-.349 1.003-.215.283-.401.604-.401.959v0c0 .333.277.599.61.58a48.1 48.1 0 005.427-.63 48.05 48.05 0 00.582-4.717.532.532 0 00-.533-.57v0c-.355 0-.676.186-.959.401-.29.221-.634.349-1.003.349-1.035 0-1.875-1.007-1.875-2.25s.84-2.25 1.875-2.25c.37 0 .713.128 1.003.349.283.215.604.401.959.401v0a.656.656 0 00.658-.663 48.422 48.422 0 00-.37-5.36c-1.886.342-3.81.574-5.766.689a.578.578 0 01-.61-.58z" /></svg>今日会话</h3>
                            <p className="text-xs text-gray-400">点击会话可展开窗口/Diff/浏览证据</p>
                        </div>
                        <div className="flex items-center gap-2">
                            <button className="text-xs px-3 py-2 rounded-lg bg-gray-50 hover:bg-gray-100 transition" onClick={() => void sessionsHook.build()} disabled={sessionsHook.loading}>
                                切分会话
                            </button>
                            <button className="text-xs px-3 py-2 rounded-lg bg-red-50 text-red-700 hover:bg-red-100 transition" onClick={() => void sessionsHook.rebuild()} disabled={sessionsHook.loading}>
                                重建会话
                            </button>
                            <button className="text-xs px-3 py-2 rounded-lg bg-amber-50 text-amber-700 hover:bg-amber-100 transition" onClick={() => void sessionsHook.enrich()} disabled={sessionsHook.loading}>
                                补全会话语义
                            </button>
                            <button className="text-xs px-3 py-2 rounded-lg bg-gray-50 hover:bg-gray-100 transition" onClick={() => void sessionsHook.reload()} disabled={sessionsHook.loading}>
                                刷新
                            </button>
                        </div>
                    </div>
                    <div className="text-xs text-gray-400 mb-2">影响范围：重建会话会改变会话口径（version 覆盖展示）；补全仅更新摘要/证据索引。</div>

                    {sessionsHook.loading && sessionsHook.sessions.length === 0 && (
                        <div className="text-sm text-gray-400">加载中...</div>
                    )}
                    {sessionsHook.error && (
                        <div className="text-sm text-red-500 mb-2">{sessionsHook.error}</div>
                    )}
                    {(!sessionsHook.loading && sessionsHook.sessions.length === 0) && (
                        <div className="text-sm text-gray-400">暂无会话记录（可先点击“切分会话”，再点击“补全会话语义”）</div>
                    )}
                    <SessionList sessions={sessionsHook.sessions} onSelect={(id) => setActiveSessionId(id)} />
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
