import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { StatusDot, StatusType } from '@/components/common/StatusDot';
import { AlertTriangle, RefreshCw, Download, Play } from 'lucide-react';
import { GetStatus, RebuildSessionsForDate, EnrichSessionsForDate, BuildSessions } from '@/api/app';
import { StatusDTO, extractHealthIndicator } from '@/types/status';
import { todayLocalISODate } from '@/lib/date';

interface CollectorRowProps {
  name: string;
  status: StatusType;
  detail: string;
  heartbeat: string;
}

function CollectorRow({ name, status, detail, heartbeat }: CollectorRowProps) {
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
        心跳: {heartbeat}
      </div>
    </div>
  );
}

export default function StatusView() {
  const [status, setStatus] = useState<StatusDTO | null>(null);
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [selectedDate, setSelectedDate] = useState<string>(() => todayLocalISODate());

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
      alert(`会话构建成功（${selectedDate}）！`);
      setStatus(await GetStatus());
    } catch (e) {
      alert(`构建失败: ${e}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleRebuild = async () => {
    const date = selectedDate;
    const ok = window.confirm(
      `风险提示：将按日期重建会话（${date}）。\n\n影响：该日期的会话切分/聚合结果将被重算，可能改变技能归因与报告口径。\n建议：必要时先导出诊断包。\n\n确认继续？`
    );
    if (!ok) return;
    setActionLoading(true);
    try {
      await RebuildSessionsForDate(date);
      alert(`会话重建成功（${date}）！`);
      setStatus(await GetStatus());
    } catch (e) {
      alert(`重建失败: ${e}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleEnrich = async () => {
    const date = selectedDate;
    const ok = window.confirm(
      `风险提示：将对 ${date} 的会话进行语义增强（可能触发 AI 调用）。\n\n影响：会写入摘要/标签/技能等语义字段；无 AI 配置会降级为规则。\n\n确认继续？`
    );
    if (!ok) return;
    setActionLoading(true);
    try {
      await EnrichSessionsForDate(date);
      alert(`增强成功（${date}）！`);
      setStatus(await GetStatus());
    } catch (e) {
      alert(`增强失败: ${e}`);
    } finally {
      setActionLoading(false);
    }
  };

  if (loading || !status) {
    return (
      <div className="flex items-center justify-center h-64 text-zinc-500">
        加载诊断信息中...
      </div>
    );
  }

  const health = extractHealthIndicator(status);
  const isOffline = status.pipeline.ai.mode === 'offline' || status.pipeline.ai.degraded;

  // 格式化心跳时间
  const formatHeartbeat = (ts: number): string => {
    if (!ts) return '无';
    const secAgo = Math.floor((Date.now() - ts) / 1000);
    if (secAgo < 60) return `${secAgo}秒前`;
    return `${Math.floor(secAgo / 60)}分钟前`;
  };

  return (
    <div className="space-y-6 max-w-4xl mx-auto animate-in fade-in duration-500">
      {/* Offline Warning */}
      {isOffline && (
        <Alert variant="destructive" className="bg-amber-500/10 border-amber-500/20 text-amber-500">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>离线模式已激活</AlertTitle>
          <AlertDescription className="text-amber-500/80">
            {status.pipeline.ai.degraded_reason || 'AI 服务不可用，已降级为规则生成模式'}
          </AlertDescription>
        </Alert>
      )}

      {/* Collector Health */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-zinc-200 font-medium flex items-center gap-2">
            采集器健康状态
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <CollectorRow
            name="窗口监视器"
            status={health.window}
            detail={`24h 事件: ${status.collectors.window.count_24h}`}
            heartbeat={formatHeartbeat(status.collectors.window.last_collected_at)}
          />
          <CollectorRow
            name="Git Diff 采集器"
            status={health.diff}
            detail={`监控路径: ${status.collectors.diff.effective_paths || 0}, 24h: ${status.collectors.diff.count_24h}`}
            heartbeat={formatHeartbeat(status.collectors.diff.last_collected_at)}
          />
          <CollectorRow
            name="浏览器监视器"
            status={status.collectors.browser.enabled ? (status.collectors.browser.running ? 'healthy' : 'warning') : 'offline'}
            detail={status.collectors.browser.enabled ? `24h 事件: ${status.collectors.browser.count_24h}` : '已禁用'}
            heartbeat={formatHeartbeat(status.collectors.browser.last_collected_at)}
          />
          <CollectorRow
            name="AI 管线"
            status={health.ai}
            detail={status.pipeline.ai.configured ? `模式: ${status.pipeline.ai.mode === 'ai' ? 'AI' : '离线'}` : '未配置'}
            heartbeat={formatHeartbeat(status.pipeline.ai.last_call_at)}
          />
        </CardContent>
      </Card>

      {/* Maintenance */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-zinc-200 font-medium">
            诊断与修复
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center justify-between bg-zinc-950/50 border border-zinc-800 rounded-lg p-3">
            <div>
              <div className="text-sm text-zinc-300 font-medium">操作日期</div>
              <div className="text-xs text-zinc-500">诊断动作将作用于该自然日</div>
            </div>
            <input
              type="date"
              value={selectedDate}
              onChange={(e) => setSelectedDate(e.target.value)}
              className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 font-mono"
              max={todayLocalISODate()}
            />
          </div>
          <button
            onClick={handleBuild}
            disabled={actionLoading}
            className="w-full flex items-center justify-between p-3 bg-emerald-500/10 hover:bg-emerald-500/20 border border-emerald-500/20 rounded-lg transition-colors text-left group disabled:opacity-50"
          >
            <div>
              <div className="text-sm text-emerald-300 font-medium">构建会话</div>
              <div className="text-xs text-zinc-500">从该日原始数据生成会话（首次使用/无会话时）</div>
            </div>
            <Play size={16} className="text-emerald-600 group-hover:text-emerald-400" />
          </button>

          <button
            onClick={handleRebuild}
            disabled={actionLoading}
            className="w-full flex items-center justify-between p-3 bg-zinc-800/30 hover:bg-zinc-800 border border-zinc-800 rounded-lg transition-colors text-left group disabled:opacity-50"
          >
            <div>
              <div className="text-sm text-zinc-300 font-medium">重建会话</div>
              <div className="text-xs text-zinc-500">重新切分该日的原始数据（有风险提示）</div>
            </div>
            <RefreshCw size={16} className="text-zinc-600 group-hover:text-zinc-300" />
          </button>

          <button
            onClick={handleEnrich}
            disabled={actionLoading}
            className="w-full flex items-center justify-between p-3 bg-zinc-800/30 hover:bg-zinc-800 border border-zinc-800 rounded-lg transition-colors text-left group disabled:opacity-50"
          >
            <div>
              <div className="text-sm text-zinc-300 font-medium">增强缺失上下文</div>
              <div className="text-xs text-zinc-500">对该日未总结的会话运行 AI 分析（有风险提示）</div>
            </div>
            <RefreshCw size={16} className="text-zinc-600 group-hover:text-indigo-400" />
          </button>

          <a
            href="/api/diagnostics/export"
            download
            className="w-full flex items-center justify-between p-3 bg-zinc-800/30 hover:bg-zinc-800 border border-zinc-800 rounded-lg transition-colors text-left group"
          >
            <div>
              <div className="text-sm text-zinc-300 font-medium">导出诊断包</div>
              <div className="text-xs text-zinc-500">日志和数据库统计（已脱敏）</div>
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
              最近错误
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
