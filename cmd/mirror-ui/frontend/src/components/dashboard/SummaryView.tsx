import React, { useMemo, useState } from 'react';

export interface DailySummary {
    date: string;
    summary: string;
    highlights: string;
    struggles: string;
    skills_gained: string[];
    total_coding: number;
    total_diffs: number;
}

export interface AppStat {
    app_name: string;
    total_duration: number;
    event_count: number;
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

export interface SummaryViewProps {
    summary: DailySummary | null;
    loading: boolean;
    error: string | null;
    onGenerate: () => void;
    skills?: SkillNode[];
    appStats?: AppStat[];
    summaryIndex?: SummaryIndex[];
    selectedDate?: string | null;
    onSelectDate?: (date: string) => void;
    onReloadIndex?: () => void;
}

// 判断是否为编码应用
const isCodeEditor = (appName: string): boolean => {
    const codeEditors = ['code', 'cursor', 'goland', 'idea', 'pycharm', 'webstorm', 'vim', 'nvim', 'sublime', 'atom', 'vscode', 'android studio'];
    const lower = appName.toLowerCase();
    return codeEditors.some(editor => lower.includes(editor));
};

// 统计卡片组件
const StatCard: React.FC<{ value: string | number; label: string; }> = ({ value, label }) => (
    <div className="stat-card">
        <div className="flex items-center gap-2 text-gray-400">
            <span className="text-xs font-medium uppercase tracking-wider">{label}</span>
        </div>
        <div className="text-4xl font-bold text-gray-900 tracking-tight">{value}</div>
    </div>
);

// 主卡片组件 - 深色背景
const MainCard: React.FC<{ title: string; subtitle: string; value: string; }> = ({ title, subtitle, value }) => (
    <div className="card-dark h-full">
        <svg className="absolute inset-0 w-full h-full" viewBox="0 0 400 200" preserveAspectRatio="none">
            <defs>
                <linearGradient id="curveGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                    <stop offset="0%" stopColor="#D4AF37" stopOpacity="0.3" />
                    <stop offset="100%" stopColor="#F6C343" stopOpacity="0.1" />
                </linearGradient>
            </defs>
            <path d="M0,150 Q100,50 200,100 T400,80" fill="none" stroke="url(#curveGradient)" strokeWidth="2" />
            <path d="M0,180 Q150,100 300,120 T400,100" fill="none" stroke="url(#curveGradient)" strokeWidth="1.5" opacity="0.5" />
        </svg>
        <div className="relative z-10">
            <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">{subtitle}</p>
            <h3 className="text-xl font-semibold text-white mb-6">{title}</h3>
            <div className="text-3xl font-bold text-gradient">{value}</div>
        </div>
    </div>
);

// 系统提醒卡片
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

// 历史侧边栏组件
const HistorySidebar: React.FC<{
    summaryIndex: SummaryIndex[];
    selectedDate: string | null;
    onSelectDate: (date: string) => void;
    onReload: () => void;
}> = ({ summaryIndex, selectedDate, onSelectDate, onReload }) => {
    // 按月分组
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

    return (
        <aside className="card h-fit sticky top-24 w-64 flex-shrink-0">
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-semibold text-gray-900">📁 日报历史</h3>
                <button className="text-xs text-gray-500 hover:text-gray-900" onClick={onReload}>刷新</button>
            </div>

            {summaryIndex.length === 0 ? (
                <div className="text-xs text-gray-400">暂无历史索引</div>
            ) : (
                <div className="space-y-1 max-h-[60vh] overflow-y-auto">
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
                                        <span className="text-xs text-gray-400">({hasSummaryCount}/{items.length})</span>
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
                                                    className={`w-full text-left px-2 py-1.5 rounded-md text-sm transition ${isActive ? 'bg-amber-50 text-amber-900' : 'hover:bg-gray-50 text-gray-700'} ${!item.has_summary ? 'opacity-60' : ''}`}
                                                    onClick={() => onSelectDate(item.date)}
                                                >
                                                    <div className="flex items-center gap-2">
                                                        <span>{item.has_summary ? '📄' : '📝'}</span>
                                                        <span>{item.date.slice(8, 10)}日</span>
                                                    </div>
                                                    {item.preview && <div className="text-xs text-gray-400 ml-6 truncate">{item.preview}...</div>}
                                                    {!item.has_summary && <div className="text-xs text-gray-400 ml-6">点击生成</div>}
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
        </aside>
    );
};

const SummaryView: React.FC<SummaryViewProps> = ({
    summary, loading, error, onGenerate, skills = [], appStats = [],
    summaryIndex = [], selectedDate = null, onSelectDate, onReloadIndex,
}) => {
    const focusStats = useMemo(() => {
        if (!appStats.length) return { focusPercent: 0, codingTime: 0, totalTime: 0 };
        let codingTime = 0, totalTime = 0;
        for (const stat of appStats) {
            totalTime += stat.total_duration;
            if (isCodeEditor(stat.app_name)) codingTime += stat.total_duration;
        }
        return { focusPercent: totalTime > 0 ? Math.round((codingTime / totalTime) * 100) : 0, codingTime, totalTime };
    }, [appStats]);

    const skillDistribution = useMemo(() => {
        if (!skills.length) return [];
        const categoryCount: Record<string, number> = {};
        for (const skill of skills) {
            const cat = skill.category || 'other';
            categoryCount[cat] = (categoryCount[cat] || 0) + 1;
        }
        const total = skills.length;
        const labels: Record<string, string> = { language: '编程语言', framework: '框架', database: '数据库', devops: 'DevOps', tool: '工具', concept: '概念', other: '其他' };
        return Object.entries(categoryCount).map(([cat, count]) => ({ category: cat, label: labels[cat] || cat, count, percent: Math.round((count / total) * 100) })).sort((a, b) => b.count - a.count).slice(0, 3);
    }, [skills]);

    const renderMainContent = () => {
        if (loading) {
            return (
                <div className="flex flex-col items-center justify-center min-h-[50vh] gap-6 animate-fade-in">
                    <div className="w-12 h-12 border-2 border-gray-200 border-t-accent-gold rounded-full animate-spin"></div>
                    <p className="text-gray-400 text-sm tracking-wider uppercase">Analyzing Codebase...</p>
                </div>
            );
        }

        if (!summary) {
            return (
                <div className="flex flex-col items-center justify-center min-h-[50vh] text-center space-y-8 animate-fade-in">
                    <div className="relative">
                        <div className="absolute inset-0 bg-accent-gold/20 blur-3xl rounded-full"></div>
                        <h2 className="text-4xl font-bold text-gray-900 relative z-10 tracking-tight">Welcome back, <span className="text-gradient">Developer</span></h2>
                    </div>
                    <p className="text-gray-500 text-lg max-w-md">从左侧选择日期查看历史日报，或生成今日总结。</p>
                    <button className="btn-gold flex items-center gap-2" onClick={onGenerate}>
                        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 00-2.456 2.456z" />
                        </svg>
                        生成今日总结
                    </button>
                    {error && <div className="px-4 py-2 bg-red-50 border border-red-200 rounded-xl text-red-600 text-sm">{error}</div>}
                </div>
            );
        }

        return (
            <div className="space-y-8 animate-slide-up">
                <header className="space-y-6">
                    <div className="flex items-end justify-between">
                        <div><h1 className="text-4xl font-bold text-gray-900 tracking-tight">Welcome back, <span className="text-gradient">Developer</span></h1><p className="text-gray-500 mt-1">{summary.date} 日报</p></div>
                        <div className="flex items-center gap-8">
                            <StatCard value={`${Math.round(summary.total_coding / 60)}h`} label="专注时间" />
                            <StatCard value={summary.total_diffs} label="代码变更" />
                            <StatCard value={summary.skills_gained?.length || 0} label="技能增长" />
                        </div>
                    </div>
                    <div className="flex items-center gap-4">
                        <div className="flex items-center gap-3"><span className="text-sm text-gray-500">编码专注度</span><span className="text-sm font-medium text-gray-900">{focusStats.focusPercent}%</span></div>
                        <div className="flex-1 h-2 bg-gray-100 rounded-full overflow-hidden"><div className="h-full rounded-full bg-gradient-gold transition-all duration-500" style={{ width: `${focusStats.focusPercent}%` }} /></div>
                    </div>
                </header>

                <div className="grid grid-cols-12 gap-5">
                    <div className="col-span-4 row-span-2"><MainCard title="日报概览" subtitle="Daily Overview" value={summary.date} /></div>
                    <div className="col-span-3"><div className="card"><div className="flex items-center justify-between mb-4"><span className="text-xs text-gray-400 uppercase tracking-wider">编码专注</span></div><div className="flex items-baseline gap-2 mb-3"><span className="text-2xl font-bold text-gray-900">{focusStats.focusPercent > 70 ? 'High' : focusStats.focusPercent > 40 ? 'Medium' : 'Low'}</span><span className="px-2 py-0.5 bg-accent-gold/20 text-accent-gold text-xs font-medium rounded-full">{focusStats.focusPercent}%</span></div><div className="text-xs text-gray-500">编码 {Math.round(focusStats.codingTime / 60)} 分钟 / 总计 {Math.round(focusStats.totalTime / 60)} 分钟</div></div></div>
                    <div className="col-span-2"><div className="card h-full flex flex-col items-center justify-center"><span className="text-xs text-gray-400 uppercase tracking-wider mb-2">专注时长</span><div className="relative w-20 h-20"><svg className="w-full h-full -rotate-90" viewBox="0 0 36 36"><circle cx="18" cy="18" r="16" fill="none" stroke="#E5E7EB" strokeWidth="2" /><circle cx="18" cy="18" r="16" fill="none" stroke="#D4AF37" strokeWidth="2" strokeDasharray={`${focusStats.focusPercent} ${100 - focusStats.focusPercent}`} /></svg><div className="absolute inset-0 flex items-center justify-center"><span className="text-lg font-bold text-gray-900">{Math.round(summary.total_coding / 60)}h</span></div></div></div></div>
                    <div className="col-span-3"><div className="card"><span className="text-xs text-gray-400 uppercase tracking-wider">技能分布</span><div className="mt-3 space-y-2">{skillDistribution.length > 0 ? skillDistribution.map((item, i) => (<div key={item.category}><div className="flex items-center justify-between text-sm"><span className="text-gray-600">{item.label}</span><span className="font-medium">{item.percent}%</span></div><div className="h-2 bg-gray-100 rounded-full overflow-hidden mt-1"><div className="h-full bg-accent-gold rounded-full transition-all duration-500" style={{ width: `${item.percent}%`, opacity: 1 - i * 0.2 }} /></div></div>)) : <div className="text-sm text-gray-400">暂无技能数据</div>}</div></div></div>
                    <div className="col-span-5"><div className="card"><div className="flex items-center justify-between mb-4"><span className="text-sm font-medium text-gray-900">核心总结</span></div><p className="text-gray-600 leading-relaxed">{summary.summary}</p></div></div>
                    <div className="col-span-3"><AlertCard alerts={[{ title: '高光时刻', subtitle: summary.highlights?.slice(0, 30) + '...' || '暂无' }, { title: '待改进', subtitle: summary.struggles?.slice(0, 30) + '...' || '无' }]} total={2} /></div>
                    <div className="col-span-12"><div className="card"><div className="flex items-center justify-between mb-4"><span className="text-sm font-medium text-gray-900">今日习得技能</span><span className="text-xs text-gray-400">{summary.skills_gained?.length || 0} skills</span></div><div className="flex flex-wrap gap-2">{summary.skills_gained?.map((skill, i) => (<span key={i} className="pill hover:pill-active transition-colors cursor-default">{skill}</span>))}{(!summary.skills_gained || summary.skills_gained.length === 0) && <span className="text-sm text-gray-400">暂无新技能</span>}</div></div></div>
                </div>
            </div>
        );
    };

    return (
        <div className="flex gap-6 pb-12">
            <HistorySidebar summaryIndex={summaryIndex} selectedDate={selectedDate} onSelectDate={onSelectDate || (() => {})} onReload={onReloadIndex || (() => {})} />
            <div className="flex-1">{renderMainContent()}</div>
        </div>
    );
};

export default SummaryView;
