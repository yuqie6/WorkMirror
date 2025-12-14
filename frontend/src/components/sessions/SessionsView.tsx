import React, { useMemo, useState } from "react";
import { useSessionsByDate } from "../../hooks/useSessionsByDate";
import SessionDetailModal from "./SessionDetailModal";
import SessionList from "./SessionList";

const SessionsView: React.FC = () => {
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10));
  const { sessions, loading, error, reload } = useSessionsByDate(date);
  const [selectedId, setSelectedId] = useState<number | null>(null);

  const coverage = useMemo(() => {
    let withDiff = 0;
    let withBrowser = 0;
    let weak = 0;
    for (const s of sessions) {
      const hasDiff = (s.diff_count || 0) > 0;
      const hasBrowser = (s.browser_count || 0) > 0;
      if (hasDiff) withDiff++;
      if (hasBrowser) withBrowser++;
      if (!hasDiff && !hasBrowser) weak++;
    }
    return { total: sessions.length, withDiff, withBrowser, weak };
  }, [sessions]);

  return (
    <div className="py-8 space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-2xl font-bold text-gray-900">Sessions</h1>
          <div className="flex flex-wrap gap-2">
            <span className="pill">总数 {coverage.total}</span>
            <span className="pill">含 diff {coverage.withDiff}</span>
            <span className="pill">含 browser {coverage.withBrowser}</span>
            <span className="pill">弱证据 {coverage.weak}</span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <input
            type="date"
            className="rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm outline-none focus:border-amber-300"
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
          <button className="btn-gold disabled:opacity-60" onClick={() => void reload()} disabled={loading}>
            {loading ? "刷新中..." : "刷新"}
          </button>
        </div>
      </div>

      {error && <div className="text-sm text-red-600">{error}</div>}

      <div className="card">
        {loading && (
          <div className="flex items-center gap-3 text-sm text-gray-500">
            <div className="w-5 h-5 border-2 border-gray-200 border-t-accent-gold rounded-full animate-spin" />
            加载中...
          </div>
        )}

        {!loading && sessions.length === 0 && (
          <div className="text-sm text-gray-400">
            该日期暂无会话。建议检查 Status 页中的采集/切分健康，并确认已配置 Diff watch paths。
          </div>
        )}

        {!loading && sessions.length > 0 && (
          <SessionList sessions={sessions} onSelect={(id) => setSelectedId(id)} />
        )}
      </div>

      {selectedId != null && <SessionDetailModal sessionId={selectedId} onClose={() => setSelectedId(null)} />}
    </div>
  );
};

export default SessionsView;
