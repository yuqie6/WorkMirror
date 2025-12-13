import React from 'react';
import TopNav from './TopNav';

interface MainLayoutProps {
    children: React.ReactNode;
    activeTab: 'summary' | 'skills' | 'trends' | 'settings';
    onTabChange: (tab: 'summary' | 'skills' | 'trends' | 'settings') => void;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children, activeTab, onTabChange }) => {
    return (
        <div className="min-h-screen bg-gradient-warm relative">
            {/* 装饰性渐变光斑 */}
            <div className="fixed top-0 right-0 w-[60vw] h-[60vh] bg-gradient-to-bl from-amber-200/30 via-orange-100/20 to-transparent rounded-full blur-3xl pointer-events-none" />
            <div className="fixed bottom-0 left-0 w-[40vw] h-[40vh] bg-gradient-to-tr from-amber-100/30 to-transparent rounded-full blur-3xl pointer-events-none" />
            
            {/* 顶部导航 */}
            <TopNav activeTab={activeTab} onTabChange={onTabChange} />
            
            {/* 主内容区 */}
            <main className="relative z-10 px-6 pb-12">
                <div className="max-w-7xl mx-auto animate-fade-in">
                    {children}
                </div>
            </main>
        </div>
    );
};

export default MainLayout;
