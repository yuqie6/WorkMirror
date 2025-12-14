import { useEffect, useMemo, useState } from 'react';
import './App.css';
import MainLayout, { TabId, SystemHealthIndicator } from '@/components/layout/MainLayout';
import DashboardView from '@/components/dashboard/DashboardView';
import SessionsView from '@/components/sessions/SessionsView';
import SkillView from '@/components/skills/SkillView';
import ReportsView from '@/components/reports/ReportsView';
import StatusView from '@/components/status/StatusView';
import SettingsView from '@/components/settings/SettingsView';
import { GetStatus } from '@/api/app';
import { StatusDTO, extractHealthIndicator } from '@/types/status';
import { useBrowserLocation } from '@/hooks/useBrowserLocation';
import { buildSessionsURL, buildSkillsURL, parseAppRoute } from '@/lib/routes';
import { todayLocalISODate } from '@/lib/date';

function App() {
  const { location, navigate } = useBrowserLocation();
  const route = useMemo(() => parseAppRoute(location.pathname, location.search), [location.pathname, location.search]);

  const activeTab: TabId = useMemo(() => {
    switch (route.kind) {
      case 'dashboard':
        return 'dashboard';
      case 'sessions':
        return 'sessions';
      case 'skills':
        return 'skills';
      case 'reports':
        return 'reports';
      case 'status':
        return 'status';
      case 'settings':
        return 'settings';
      case 'root':
      case 'not_found':
      default:
        return 'dashboard';
    }
  }, [route.kind]);

  const [systemIndicator, setSystemIndicator] = useState<SystemHealthIndicator | null>(null);

  const [lastSessionsDate, setLastSessionsDate] = useState<string>(() => todayLocalISODate());
  const [lastSkillId, setLastSkillId] = useState<string | null>(null);

  // 加载系统健康状态
  const refreshSystemIndicator = async () => {
    try {
      const status: StatusDTO = await GetStatus();
      setSystemIndicator(extractHealthIndicator(status));
    } catch (e) {
      console.error('Failed to load status:', e);
    }
  };

  // 初始加载
  useEffect(() => {
    refreshSystemIndicator();
  }, []);

  // 根路径重定向
  useEffect(() => {
    if (route.kind === 'root') {
      navigate('/dashboard', { replace: true });
    }
    if (route.kind === 'not_found') {
      navigate('/dashboard', { replace: true });
    }
  }, [navigate, route.kind]);

  // 记住 sessions/skills 最近选中态（用于跨页面返回）
  useEffect(() => {
    if (route.kind === 'sessions') {
      if (typeof route.date === 'string' && route.date.trim() !== '') {
        setLastSessionsDate(route.date);
      }
    }
    if (route.kind === 'skills') {
      if (typeof route.skillId === 'string' && route.skillId.trim() !== '') {
        setLastSkillId(route.skillId);
      }
    }
  }, [route]);

  // 订阅 Agent 实时事件
  useEffect(() => {
    const es = new EventSource('/api/events');

    const refresh = () => {
      void refreshSystemIndicator();
    };

    es.addEventListener('data_changed', refresh);
    es.addEventListener('settings_updated', refresh);
    es.addEventListener('pipeline_status_changed', refresh);

    es.onerror = () => {
      // 浏览器会自动重连
    };

    return () => {
      es.close();
    };
  }, []);

  const handleTabChange = (tab: TabId) => {
    switch (tab) {
      case 'dashboard':
        navigate('/dashboard');
        return;
      case 'sessions':
        navigate(buildSessionsURL({ date: lastSessionsDate }));
        return;
      case 'skills':
        navigate(buildSkillsURL(lastSkillId || undefined));
        return;
      case 'reports':
        navigate('/reports');
        return;
      case 'status':
        navigate('/status');
        return;
      case 'settings':
        navigate('/settings');
        return;
      default:
        navigate('/dashboard');
    }
  };

  const handleNavigateToSession = (sessionId: number, date: string) => {
    const d = typeof date === 'string' && date.trim() !== '' ? date : lastSessionsDate;
    setLastSessionsDate(d);
    navigate(buildSessionsURL({ sessionId, date: d }));
  };

  const handleNavigateToSkill = (skillId: string) => {
    const id = typeof skillId === 'string' ? skillId.trim() : '';
    if (!id) {
      navigate('/skills');
      return;
    }
    setLastSkillId(id);
    navigate(buildSkillsURL(id));
  };

  const handleSessionsDateChange = (date: string, sessionId?: number | null) => {
    const d = typeof date === 'string' && date.trim() !== '' ? date : todayLocalISODate();
    setLastSessionsDate(d);
    navigate(buildSessionsURL({ date: d, sessionId: sessionId ?? undefined }));
  };

  // 视图渲染
  const renderContent = () => {
    switch (activeTab) {
      case 'dashboard':
        return <DashboardView onNavigate={(tab) => handleTabChange(tab as TabId)} />;
      case 'sessions':
        return (
          <SessionsView
            initialDate={route.kind === 'sessions' ? route.date : undefined}
            selectedSessionId={route.kind === 'sessions' ? route.sessionId ?? null : null}
            onOpenSession={(id, date) => handleNavigateToSession(id, date)}
            onCloseSession={(date) => handleSessionsDateChange(date, null)}
            onDateChange={(date, id) => handleSessionsDateChange(date, id)}
          />
        );
      case 'skills':
        return (
          <SkillView
            selectedSkillId={route.kind === 'skills' ? route.skillId ?? null : null}
            onSelectSkill={handleNavigateToSkill}
            onNavigateToSession={handleNavigateToSession}
          />
        );
      case 'reports':
        return <ReportsView onNavigateToSession={handleNavigateToSession} />;
      case 'status':
        return <StatusView />;
      case 'settings':
        return <SettingsView />;
      default:
        return null;
    }
  };

  return (
    <MainLayout
      activeTab={activeTab}
      onTabChange={handleTabChange}
      systemIndicator={systemIndicator}
    >
      {renderContent()}
    </MainLayout>
  );
}

export default App;
