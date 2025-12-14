import React, { useEffect, useMemo, useState } from "react";
import { EnrichSessionsForDate, GetStatus, RebuildSessionsForDate } from "../../api/app";
import { useToast } from "../common/Toast";

type StatusDTO = {
  app: { name: string; version: string; started_at: string; uptime_sec: number; safe_mode: boolean; config_path?: string };
  storage: { db_path: string; schema_version: number; safe_mode_reason?: string };
  privacy: { enabled: boolean; pattern_count: number };
  collectors: {
    window: { enabled: boolean; running: boolean; last_collected_at: number; last_persisted_at: number; count_24h: number; dropped_events?: number; dropped_batches?: number };
    diff: { enabled: boolean; running: boolean; last_collected_at: number; last_persisted_at: number; count_24h: number; dropped_events?: number; skipped?: number; watch_paths?: string[]; effective_paths?: number };
    browser: { enabled: boolean; running: boolean; last_collected_at: number; last_persisted_at: number; count_24h: number; dropped_events?: number; history_path?: string; sanitized_enabled?: boolean };
  };
  pipeline: {
    sessions: { last_split_at: number; sessions_24h: number; pending_semantic_24h: number; last_semantic_at: number };
    ai: { configured: boolean; mode: "ai" | "offline"; last_call_at: number; last_error?: string; last_error_at?: number; degraded: boolean; degraded_reason?: string };
    rag: { enabled: boolean; index_count?: number; last_error?: string; last_error_at?: number };
  };
  evidence: { sessions_24h: number; with_diff: number; with_browser: number; with_diff_and_browser: number; weak_evidence: number };
  recent_errors: Array<{ time?: string; level?: string; message: string; raw?: string }>;
};

const formatTs = (ms?: number) => {
  if (!ms) return "—";
  try {
    return new Date(ms).toLocaleString();
  } catch {
    return String(ms);
  }
};

