import { useState, useEffect, useRef } from 'react';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs';
import { Sparkles, Cog, AlertTriangle, ChevronDown, ChevronRight, FileCode, Plus, Minus, MonitorSmartphone, Globe, Clock, GripVertical, Calendar, ExternalLink } from 'lucide-react';
import { cn } from '@/lib/utils';
import { GetSessionsByDate, GetSessionDetail, GetSessionEvents, GetDiffDetail } from '@/api/app';
import { ISession, SessionDTO, SessionDetailDTO, SessionWindowEventDTO, toISession } from '@/types/session';
import { parseLocalISODate, todayLocalISODate } from '@/lib/date';
import { useTranslation } from '@/lib/i18n';

interface DiffDetail {
  id: number;
  file_name: string;
  language: string;
  diff_content: string;
  insight: string;
  skills: string[];
  lines_added: number;
  lines_deleted: number;
  timestamp: number;
}

interface SessionsViewProps {
  initialDate?: string;
  selectedSessionId?: number | null;
  onOpenSession?: (sessionId: number, date: string) => void;
  onCloseSession?: (date: string) => void;
  onDateChange?: (date: string, sessionId?: number | null) => void;
}

function parseOrToday(s: string | undefined): string {
  if (typeof s === 'string' && s.trim() !== '' && parseLocalISODate(s)) return s;
  return todayLocalISODate();
}

