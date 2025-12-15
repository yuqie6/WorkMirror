import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { StatusDot, StatusType } from '@/components/common/StatusDot';
import { AlertTriangle, RefreshCw, Download, Play } from 'lucide-react';
import { GetStatus, RebuildSessionsForDate, EnrichSessionsForDate, BuildSessions, RepairEvidenceForDate } from '@/api/app';
import { StatusDTO, extractHealthIndicator } from '@/types/status';
import { todayLocalISODate } from '@/lib/date';
import { useTranslation } from '@/lib/i18n';

interface CollectorRowProps {
  name: string;
  status: StatusType;
  detail: string;
  heartbeat: string;
  heartbeatLabel: string;
}

function CollectorRow({ name, status, detail, heartbeat, heartbeatLabel }: CollectorRowProps) {
  return (
    <div className="bg-zinc-950/50 border border-zinc-800 p-4 rounded-lg flex items-center justify-between">
      <div className="flex items-center gap-4">
        <StatusDot status={status} showGlow />
        <div>
          <div className="font-medium text-zinc-200">{name}</div>
          <div className="text-xs text-zinc-500 font-mono">{detail}</div>
        </div>
      </div>
      <div className="text-xs font-mono text-zinc-600 bg-zinc-900 px-2 py-1 rounded">
        {heartbeatLabel}: {heartbeat}
      </div>
    </div>
  );
}

