import { ReactNode } from 'react';
import {
  LayoutDashboard,
  History,
  Zap,
  FileText,
  Activity,
  Settings,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { StatusDot, StatusType } from '@/components/common/StatusDot';

// 导航项定义
const navItems = [
  { id: 'dashboard', label: '仪表盘', icon: LayoutDashboard },
  { id: 'sessions', label: '会话流', icon: History },
  { id: 'skills', label: '技能树', icon: Zap },
  { id: 'reports', label: '报告', icon: FileText },
  { id: 'status', label: '系统诊断', icon: Activity },
] as const;

export type TabId = (typeof navItems)[number]['id'] | 'settings';

export interface SystemHealthIndicator {
  window: StatusType;
  diff: StatusType;
  ai: StatusType;
  lastHeartbeat: string;
}

interface MainLayoutProps {
  children: ReactNode;
  activeTab: TabId;
  onTabChange: (tab: TabId) => void;
  systemIndicator?: SystemHealthIndicator | null;
}

const pageTitles: Record<TabId, string> = {
  dashboard: '仪表盘',
  sessions: '会话流',
  skills: '技能树',
  reports: '报告',
  status: '系统诊断',
  settings: '设置',
};

export default function MainLayout({
  children,
  activeTab,
  onTabChange,
  systemIndicator,
}: MainLayoutProps) {
  return (
    <div className="flex h-screen bg-zinc-950 text-zinc-200 font-sans overflow-hidden selection:bg-indigo-500/30">
      {/* Sidebar */}
      <aside className="w-64 border-r border-zinc-800 bg-zinc-950 flex flex-col flex-shrink-0 z-20">
        {/* Logo */}
        <div className="p-4 border-b border-zinc-800">
          <h1 className="text-lg font-bold tracking-tight text-white flex items-center gap-2">
            <span className="w-2 h-6 bg-indigo-500 rounded-sm"></span>
            Project Mirror
          </h1>
          <p className="text-xs text-zinc-500 mt-1 font-mono">v0.2-alpha</p>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-2 space-y-1 overflow-y-auto">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => onTabChange(item.id)}
                className={cn(
                  'w-full text-left px-3 py-2 rounded-md text-sm font-medium flex items-center gap-3 transition-colors',
                  isActive
                    ? 'bg-zinc-800 text-white'
                    : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900'
                )}
              >
                <Icon size={18} className="opacity-70" />
                {item.label}
              </button>
            );
          })}
        </nav>

        {/* P0 System Health (设计规范 3.1) */}
        <div
          onClick={() => onTabChange('status')}
          className="p-4 border-t border-zinc-800 bg-zinc-900/50 cursor-pointer hover:bg-zinc-900 transition-colors"
        >
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-semibold uppercase tracking-wider text-zinc-500">
              系统健康
            </span>
            <span className="text-[10px] text-zinc-600 font-mono">
              心跳: {systemIndicator?.lastHeartbeat ?? '--'}
            </span>
          </div>
          <div className="flex gap-4">
            <StatusDot
              status={systemIndicator?.window ?? 'offline'}
              label="WIN"
            />
            <StatusDot
              status={systemIndicator?.diff ?? 'offline'}
              label="DIFF"
            />
            <StatusDot
              status={systemIndicator?.ai ?? 'offline'}
              label="AI"
            />
          </div>
        </div>

        {/* Settings */}
        <div className="p-4 border-t border-zinc-800">
          <button
            onClick={() => onTabChange('settings')}
            className={cn(
              'flex items-center gap-2 text-sm px-2 transition-colors w-full',
              activeTab === 'settings'
                ? 'text-white'
                : 'text-zinc-500 hover:text-zinc-300'
            )}
          >
            <Settings size={16} /> 设置
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col relative min-w-0">
        {/* Header */}
        <header className="h-14 border-b border-zinc-800 bg-zinc-950/80 backdrop-blur flex items-center justify-between px-6 z-10 sticky top-0">
          <h2 className="text-sm font-medium text-zinc-100">
            {pageTitles[activeTab]}
          </h2>
          <div className="flex items-center gap-4">
            <div className="w-6 h-6 rounded bg-gradient-to-tr from-indigo-500 to-purple-500"></div>
          </div>
        </header>

        {/* View Container */}
        <div className="flex-1 overflow-y-auto p-6 scroll-smooth">
          {children}
        </div>
      </main>
    </div>
  );
}