export default function SessionsView({
  initialDate,
  selectedSessionId,
  onOpenSession,
  onCloseSession,
  onDateChange,
}: SessionsViewProps) {
  const [sessions, setSessions] = useState<ISession[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedSession, setSelectedSession] = useState<SessionDetailDTO | null>(null);
  const [expandedDiffs, setExpandedDiffs] = useState<Set<number>>(new Set());
  const [currentDate, setCurrentDate] = useState(() => parseOrToday(initialDate));
  const { t } = useTranslation();
  
  // 窗口事件
  const [windowEvents, setWindowEvents] = useState<SessionWindowEventDTO[]>([]);
  const [loadingEvents, setLoadingEvents] = useState(false);
  
  // Diff 详情
  const [diffDetails, setDiffDetails] = useState<Map<number, DiffDetail>>(new Map());
  const [loadingDiff, setLoadingDiff] = useState<number | null>(null);

  // 可拖拽分栏
  const [leftWidth, setLeftWidth] = useState(40);
  const containerRef = useRef<HTMLDivElement>(null);
  const isDragging = useRef(false);

  // 加载会话列表
  useEffect(() => {
    const loadSessions = async () => {
      setLoading(true);
      try {
        const data: SessionDTO[] = await GetSessionsByDate(currentDate);
        setSessions(data.map(toISession));
      } catch (e) {
        console.error('Failed to load sessions:', e);
      } finally {
        setLoading(false);
      }
    };
    loadSessions();
  }, [currentDate]);

  useEffect(() => {
    const d = typeof initialDate === 'string' && initialDate.trim() !== '' ? initialDate : '';
    if (d && d !== currentDate && parseLocalISODate(d)) {
      setCurrentDate(d);
    }
  }, [currentDate, initialDate]);

  const openSessionByID = async (sessionId: number) => {
    try {
      const detail: SessionDetailDTO = await GetSessionDetail(sessionId);
      setSelectedSession(detail);
      setExpandedDiffs(new Set());
      setWindowEvents([]);
      setDiffDetails(new Map());
      loadWindowEvents(sessionId);
    } catch (e) {
      console.error('Failed to load session detail:', e);
    }
  };

  // URL 选中态：打开/关闭会话详情
  useEffect(() => {
    if (!selectedSessionId) {
      setSelectedSession(null);
      return;
    }
    void openSessionByID(selectedSessionId);
  }, [selectedSessionId]);

  const handleSessionClick = async (session: ISession) => {
    await openSessionByID(session.id);
    onOpenSession?.(session.id, currentDate);
  };

  const loadWindowEvents = async (sessionId: number) => {
    setLoadingEvents(true);
    try {
      const events = await GetSessionEvents(sessionId, 100);
      const windowEvts = (events?.window_events || events || []) as SessionWindowEventDTO[];
      setWindowEvents(windowEvts);
    } catch (e) {
      console.error('Failed to load window events:', e);
    } finally {
      setLoadingEvents(false);
    }
  };

  const loadDiffDetail = async (diffId: number) => {
    if (diffDetails.has(diffId)) return;
    setLoadingDiff(diffId);
    try {
      const detail: DiffDetail = await GetDiffDetail(diffId);
      setDiffDetails((prev) => new Map(prev).set(diffId, detail));
    } catch (e) {
      console.error('Failed to load diff detail:', e);
    } finally {
      setLoadingDiff(null);
    }
  };

  const toggleDiffExpand = async (diffId: number) => {
    setExpandedDiffs((prev) => {
      const next = new Set(prev);
      if (next.has(diffId)) {
        next.delete(diffId);
      } else {
        next.add(diffId);
        loadDiffDetail(diffId);
      }
      return next;
    });
  };

  const formatTimestamp = (ts: number): string => {
    if (!ts) return '--';
    return new Date(ts).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  };

  const formatDuration = (seconds: number): string => {
    const sec = Number.isFinite(seconds) ? Math.max(0, Math.floor(seconds)) : 0;
    if (sec < 60) return `${sec}${t('common.seconds')}`;
    const hours = Math.floor(sec / 3600);
    const minutes = Math.floor((sec % 3600) / 60);
    const remainSec = sec % 60;
    if (hours > 0) {
      if (minutes > 0) return `${hours}${t('common.hours')}${minutes}${t('common.minutes')}`;
      return `${hours}${t('common.hours')}`;
    }
    if (minutes > 0 && remainSec > 0) return `${minutes}${t('common.minutes')}${remainSec}${t('common.seconds')}`;
    return `${minutes}${t('common.minutes')}`;
  };

  const navigateDate = (direction: number) => {
    const base = parseLocalISODate(currentDate) || new Date();
    base.setDate(base.getDate() + direction);
    const next = `${base.getFullYear()}-${String(base.getMonth() + 1).padStart(2, '0')}-${String(base.getDate()).padStart(2, '0')}`;
    setCurrentDate(next);
    // 切换日期时清空选中态，避免 URL 指向不一致
    onDateChange?.(next, null);
  };

  // 拖拽逻辑
  const handleMouseDown = () => {
    isDragging.current = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
  };

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!isDragging.current || !containerRef.current) return;
      const rect = containerRef.current.getBoundingClientRect();
      const newWidth = ((e.clientX - rect.left) / rect.width) * 100;
      setLeftWidth(Math.max(25, Math.min(70, newWidth)));
    };

    const handleMouseUp = () => {
      isDragging.current = false;
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, []);

  if (loading) {
    return <div className="flex items-center justify-center h-64 text-zinc-500">{t('sessions.loading')}</div>;
  }

  return (
    <div ref={containerRef} className="flex h-[calc(100vh-8rem)] animate-in slide-in-from-bottom-2 duration-500">
      {/* 左侧：会话列表 */}
      <div style={{ width: `${leftWidth}%` }} className="pr-2 overflow-y-auto border-r border-zinc-800">
        {/* 日期选择 */}
        <div className="flex justify-between items-center mb-4 sticky top-0 bg-zinc-950 py-2 z-10">
          <h3 className="text-zinc-200 font-medium">{t('sessions.timeline')}</h3>
          <div className="flex items-center gap-2 text-zinc-400 bg-zinc-900 px-2 py-1 rounded-lg border border-zinc-800">
            <button onClick={() => navigateDate(-1)} className="p-1 hover:text-white transition-colors">
              <ChevronRight size={14} className="rotate-180" />
            </button>
            <span className="text-xs font-mono flex items-center gap-1">
              <Calendar size={12} /> {currentDate.slice(5)}
            </span>
            <button 
              onClick={() => navigateDate(1)} 
              className="p-1 hover:text-white transition-colors"
              disabled={currentDate >= todayLocalISODate()}
            >
              <ChevronRight size={14} />
            </button>
          </div>
        </div>

        {sessions.length === 0 ? (
          <div className="text-center text-zinc-500 py-12">{t('sessions.noRecords')}</div>
        ) : (
          <div className="space-y-2">
            {sessions.map((session) => (
              <div
                key={session.id}
                onClick={() => handleSessionClick(session)}
                className={cn(
                  'group relative pl-3 border-l-2 cursor-pointer hover:bg-zinc-900/50 rounded-r-lg p-3 transition-all',
                  session.type === 'ai' ? 'border-indigo-500' : 'border-zinc-700',
                  selectedSession?.id === session.id && 'bg-zinc-900 border-indigo-400'
                )}
              >
                <div className="flex justify-between items-start mb-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-mono text-zinc-500">{session.duration}</span>
                    <span title={session.type === 'ai' ? t('sessions.aiAnalysis') : t('sessions.ruleGenerated')}>
                      {session.type === 'ai' ? <Sparkles size={12} className="text-indigo-400" /> : <Cog size={12} className="text-zinc-600" />}
                    </span>
                  </div>
                </div>
                <h4 className="text-zinc-200 text-sm font-medium group-hover:text-white transition-colors">
                  {session.title}
                </h4>
                <p className="text-xs text-zinc-500 mt-1 line-clamp-2">{session.summary}</p>
                {session.evidenceStrength === 'weak' && (
                  <div
                    title={t('sessions.weakEvidenceHint')}
                    className="mt-1 text-[10px] text-amber-500 flex items-center gap-1"
                  >
                    <AlertTriangle size={10} /> {t('sessions.weakEvidence')}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 拖拽手柄 */}
      <div
        onMouseDown={handleMouseDown}
        className="w-2 flex items-center justify-center cursor-col-resize hover:bg-zinc-800 transition-colors group flex-shrink-0"
      >
        <GripVertical size={12} className="text-zinc-600 group-hover:text-zinc-400" />
      </div>

      {/* 右侧：会话详情 */}
      <div style={{ width: `${100 - leftWidth - 1}%` }} className="pl-2 overflow-y-auto">
        {selectedSession ? (
          <div className="space-y-4">
            {/* 会话头部 */}
            <div className="border-b border-zinc-800 pb-4">
              <div className="flex justify-end">
                <button
                  onClick={() => {
                    setSelectedSession(null);
                    onCloseSession?.(currentDate);
                  }}
                  className="text-xs text-zinc-500 hover:text-zinc-200 transition-colors"
                >
                  {t('sessions.backToList')}
                </button>
              </div>
              <div className="flex items-center gap-2 mb-2">
                <Badge variant={selectedSession.semantic_source === 'ai' ? 'default' : 'secondary'}>
                  {selectedSession.semantic_source === 'ai' ? t('sessions.aiAnalysis') : t('sessions.ruleGenerated')}
                </Badge>
                <Badge variant="outline">{selectedSession.time_range}</Badge>
              </div>
              <h2 className="text-xl font-bold text-white">{selectedSession.category || t('sessions.sessionDetail')}</h2>
              <p className="text-zinc-400 text-sm mt-2">{selectedSession.summary}</p>
            </div>

            {/* Tabs */}
            <Tabs defaultValue="diffs" className="w-full">
              <TabsList className="w-full bg-zinc-900 border border-zinc-800 p-1">
                <TabsTrigger value="diffs" className="flex-1 text-xs">
                  <FileCode size={12} className="mr-1" /> {t('sessions.codeChanges')}
                </TabsTrigger>
                <TabsTrigger value="timeline" className="flex-1 text-xs">
                  <Clock size={12} className="mr-1" /> {t('sessions.activityTimeline')}
                </TabsTrigger>
                <TabsTrigger value="browser" className="flex-1 text-xs">
                  <Globe size={12} className="mr-1" /> {t('sessions.browserActivity')}
                </TabsTrigger>
                <TabsTrigger value="apps" className="flex-1 text-xs">
                  <MonitorSmartphone size={12} className="mr-1" /> {t('sessions.appUsage')}
                </TabsTrigger>
              </TabsList>

              {/* 代码变更 */}
              <TabsContent value="diffs" className="mt-4">
                {selectedSession.diffs.length > 0 ? (
                  <div className="space-y-2">
                    {selectedSession.diffs.map((diff) => {
                      const isExpanded = expandedDiffs.has(diff.id);
                      const detail = diffDetails.get(diff.id);
                      return (
                        <Collapsible key={diff.id} open={isExpanded} onOpenChange={() => toggleDiffExpand(diff.id)}>
                          <div className="bg-zinc-900 border border-zinc-800 rounded overflow-hidden">
                            <CollapsibleTrigger className="w-full p-3 flex items-center justify-between hover:bg-zinc-800/50 transition-colors">
                              <div className="flex items-center gap-2 font-mono text-xs">
                                <FileCode size={14} className="text-indigo-400" />
                                <span className="text-zinc-300">{diff.file_name}</span>
                                <span className="text-zinc-600">{diff.language}</span>
                              </div>
                              <div className="flex items-center gap-3">
                                <span className="text-emerald-500 text-xs flex items-center gap-0.5"><Plus size={10} /> {diff.lines_added}</span>
                                <span className="text-rose-500 text-xs flex items-center gap-0.5"><Minus size={10} /> {diff.lines_deleted}</span>
                                {isExpanded ? <ChevronDown size={14} className="text-zinc-500" /> : <ChevronRight size={14} className="text-zinc-500" />}
                              </div>
                            </CollapsibleTrigger>
                            <CollapsibleContent>
                              <div className="border-t border-zinc-800 p-3 bg-zinc-950">
                                {loadingDiff === diff.id ? (
                                  <div className="text-zinc-500 text-sm">{t('common.loading')}</div>
                                ) : (
                                  <>
                                    {(detail?.insight || diff.insight) && (
                                      <div className="mb-3 p-2 bg-indigo-500/10 border border-indigo-500/20 rounded text-sm text-indigo-200">
                                        <span className="text-indigo-400 font-medium">{t('sessions.aiInsight')}</span> {detail?.insight || diff.insight}
                                      </div>
                                    )}
                                    {detail?.diff_content && (
                                      <pre className="text-xs font-mono bg-zinc-900 border border-zinc-800 rounded p-3 overflow-x-auto max-h-64 overflow-y-auto">
                                        {detail.diff_content.split('\n').map((line, idx) => (
                                          <div
                                            key={idx}
                                            className={cn(
                                              line.startsWith('+') && !line.startsWith('+++') ? 'text-emerald-400 bg-emerald-500/10' :
                                              line.startsWith('-') && !line.startsWith('---') ? 'text-rose-400 bg-rose-500/10' :
                                              line.startsWith('@@') ? 'text-sky-400' : 'text-zinc-400'
                                            )}
                                          >
                                            {line}
                                          </div>
                                        ))}
                                      </pre>
                                    )}
                                    {diff.skills && diff.skills.length > 0 && (
                                      <div className="mt-2 flex gap-1 flex-wrap">
                                        <span className="text-zinc-500 text-xs">{t('sessions.relatedSkills')}</span>
                                        {diff.skills.map((skill: string) => (
                                          <span key={skill} className="px-1.5 py-0.5 bg-indigo-500/20 text-indigo-300 rounded text-[10px]">{skill}</span>
                                        ))}
                                      </div>
                                    )}
                                  </>
                                )}
                              </div>
                            </CollapsibleContent>
                          </div>
                        </Collapsible>
                      );
                    })}
                  </div>
                ) : (
                  <div className="text-zinc-500 text-sm italic text-center py-8">{t('sessions.noDiffRecords')}</div>
                )}
              </TabsContent>

              {/* 活动时间轴 */}
              <TabsContent value="timeline" className="mt-4">
                {loadingEvents ? (
                  <div className="text-zinc-500 text-sm text-center py-8">{t('common.loading')}</div>
                ) : windowEvents.length > 0 ? (
                  <div className="space-y-1">
                    {windowEvents.map((evt, idx) => (
                      <div key={idx} className="p-2 bg-zinc-900 border border-zinc-800 rounded text-sm">
                        <div className="flex items-center gap-3">
                          <span className="text-xs font-mono text-zinc-600 w-12">{formatTimestamp(evt.timestamp)}</span>
                          <MonitorSmartphone size={12} className="text-zinc-500" />
                          <span className="text-zinc-300 truncate flex-1">{evt.app_name}</span>
                          {evt.duration > 0 && <span className="text-xs text-zinc-600">{formatDuration(evt.duration)}</span>}
                        </div>
                        {evt.title && (
                          <div className="mt-1 text-xs text-zinc-500 pl-[60px] truncate">
                            {evt.title}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-zinc-500 text-sm italic text-center py-8">{t('sessions.noWindowEvents')}</div>
                )}
              </TabsContent>

              {/* 浏览器证据 */}
              <TabsContent value="browser" className="mt-4">
                {selectedSession.browser && selectedSession.browser.length > 0 ? (
                  <div className="space-y-2">
                    {selectedSession.browser.slice(0, 100).map((evt, idx) => (
                      <div key={idx} className="p-2 bg-zinc-900 border border-zinc-800 rounded text-sm">
                        <div className="flex items-center gap-3">
                          <span className="text-xs font-mono text-zinc-600 w-12">{formatTimestamp(evt.timestamp)}</span>
                          <Globe size={12} className="text-sky-500" />
                          <span className="text-zinc-400">{evt.domain}</span>
                          {evt.duration > 0 && <span className="text-xs text-zinc-600">{formatDuration(evt.duration)}</span>}
                        </div>
                        <div className="mt-1 text-xs text-zinc-500 pl-[60px] space-y-1">
                          {evt.title && <div className="truncate">{evt.title}</div>}
                          {evt.url && (
                            <a
                              href={evt.url}
                              target="_blank"
                              rel="noreferrer"
                              className="inline-flex items-center gap-1 text-sky-400 hover:text-sky-300 truncate max-w-full"
                              title={evt.url}
                            >
                              <ExternalLink size={12} />
                              <span className="truncate">{evt.url}</span>
                            </a>
                          )}
                        </div>
                      </div>
                    ))}
                    {selectedSession.browser.length > 100 && (
                      <div className="text-xs text-zinc-600 text-center py-2">{t('sessions.browserEvidenceTruncated')}</div>
                    )}
                  </div>
                ) : (
                  <div className="text-zinc-500 text-sm italic text-center py-8">{t('sessions.noBrowserEvents')}</div>
                )}
              </TabsContent>

              {/* 应用使用 */}
              <TabsContent value="apps" className="mt-4">
                {selectedSession.app_usage.length > 0 ? (
                  <div className="space-y-3">
                    {selectedSession.app_usage.map((app, idx: number) => {
                      const totalDuration = selectedSession.app_usage.reduce((sum: number, a: { total_duration: number }) => sum + a.total_duration, 0);
                      const percent = totalDuration > 0 ? Math.round((app.total_duration / totalDuration) * 100) : 0;
                      return (
                        <div key={idx} className="flex items-center gap-3">
                          <div className="flex-1 text-sm text-zinc-400 text-right w-24">{app.app_name}</div>
                          <div className="flex-[3]"><Progress value={percent} className="h-2" /></div>
                          <div className="w-12 text-xs text-zinc-500">{percent}%</div>
                          <div className="w-24 text-xs text-zinc-600 text-right whitespace-nowrap">{formatDuration(app.total_duration)}</div>
                        </div>
                      );
                    })}
                  </div>
                ) : (
                  <div className="text-zinc-500 text-sm italic text-center py-8">{t('sessions.noAppUsageData')}</div>
                )}
              </TabsContent>
            </Tabs>
          </div>
        ) : (
          <div className="flex items-center justify-center h-full text-zinc-500">
            {t('sessions.selectSession')}
          </div>
        )}
      </div>
    </div>
  );
}