export default function StatusView() {
  const { t } = useTranslation();
  const [status, setStatus] = useState<StatusDTO | null>(null);
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [selectedDate, setSelectedDate] = useState<string>(() => todayLocalISODate());
  const [attachGapMinutes, setAttachGapMinutes] = useState<number>(10);

  useEffect(() => {
    const loadStatus = async () => {
      setLoading(true);
      try {
        const data: StatusDTO = await GetStatus();
        setStatus(data);
      } catch (e) {
        console.error('Failed to load status:', e);
      } finally {
        setLoading(false);
      }
    };
    loadStatus();
  }, []);

  const handleBuild = async () => {
    setActionLoading(true);
    try {
      await BuildSessions(selectedDate);
      alert(`${t('status.buildSuccess')}（${selectedDate}）！`);
      setStatus(await GetStatus());
    } catch (e) {
      alert(`${t('status.buildFailed')}: ${e}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleRebuild = async () => {
    const date = selectedDate;
    const ok = window.confirm(
      `${t('status.rebuildWarning')}（${date}）。\n\n${t('status.rebuildImpact')}\n${t('status.rebuildAdvice')}\n\n${t('status.confirmContinue')}`
    );
    if (!ok) return;
    setActionLoading(true);
    try {
      await RebuildSessionsForDate(date);
      alert(`${t('status.rebuildSuccess')}（${date}）！`);
      setStatus(await GetStatus());
    } catch (e) {
      alert(`${t('status.rebuildFailed')}: ${e}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleEnrich = async () => {
    const date = selectedDate;
    const ok = window.confirm(
      `${t('status.enrichWarning')}\n\n${t('status.enrichImpact')}\n\n${t('status.confirmContinue')}`
    );
    if (!ok) return;
    setActionLoading(true);
    try {
      await EnrichSessionsForDate(date);
      alert(`${t('status.enrichSuccess')}（${date}）！`);
      setStatus(await GetStatus());
    } catch (e) {
      alert(`${t('status.enrichFailed')}: ${e}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleRepairEvidence = async () => {
    const date = selectedDate;
    const ok = window.confirm(
      `${t('status.repairEvidenceWarning')}（${date}）。\n\n${t('status.repairEvidenceImpact')}\n\n${t('status.confirmContinue')}`
    );
    if (!ok) return;
    setActionLoading(true);
    try {
      const res = await RepairEvidenceForDate(date, attachGapMinutes);
      const updatedSessions = (res && typeof res.updated_sessions === 'number') ? res.updated_sessions : 0;
      const attachedDiffs = (res && typeof res.attached_diffs === 'number') ? res.attached_diffs : 0;
      const attachedBrowser = (res && typeof res.attached_browser === 'number') ? res.attached_browser : 0;
      alert(`${t('status.repairEvidenceSuccess')}（${date}）：sessions=${updatedSessions}, diffs=${attachedDiffs}, browser=${attachedBrowser}`);
      setStatus(await GetStatus());
    } catch (e) {
      alert(`${t('status.repairEvidenceFailed')}: ${e}`);
    } finally {
      setActionLoading(false);
    }
  };

  if (loading || !status) {
    return (
      <div className="flex items-center justify-center h-64 text-zinc-500">
        {t('status.loading')}
      </div>
    );
  }

  const health = extractHealthIndicator(status);
  const isOffline = status.pipeline.ai.mode === 'offline' || status.pipeline.ai.degraded;

  // 格式化心跳时间
  const formatHeartbeat = (ts: number): string => {
    if (!ts) return t('status.noHeartbeat');
    const secAgo = Math.floor((Date.now() - ts) / 1000);
    if (secAgo < 60) return `${secAgo}${t('common.secondsAgo')}`;
    return `${Math.floor(secAgo / 60)}${t('common.minutesAgo')}`;
  };

  return (
    <div className="space-y-6 max-w-4xl mx-auto animate-in fade-in duration-500">
      {/* Offline Warning */}
      {isOffline && (
        <Alert variant="destructive" className="bg-amber-500/10 border-amber-500/20 text-amber-500">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>{t('status.offlineModeActive')}</AlertTitle>
          <AlertDescription className="text-amber-500/80">
            {status.pipeline.ai.degraded_reason || t('status.aiUnavailable')}
          </AlertDescription>
        </Alert>
      )}

      {/* Collector Health */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-zinc-200 font-medium flex items-center gap-2">
            {t('status.collectorHealth')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <CollectorRow
            name={t('status.windowMonitor')}
            status={health.window}
            detail={`${t('status.events24h')}: ${status.collectors.window.count_24h}`}
            heartbeat={formatHeartbeat(status.collectors.window.last_collected_at)}
            heartbeatLabel={t('systemHealth.heartbeat')}
          />
          <CollectorRow
            name={t('status.gitDiffCollector')}
            status={health.diff}
            detail={`${t('status.watchPaths')}: ${status.collectors.diff.effective_paths || 0}, 24h: ${status.collectors.diff.count_24h}`}
            heartbeat={formatHeartbeat(status.collectors.diff.last_collected_at)}
            heartbeatLabel={t('systemHealth.heartbeat')}
          />
          <CollectorRow
            name={t('status.browserMonitor')}
            status={status.collectors.browser.enabled ? (status.collectors.browser.running ? 'healthy' : 'warning') : 'offline'}
            detail={status.collectors.browser.enabled ? `${t('status.events24h')}: ${status.collectors.browser.count_24h}` : t('status.disabled')}
            heartbeat={formatHeartbeat(status.collectors.browser.last_collected_at)}
            heartbeatLabel={t('systemHealth.heartbeat')}
          />
          <CollectorRow
            name={t('status.aiPipeline')}
            status={health.ai}
            detail={status.pipeline.ai.configured ? `${t('status.mode')}: ${status.pipeline.ai.mode === 'ai' ? t('status.modeAI') : t('status.modeOffline')}` : t('status.notConfigured')}
            heartbeat={formatHeartbeat(status.pipeline.ai.last_call_at)}
            heartbeatLabel={t('systemHealth.heartbeat')}
          />
        </CardContent>
      </Card>

      {/* Evidence Coverage */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-zinc-200 font-medium">
            {t('status.evidenceCoverage')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <div className="flex items-center justify-between bg-zinc-950/50 border border-zinc-800 px-3 py-2 rounded">
            <div className="text-zinc-400">{t('status.sessions24h')}</div>
            <div className="font-mono text-zinc-200">{status.evidence.sessions_24h}</div>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <div className="bg-zinc-950/50 border border-zinc-800 px-3 py-2 rounded flex items-center justify-between">
              <div className="text-zinc-400">{t('status.withDiff')}</div>
              <div className="font-mono text-zinc-200">{status.evidence.with_diff}</div>
            </div>
            <div className="bg-zinc-950/50 border border-zinc-800 px-3 py-2 rounded flex items-center justify-between">
              <div className="text-zinc-400">{t('status.withBrowser')}</div>
              <div className="font-mono text-zinc-200">{status.evidence.with_browser}</div>
            </div>
            <div className="bg-zinc-950/50 border border-zinc-800 px-3 py-2 rounded flex items-center justify-between">
              <div className="text-zinc-400">{t('status.withDiffBrowser')}</div>
              <div className="font-mono text-zinc-200">{status.evidence.with_diff_and_browser}</div>
            </div>
            <div className="bg-zinc-950/50 border border-zinc-800 px-3 py-2 rounded flex items-center justify-between">
              <div className="text-zinc-400">{t('status.weakEvidence')}</div>
              <div className="font-mono text-zinc-200">{status.evidence.weak_evidence}</div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-2">
            <div className="bg-zinc-950/50 border border-zinc-800 px-3 py-2 rounded flex items-center justify-between">
              <div className="text-zinc-400">{t('status.orphanDiffs24h')}</div>
              <div className="font-mono text-zinc-200">{status.evidence.orphan_diffs_24h || 0}</div>
            </div>
            <div className="bg-zinc-950/50 border border-zinc-800 px-3 py-2 rounded flex items-center justify-between">
              <div className="text-zinc-400">{t('status.orphanBrowser24h')}</div>
              <div className="font-mono text-zinc-200">{status.evidence.orphan_browser_24h || 0}</div>
            </div>
          </div>

          {(status.evidence.orphan_diffs_24h || status.evidence.orphan_browser_24h) ? (
            <Alert className="bg-amber-500/10 border-amber-500/20 text-amber-500">
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>{t('status.orphanEvidenceTitle')}</AlertTitle>
              <AlertDescription className="text-amber-500/80">
                {t('status.orphanEvidenceHint')}
              </AlertDescription>
            </Alert>
          ) : null}
        </CardContent>
      </Card>

      {/* Maintenance */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-zinc-200 font-medium">
            {t('status.diagnosisRepair')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center justify-between bg-zinc-950/50 border border-zinc-800 rounded-lg p-3">
            <div>
              <div className="text-sm text-zinc-300 font-medium">{t('status.operationDate')}</div>
              <div className="text-xs text-zinc-500">{t('status.operationDateHint')}</div>
            </div>
            <input
              type="date"
              value={selectedDate}
              onChange={(e) => setSelectedDate(e.target.value)}
              className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 font-mono"
              max={todayLocalISODate()}
            />
          </div>
          <div className="flex items-center justify-between bg-zinc-950/50 border border-zinc-800 rounded-lg p-3">
            <div>
              <div className="text-sm text-zinc-300 font-medium">{t('status.attachGapMinutes')}</div>
              <div className="text-xs text-zinc-500">{t('status.attachGapMinutesHint')}</div>
            </div>
            <input
              type="number"
              min={1}
              max={180}
              value={Number.isFinite(attachGapMinutes) ? attachGapMinutes : 10}
              onChange={(e) => {
                const n = Math.floor(Number(e.target.value));
                setAttachGapMinutes(Number.isFinite(n) ? Math.max(1, Math.min(180, n)) : 10);
              }}
              className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 font-mono w-24 text-right"
            />
          </div>
          <button
            onClick={handleBuild}
            disabled={actionLoading}
            className="w-full flex items-center justify-between p-3 bg-emerald-500/10 hover:bg-emerald-500/20 border border-emerald-500/20 rounded-lg transition-colors text-left group disabled:opacity-50"
          >
            <div>
              <div className="text-sm text-emerald-300 font-medium">{t('status.buildSessions')}</div>
              <div className="text-xs text-zinc-500">{t('status.buildSessionsHint')}</div>
            </div>
            <Play size={16} className="text-emerald-600 group-hover:text-emerald-400" />
          </button>

          <button
            onClick={handleRebuild}
            disabled={actionLoading}
            className="w-full flex items-center justify-between p-3 bg-zinc-800/30 hover:bg-zinc-800 border border-zinc-800 rounded-lg transition-colors text-left group disabled:opacity-50"
          >
            <div>
              <div className="text-sm text-zinc-300 font-medium">{t('status.rebuildSessions')}</div>
              <div className="text-xs text-zinc-500">{t('status.rebuildSessionsHint')}</div>
            </div>
            <RefreshCw size={16} className="text-zinc-600 group-hover:text-zinc-300" />
          </button>

          <button
            onClick={handleEnrich}
            disabled={actionLoading}
            className="w-full flex items-center justify-between p-3 bg-zinc-800/30 hover:bg-zinc-800 border border-zinc-800 rounded-lg transition-colors text-left group disabled:opacity-50"
          >
            <div>
              <div className="text-sm text-zinc-300 font-medium">{t('status.enrichMissing')}</div>
              <div className="text-xs text-zinc-500">{t('status.enrichMissingHint')}</div>
            </div>
            <RefreshCw size={16} className="text-zinc-600 group-hover:text-indigo-400" />
          </button>

          <button
            onClick={handleRepairEvidence}
            disabled={actionLoading}
            className="w-full flex items-center justify-between p-3 bg-amber-500/10 hover:bg-amber-500/20 border border-amber-500/20 rounded-lg transition-colors text-left group disabled:opacity-50"
          >
            <div>
              <div className="text-sm text-amber-300 font-medium">{t('status.repairEvidence')}</div>
              <div className="text-xs text-zinc-500">{t('status.repairEvidenceHint')}</div>
            </div>
            <RefreshCw size={16} className="text-amber-600 group-hover:text-amber-400" />
          </button>

          <a
            href="/api/diagnostics/export"
            download
            className="w-full flex items-center justify-between p-3 bg-zinc-800/30 hover:bg-zinc-800 border border-zinc-800 rounded-lg transition-colors text-left group"
          >
            <div>
              <div className="text-sm text-zinc-300 font-medium">{t('status.exportDiagnostics')}</div>
              <div className="text-xs text-zinc-500">{t('status.exportDiagnosticsHint')}</div>
            </div>
            <Download size={16} className="text-zinc-600 group-hover:text-zinc-300" />
          </a>
        </CardContent>
      </Card>

      {/* Recent Errors */}
      {status.recent_errors.length > 0 && (
        <Card className="bg-zinc-900 border-zinc-800">
          <CardHeader>
            <CardTitle className="text-zinc-200 font-medium text-rose-400">
              {t('status.recentErrors')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {status.recent_errors.slice(0, 5).map((err, idx: number) => (
              <div key={idx} className="p-3 bg-rose-500/5 border border-rose-500/20 rounded text-sm">
                <div className="text-rose-400 font-mono text-xs">{err.time}</div>
                <div className="text-zinc-300">{err.message}</div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
