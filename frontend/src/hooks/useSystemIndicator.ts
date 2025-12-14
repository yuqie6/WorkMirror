import { useCallback, useEffect, useState } from "react";
import { GetStatus } from "../api/app";

export type SystemIndicatorTone = "info" | "warn" | "danger";

export type SystemIndicator = { text: string; tone: SystemIndicatorTone } | null;

const toIndicator = (st: any): SystemIndicator => {
  const safe = !!st?.app?.safe_mode;
  const mode = String(st?.pipeline?.ai?.mode || "");
  const degraded = !!st?.pipeline?.ai?.degraded;
  if (safe) return { text: "安全模式（只读）", tone: "danger" };
  if (mode === "offline") return { text: "离线模式（规则口径）", tone: "warn" };
  if (degraded) return { text: "降级模式", tone: "warn" };
  return { text: "AI 模式", tone: "info" };
};

export function useSystemIndicator() {
  const [systemIndicator, setSystemIndicator] = useState<SystemIndicator>(null);

  const refreshSystemIndicator = useCallback(async () => {
    try {
      const st = await GetStatus();
      setSystemIndicator(toIndicator(st));
    } catch {
      // ignore: Status 页可单独诊断
    }
  }, []);

  useEffect(() => {
    void refreshSystemIndicator();
  }, [refreshSystemIndicator]);

  return { systemIndicator, refreshSystemIndicator };
}

