import React, { useState, useMemo, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { GetSkillEvidence, GetDiffDetail, GetSkillSessions } from '../../api/app';
import SessionDetailModal from '../sessions/SessionDetailModal';
import type { SessionDTO } from '../../types/session';
import EmptyState, { SkillTreeIcon } from '../common/EmptyState';

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

interface SkillEvidence {
    source: string;
    evidence_id: number;
    timestamp: number;
    contribution_context: string;
    file_name: string;
}

interface DiffDetail {
    id: number;
    file_name: string;
    language: string;
    diff_content: string;
    insight: string;
    timestamp: number;
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
    onSelect: (skill: SkillNode) => void;
}

const TreeNode: React.FC<TreeNodeProps> = ({ skill, children, allSkills, depth, getStatus, getCategory, onSelect }) => {
    const [isExpanded, setIsExpanded] = useState(false); // 默认折叠
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

                <div
                    className="flex-1 card !p-4 hover:shadow-card-lg transition-all duration-300 group-hover:-translate-y-0.5 cursor-pointer"
                    onClick={() => onSelect(skill)}
                >
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
                                onSelect={onSelect}
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
    const [selectedSkill, setSelectedSkill] = useState<SkillNode | null>(null);
    const [evidences, setEvidences] = useState<SkillEvidence[]>([]);
    const [evidenceLoading, setEvidenceLoading] = useState(false);
    const [evidenceError, setEvidenceError] = useState<string | null>(null);
    const [skillSessions, setSkillSessions] = useState<SessionDTO[]>([]);
    const [sessionsLoading, setSessionsLoading] = useState(false);
    const [sessionsError, setSessionsError] = useState<string | null>(null);
    const [activeSessionId, setActiveSessionId] = useState<number | null>(null);
    
    // New state for Diff Detail
    const [diffDetail, setDiffDetail] = useState<DiffDetail | null>(null);
    const [diffLoading, setDiffLoading] = useState(false);
    
    const getStatus = (status: string) => statusConfig[status] || statusConfig.stable;
    const getCategory = (category: string) => categoryConfig[category] || categoryConfig.other;

    const loadEvidence = async (skill: SkillNode) => {
        setSelectedSkill(skill);
        setEvidenceLoading(true);
        setEvidenceError(null);
        setSessionsLoading(true);
        setSessionsError(null);
        try {
            const [ev, sess] = await Promise.all([
                GetSkillEvidence(skill.key),
                GetSkillSessions(skill.key),
            ]);
            setEvidences(ev || []);
            setSkillSessions(sess || []);
        } catch (err: any) {
            setEvidenceError(err?.message || '获取证据失败');
            setEvidences([]);
            setSessionsError(err?.message || '获取相关会话失败');
            setSkillSessions([]);
        } finally {
            setEvidenceLoading(false);
            setSessionsLoading(false);
        }
    };

    const loadDiffDetail = async (id: number) => {
        setDiffLoading(true);
        try {
            const res = await GetDiffDetail(id);
            setDiffDetail(res);
        } catch (err: any) {
            console.error("Failed to load diff detail", err);
        } finally {
            setDiffLoading(false);
        }
    };

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

    // Pagination - different strategies for tree vs grid
    const PAGE_SIZE_TREE = 8;  // 树形视图每页显示的根节点数
    const PAGE_SIZE_GRID = 12; // 网格视图每页显示的技能数
    const [currentPage, setCurrentPage] = useState(1);
    
    // Reset page when filter changes
    useEffect(() => {
        setCurrentPage(1);
    }, [activeCategory, viewMode]);

    // 树形视图：按根节点分页，保持完整的父子层级
    const treeTotalPages = useMemo(() => 
        Math.ceil(rootSkills.length / PAGE_SIZE_TREE),
        [rootSkills]
    );

    const paginatedRootSkills = useMemo(() => {
        const start = (currentPage - 1) * PAGE_SIZE_TREE;
        return rootSkills.slice(start, start + PAGE_SIZE_TREE);
    }, [rootSkills, currentPage]);

    // 网格视图：按技能数分页
    const gridTotalPages = useMemo(() => 
        Math.ceil(filteredSkills.length / PAGE_SIZE_GRID),
        [filteredSkills]
    );

    const paginatedSkills = useMemo(() => {
        const start = (currentPage - 1) * PAGE_SIZE_GRID;
        return filteredSkills.slice(start, start + PAGE_SIZE_GRID);
    }, [filteredSkills, currentPage]);

    // 当前视图的总页数
    const totalPages = viewMode === 'tree' ? treeTotalPages : gridTotalPages;

    if (!skills || skills.length === 0) {
        return (
            <EmptyState
                icon={<SkillTreeIcon />}
                title="技能树空空如也"
                description="开始编写代码，您的技能树将自动成长。系统会分析您的代码变更并提取相关技能。"
                action={
                    <a 
                        href="#" 
                        className="text-sm text-accent-gold hover:underline"
                        onClick={(e) => e.preventDefault()}
                    >
                        了解技能树如何工作
                    </a>
                }
            />
        );
    }

    return (
        <div className="h-[calc(100vh-80px)] flex flex-col animate-slide-up">
            {/* 固定头部区域 */}
            <div className="flex-shrink-0 space-y-4 pb-4">
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

                {/* 分页控件 */}
                {totalPages > 1 && (
                    <div className="flex items-center justify-between bg-white rounded-xl px-4 py-3 shadow-sm">
                        <div className="text-sm text-gray-500">
                            {viewMode === 'tree' 
                                ? `共 ${rootSkills.length} 个技能组，第 ${currentPage}/${totalPages} 页`
                                : `共 ${filteredSkills.length} 个技能，第 ${currentPage}/${totalPages} 页`
                            }
                        </div>
                        <div className="flex items-center gap-2">
                            <button
                                className="px-3 py-1.5 text-sm font-medium rounded-lg bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                                disabled={currentPage === 1}
                            >
                                ← 上一页
                            </button>
                            <div className="flex items-center gap-1">
                                {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
                                    let page: number;
                                    if (totalPages <= 5) {
                                        page = i + 1;
                                    } else if (currentPage <= 3) {
                                        page = i + 1;
                                    } else if (currentPage >= totalPages - 2) {
                                        page = totalPages - 4 + i;
                                    } else {
                                        page = currentPage - 2 + i;
                                    }
                                    return (
                                        <button
                                            key={page}
                                            className={`w-8 h-8 rounded-lg text-sm font-medium transition-colors ${
                                                currentPage === page
                                                    ? 'bg-accent-gold text-white'
                                                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                                            }`}
                                            onClick={() => setCurrentPage(page)}
                                        >
                                            {page}
                                        </button>
                                    );
                                })}
                            </div>
                            <button
                                className="px-3 py-1.5 text-sm font-medium rounded-lg bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                                disabled={currentPage === totalPages}
                            >
                                下一页 →
                            </button>
                        </div>
                    </div>
                )}
            </div>

            {/* 主内容区 - 独立滚动 */}
            <div className="flex-1 overflow-hidden grid grid-cols-12 gap-6">
                {/* 技能树/网格 - 左侧独立滚动 */}
                <div className="col-span-9 overflow-auto pr-2">
                    {viewMode === 'tree' ? (
                        <div className="card !p-6">
                            <div className="space-y-1">
                                {/* 按根节点分页，子节点从完整列表获取 */}
                                {paginatedRootSkills.map(skill => {
                                    // 从完整的 filteredSkills 中获取子节点，保持完整的层级
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
                                            onSelect={loadEvidence}
                                        />
                                    );
                                })}
                            </div>
                        </div>
                    ) : (
                        <>
                            <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
                                {paginatedSkills.map((skill) => {
                                    const status = getStatus(skill.status);
                                    const category = getCategory(skill.category);
                                    
                                    return (
                                        <div
                                            key={skill.key}
                                            className="card group hover:shadow-card-lg transition-all duration-300 cursor-pointer"
                                            onClick={() => loadEvidence(skill)}
                                        >
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
                        </>
                    )}
                </div>

                {/* 成长记录侧边栏 - 固定在右侧，自己可滚动 */}
                <div className="col-span-3 overflow-auto">
                    <div className="space-y-4">
                        <GrowthTimeline skills={skills} />

                        <div className="card !p-5">
                        <div className="flex items-center justify-between mb-3">
                            <h3 className="font-semibold text-gray-900">证据链</h3>
                            {selectedSkill && (
                                <span className="text-xs text-gray-400 truncate max-w-[120px]">{selectedSkill.name}</span>
                            )}
                        </div>

                        {!selectedSkill && (
                            <div className="text-sm text-gray-500">点选一个技能查看“为什么我认为你在提升它”。</div>
                        )}

                        {selectedSkill && evidenceLoading && (
                            <div className="text-sm text-gray-500">加载中...</div>
                        )}

                        {selectedSkill && evidenceError && (
                            <div className="text-sm text-red-500">{evidenceError}</div>
                        )}

                        {selectedSkill && !evidenceLoading && !evidenceError && (
                            <div className="space-y-2">
                                {evidences.length === 0 && (
                                    <div className="text-sm text-gray-500">暂无最近证据</div>
                                )}
                                {evidences.map((e) => {
                                    const t = new Date(e.timestamp);
                                    const timeLabel = t.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
                                    return (
                                        <div 
                                            key={e.evidence_id} 
                                            className="p-3 rounded-xl bg-gray-50 hover:bg-gray-100 cursor-pointer transition-colors border border-transparent hover:border-gray-200 group relative"
                                            onClick={() => loadDiffDetail(e.evidence_id)}
                                        >
                                            <div className="text-xs text-gray-400 mb-1 flex justify-between">
                                                <span>{timeLabel}</span>
                                                <span className="opacity-0 group-hover:opacity-100 text-accent-gold">查看 &rarr;</span>
                                            </div>
                                            <div className="text-sm text-gray-900 leading-snug">{e.contribution_context}</div>
                                            {e.file_name && <div className="text-xs text-gray-500 mt-1 truncate">{e.file_name}</div>}
                                        </div>
                                    );
                                })}

                                <div className="pt-3 border-t border-gray-100">
                                    <div className="flex items-center justify-between mb-2">
                                        <h4 className="text-sm font-semibold text-gray-900">相关会话</h4>
                                        <span className="text-xs text-gray-400">{skillSessions.length}</span>
                                    </div>

                                    {sessionsLoading && <div className="text-sm text-gray-500">加载中...</div>}
                                    {sessionsError && <div className="text-sm text-red-500">{sessionsError}</div>}

                                    {!sessionsLoading && !sessionsError && skillSessions.length === 0 && (
                                        <div className="text-sm text-gray-400">暂无会话关联（可在“日报”页或“状态”页补全会话语义后再试）。</div>
                                    )}

                                    {!sessionsLoading && !sessionsError && skillSessions.length > 0 && (
                                        <div className="space-y-2">
                                            {skillSessions.map((s) => (
                                                <button
                                                    key={s.id}
                                                    className="w-full text-left p-3 rounded-xl bg-gray-50 hover:bg-gray-100 transition-colors border border-transparent hover:border-gray-200"
                                                    onClick={() => setActiveSessionId(s.id)}
                                                >
                                                    <div className="flex items-center justify-between gap-3">
                                                        <div className="min-w-0">
                                                            <div className="text-sm font-medium text-gray-900 truncate">{s.time_range || s.date}</div>
                                                            <div className="text-xs text-gray-400 truncate">{s.primary_app || ''}</div>
                                                        </div>
                                                        <div className="text-xs text-gray-500 line-clamp-2 max-w-[55%]">{s.summary || '（未生成摘要）'}</div>
                                                    </div>
                                                </button>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            </div>
                        )}
                        </div>
                    </div>
                </div>
            </div>

            {activeSessionId && (
                <SessionDetailModal sessionId={activeSessionId} onClose={() => setActiveSessionId(null)} />
            )}

            {/* Diff Detail Modal - 使用 Portal 确保在 body 根部渲染 */}
            {diffDetail && createPortal(
                <div 
                    className="fixed inset-0 flex items-center justify-center bg-black/60 p-4 animate-fade-in"
                    style={{ zIndex: 9999 }}
                    onClick={() => setDiffDetail(null)}
                >
                    <div 
                        className="bg-white rounded-3xl w-full max-w-4xl max-h-[85vh] overflow-hidden flex flex-col animate-slide-up border border-gray-200/50"
                        style={{ boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(255,255,255,0.1)' }}
                        onClick={e => e.stopPropagation()}
                    >
                        <div className="p-4 border-b border-gray-100 flex justify-between items-center bg-gray-50/50">
                            <div>
                                <h3 className="text-lg font-bold text-gray-900 flex items-center gap-3">
                                    <span className="text-xs font-medium text-amber-700 bg-amber-100 px-2.5 py-1 rounded-lg uppercase tracking-wide">{diffDetail.language}</span>
                                    <span className="truncate">{diffDetail.file_name}</span>
                                </h3>
                                <div className="text-sm text-gray-500 mt-2 flex items-center gap-2">
                                    <svg className="w-4 h-4 text-accent-gold flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                                    </svg>
                                    <span className="line-clamp-2">{diffDetail.insight || "AI 正在分析..."}</span>
                                </div>
                            </div>
                            <button onClick={() => setDiffDetail(null)} className="p-2.5 hover:bg-gray-100 rounded-xl transition-colors text-gray-400 hover:text-gray-700">
                                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        </div>
                        <div className="flex-1 overflow-auto bg-[#0d1117] rounded-b-none">
                            <pre className="text-[#c9d1d9] font-mono text-xs p-5 overflow-x-auto leading-relaxed">
                                <code>{diffDetail.diff_content}</code>
                            </pre>
                        </div>
                        <div className="px-5 py-4 bg-white border-t border-gray-100 flex justify-end rounded-b-3xl">
                             <button onClick={() => setDiffDetail(null)} className="px-5 py-2 bg-gray-900 text-white rounded-xl text-sm font-medium hover:bg-gray-800 transition-colors">
                                关闭
                             </button>
                        </div>
                    </div>
                </div>,
                document.body
            )}
        </div>
    );
};

export default SkillView;
