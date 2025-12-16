import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { FileText, ArrowRight, AlertTriangle, ExternalLink } from 'lucide-react';
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
} from 'recharts';
import { GetTrends, GetAppStats, GetTodaySummary, GetStatus } from '@/api/app';
import { EvidenceStatusDTO } from '@/types/status';
import { todayLocalISODate } from '@/lib/date';
import { useTranslation } from '@/lib/i18n';

// 后端实际返回的 TrendReportDTO 结构 - 匹配 internal/dto/httpapi.go:61
interface TrendReportDTO {
  period: string;
  start_date: string;
  end_date: string;
  total_diffs: number;
  total_coding_mins: number;
  avg_diffs_per_day: number;
  top_languages: Array<{ language: string; diff_count: number; percentage: number }>;
  top_skills: Array<{ skill_name: string; status: string; days_active: number }>;
  bottlenecks: string[];
  daily_stats?: Array<{
    date: string;
    total_diffs: number;
    total_coding_mins: number;
    session_count: number;
  }>;
}

// 后端 AppStatsDTO - 匹配 internal/dto/httpapi.go:89
interface AppStat {
  app_name: string;
  total_duration: number;
  event_count: number;
  is_code_editor: boolean;
}

interface DashboardViewProps {
  onNavigate?: (tab: string) => void;
}

