import React, { useState, useMemo } from 'react';

export interface SkillNode {
    key: string;
    name: string;
    category: string;
    parent_key?: string;
    level: number;
    experience: number;
    progress: number;
    status: string;
    last_active?: number;
    created_at?: string;
}

interface SkillViewProps {
    skills: SkillNode[];
}

// 状态配置
const statusConfig: Record<string, { color: string; bgColor: string; label: string }> = {
    growing: { color: '#22C55E', bgColor: '#DCFCE7', label: '成长中' },
    declining: { color: '#EF4444', bgColor: '#FEE2E2', label: '下滑' },
    stable: { color: '#6B7280', bgColor: '#F3F4F6', label: '稳定' },
    up: { color: '#22C55E', bgColor: '#DCFCE7', label: '成长中' },
    down: { color: '#EF4444', bgColor: '#FEE2E2', label: '下滑' },
};

// 分类配置 - 使用字母缩写代替 emoji
const categoryConfig: Record<string, { abbr: string; label: string; color: string }> = {
    language: { abbr: 'L', label: '编程语言', color: '#3B82F6' },
    framework: { abbr: 'F', label: '框架', color: '#8B5CF6' },
    database: { abbr: 'D', label: '数据库', color: '#06B6D4' },
    devops: { abbr: 'O', label: 'DevOps', color: '#F59E0B' },
    tool: { abbr: 'T', label: '工具', color: '#10B981' },
    concept: { abbr: 'C', label: '概念', color: '#EC4899' },
    other: { abbr: '?', label: '其他', color: '#6B7280' },
};

// 树节点组件
interface TreeNodeProps {
    skill: SkillNode;
    children: SkillNode[];
    allSkills: SkillNode[];
    depth: number;
    getStatus: (status: string) => { color: string; bgColor: string; label: string };
    getCategory: (category: string) => { abbr: string; label: string; color: string };
}

const TreeNode: React.FC<TreeNodeProps> = ({ skill, children, allSkills, depth, getStatus, getCategory }) => {
    const [isExpanded, setIsExpanded] = useState(true);
    const status = getStatus(skill.status);
    const category = getCategory(skill.category);
    const hasChildren = children.length > 0;

    const formatLastActive = (timestamp?: number) => {
        if (!timestamp) return '暂无记录';
        const date = new Date(timestamp);
        const now = new Date();
        const diff = now.getTime() - date.getTime();
        const days = Math.floor(diff / (1000 * 60 * 60 * 24));
        
        if (days === 0) return '今天';
        if (days === 1) return '昨天';
        if (days < 7) return `${days}天前`;
        if (days < 30) return `${Math.floor(days / 7)}周前`;
        return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
    };

    return (
        <div className="relative">
            {depth > 0 && (
                <div 
                    className="absolute left-0 top-0 bottom-0 w-px bg-gray-200"
                    style={{ left: `${(depth - 1) * 32 + 16}px` }}
                />
            )}
            
            <div 
                className="flex items-center gap-3 py-2 group"
                style={{ paddingLeft: `${depth * 32}px` }}
            >
                {hasChildren ? (
                    <button
                        onClick={() => setIsExpanded(!isExpanded)}
                        className="w-6 h-6 rounded-lg bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors flex-shrink-0"
                    >
                        <svg 
                            className={`w-4 h-4 text-gray-500 transition-transform duration-200 ${isExpanded ? 'rotate-90' : ''}`}
                            fill="none" 
                            viewBox="0 0 24 24" 
                            stroke="currentColor"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                    </button>
                ) : (
                    <div className="w-6 h-6 flex items-center justify-center flex-shrink-0">
                        <div className="w-2 h-2 rounded-full bg-gray-300" />
                    </div>
                )}

                <div className="flex-1 card !p-4 hover:shadow-card-lg transition-all duration-300 group-hover:-translate-y-0.5">
                    <div className="flex items-center gap-4">
                        <div 
                            className="w-10 h-10 rounded-xl flex items-center justify-center text-sm font-bold text-white flex-shrink-0"
                            style={{ backgroundColor: category.color }}
                        >
                            {category.abbr}
                        </div>

                        <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 mb-1">
                                <h4 className="font-semibold text-gray-900 truncate">{skill.name}</h4>
                                <span 
                                    className="px-2 py-0.5 rounded-full text-xs font-medium flex-shrink-0"
                                    style={{ backgroundColor: status.bgColor, color: status.color }}
                                >
                                    {status.label}
                                </span>
                            </div>
                            <div className="flex items-center gap-3 text-xs text-gray-400">
                                <span>Lv.{skill.level}</span>
                                <span>{skill.experience} XP</span>
                                <span>{formatLastActive(skill.last_active)}</span>
                            </div>
                        </div>

                        <div className="w-24 flex-shrink-0">
                            <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                                <div 
                                    className="h-full rounded-full bg-gradient-gold transition-all duration-700"
                                    style={{ width: `${Math.min(skill.progress, 100)}%` }}
                                />
                            </div>
                            <div className="text-xs text-gray-400 text-right mt-1">{skill.progress}%</div>
                        </div>

                        {hasChildren && (
                            <div className="px-2.5 py-1 bg-gray-100 rounded-full text-xs font-medium text-gray-500 flex-shrink-0">
                                {children.length} 子技能
                            </div>
                        )}
                    </div>
                </div>
            </div>

            {hasChildren && isExpanded && (
                <div className="relative">
                    {children.map(child => {
                        const grandChildren = allSkills.filter(s => s.parent_key === child.key);
                        return (
                            <TreeNode
                                key={child.key}
                                skill={child}
                                children={grandChildren}
                                allSkills={allSkills}
                                depth={depth + 1}
                                getStatus={getStatus}
                                getCategory={getCategory}
                            />
                        );
                    })}
                </div>
            )}
        </div>
    );
};

