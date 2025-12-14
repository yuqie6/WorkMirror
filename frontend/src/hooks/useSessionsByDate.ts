import { useCallback, useEffect, useMemo, useState } from "react";
import { BuildSessionsForDate, EnrichSessionsForDate, GetSessionsByDate, RebuildSessionsForDate } from "../api/app";
import type { SessionDTO } from "../types/session";

export function useSessionsByDate(date: string | null | undefined) {
  const [sessions, setSessions] = useState<SessionDTO[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canLoad = useMemo(() => !!date && String(date).trim().length > 0, [date]);

  const reload = useCallback(async () => {
    if (!canLoad) {
      setSessions([]);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const res = (await GetSessionsByDate(String(date))) as SessionDTO[];
      setSessions((res || []) as SessionDTO[]);
    } catch (e: any) {
      setError(e?.message || "加载会话失败");
      setSessions([]);
    } finally {
      setLoading(false);
    }
  }, [canLoad, date]);

  useEffect(() => {
    void reload();
  }, [reload]);

  const build = useCallback(async () => {
    if (!canLoad) return;
    setLoading(true);
    setError(null);
    try {
      await BuildSessionsForDate(String(date));
      await reload();
    } catch (e: any) {
      setError(e?.message || "切分会话失败");
    } finally {
      setLoading(false);
    }
  }, [canLoad, date, reload]);

  const rebuild = useCallback(async () => {
    if (!canLoad) return;
    setLoading(true);
    setError(null);
    try {
      await RebuildSessionsForDate(String(date));
      await reload();
    } catch (e: any) {
      setError(e?.message || "重建会话失败");
    } finally {
      setLoading(false);
    }
  }, [canLoad, date, reload]);

  const enrich = useCallback(async () => {
    if (!canLoad) return;
    setLoading(true);
    setError(null);
    try {
      await EnrichSessionsForDate(String(date));
      await reload();
    } catch (e: any) {
      setError(e?.message || "补全会话语义失败");
    } finally {
      setLoading(false);
    }
  }, [canLoad, date, reload]);

  return { sessions, loading, error, reload, build, rebuild, enrich };
}
