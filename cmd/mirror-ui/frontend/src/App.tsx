import { useState, useEffect } from 'react';
import './App.css';
// @ts-ignore
import { GetTodaySummary, GetDailySummary, ListSummaryIndex, GetSkillTree, GetAppStats } from "../wailsjs/go/main/App";
import MainLayout from './components/layout/MainLayout';
import SummaryView, { DailySummary, AppStat, SummaryIndex } from './components/dashboard/SummaryView';
import SkillView, { SkillNode } from './components/skills/SkillView';
import TrendsView from './components/dashboard/TrendsView';

function App() {
    // 状态管理
    const [activeTab, setActiveTab] = useState<'summary' | 'skills' | 'trends'>('summary');
    
    // 数据状态
    const [summary, setSummary] = useState<DailySummary | null>(null);
    const [summaryIndex, setSummaryIndex] = useState<SummaryIndex[]>([]);
    const [selectedDate, setSelectedDate] = useState<string | null>(null);
    const [skills, setSkills] = useState<SkillNode[]>([]);
    const [appStats, setAppStats] = useState<AppStat[]>([]);
    
    // UI状态
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // 加载指定日期总结（为空则加载今日）
    const loadSummary = async (date?: string) => {
        setLoading(true);
        setError(null);
        try {
            const targetDate = date || new Date().toISOString().slice(0, 10);
            const result = date ? await GetDailySummary(targetDate) : await GetTodaySummary();
            setSummary(result);
            setSelectedDate(targetDate);
        } catch (e: any) {
            setError(e.message || '加载失败');
        } finally {
            setLoading(false);
        }
    };

    // 加载历史索引
    const loadSummaryIndex = async (days: number = 60) => {
        try {
            const result = await ListSummaryIndex(days);
            setSummaryIndex(result || []);
        } catch (e: any) {
            console.error('加载历史索引失败:', e);
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
    const loadAppStats = async () => {
        try {
            const result = await GetAppStats();
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
    }, []);

    // 视图渲染逻辑
    const renderContent = () => {
        switch (activeTab) {
            case 'summary':
                return (
                    <SummaryView 
                        summary={summary} 
                        loading={loading} 
                        error={error} 
                        onGenerate={() => { void loadSummary(); }}
                        skills={skills}
                        appStats={appStats}
                        summaryIndex={summaryIndex}
                        selectedDate={selectedDate}
                        onSelectDate={(date: string) => { void loadSummary(date); }}
                        onReloadIndex={() => { void loadSummaryIndex(); }}
                    />
                );
            case 'skills':
                return <SkillView skills={skills} />;
            case 'trends':
                return <TrendsView />;
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