// 成长记录卡片
const GrowthTimeline: React.FC<{ skills: SkillNode[] }> = ({ skills }) => {
    const recentSkills = useMemo(() => {
        return [...skills]
            .filter(s => s.last_active)
            .sort((a, b) => (b.last_active || 0) - (a.last_active || 0))
            .slice(0, 5);
    }, [skills]);

    const formatTime = (timestamp: number) => {
        const date = new Date(timestamp);
        return date.toLocaleString('zh-CN', { 
            month: 'short', 
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    if (recentSkills.length === 0) return null;

    return (
        <div className="card-dark">
            <div className="flex items-center justify-between mb-4">
                <h3 className="font-semibold text-white">成长记录</h3>
                <span className="text-xs text-gray-400">最近活动</span>
            </div>
            <div className="space-y-3">
                {recentSkills.map((skill) => (
                    <div key={skill.key} className="flex items-center gap-3">
                        <div className="w-2 h-2 rounded-full bg-accent-gold flex-shrink-0" />
                        <div className="flex-1 min-w-0">
                            <div className="text-sm font-medium text-white truncate">{skill.name}</div>
                            <div className="text-xs text-gray-400">
                                +{skill.experience} XP / Lv.{skill.level}
                            </div>
                        </div>
                        <span className="text-xs text-gray-500 flex-shrink-0">
                            {skill.last_active ? formatTime(skill.last_active) : ''}
                        </span>
                    </div>
                ))}
            </div>
        </div>
    );
};

// 统计摘要
const SkillStats: React.FC<{ skills: SkillNode[] }> = ({ skills }) => {
    const stats = useMemo(() => {
        const totalExp = skills.reduce((sum, s) => sum + s.experience, 0);
        const avgLevel = skills.length > 0 
            ? Math.round(skills.reduce((sum, s) => sum + s.level, 0) / skills.length * 10) / 10
            : 0;
        const growingCount = skills.filter(s => s.status === 'growing' || s.status === 'up').length;
        
        return { totalExp, avgLevel, growingCount };
    }, [skills]);

    return (
        <div className="grid grid-cols-4 gap-4">
            <div className="card text-center">
                <div className="text-2xl font-bold text-gray-900">{skills.length}</div>
                <div className="text-xs text-gray-400">技能总数</div>
            </div>
            <div className="card text-center">
                <div className="text-2xl font-bold text-accent-gold">{stats.totalExp}</div>
                <div className="text-xs text-gray-400">总经验值</div>
            </div>
            <div className="card text-center">
                <div className="text-2xl font-bold text-gray-900">{stats.avgLevel}</div>
                <div className="text-xs text-gray-400">平均等级</div>
            </div>
            <div className="card text-center">
                <div className="text-2xl font-bold text-green-500">{stats.growingCount}</div>
                <div className="text-xs text-gray-400">成长中</div>
            </div>
        </div>
    );
};

const SkillView: React.FC<SkillViewProps> = ({ skills }) => {
    const [viewMode, setViewMode] = useState<'tree' | 'grid'>('tree');
    const [activeCategory, setActiveCategory] = useState<string | null>(null);
    
    const getStatus = (status: string) => statusConfig[status] || statusConfig.stable;
    const getCategory = (category: string) => categoryConfig[category] || categoryConfig.other;

    const categories = useMemo(() => 
        Array.from(new Set(skills.map(s => s.category))), 
        [skills]
    );

    const filteredSkills = useMemo(() => 
        activeCategory ? skills.filter(s => s.category === activeCategory) : skills,
        [skills, activeCategory]
    );

    const rootSkills = useMemo(() => 
        filteredSkills.filter(s => !s.parent_key || !filteredSkills.find(p => p.key === s.parent_key)),
        [filteredSkills]
    );

    if (!skills || skills.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] text-center space-y-6 animate-fade-in">
                <div className="w-20 h-20 rounded-3xl bg-gray-100 flex items-center justify-center">
                    <svg className="w-10 h-10 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z" />
                    </svg>
                </div>
                <div>
                    <h2 className="text-2xl font-bold text-gray-900 mb-2">技能树空空如也</h2>
                    <p className="text-gray-500">开始编写代码，您的技能树将自动成长</p>
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-6 pb-12 animate-slide-up">
            {/* 头部 */}
            <header className="flex items-end justify-between">
                <div>
                    <h1 className="text-4xl font-bold text-gray-900 tracking-tight">技能树</h1>
                    <p className="text-gray-500 mt-1">Skill Tree</p>
                </div>
                <div className="flex items-center gap-3">
                    {/* 视图切换 */}
                    <div className="flex items-center bg-gray-100 rounded-full p-1">
                        <button
                            className={`px-3 py-1.5 text-xs font-medium rounded-full transition-colors ${
                                viewMode === 'tree' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500'
                            }`}
                            onClick={() => setViewMode('tree')}
                        >
                            树形
                        </button>
                        <button
                            className={`px-3 py-1.5 text-xs font-medium rounded-full transition-colors ${
                                viewMode === 'grid' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500'
                            }`}
                            onClick={() => setViewMode('grid')}
                        >
                            网格
                        </button>
                    </div>
                    <div className="px-4 py-2 bg-white rounded-full shadow-sm text-sm font-medium text-gray-600">
                        {skills.length} 个技能
                    </div>
                </div>
            </header>

            {/* 统计摘要 */}
            <SkillStats skills={skills} />

            {/* 分类筛选 */}
            <div className="flex items-center gap-2 flex-wrap">
                <button
                    className={`pill ${!activeCategory ? 'pill-active' : ''} transition-colors`}
                    onClick={() => setActiveCategory(null)}
                >
                    全部
                </button>
                {categories.map(cat => {
                    const config = getCategory(cat);
                    return (
                        <button
                            key={cat}
                            className={`pill ${activeCategory === cat ? 'pill-active' : ''} transition-colors`}
                            onClick={() => setActiveCategory(cat)}
                        >
                            {config.label}
                        </button>
                    );
                })}
            </div>

            {/* 主内容区 */}
            <div className="grid grid-cols-12 gap-6">
                {/* 技能树/网格 */}
                <div className="col-span-9">
                    {viewMode === 'tree' ? (
                        <div className="card !p-6">
                            <div className="space-y-1">
                                {rootSkills.map(skill => {
                                    const children = filteredSkills.filter(s => s.parent_key === skill.key);
                                    return (
                                        <TreeNode
                                            key={skill.key}
                                            skill={skill}
                                            children={children}
                                            allSkills={filteredSkills}
                                            depth={0}
                                            getStatus={getStatus}
                                            getCategory={getCategory}
                                        />
                                    );
                                })}
                            </div>
                        </div>
                    ) : (
                        <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
                            {filteredSkills.map((skill) => {
                                const status = getStatus(skill.status);
                                const category = getCategory(skill.category);
                                
                                return (
                                    <div key={skill.key} className="card group hover:shadow-card-lg transition-all duration-300">
                                        <div className="flex justify-between items-start mb-4">
                                            <div 
                                                className="w-12 h-12 rounded-2xl flex items-center justify-center text-lg font-bold text-white"
                                                style={{ backgroundColor: category.color }}
                                            >
                                                {category.abbr}
                                            </div>
                                            <span 
                                                className="px-2 py-1 rounded-full text-xs font-medium"
                                                style={{ backgroundColor: status.bgColor, color: status.color }}
                                            >
                                                {status.label}
                                            </span>
                                        </div>
                                        
                                        <h3 className="text-lg font-semibold text-gray-900 mb-1 truncate">{skill.name}</h3>
                                        <div className="text-xs text-gray-400 mb-4">{category.label}</div>

                                        <div className="flex justify-between items-center mb-2 text-sm">
                                            <span className="font-semibold text-gray-900">Lv.{skill.level}</span>
                                            <span className="text-gray-400">{skill.experience} XP</span>
                                        </div>

                                        <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                                            <div 
                                                className="h-full rounded-full bg-gradient-gold transition-all duration-700"
                                                style={{ width: `${Math.min(skill.progress, 100)}%` }}
                                            />
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>

                {/* 成长记录侧边栏 */}
                <div className="col-span-3">
                    <GrowthTimeline skills={skills} />
                </div>
            </div>
        </div>
    );
};

export default SkillView;