export default function DashboardView({ onNavigate }: DashboardViewProps) {
  const { t } = useTranslation();
  const [trends, setTrends] = useState<TrendReportDTO | null>(null);
  const [appStats, setAppStats] = useState<AppStat[]>([]);
  const [evidence, setEvidence] = useState<EvidenceStatusDTO | null>(null);
  const [hasDailyReport, setHasDailyReport] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      setError(null);
      try {
        const [trendsData, appData, summaryData, statusData] = await Promise.all([
          GetTrends(30),
          GetAppStats(),
          GetTodaySummary().catch(() => null),
          GetStatus(),
        ]);
        setTrends(trendsData);
        setAppStats(appData || []);
        setHasDailyReport(!!summaryData?.summary);
        setEvidence(statusData?.evidence || null);
      } catch (e: any) {
        console.error('Failed to load dashboard data:', e);
        setError(e?.message || t('common.error'));
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, []);

  // 计算专注分布
  const focusData = (() => {
    if (!appStats || appStats.length === 0) return [];
    const colors = ['#6366f1', '#0ea5e9', '#f59e0b', '#10b981', '#f43f5e'];
    const totalDuration = appStats.reduce((sum, app) => sum + app.total_duration, 0);
    return appStats.slice(0, 5).map((app, idx) => ({
      name: app.app_name,
      value: totalDuration > 0 ? Math.round((app.total_duration / totalDuration) * 100) : 0,
      color: colors[idx % colors.length],
    }));
  })();

  // 今日会话数：优先使用 trends.daily_stats；fallback 到 evidence.sessions_24h（口径为最近24h）
  const todayDate = todayLocalISODate();
  const todaySessions = trends?.daily_stats?.find((d) => d.date === todayDate)?.session_count ?? (evidence?.sessions_24h || 0);

  // 证据覆盖率：使用后端 evidence 数据
  const evidenceCoverage = (() => {
    if (!evidence || evidence.sessions_24h === 0) return 0;
    const covered = evidence.with_diff + evidence.with_browser - evidence.with_diff_and_browser;
    return Math.min(100, Math.round((covered / evidence.sessions_24h) * 100));
  })();

  const weakEvidenceCount = evidence?.weak_evidence || 0;

  // 生成热力图数据 - 使用后端 daily_stats（自然日）
  const heatmapData = (() => {
    const stats = trends?.daily_stats || [];
    if (stats.length === 0) return [];
    const maxSessions = Math.max(...stats.map((d) => d.session_count || 0), 1);
    return stats.map((d) => ({
      intensity: (d.session_count || 0) / maxSessions,
      date: d.date,
      sessions: d.session_count || 0,
      diffs: d.total_diffs || 0,
      codingMins: d.total_coding_mins || 0,
    }));
  })();

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64 text-zinc-500">
        {t('dashboard.loadingData')}
      </div>
    );
  }

  // 错误状态引导到 Status 页
  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-zinc-500 gap-4">
        <AlertTriangle size={32} className="text-amber-500" />
        <div>{error}</div>
        <button
          onClick={() => onNavigate?.('status')}
          className="flex items-center gap-2 px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded-lg text-sm text-zinc-300 transition-colors"
        >
          <ExternalLink size={14} /> {t('dashboard.goToStatus')}
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      {/* Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Session Counter */}
        <Card className="bg-zinc-900 border-zinc-800">
          <CardContent className="p-6">
            <h3 className="text-zinc-400 text-sm font-medium mb-1">{t('dashboard.todaySessions')}</h3>
            <div className="flex items-baseline gap-2">
              <span className="text-3xl font-bold text-white">{todaySessions}</span>
            </div>
            <div className="mt-4 text-xs text-zinc-500">
              {t('dashboard.thirtyDayDiffs')}: {trends?.total_diffs || 0} {t('dashboard.times')}
            </div>
            <div className="text-xs text-zinc-600">
              {t('dashboard.avgPerDay')}: {trends?.avg_diffs_per_day?.toFixed(1) || 0} {t('dashboard.perDay')}
            </div>
          </CardContent>
        </Card>

        {/* Evidence Coverage */}
        <Card className="bg-zinc-900 border-zinc-800">
          <CardContent className="p-6">
            <h3 className="text-zinc-400 text-sm font-medium mb-1">{t('dashboard.evidenceCoverage')}</h3>
            <div className="flex items-baseline gap-2">
              <span className={`text-3xl font-bold ${
                evidenceCoverage >= 70 ? 'text-emerald-400' : 
                evidenceCoverage >= 40 ? 'text-amber-400' : 'text-rose-400'
              }`}>
                {evidenceCoverage}%
              </span>
              <span className="text-sm text-zinc-500">
                {evidenceCoverage >= 70 ? t('dashboard.highCredibility') : evidenceCoverage >= 40 ? t('dashboard.medium') : t('dashboard.needsImprovement')}
              </span>
            </div>
            {weakEvidenceCount > 0 && (
              <div
                title={t('dashboard.weakEvidenceTooltip')}
                className="mt-4 text-xs text-amber-500 flex items-center gap-1"
              >
                <AlertTriangle size={12} /> {weakEvidenceCount} {t('dashboard.weakEvidenceSessions')}
              </div>
            )}
            {evidence && (
              <div className="mt-2 text-xs text-zinc-600">
                24h: {evidence.sessions_24h} {t('common.sessions')} | 
                {t('dashboard.withDiff')}: {evidence.with_diff} | 
                {t('dashboard.withBrowser')}: {evidence.with_browser}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Focus Distribution */}
        <Card className="bg-zinc-900 border-zinc-800">
          <CardContent className="p-6">
            <h3 className="text-zinc-400 text-sm font-medium mb-2">{t('dashboard.focusDistribution')}</h3>
            {focusData.length > 0 ? (
              <div className="flex items-center gap-4">
                <div className="w-20 h-20">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={focusData}
                        cx="50%"
                        cy="50%"
                        innerRadius={25}
                        outerRadius={35}
                        paddingAngle={2}
                        dataKey="value"
                      >
                        {focusData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <div className="space-y-1 text-xs flex-1">
                  {focusData.map((item) => (
                    <div key={item.name} className="flex items-center gap-2">
                      <span
                        className="w-2 h-2 rounded-full flex-shrink-0"
                        style={{ backgroundColor: item.color }}
                      />
                      <span className="truncate">{item.name}</span>
                      <span className="text-zinc-500 ml-auto">{item.value}%</span>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              <div className="text-zinc-500 text-sm">{t('dashboard.noAppData')}</div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Daily Report Quick Action */}
      {hasDailyReport && (
        <div
          onClick={() => onNavigate?.('reports')}
          className="bg-gradient-to-r from-indigo-900/20 to-purple-900/20 border border-indigo-500/20 p-4 rounded-xl flex items-center justify-between cursor-pointer hover:border-indigo-500/40 transition-colors"
        >
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-full bg-indigo-500/20 flex items-center justify-center text-indigo-400">
              <FileText size={20} />
            </div>
            <div>
              <h3 className="text-indigo-100 font-medium">{t('dashboard.dailySummaryReady')}</h3>
              <p className="text-sm text-zinc-400">
                {t('dashboard.viewTodaySummary')}
              </p>
            </div>
          </div>
          <ArrowRight size={20} className="text-indigo-400" />
        </div>
      )}

      {/* Top Skills from Trends */}
      {trends?.top_skills && trends.top_skills.length > 0 && (
        <Card className="bg-zinc-900 border-zinc-800">
          <CardHeader className="pb-2">
            <CardTitle className="text-zinc-200 text-base font-medium">{t('dashboard.hotSkills30Days')}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              {trends.top_skills.slice(0, 8).map((skill) => (
                <span
                  key={skill.skill_name}
                  onClick={() => onNavigate?.('skills')}
                  className="px-3 py-1.5 bg-zinc-950 border border-zinc-800 rounded-lg text-sm text-zinc-300 hover:border-indigo-500/50 cursor-pointer transition-colors"
                >
                  {skill.skill_name}
                  <span className="ml-2 text-xs text-zinc-600">{skill.days_active}{t('common.days')}</span>
                </span>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Activity Heatmap */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-zinc-200 text-base font-medium">
              {t('dashboard.activityHeatmap')}
            </CardTitle>
            <span className="text-xs text-zinc-600">{t('dashboard.last30Days')}</span>
          </div>
        </CardHeader>
        <CardContent>
          {heatmapData.length > 0 ? (
            <div className="flex flex-wrap gap-1">
              {heatmapData.map((item, i) => {
                const intensity = item.intensity;
                const dateStr = item.date.slice(5);

                let color = 'bg-zinc-800';
                if (intensity > 0.8) color = 'bg-emerald-500';
                else if (intensity > 0.6) color = 'bg-emerald-600/80';
                else if (intensity > 0.3) color = 'bg-emerald-700/60';
                else if (intensity > 0) color = 'bg-emerald-900/40';

                return (
                  <div
                    key={i}
                    className={`w-4 h-4 rounded-sm ${color} cursor-pointer hover:ring-1 hover:ring-zinc-600`}
                    title={`${dateStr}: ${item.sessions} ${t('common.sessions')} | ${item.diffs} ${t('dashboard.codeChangesShort')} | ${item.codingMins}m`}
                  />
                );
              })}
            </div>
          ) : (
            <div className="text-zinc-500 text-sm">{t('dashboard.noHeatmapData')}</div>
          )}
          <div className="flex justify-between mt-2 text-[10px] text-zinc-600">
            <span>30{t('dashboard.daysAgo')}</span>
            <span>{t('common.today')}</span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
