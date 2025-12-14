import React from 'react';
import Tooltip from '../common/Tooltip';

interface TopNavProps {
    activeTab: 'summary' | 'sessions' | 'skills' | 'trends' | 'status' | 'settings';
    onTabChange: (tab: 'summary' | 'sessions' | 'skills' | 'trends' | 'status' | 'settings') => void;
    systemIndicator?: { text: string; tone: 'info' | 'warn' | 'danger' } | null;
}

const TopNav: React.FC<TopNavProps> = ({ activeTab, onTabChange, systemIndicator }) => {
    const menuItems = [
        { id: 'summary', label: '今日总结' },
        { id: 'sessions', label: '会话' },
        { id: 'skills', label: '技能树' },
        { id: 'trends', label: '趋势分析' },
        { id: 'status', label: '状态' },
        { id: 'settings', label: '设置' },
    ];

    return (
        <header className="sticky top-0 z-50 px-6 py-4 bg-[#FDF8F3]/95 backdrop-blur-md border-b border-amber-100/50">
            <div className="max-w-7xl mx-auto flex items-center justify-between">
                {/* Logo */}
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-2xl bg-gradient-gold flex items-center justify-center text-white font-bold text-lg shadow-lg">
                        M
                    </div>
                    <span className="text-xl font-semibold text-gray-900">Mirror</span>
                </div>

                {/* 导航菜单 */}
                <nav className="flex items-center gap-1 bg-surface-warm/80 backdrop-blur-sm rounded-full px-2 py-1.5 shadow-sm">
                    {menuItems.map((item) => (
                        <button
                            key={item.id}
                            className={`nav-item ${activeTab === item.id ? 'active' : ''}`}
                            onClick={() => onTabChange(item.id as any)}
                        >
                            {item.label}
                        </button>
                    ))}
                </nav>

                {/* 右侧操作区 */}
                <div className="flex items-center gap-4">
                    {systemIndicator?.text && (
                        <span
                            className={`px-3 py-1.5 rounded-full border text-xs font-medium ${
                                systemIndicator.tone === 'danger'
                                    ? 'bg-red-50 text-red-700 border-red-100'
                                    : systemIndicator.tone === 'warn'
                                        ? 'bg-amber-50 text-amber-700 border-amber-100'
                                        : 'bg-emerald-50 text-emerald-700 border-emerald-100'
                            }`}
                            title="运行模式"
                        >
                            {systemIndicator.text}
                        </span>
                    )}
                    {/* 语言切换 */}
                    <button className="px-3 py-1.5 text-sm font-medium text-gray-600 bg-white/60 rounded-full hover:bg-white transition-colors">
                        中文
                    </button>
                    
                    {/* 主题/通知按钮 */}
                    <Tooltip content="切换主题">
                        <button className="w-9 h-9 rounded-full bg-white/60 flex items-center justify-center text-gray-600 hover:bg-white transition-colors">
                            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z" />
                            </svg>
                        </button>
                    </Tooltip>

                    {/* 通知 */}
                    <Tooltip content="通知">
                        <button className="w-9 h-9 rounded-full bg-white/60 flex items-center justify-center text-gray-600 hover:bg-white transition-colors relative">
                            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />
                            </svg>
                        </button>
                    </Tooltip>

                    {/* 用户头像 */}
                    <Tooltip content="个人中心">
                        <div className="w-10 h-10 rounded-full bg-gradient-gold flex items-center justify-center text-white shadow-lg cursor-pointer hover:shadow-xl transition-shadow">
                            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
                            </svg>
                        </div>
                    </Tooltip>
                </div>
            </div>
        </header>
    );
};

export default TopNav;
