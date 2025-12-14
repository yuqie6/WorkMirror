import { useState, useEffect, useRef } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import {
  ChevronRight,
  ChevronDown,
  TrendingUp,
  TrendingDown,
  Minus,
  ExternalLink,
  FileCode,
  History,
  GripVertical,
  Server,
  Layout,
  Database,
  Box,
  Layers,
  Globe,
  Cpu,
  Code2,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { GetSkillTree, GetSkillEvidence, GetSkillSessions } from '@/api/app';
import { ISkillNode, SkillNodeDTO, buildSkillTree } from '@/types/skill';

interface SkillEvidence {
  source: string;
  evidence_id: number;
  timestamp: number;
  contribution_context: string;
  file_name: string;
}

interface SkillSession {
  id: number;
  category: string;
  summary: string;
  time_range: string;
  date: string;
}

interface SkillTreeItemProps {
  node: ISkillNode;
  selectedId: string | null;
  onSelect: (node: ISkillNode) => void;
  depth?: number;
}

// 根据技能名称推断图标
function getSkillIcon(name: string, type: string) {
  const lowerName = name.toLowerCase();
  if (type === 'domain') {
    if (lowerName.includes('backend') || lowerName.includes('后端')) return Server;
    if (lowerName.includes('frontend') || lowerName.includes('前端')) return Layout;
    if (lowerName.includes('database') || lowerName.includes('数据')) return Database;
    if (lowerName.includes('devops') || lowerName.includes('运维')) return Cpu;
    if (lowerName.includes('web')) return Globe;
    return Layers;
  }
  if (type === 'skill') {
    return Box;
  }
  return Code2;
}

// 层级样式配置
const levelStyles = {
  domain: {
    container: 'py-2.5 text-base font-bold',
    text: 'text-zinc-100 uppercase tracking-wide text-sm',
    icon: 'text-indigo-400',
    progress: 'bg-indigo-500',
  },
  skill: {
    container: 'py-2 text-[15px] font-medium',
    text: 'text-zinc-300',
    icon: 'text-emerald-400',
    progress: 'bg-emerald-500',
  },
  topic: {
    container: 'py-1.5 text-sm',
    text: 'text-zinc-400',
    icon: 'text-zinc-500',
    progress: 'bg-zinc-500',
  },
};

function SkillTreeItem({ node, selectedId, onSelect, depth = 0 }: SkillTreeItemProps) {
  const [open, setOpen] = useState(depth < 2);
  const hasChildren = node.children && node.children.length > 0;
  const isSelected = selectedId === node.id;

  const isRecent = node.lastActive === 'Today' || node.lastActive === 'Yesterday' || node.lastActive === '今天' || node.lastActive === '昨天';
  const TrendIcon = node.trend === 'up' ? TrendingUp : node.trend === 'down' ? TrendingDown : Minus;
  const trendColor = node.trend === 'up' ? 'text-emerald-500' : node.trend === 'down' ? 'text-rose-500' : 'text-zinc-500';

  const style = levelStyles[node.type] || levelStyles.topic;
  const Icon = getSkillIcon(node.name, node.type);

  return (
    <li className="relative">
      {/* 视觉引导线 - 非根节点显示 */}
      {depth > 0 && (
        <div 
          className="absolute left-0 top-0 w-px bg-zinc-800"
          style={{ height: hasChildren && open ? '100%' : '50%' }}
        />
      )}
      
      <Collapsible open={open} onOpenChange={setOpen}>
        <div
          className={cn(
            'flex items-center gap-3 px-3 rounded-lg cursor-pointer transition-all group',
            style.container,
            isSelected 
              ? 'bg-zinc-800/80 ring-1 ring-zinc-700 shadow-sm' 
              : 'hover:bg-zinc-800/50'
          )}
          onClick={() => onSelect(node)}
        >
          {/* 展开/折叠按钮或叶子节点圆点 */}
          <div className="w-5 h-5 flex items-center justify-center shrink-0">
            {hasChildren ? (
              <CollapsibleTrigger asChild onClick={(e: React.MouseEvent) => e.stopPropagation()}>
                <button className="text-zinc-500 hover:text-zinc-300 transition-colors">
                  {open ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                </button>
              </CollapsibleTrigger>
            ) : (
              <div className={cn(
                'w-1.5 h-1.5 rounded-full transition-colors',
                isSelected ? 'bg-white' : 'bg-zinc-700 group-hover:bg-zinc-500'
              )} />
            )}
          </div>

          {/* 节点图标 - 仅 domain 和 skill 显示 */}
          {node.type !== 'topic' && (
            <div className={cn(style.icon, 'opacity-80 shrink-0')}>
              <Icon size={node.type === 'domain' ? 18 : 16} />
            </div>
          )}

          {/* 文本内容 */}
          <div className="flex-1 flex justify-between items-center min-w-0 gap-3">
            <span className={cn(
              'truncate',
              style.text,
              !isRecent && node.type !== 'domain' && 'opacity-60'
            )}>
              {node.name}
            </span>
            
            {/* 右侧：Level + 进度条 + 趋势 */}
            <div className="flex items-center gap-3 shrink-0">
              <TrendIcon size={12} className={trendColor} />
              {node.type !== 'topic' && (
                <span className="text-xs font-mono text-zinc-500 bg-zinc-900/50 px-1.5 py-0.5 rounded border border-zinc-800">
                  Lv.{node.level}
                </span>
              )}
              {node.type === 'topic' && (
                <span className="text-[10px] font-mono text-zinc-600">Lv.{node.level}</span>
              )}
              <div className="w-14 h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                <div 
                  className={cn('h-full transition-all duration-500', style.progress)}
                  style={{ width: `${node.progress}%` }}
                />
              </div>
            </div>
          </div>
        </div>

        {hasChildren && (
          <CollapsibleContent>
            <ul className="ml-5 pl-4 mt-1 space-y-0.5 relative">
              {node.children!.map((child: ISkillNode) => (
                <SkillTreeItem key={child.id} node={child} selectedId={selectedId} onSelect={onSelect} depth={depth + 1} />
              ))}
            </ul>
          </CollapsibleContent>
        )}
      </Collapsible>
    </li>
  );
}

interface SkillViewProps {
  selectedSkillId?: string | null;
  onSelectSkill?: (skillId: string) => void;
  onNavigateToSession?: (sessionId: number, date: string) => void;
}

function findSkillNodeByID(nodes: ISkillNode[], id: string): ISkillNode | null {
  for (const n of nodes) {
    if (n.id === id) return n;
    if (n.children && n.children.length > 0) {
      const found = findSkillNodeByID(n.children, id);
      if (found) return found;
    }
  }
  return null;
}

export default function SkillView({ selectedSkillId, onSelectSkill, onNavigateToSession }: SkillViewProps) {
  const [skills, setSkills] = useState<ISkillNode[]>([]);
  const [selectedSkill, setSelectedSkill] = useState<ISkillNode | null>(null);
  const [loading, setLoading] = useState(false);
  const [evidence, setEvidence] = useState<SkillEvidence[]>([]);
  const [sessions, setSessions] = useState<SkillSession[]>([]);
  const [loadingEvidence, setLoadingEvidence] = useState(false);
  
  // 可拖拽分栏
  const [leftWidth, setLeftWidth] = useState(33);
  const containerRef = useRef<HTMLDivElement>(null);
  const isDragging = useRef(false);

  useEffect(() => {
    const loadSkills = async () => {
      setLoading(true);
      try {
        const data: SkillNodeDTO[] = await GetSkillTree();
        const tree = buildSkillTree(data);
        setSkills(tree);
        const preferred = typeof selectedSkillId === 'string' && selectedSkillId.trim() !== '' ? selectedSkillId.trim() : '';
        const next =
          (preferred ? findSkillNodeByID(tree, preferred) : null) ||
          (tree.length > 0 && tree[0].children && tree[0].children.length > 0 ? tree[0].children[0] : null);
        if (next) {
          setSelectedSkill(next);
          if (!preferred) onSelectSkill?.(next.id);
        }
      } catch (e) {
        console.error('Failed to load skills:', e);
      } finally {
        setLoading(false);
      }
    };
    loadSkills();
  }, []);

  useEffect(() => {
    const preferred = typeof selectedSkillId === 'string' && selectedSkillId.trim() !== '' ? selectedSkillId.trim() : '';
    if (!preferred || skills.length === 0) return;
    const found = findSkillNodeByID(skills, preferred);
    if (found && found.id !== selectedSkill?.id) {
      setSelectedSkill(found);
    }
  }, [selectedSkill?.id, selectedSkillId, skills]);

  useEffect(() => {
    if (!selectedSkill) return;
    
    const loadEvidence = async () => {
      setLoadingEvidence(true);
      try {
        const [evidenceData, sessionsData] = await Promise.all([
          GetSkillEvidence(selectedSkill.id).catch(() => []),
          GetSkillSessions(selectedSkill.id).catch(() => []),
        ]);
        setEvidence(evidenceData || []);
        setSessions(sessionsData || []);
      } catch (e) {
        console.error('Failed to load evidence:', e);
      } finally {
        setLoadingEvidence(false);
      }
    };
    loadEvidence();
  }, [selectedSkill]);

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
      setLeftWidth(Math.max(20, Math.min(60, newWidth)));
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

  const getTrendText = (trend: string) => {
    if (trend === 'up') return '↗ 上升中';
    if (trend === 'down') return '↘ 下降中';
    return '→ 稳定';
  };

  const formatTimestamp = (ts: number): string => {
    if (!ts) return '--';
    return new Date(ts).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
  };

  // 根据技能类型生成详情卡片的背景样式 - 增加视觉特点
  const getDetailHeaderStyle = () => {
    if (!selectedSkill) return {};
    const type = selectedSkill.type;
    if (type === 'domain') {
      return 'bg-gradient-to-br from-zinc-900 via-indigo-950/20 to-zinc-900';
    }
    if (type === 'skill') {
      return 'bg-gradient-to-br from-zinc-900 via-emerald-950/20 to-zinc-900';
    }
    return 'bg-zinc-900';
  };

  if (loading) {
    return <div className="flex items-center justify-center h-64 text-zinc-500">加载技能树中...</div>;
  }

  const handleSelectSkill = (node: ISkillNode) => {
    setSelectedSkill(node);
    onSelectSkill?.(node.id);
  };

  return (
    <div ref={containerRef} className="flex h-[calc(100vh-8rem)] animate-in fade-in duration-500">
      {/* Tree Explorer - 可拖拽宽度 */}
      <div style={{ width: `${leftWidth}%` }} className="pr-2 overflow-y-auto border-r border-zinc-800">
        <ul className="space-y-1">
          {skills.map((node) => (
            <SkillTreeItem key={node.id} node={node} selectedId={selectedSkill?.id ?? null} onSelect={handleSelectSkill} />
          ))}
        </ul>
      </div>

      {/* 拖拽手柄 */}
      <div
        onMouseDown={handleMouseDown}
        className="w-2 flex items-center justify-center cursor-col-resize hover:bg-zinc-800 transition-colors group"
      >
        <GripVertical size={12} className="text-zinc-600 group-hover:text-zinc-400" />
      </div>

      {/* Detail Pane */}
      <div style={{ width: `${100 - leftWidth - 1}%` }} className="pl-2 overflow-y-auto">
        {selectedSkill ? (
          <>
            {/* Header Card - 增强背景视觉 */}
            <Card className={cn('border-zinc-800 mb-6 relative overflow-hidden', getDetailHeaderStyle())}>
              {/* 装饰性网格背景 */}
              <div className="absolute inset-0 opacity-[0.03]" style={{
                backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
              }} />
              {/* 大号装饰文字 */}
              <div className="absolute -top-4 -right-4 p-8 opacity-[0.03] font-black text-[10rem] select-none leading-none">
                {selectedSkill.type === 'domain' ? '◆' : selectedSkill.type === 'skill' ? '◇' : '○'}
              </div>
              <CardContent className="p-8 relative z-10">
                <div className="flex items-center gap-3 mb-2">
                  <h2 className="text-3xl font-bold text-white">{selectedSkill.name}</h2>
                  <Badge variant="default">Lv.{selectedSkill.level}</Badge>
                </div>

                <div className="flex gap-8 mb-6">
                  <div>
                    <div className="text-xs text-zinc-500 uppercase">总经验值</div>
                    <div className="text-2xl font-mono text-indigo-400">
                      {selectedSkill.xp} <span className="text-sm text-zinc-600">({selectedSkill.progress}%)</span>
                    </div>
                  </div>
                  <div>
                    <div className="text-xs text-zinc-500 uppercase">趋势</div>
                    <div className={cn('text-2xl font-mono', selectedSkill.trend === 'up' ? 'text-emerald-500' : 'text-zinc-400')}>
                      {getTrendText(selectedSkill.trend)}
                    </div>
                  </div>
                  <div>
                    <div className="text-xs text-zinc-500 uppercase">最近活跃</div>
                    <div className="text-2xl font-mono text-zinc-300">{selectedSkill.lastActive}</div>
                  </div>
                </div>

                <div className="w-full h-2 bg-zinc-950 rounded-full overflow-hidden border border-zinc-800">
                  <div className="h-full bg-gradient-to-r from-indigo-500 to-purple-500" style={{ width: `${selectedSkill.progress}%` }} />
                </div>
              </CardContent>
            </Card>

            {/* 相关会话 - 可点击跳转 */}
            <Card className="bg-zinc-900 border-zinc-800 mb-6">
              <CardHeader>
                <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-wider flex items-center gap-2">
                  <History size={14} /> 相关会话
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {loadingEvidence ? (
                  <div className="text-zinc-500 text-sm">加载中...</div>
                ) : sessions.length > 0 ? (
                  sessions.slice(0, 5).map((session) => (
                    <div
                      key={session.id}
                      onClick={() => onNavigateToSession?.(session.id, session.date)}
                      className="p-3 bg-zinc-950 border border-zinc-800 rounded text-sm cursor-pointer hover:bg-zinc-900 hover:border-zinc-700 transition-colors group"
                    >
                      <div className="flex justify-between items-start">
                        <div>
                          <div className="font-mono text-xs text-indigo-400 mb-1">会话 #{session.id} • {session.date}</div>
                          <div className="text-zinc-300">{session.category || session.summary}</div>
                          <div className="text-xs text-zinc-500 mt-1">{session.time_range}</div>
                        </div>
                        <ExternalLink size={14} className="text-zinc-600 group-hover:text-indigo-400 transition-colors" />
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-zinc-500 text-sm italic">暂无相关会话</div>
                )}
              </CardContent>
            </Card>

            {/* 代码证据 - 使用后端正确字段 */}
            <Card className="bg-zinc-900 border-zinc-800">
              <CardHeader>
                <CardTitle className="text-sm font-bold text-zinc-400 uppercase tracking-wider flex items-center gap-2">
                  <FileCode size={14} /> 代码证据
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {loadingEvidence ? (
                  <div className="text-zinc-500 text-sm">加载中...</div>
                ) : evidence.length > 0 ? (
                  evidence.slice(0, 10).map((ev) => (
                    <Collapsible key={ev.evidence_id}>
                      <div className="bg-zinc-950 border border-zinc-800 rounded overflow-hidden">
                        <CollapsibleTrigger className="w-full p-3 flex items-center justify-between hover:bg-zinc-900/50 transition-colors text-left">
                          <div className="flex items-center gap-2 font-mono text-xs">
                            <FileCode size={14} className="text-emerald-400" />
                            <span className="text-zinc-300 truncate max-w-[200px]">{ev.file_name || ev.source}</span>
                            <Badge variant="outline" className="text-[10px]">{ev.source}</Badge>
                          </div>
                          <ChevronDown size={14} className="text-zinc-500" />
                        </CollapsibleTrigger>
                        <CollapsibleContent>
                          <div className="border-t border-zinc-800 p-3">
                            {ev.contribution_context && (
                              <div className="mb-2 p-2 bg-indigo-500/10 border border-indigo-500/20 rounded text-sm text-indigo-200">
                                <span className="text-indigo-400 font-medium">上下文：</span> {ev.contribution_context}
                              </div>
                            )}
                            <div className="text-xs text-zinc-500">
                              来源 ID: {ev.evidence_id}
                              <br />
                              时间: {formatTimestamp(ev.timestamp)}
                            </div>
                          </div>
                        </CollapsibleContent>
                      </div>
                    </Collapsible>
                  ))
                ) : (
                  <div className="text-zinc-500 text-sm italic">暂无代码证据</div>
                )}
              </CardContent>
            </Card>
          </>
        ) : (
          <div className="text-zinc-500 text-center mt-20">选择一个技能查看详情</div>
        )}
      </div>
    </div>
  );
}