const StatusView: React.FC = () => {
  const { showToast } = useToast();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<StatusDTO | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [maintaining, setMaintaining] = useState(false);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = (await GetStatus()) as StatusDTO;
      setData(res);
    } catch (e: any) {
      setError(e?.message || "加载状态失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const modeBadge = useMemo(() => {
    if (!data) return null;
    if (data.app.safe_mode) return { text: "安全模式（只读）", tone: "bg-red-50 text-red-700 border-red-100" };
    if (data.pipeline.ai.mode === "offline") return { text: "离线模式（规则口径）", tone: "bg-amber-50 text-amber-700 border-amber-100" };
    if (data.pipeline.ai.degraded) return { text: "降级模式", tone: "bg-amber-50 text-amber-700 border-amber-100" };
    return { text: "AI 模式", tone: "bg-emerald-50 text-emerald-700 border-emerald-100" };
  }, [data]);

  const runMaintenance = async (kind: "rebuild" | "enrich") => {
    if (!date) return;
    const actionLabel = kind === "rebuild" ? "重建会话（Session）" : "补全会话语义";
    const ok = window.confirm(`确认对 ${date} 执行：${actionLabel}？`);
    if (!ok) return;
    setMaintaining(true);
    try {
      if (kind === "rebuild") {
        const res = await RebuildSessionsForDate(date);
        showToast(`重建完成：created=${res?.created ?? 0} enriched=${res?.enriched ?? 0}`, "success");
      } else {
        const res = await EnrichSessionsForDate(date);
        showToast(`补全完成：enriched=${res?.enriched ?? 0}`, "success");
      }
      await load();
    } catch (e: any) {
      showToast(e?.message || "操作失败", "error");
    } finally {
      setMaintaining(false);
    }
  };

  const downloadDiagnostics = () => {
    window.location.href = "/api/diagnostics/export";
  };

  return (
    <div className="py-8 space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-2xl font-bold text-gray-900">系统状态</h1>
          <div className="flex items-center gap-2">
            {modeBadge && <span className={`px-3 py-1 rounded-full border text-xs font-medium ${modeBadge.tone}`}>{modeBadge.text}</span>}
            {data?.privacy?.enabled && <span className="pill">脱敏已启用</span>}
            {data?.pipeline?.rag?.enabled && <span className="pill">RAG 已启用</span>}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button className="btn-gold disabled:opacity-60" onClick={() => void load()} disabled={loading}>
            {loading ? "刷新中..." : "刷新"}
          </button>
          <button className="px-4 py-2 rounded-xl border border-gray-200 bg-white text-sm hover:bg-gray-50" onClick={downloadDiagnostics}>
            导出诊断包
          </button>
        </div>
      </div>

      {error && <div className="text-sm text-red-600">{error}</div>}

      {data && (
        <>
          <div className="grid grid-cols-12 gap-4">
            <div className="col-span-12 md:col-span-6">
              <div className="card">
                <h3 className="text-sm font-semibold text-gray-900 mb-3">基础信息</h3>
                <div className="text-sm text-gray-600 space-y-2">
                  <div className="flex items-center justify-between"><span>应用</span><span className="text-gray-900">{data.app.name} {data.app.version}</span></div>
                  <div className="flex items-center justify-between"><span>启动时间</span><span className="text-gray-900">{data.app.started_at}</span></div>
                  <div className="flex items-center justify-between"><span>运行时长</span><span className="text-gray-900">{Math.floor((data.app.uptime_sec || 0) / 60)} 分钟</span></div>
                  <div className="flex items-center justify-between"><span>配置文件</span><span className="text-gray-900 truncate max-w-[65%]" title={data.app.config_path || ""}>{data.app.config_path || "—"}</span></div>
                </div>
              </div>
            </div>
            <div className="col-span-12 md:col-span-6">
              <div className="card">
                <h3 className="text-sm font-semibold text-gray-900 mb-3">存储与迁移</h3>
                <div className="text-sm text-gray-600 space-y-2">
                  <div className="flex items-center justify-between"><span>DB 路径</span><span className="text-gray-900 truncate max-w-[65%]" title={data.storage.db_path}>{data.storage.db_path}</span></div>
                  <div className="flex items-center justify-between"><span>schema_version</span><span className="text-gray-900">{data.storage.schema_version}</span></div>
                  <div className="flex items-center justify-between"><span>安全模式</span><span className={data.app.safe_mode ? "text-red-600 font-medium" : "text-gray-900"}>{data.app.safe_mode ? "是" : "否"}</span></div>
                  {data.storage.safe_mode_reason && (
                    <div className="text-xs text-red-600 whitespace-pre-wrap break-words">原因：{data.storage.safe_mode_reason}</div>
                  )}
                </div>
              </div>
            </div>
          </div>

          <div className="card">
            <h3 className="text-sm font-semibold text-gray-900 mb-4">采集健康（近 24h）</h3>
            <div className="grid grid-cols-12 gap-4">
              <div className="col-span-12 md:col-span-4">
                <div className="p-4 rounded-2xl bg-white border border-gray-100">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium text-gray-900">Window</span>
                    <span className={`text-xs ${data.collectors.window.running ? "text-emerald-600" : "text-gray-400"}`}>{data.collectors.window.running ? "运行中" : "未运行"}</span>
                  </div>
                  <div className="text-xs text-gray-500 space-y-1">
                    <div className="flex justify-between"><span>写入数</span><span className="text-gray-900">{data.collectors.window.count_24h}</span></div>
                    <div className="flex justify-between"><span>最近采集</span><span className="text-gray-900">{formatTs(data.collectors.window.last_collected_at)}</span></div>
                    <div className="flex justify-between"><span>最近落库</span><span className="text-gray-900">{formatTs(data.collectors.window.last_persisted_at)}</span></div>
                    {typeof data.collectors.window.dropped_events === "number" && <div className="flex justify-between"><span>丢弃事件</span><span className="text-gray-900">{data.collectors.window.dropped_events}</span></div>}
                    {typeof data.collectors.window.dropped_batches === "number" && <div className="flex justify-between"><span>丢弃批次</span><span className="text-gray-900">{data.collectors.window.dropped_batches}</span></div>}
                  </div>
                </div>
              </div>
              <div className="col-span-12 md:col-span-4">
                <div className="p-4 rounded-2xl bg-white border border-gray-100">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium text-gray-900">Diff</span>
                    <span className={`text-xs ${data.collectors.diff.running ? "text-emerald-600" : "text-gray-400"}`}>{data.collectors.diff.running ? "运行中" : "未运行"}</span>
                  </div>
                  <div className="text-xs text-gray-500 space-y-1">
                    <div className="flex justify-between"><span>diff 数</span><span className="text-gray-900">{data.collectors.diff.count_24h}</span></div>
                    <div className="flex justify-between"><span>watch paths</span><span className="text-gray-900">{data.collectors.diff.effective_paths ?? (data.collectors.diff.watch_paths?.length ?? 0)}</span></div>
                    <div className="flex justify-between"><span>最近采集</span><span className="text-gray-900">{formatTs(data.collectors.diff.last_collected_at)}</span></div>
                    <div className="flex justify-between"><span>最近落库</span><span className="text-gray-900">{formatTs(data.collectors.diff.last_persisted_at)}</span></div>
                    {typeof data.collectors.diff.skipped === "number" && <div className="flex justify-between"><span>非 Git 跳过</span><span className="text-gray-900">{data.collectors.diff.skipped}</span></div>}
                  </div>
                  {data.collectors.diff.watch_paths?.length ? (
                    <div className="mt-3 flex flex-wrap gap-2">
                      {data.collectors.diff.watch_paths.slice(0, 6).map((p) => <span key={p} className="pill">{p}</span>)}
                      {data.collectors.diff.watch_paths.length > 6 && <span className="pill">+{data.collectors.diff.watch_paths.length - 6}</span>}
                    </div>
                  ) : (
                    <div className="mt-3 text-xs text-amber-700 bg-amber-50 border border-amber-100 rounded-xl p-3">
                      未配置 Diff 监控目录：建议在设置里补充 watch paths，否则证据链会偏弱。
                    </div>
                  )}
                </div>
              </div>
              <div className="col-span-12 md:col-span-4">
                <div className="p-4 rounded-2xl bg-white border border-gray-100">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium text-gray-900">Browser</span>
                    <span className={`text-xs ${data.collectors.browser.running ? "text-emerald-600" : "text-gray-400"}`}>{data.collectors.browser.running ? "运行中" : "未运行"}</span>
                  </div>
                  <div className="text-xs text-gray-500 space-y-1">
                    <div className="flex justify-between"><span>events 数</span><span className="text-gray-900">{data.collectors.browser.count_24h}</span></div>
                    <div className="flex justify-between"><span>history</span><span className="text-gray-900 truncate max-w-[65%]" title={data.collectors.browser.history_path || ""}>{data.collectors.browser.history_path || "—"}</span></div>
                    <div className="flex justify-between"><span>最近采集</span><span className="text-gray-900">{formatTs(data.collectors.browser.last_collected_at)}</span></div>
                    <div className="flex justify-between"><span>最近落库</span><span className="text-gray-900">{formatTs(data.collectors.browser.last_persisted_at)}</span></div>
                    {typeof data.collectors.browser.dropped_events === "number" && <div className="flex justify-between"><span>丢弃事件</span><span className="text-gray-900">{data.collectors.browser.dropped_events}</span></div>}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-12 gap-4">
            <div className="col-span-12 md:col-span-6">
              <div className="card">
                <h3 className="text-sm font-semibold text-gray-900 mb-3">管道健康</h3>
                <div className="text-sm text-gray-600 space-y-2">
                  <div className="flex items-center justify-between"><span>Sessions 近 24h</span><span className="text-gray-900">{data.pipeline.sessions.sessions_24h}</span></div>
                  <div className="flex items-center justify-between"><span>待补全语义 近 24h</span><span className="text-gray-900">{data.pipeline.sessions.pending_semantic_24h}</span></div>
                  <div className="flex items-center justify-between"><span>最近切分</span><span className="text-gray-900">{formatTs(data.pipeline.sessions.last_split_at)}</span></div>
                  <div className="flex items-center justify-between"><span>最近补全</span><span className="text-gray-900">{formatTs(data.pipeline.sessions.last_semantic_at)}</span></div>
                  <div className="pt-2 border-t border-gray-100" />
                  <div className="flex items-center justify-between"><span>AI</span><span className="text-gray-900">{data.pipeline.ai.mode === "ai" ? "已配置" : "未配置（离线）"}</span></div>
                  {data.pipeline.ai.last_error && <div className="text-xs text-amber-700 whitespace-pre-wrap break-words">最近错误：{data.pipeline.ai.last_error}</div>}
                </div>
              </div>
            </div>
            <div className="col-span-12 md:col-span-6">
              <div className="card">
                <h3 className="text-sm font-semibold text-gray-900 mb-3">证据覆盖率（近 24h 会话）</h3>
                <div className="text-sm text-gray-600 space-y-2">
                  <div className="flex items-center justify-between"><span>会话数</span><span className="text-gray-900">{data.evidence.sessions_24h}</span></div>
                  <div className="flex items-center justify-between"><span>包含 diff</span><span className="text-gray-900">{data.evidence.with_diff}</span></div>
                  <div className="flex items-center justify-between"><span>包含 browser</span><span className="text-gray-900">{data.evidence.with_browser}</span></div>
                  <div className="flex items-center justify-between"><span>diff + browser</span><span className="text-gray-900">{data.evidence.with_diff_and_browser}</span></div>
                  <div className="flex items-center justify-between"><span>弱证据（无 diff/浏览）</span><span className="text-gray-900">{data.evidence.weak_evidence}</span></div>
                </div>
              </div>
            </div>
          </div>

          <div className="card">
            <h3 className="text-sm font-semibold text-gray-900 mb-3">维护操作</h3>
            <div className="flex flex-col md:flex-row md:items-center gap-3">
              <div className="flex items-center gap-2">
                <span className="text-sm text-gray-600">日期</span>
                <input
                  type="date"
                  className="rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm outline-none focus:border-amber-300"
                  value={date}
                  onChange={(e) => setDate(e.target.value)}
                />
              </div>
              <div className="flex items-center gap-2">
                <button className="px-4 py-2 rounded-xl border border-gray-200 bg-white text-sm hover:bg-gray-50 disabled:opacity-60" onClick={() => void runMaintenance("rebuild")} disabled={maintaining}>
                  重建会话（Session）
                </button>
                <button className="px-4 py-2 rounded-xl border border-gray-200 bg-white text-sm hover:bg-gray-50 disabled:opacity-60" onClick={() => void runMaintenance("enrich")} disabled={maintaining}>
                  补全会话语义
                </button>
              </div>
              <div className="text-xs text-gray-400">影响范围：重建会改变会话口径（version 覆盖展示）；补全仅更新摘要/证据索引。安全模式下禁用写入。</div>
            </div>
          </div>

          <div className="card">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-semibold text-gray-900">最近错误（从日志提取）</h3>
              <span className="text-xs text-gray-400">{data.recent_errors?.length || 0}</span>
            </div>
            {data.recent_errors?.length ? (
              <div className="space-y-2">
                {data.recent_errors.map((e, i) => (
                  <div key={i} className="text-xs text-gray-600 border border-gray-100 rounded-xl p-3">
                    <div className="flex items-center gap-2 mb-1">
                      {e.level && <span className="pill">{e.level}</span>}
                      {e.time && <span className="text-gray-400">{e.time}</span>}
                    </div>
                    <div className="whitespace-pre-wrap break-words">{e.message || e.raw}</div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-sm text-gray-400">暂无</div>
            )}
          </div>
        </>
      )}
    </div>
  );
};

export default StatusView;
