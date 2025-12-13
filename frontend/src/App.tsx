import { useState, useEffect, useRef } from 'react';
import './App.css';
import { GetTodaySummary, GetDailySummary, ListSummaryIndex, GetPeriodSummary, ListPeriodSummaryIndex, GetSkillTree, GetAppStats } from "./api/app";
import MainLayout from './components/layout/MainLayout';
import SummaryView, { DailySummary, AppStat, SummaryIndex, PeriodSummary, PeriodSummaryIndex } from './components/dashboard/SummaryView';
import SkillView, { SkillNode } from './components/skills/SkillView';
import TrendsView from './components/dashboard/TrendsView';
import SettingsView from './components/settings/SettingsView';

const normalizePeriodType = (value: string, fallback: 'week' | 'month'): 'week' | 'month' => {
    if (value === 'week' || value === 'month') return value;
    return fallback;
};

const normalizePeriodSummaryIndex = (
    items: Array<{ type: string; start_date: string; end_date: string; }>,
    fallbackType: 'week' | 'month',
): PeriodSummaryIndex[] => {
    return (items || []).map((it) => ({
        type: normalizePeriodType(it.type, fallbackType),
        start_date: it.start_date,
        end_date: it.end_date,
    }));
};

function App() {
    const [activeTab, setActiveTab] = useState<'summary' | 'skills' | 'trends' | 'settings'>('summary');
    
    // 数据状态
    const [summary, setSummary] = useState<DailySummary | null>(null);
    const [periodSummary, setPeriodSummary] = useState<PeriodSummary | null>(null);
    const [summaryIndex, setSummaryIndex] = useState<SummaryIndex[]>([]);
    const [weekSummaryIndex, setWeekSummaryIndex] = useState<PeriodSummaryIndex[]>([]);
    const [monthSummaryIndex, setMonthSummaryIndex] = useState<PeriodSummaryIndex[]>([]);
    const [selectedDate, setSelectedDate] = useState<string | null>(null);
    const [skills, setSkills] = useState<SkillNode[]>([]);
    const [appStats, setAppStats] = useState<AppStat[]>([]);
    const selectedDateRef = useRef<string | null>(null);
    
    // UI状态
    const [loading, setLoading] = useState(false);
    const [periodLoading, setPeriodLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // 加载指定日期总结
    const loadSummary = async (date?: string) => {
        setLoading(true);
        setError(null);
        setPeriodSummary(null); // 清除阶段汇总
        try {
            const targetDate = date || new Date().toISOString().slice(0, 10);
            const result = date ? await GetDailySummary(targetDate) : await GetTodaySummary(true);
            setSummary(result);
            setSelectedDate(targetDate);
            await loadAppStats(targetDate);
        } catch (e: any) {
            setError(e.message || '加载失败');
        } finally {
            setLoading(false);
        }
    };

    // 加载阶段汇总
    const loadPeriodSummary = async (periodType: 'week' | 'month', startDate?: string) => {
        setPeriodLoading(true);
        setError(null);
        setPeriodSummary(null);
        setSummary(null); // 清除日报
        setSelectedDate(null);
        try {
            const result = await GetPeriodSummary(periodType, startDate || "", true); // 用户点击时强制刷新
            setPeriodSummary(result);
            await loadPeriodSummaryIndex(periodType);
        } catch (e: any) {
            setError(e.message || '生成汇总失败');
        } finally {
            setPeriodLoading(false);
        }
    };

    // 加载历史索引
    const loadSummaryIndex = async (days: number = 365) => {
        try {
            const result = await ListSummaryIndex(days);
            setSummaryIndex(result || []);
        } catch (e: any) {
            console.error('加载历史索引失败:', e);
        }
    };

    // 加载周/月汇总历史索引
    const loadPeriodSummaryIndex = async (periodType: 'week' | 'month', limit: number = 60) => {
        try {
            const result = await ListPeriodSummaryIndex(periodType, limit);
            const normalized = normalizePeriodSummaryIndex(result || [], periodType);
            if (periodType === 'week') setWeekSummaryIndex(normalized);
            else setMonthSummaryIndex(normalized);
        } catch (e: any) {
            console.error('加载阶段汇总历史索引失败:', e);
        }
    };

    // 加载技能树
    const loadSkills = async () => {
        try {
            const result = await GetSkillTree();
            setSkills(result || []);
        } catch (e: any) {
            console.error('加载技能失败:', e);
        }
    };

    // 加载应用统计
    const loadAppStats = async (date?: string) => {
        try {
            const effectiveDate = date || selectedDateRef.current || undefined;
            const result = await GetAppStats(effectiveDate);
            setAppStats(result || []);
        } catch (e: any) {
            console.error('加载应用统计失败:', e);
        }
    };

    // 初始加载
    useEffect(() => {
        loadSkills();
        loadAppStats();
        loadSummaryIndex();
        loadPeriodSummaryIndex('week');
        loadPeriodSummaryIndex('month');
    }, []);

    // 订阅 Agent 实时事件：用于触发轻量刷新（避免 UI 侧重复逻辑）
    useEffect(() => {
        const es = new EventSource("/api/events");

        const refresh = () => {
            void loadAppStats();
            void loadSummaryIndex();
        };

        es.addEventListener("data_changed", refresh);

        es.onerror = () => {
            // 浏览器会自动重连；避免噪音
        };

        return () => {
            es.close();
        };
    }, []);

    useEffect(() => {
        selectedDateRef.current = selectedDate;
    }, [selectedDate]);

    // 视图渲染
    const renderContent = () => {
        switch (activeTab) {
            case 'summary':
                return (
                    <SummaryView 
                        summary={summary} 
                        periodSummary={periodSummary}
                        loading={loading || periodLoading} 
                        error={error} 
                        onGenerate={() => { void loadSummary(); }}
                        onGeneratePeriod={(type) => { void loadPeriodSummary(type); }}
                        skills={skills}
                        appStats={appStats}
                        summaryIndex={summaryIndex}
                        weekSummaryIndex={weekSummaryIndex}
                        monthSummaryIndex={monthSummaryIndex}
                        selectedDate={selectedDate}
                        onSelectDate={(date: string) => { void loadSummary(date); }}
                        onReloadIndex={() => { void loadSummaryIndex(); }}
                        onSelectPeriod={(type, startDate) => { void loadPeriodSummary(type, startDate); }}
                        onReloadPeriodIndex={(type) => { void loadPeriodSummaryIndex(type); }}
                    />
                );
            case 'skills':
                return <SkillView skills={skills} />;
            case 'trends':
                return <TrendsView />;
            case 'settings':
                return <SettingsView />;
            default:
                return null;
        }
    };

    return (
        <MainLayout activeTab={activeTab} onTabChange={setActiveTab}>
            {renderContent()}
        </MainLayout>
    );
}

export default App;
