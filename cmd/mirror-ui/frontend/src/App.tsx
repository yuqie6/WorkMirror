import { useState, useEffect } from 'react';
import './App.css';
import { GetTodaySummary, GetSkillTree, GetTrends } from "../wailsjs/go/main/App";

interface DailySummary {
    date: string;
    summary: string;
    highlights: string;
    struggles: string;
    skills_gained: string[];
    total_coding: number;
    total_diffs: number;
}

interface SkillNode {
    key: string;
    name: string;
    category: string;
    level: number;
    experience: number;
    progress: number;
    status: string;
}

function App() {
    const [summary, setSummary] = useState<DailySummary | null>(null);
    const [skills, setSkills] = useState<SkillNode[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [activeTab, setActiveTab] = useState<'summary' | 'skills' | 'trends'>('summary');

    const loadSummary = async () => {
        setLoading(true);
        setError(null);
        try {
            const result = await GetTodaySummary();
            setSummary(result);
        } catch (e: any) {
            setError(e.message || '加载失败');
        } finally {
            setLoading(false);
        }
    };

    const loadSkills = async () => {
        try {
            const result = await GetSkillTree();
            setSkills(result || []);
        } catch (e: any) {
            console.error('加载技能失败:', e);
        }
    };

    useEffect(() => {
        loadSkills();
    }, []);

    const getStatusEmoji = (status: string) => {
        switch (status) {
            case 'growing': return '🔼';
            case 'declining': return '🔽';
            default: return '➡️';
        }
    };

    return (
        <div id="App">
            <header className="header">
                <h1>🪞 Mirror</h1>
                <p className="subtitle">个人成长量化系统</p>
            </header>

            <nav className="tabs">
                <button 
                    className={activeTab === 'summary' ? 'active' : ''} 
                    onClick={() => setActiveTab('summary')}
                >
                    📝 今日总结
                </button>
                <button 
                    className={activeTab === 'skills' ? 'active' : ''} 
                    onClick={() => setActiveTab('skills')}
                >
                    🎯 技能树
                </button>
            </nav>

            <main className="content">
                {activeTab === 'summary' && (
                    <div className="summary-panel">
                        {!summary && !loading && (
                            <button className="btn-primary" onClick={loadSummary}>
                                生成今日总结
                            </button>
                        )}
                        {loading && <div className="loading">⏳ AI 正在分析...</div>}
                        {error && <div className="error">❌ {error}</div>}
                        {summary && (
                            <div className="summary-content">
                                <div className="summary-header">
                                    <h2>📅 {summary.date}</h2>
                                    <div className="stats">
                                        <span>⏱️ {summary.total_coding}分钟</span>
                                        <span>📝 {summary.total_diffs}次变更</span>
                                    </div>
                                </div>
                                <div className="summary-body">
                                    <div className="section">
                                        <h3>📋 总结</h3>
                                        <p>{summary.summary}</p>
                                    </div>
                                    <div className="section">
                                        <h3>🌟 亮点</h3>
                                        <p>{summary.highlights}</p>
                                    </div>
                                    {summary.struggles && summary.struggles !== '无' && (
                                        <div className="section">
                                            <h3>💪 挑战</h3>
                                            <p>{summary.struggles}</p>
                                        </div>
                                    )}
                                    <div className="section">
                                        <h3>🎯 技能</h3>
                                        <div className="tags">
                                            {summary.skills_gained.map((skill, i) => (
                                                <span key={i} className="tag">{skill}</span>
                                            ))}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                )}

                {activeTab === 'skills' && (
                    <div className="skills-panel">
                        <h2>🎯 技能树</h2>
                        {skills.length === 0 ? (
                            <p className="empty">暂无技能数据，开始编码吧！</p>
                        ) : (
                            <div className="skill-list">
                                {skills.map((skill, i) => (
                                    <div key={i} className="skill-card">
                                        <div className="skill-header">
                                            <span className="skill-name">{skill.name}</span>
                                            <span className="skill-status">{getStatusEmoji(skill.status)}</span>
                                        </div>
                                        <div className="skill-category">{skill.category}</div>
                                        <div className="skill-level">Lv.{skill.level}</div>
                                        <div className="progress-bar">
                                            <div 
                                                className="progress-fill" 
                                                style={{ width: `${Math.min(skill.progress, 100)}%` }}
                                            />
                                        </div>
                                        <div className="skill-exp">{skill.experience} EXP</div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                )}
            </main>
        </div>
    );
}

export default App;
