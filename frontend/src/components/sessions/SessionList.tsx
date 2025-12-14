import React from "react";
import type { SessionDTO } from "../../types/session";

const sessionCategoryLabel = (cat: string): string => {
  switch ((cat || "").toLowerCase()) {
    case "technical":
      return "技术";
    case "learning":
      return "学习";
    case "exploration":
      return "探索";
    case "other":
      return "其他";
    default:
      return cat || "其他";
  }
};

const SessionList: React.FC<{ sessions: SessionDTO[]; onSelect: (id: number) => void }> = ({ sessions, onSelect }) => {
  if (!sessions?.length) return null;
  return (
    <div className="space-y-2">
      {sessions.map((s) => (
        <button
          key={s.id}
          className="w-full text-left p-3 rounded-xl border border-gray-100 hover:border-amber-200 hover:bg-amber-50/40 transition"
          onClick={() => onSelect(s.id)}
        >
          <div className="flex items-center justify-between gap-3">
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-sm font-semibold text-gray-900">{s.time_range || "会话"}</span>
                {s.category && <span className="pill">{sessionCategoryLabel(s.category)}</span>}
                {(s.diff_count || 0) > 0 && <span className="text-xs text-gray-400">Diff {s.diff_count}</span>}
                {(s.browser_count || 0) > 0 && <span className="text-xs text-gray-400">Browser {s.browser_count}</span>}
              </div>
              <div className="text-xs text-gray-400 truncate">{s.primary_app || ""}</div>
            </div>
            <div className="text-sm text-gray-700 line-clamp-2 max-w-[55%]">{s.summary || "（未生成摘要）"}</div>
          </div>
        </button>
      ))}
    </div>
  );
};

export default SessionList;
