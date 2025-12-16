import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { Settings as SettingsIcon, Moon, Shield, Save, RefreshCw, Plus, X, Key, Server, Database, Eye } from 'lucide-react';
import { GetSettings, SaveSettings } from '@/api/app';
import { useTranslation } from '@/lib/i18n';

// 匹配后端 dto/httpapi.go SettingsDTO 完整字段
interface SettingsData {
  config_path: string;

  language: string; // AI Prompt 语言偏好：zh/en

  ai: {
    provider: 'default' | 'openai' | 'anthropic' | 'google' | 'zhipu' | string;
    default: {
      enabled: boolean;
      api_key_set: boolean;
      base_url: string;
      model: string;
      api_key_locked?: boolean;
      base_url_locked?: boolean;
      model_locked?: boolean;
    };
    openai: {
      api_key_set: boolean;
      base_url: string;
      model: string;
    };
    anthropic: {
      api_key_set: boolean;
      base_url: string;
      model: string;
    };
    google: {
      api_key_set: boolean;
      base_url: string;
      model: string;
    };
    zhipu: {
      api_key_set: boolean;
      base_url: string;
      model: string;
    };
    siliconflow: {
      api_key_set: boolean;
      base_url: string;
      embedding_model: string;
      reranker_model: string;
    };
  };

  db_path: string;
  diff_enabled: boolean;
  diff_watch_paths: string[];
  browser_enabled: boolean;
  browser_history_path: string;

  privacy_enabled: boolean;
  privacy_patterns: string[];
}

// 匹配后端 SaveSettingsRequestDTO
interface SaveSettingsRequest {
  language?: string; // AI Prompt 语言偏好：zh/en

  ai?: {
    provider?: 'default' | 'openai' | 'anthropic' | 'google' | 'zhipu' | string;
    default?: {
      enabled?: boolean;
      api_key?: string;
      base_url?: string;
      model?: string;
    };
    openai?: {
      api_key?: string;
      base_url?: string;
      model?: string;
    };
    anthropic?: {
      api_key?: string;
      base_url?: string;
      model?: string;
    };
    google?: {
      api_key?: string;
      base_url?: string;
      model?: string;
    };
    zhipu?: {
      api_key?: string;
      base_url?: string;
      model?: string;
    };
    siliconflow?: {
      api_key?: string;
      base_url?: string;
      embedding_model?: string;
      reranker_model?: string;
    };
  };

  db_path?: string;
  diff_enabled?: boolean;
  diff_watch_paths?: string[];
  browser_enabled?: boolean;
  browser_history_path?: string;

  privacy_enabled?: boolean;
  privacy_patterns?: string[];
}

export default function SettingsView() {
  const { t } = useTranslation();
  const [settings, setSettings] = useState<SettingsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [pendingChanges, setPendingChanges] = useState<SaveSettingsRequest>({});
  
  // 输入状态
  const [newDefaultApiKey, setNewDefaultApiKey] = useState('');
  const [newOpenAICompatibleApiKey, setNewOpenAICompatibleApiKey] = useState('');
  const [newAnthropicApiKey, setNewAnthropicApiKey] = useState('');
  const [newGoogleApiKey, setNewGoogleApiKey] = useState('');
  const [newZhipuApiKey, setNewZhipuApiKey] = useState('');
  const [newSiliconFlowApiKey, setNewSiliconFlowApiKey] = useState('');
  const [newWatchPath, setNewWatchPath] = useState('');
  const [newPrivacyPattern, setNewPrivacyPattern] = useState('');
  const [privacySample, setPrivacySample] = useState(
    'https://example.com/callback?token=abc123&email=user@example.com#access_token=xyz987'
  );

  useEffect(() => {
    const loadSettings = async () => {
      setLoading(true);
      try {
        const data = await GetSettings();
        setSettings(data);
      } catch (e) {
        console.error('Failed to load settings:', e);
      } finally {
        setLoading(false);
      }
    };
    loadSettings();
  }, []);

  const handleSave = async () => {
    if (Object.keys(pendingChanges).length === 0) return;
    setSaving(true);
    try {
      const resp = await SaveSettings(pendingChanges as any);
      if (resp.restart_required) {
        alert(t('settings.settingsSavedRestart'));
      } else {
        alert(t('settings.settingsSaved'));
      }
      setPendingChanges({});
      const data = await GetSettings();
      setSettings(data);
    } catch (e) {
      alert(`${t('settings.saveFailed')}: ${e}`);
    } finally {
      setSaving(false);
    }
  };

  const updatePending = <K extends keyof SaveSettingsRequest>(key: K, value: SaveSettingsRequest[K]) => {
    setPendingChanges((prev) => ({ ...prev, [key]: value }));
  };

  const updatePendingAI = (patch: NonNullable<SaveSettingsRequest['ai']>) => {
    setPendingChanges((prev) => ({
      ...prev,
      ai: {
        ...(prev.ai || {}),
        ...patch,
      },
    }));
  };

  const updatePendingAIProvider = (
    provider: 'default' | 'openai' | 'anthropic' | 'google' | 'zhipu' | 'siliconflow',
    patch: Record<string, any>
  ) => {
    setPendingChanges((prev) => ({
      ...prev,
      ai: {
        ...(prev.ai || {}),
        [provider]: {
          ...((prev.ai as any)?.[provider] || {}),
          ...patch,
        },
      },
    }));
  };

  // 监控路径管理
  const displayWatchPaths = pendingChanges.diff_watch_paths ?? settings?.diff_watch_paths ?? [];
  const addWatchPath = () => {
    if (!newWatchPath.trim()) return;
    updatePending('diff_watch_paths', [...displayWatchPaths, newWatchPath.trim()]);
    setNewWatchPath('');
  };
  const removeWatchPath = (path: string) => {
    updatePending('diff_watch_paths', displayWatchPaths.filter((p) => p !== path));
  };

  // 隐私规则管理
  const displayPrivacyPatterns = pendingChanges.privacy_patterns ?? settings?.privacy_patterns ?? [];
  const addPrivacyPattern = () => {
    if (!newPrivacyPattern.trim()) return;
    updatePending('privacy_patterns', [...displayPrivacyPatterns, newPrivacyPattern.trim()]);
    setNewPrivacyPattern('');
  };
  const removePrivacyPattern = (pattern: string) => {
    updatePending('privacy_patterns', displayPrivacyPatterns.filter((p) => p !== pattern));
  };

  const hasPendingChanges = Object.keys(pendingChanges).length > 0;
  const privacyEnabled = pendingChanges.privacy_enabled ?? settings?.privacy_enabled ?? false;
  const aiProvider = (pendingChanges.ai?.provider ?? settings?.ai.provider ?? 'default') as string;

  const previewPrivacy = (text: string): string => {
    const input = typeof text === 'string' ? text : '';
    if (!privacyEnabled) return input;

    // 最小口径：默认去掉 URL query/fragment，避免直接暴露 token/email 等
    let out = input.replace(/(https?:\/\/[^\s?#]+)\?[^\s#]*/g, '$1?***');
    out = out.replace(/(https?:\/\/[^\s?#]+)#[^\s]*/g, '$1#***');

    for (const p of displayPrivacyPatterns) {
      const pattern = typeof p === 'string' ? p.trim() : '';
      if (!pattern) continue;
      try {
        const re = new RegExp(pattern, 'gi');
        out = out.replace(re, '***');
      } catch {
        // 忽略无效正则（后端保存时会校验路径等；正则本身允许用户自定义）
      }
    }
    return out;
  };

  if (loading) {
    return <div className="flex items-center justify-center h-64 text-zinc-500">{t('settings.loadingSettings')}</div>;
  }

  if (!settings) {
    return <div className="flex items-center justify-center h-64 text-zinc-500">{t('settings.loadFailed')}</div>;
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6 animate-in fade-in duration-500">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold text-zinc-100">{t('settings.title')}</h2>
        <button
          onClick={handleSave}
          disabled={saving || !hasPendingChanges}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg transition-colors text-sm font-medium ${
            hasPendingChanges
              ? 'bg-indigo-500 hover:bg-indigo-600 text-white'
              : 'bg-zinc-800 text-zinc-500 cursor-not-allowed'
          } disabled:opacity-50`}
        >
          {saving ? <RefreshCw size={16} className="animate-spin" /> : <Save size={16} />}
          {hasPendingChanges ? t('settings.saveChanges') : t('settings.noChanges')}
        </button>
      </div>

      {/* 新手引导（最小可重复口径：强提示 diff watch paths） */}
      {(pendingChanges.diff_enabled ?? settings.diff_enabled) &&
        displayWatchPaths.length === 0 && (
          <Card className="bg-amber-500/10 border-amber-500/20">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-amber-300 flex items-center gap-2">
                <Eye size={14} /> {t('settings.firstTimeSetup')}
              </CardTitle>
            </CardHeader>
            <CardContent className="text-xs text-amber-200/80 space-y-2">
              <div>{t('settings.firstTimeHint')}</div>
              <div>{t('settings.firstTimeAdvice')}</div>
            </CardContent>
          </Card>
        )}

      {/* AI Prompt 语言偏好 */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
            <SettingsIcon size={18} /> {t('settings.aiOutputLanguage')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <div className="text-sm text-zinc-300">{t('settings.aiOutputLanguage')}</div>
              <div className="text-xs text-zinc-500">
                {t('settings.aiOutputLanguageHint')}
              </div>
            </div>
            <div className="flex gap-2 bg-zinc-950 p-1 rounded-lg border border-zinc-800">
              <button
                onClick={() => updatePending('language', 'zh')}
                className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${
                  (pendingChanges.language ?? settings.language) === 'zh'
                    ? 'bg-zinc-800 text-white shadow-sm'
                    : 'text-zinc-500 hover:text-zinc-300'
                }`}
              >
                {t('language.zh')}
              </button>
              <button
                onClick={() => updatePending('language', 'en')}
                className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${
                  (pendingChanges.language ?? settings.language) === 'en'
                    ? 'bg-zinc-800 text-white shadow-sm'
                    : 'text-zinc-500 hover:text-zinc-300'
                }`}
              >
                {t('language.en')}
              </button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* LLM Provider */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
            <Server size={18} /> {t('settings.llmProvider')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="text-xs text-zinc-500">{t('settings.llmProviderHint')}</div>
          <div className="flex items-center justify-between">
            <div className="text-sm text-zinc-300">{t('settings.llmProvider')}</div>
            <select
              value={aiProvider}
              onChange={(e) => updatePendingAI({ provider: e.target.value })}
              className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300"
            >
              <option value="default">{t('settings.providerDefault')}</option>
              <option value="openai">{t('settings.providerOpenAI')}</option>
              <option value="anthropic">{t('settings.providerAnthropic')}</option>
              <option value="google">{t('settings.providerGoogle')}</option>
              <option value="zhipu">{t('settings.providerZhipu')}</option>
            </select>
          </div>
        </CardContent>
      </Card>

      {/* Built-in / Default provider */}
      {(aiProvider === 'default') && settings && (
        <Card className="bg-zinc-900 border-zinc-800">
          <CardHeader>
            <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
              <Server size={18} /> {t('settings.defaultProviderConfig')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.defaultEnabled')}</div>
              <Switch
                checked={(pendingChanges.ai?.default?.enabled ?? settings.ai.default.enabled) as boolean}
                onCheckedChange={(checked: boolean) => updatePendingAIProvider('default', { enabled: checked })}
              />
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.apiKey')}</div>
              <div className="flex items-center gap-2">
                {settings.ai.default.api_key_set ? (
                  <Badge variant="default" className="text-xs"><Key size={10} className="mr-1" /> {t('settings.apiKeySet')}</Badge>
                ) : null}
                <input
                  type="password"
                  placeholder={settings.ai.default.api_key_set ? t('settings.enterApiKeyReplace') : t('settings.enterApiKey')}
                  value={newDefaultApiKey}
                  disabled={Boolean(settings.ai.default.api_key_locked)}
                  onChange={(e) => setNewDefaultApiKey(e.target.value)}
                  onBlur={() => {
                    if (settings.ai.default.api_key_locked) return;
                    const v = newDefaultApiKey.trim();
                    if (v) updatePendingAIProvider('default', { api_key: v });
                  }}
                  className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 w-56 disabled:opacity-60 disabled:cursor-not-allowed"
                />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.baseUrl')}</div>
              <input
                type="text"
                defaultValue={settings.ai.default.base_url}
                disabled={Boolean(settings.ai.default.base_url_locked)}
                onBlur={(e) => {
                  if (settings.ai.default.base_url_locked) return;
                  const v = e.target.value.trim();
                  if (v !== settings.ai.default.base_url) updatePendingAIProvider('default', { base_url: v });
                }}
                className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-80 disabled:opacity-60 disabled:cursor-not-allowed"
              />
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.model')}</div>
              <input
                type="text"
                defaultValue={settings.ai.default.model}
                disabled={Boolean(settings.ai.default.model_locked)}
                onBlur={(e) => {
                  if (settings.ai.default.model_locked) return;
                  const v = e.target.value.trim();
                  if (v !== settings.ai.default.model) updatePendingAIProvider('default', { model: v });
                }}
                className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-56 disabled:opacity-60 disabled:cursor-not-allowed"
              />
            </div>
          </CardContent>
        </Card>
      )}

      {/* OpenAI compatible */}
      {(aiProvider === 'openai') && settings && (
        <Card className="bg-zinc-900 border-zinc-800">
          <CardHeader>
            <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
              <Server size={18} /> {t('settings.openaiProviderConfig')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.apiKey')}</div>
              <div className="flex items-center gap-2">
                {settings.ai.openai.api_key_set ? (
                  <Badge variant="default" className="text-xs"><Key size={10} className="mr-1" /> {t('settings.apiKeySet')}</Badge>
                ) : null}
                <input
                  type="password"
                  placeholder={settings.ai.openai.api_key_set ? t('settings.enterApiKeyReplace') : t('settings.enterApiKey')}
                  value={newOpenAICompatibleApiKey}
                  onChange={(e) => setNewOpenAICompatibleApiKey(e.target.value)}
                  onBlur={() => {
                    const v = newOpenAICompatibleApiKey.trim();
                    if (v) updatePendingAIProvider('openai', { api_key: v });
                  }}
                  className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 w-56"
                />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.baseUrl')}</div>
              <input
                type="text"
                defaultValue={settings.ai.openai.base_url}
                onBlur={(e) => {
                  const v = e.target.value.trim();
                  if (v !== settings.ai.openai.base_url) updatePendingAIProvider('openai', { base_url: v });
                }}
                className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-80"
              />
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.model')}</div>
              <input
                type="text"
                defaultValue={settings.ai.openai.model}
                onBlur={(e) => {
                  const v = e.target.value.trim();
                  if (v !== settings.ai.openai.model) updatePendingAIProvider('openai', { model: v });
                }}
                className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-56"
              />
            </div>
          </CardContent>
        </Card>
      )}

      {/* Anthropic */}
      {(aiProvider === 'anthropic') && settings && (
        <Card className="bg-zinc-900 border-zinc-800">
          <CardHeader>
            <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
              <Server size={18} /> {t('settings.anthropicProviderConfig')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.apiKey')}</div>
              <div className="flex items-center gap-2">
                {settings.ai.anthropic.api_key_set ? (
                  <Badge variant="default" className="text-xs"><Key size={10} className="mr-1" /> {t('settings.apiKeySet')}</Badge>
                ) : null}
                <input
                  type="password"
                  placeholder={settings.ai.anthropic.api_key_set ? t('settings.enterApiKeyReplace') : t('settings.enterApiKey')}
                  value={newAnthropicApiKey}
                  onChange={(e) => setNewAnthropicApiKey(e.target.value)}
                  onBlur={() => {
                    const v = newAnthropicApiKey.trim();
                    if (v) updatePendingAIProvider('anthropic', { api_key: v });
                  }}
                  className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 w-56"
                />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.baseUrl')}</div>
              <input
                type="text"
                defaultValue={settings.ai.anthropic.base_url}
                onBlur={(e) => {
                  const v = e.target.value.trim();
                  if (v !== settings.ai.anthropic.base_url) updatePendingAIProvider('anthropic', { base_url: v });
                }}
                className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-80"
              />
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.model')}</div>
              <input
                type="text"
                defaultValue={settings.ai.anthropic.model}
                onBlur={(e) => {
                  const v = e.target.value.trim();
                  if (v !== settings.ai.anthropic.model) updatePendingAIProvider('anthropic', { model: v });
                }}
                className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-56"
              />
            </div>
          </CardContent>
        </Card>
      )}

      {/* Google */}
      {(aiProvider === 'google') && settings && (
        <Card className="bg-zinc-900 border-zinc-800">
          <CardHeader>
            <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
              <Server size={18} /> {t('settings.googleProviderConfig')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.apiKey')}</div>
              <div className="flex items-center gap-2">
                {settings.ai.google.api_key_set ? (
                  <Badge variant="default" className="text-xs"><Key size={10} className="mr-1" /> {t('settings.apiKeySet')}</Badge>
                ) : null}
                <input
                  type="password"
                  placeholder={settings.ai.google.api_key_set ? t('settings.enterApiKeyReplace') : t('settings.enterApiKey')}
                  value={newGoogleApiKey}
                  onChange={(e) => setNewGoogleApiKey(e.target.value)}
                  onBlur={() => {
                    const v = newGoogleApiKey.trim();
                    if (v) updatePendingAIProvider('google', { api_key: v });
                  }}
                  className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 w-56"
                />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.baseUrl')}</div>
              <input
                type="text"
                defaultValue={settings.ai.google.base_url}
                onBlur={(e) => {
                  const v = e.target.value.trim();
                  if (v !== settings.ai.google.base_url) updatePendingAIProvider('google', { base_url: v });
                }}
                className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-80"
              />
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.model')}</div>
              <input
                type="text"
                defaultValue={settings.ai.google.model}
                onBlur={(e) => {
                  const v = e.target.value.trim();
                  if (v !== settings.ai.google.model) updatePendingAIProvider('google', { model: v });
                }}
                className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-56"
              />
            </div>
          </CardContent>
        </Card>
      )}

      {/* Zhipu */}
      {(aiProvider === 'zhipu') && settings && (
        <Card className="bg-zinc-900 border-zinc-800">
          <CardHeader>
            <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
              <Server size={18} /> {t('settings.zhipuProviderConfig')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.apiKey')}</div>
              <div className="flex items-center gap-2">
                {settings.ai.zhipu.api_key_set ? (
                  <Badge variant="default" className="text-xs"><Key size={10} className="mr-1" /> {t('settings.apiKeySet')}</Badge>
                ) : null}
                <input
                  type="password"
                  placeholder={settings.ai.zhipu.api_key_set ? t('settings.enterApiKeyReplace') : t('settings.enterApiKey')}
                  value={newZhipuApiKey}
                  onChange={(e) => setNewZhipuApiKey(e.target.value)}
                  onBlur={() => {
                    const v = newZhipuApiKey.trim();
                    if (v) updatePendingAIProvider('zhipu', { api_key: v });
                  }}
                  className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 w-56"
                />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.baseUrl')}</div>
              <input
                type="text"
                defaultValue={settings.ai.zhipu.base_url}
                onBlur={(e) => {
                  const v = e.target.value.trim();
                  if (v !== settings.ai.zhipu.base_url) updatePendingAIProvider('zhipu', { base_url: v });
                }}
                className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-80"
              />
            </div>
            <div className="flex items-center justify-between">
              <div className="text-sm text-zinc-300">{t('settings.model')}</div>
              <input
                type="text"
                defaultValue={settings.ai.zhipu.model}
                onBlur={(e) => {
                  const v = e.target.value.trim();
                  if (v !== settings.ai.zhipu.model) updatePendingAIProvider('zhipu', { model: v });
                }}
                className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-56"
              />
            </div>
          </CardContent>
        </Card>
      )}

      {/* SiliconFlow 配置 */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
            <Server size={18} /> {t('settings.siliconflowConfig')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="text-sm text-zinc-300">{t('settings.apiKey')}</div>
            <div className="flex items-center gap-2">
              {settings.ai.siliconflow.api_key_set ? (
                <Badge variant="default" className="text-xs"><Key size={10} className="mr-1" /> {t('settings.apiKeySet')}</Badge>
              ) : null}
              <input
                type="password"
                placeholder={settings.ai.siliconflow.api_key_set ? t('settings.enterApiKeyReplace') : t('settings.enterApiKey')}
                value={newSiliconFlowApiKey}
                onChange={(e) => setNewSiliconFlowApiKey(e.target.value)}
                onBlur={() => {
                  const v = newSiliconFlowApiKey.trim();
                  if (v) updatePendingAIProvider('siliconflow', { api_key: v });
                }}
                className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 w-56"
              />
            </div>
          </div>
          <div className="flex items-center justify-between">
            <div className="text-sm text-zinc-300">{t('settings.baseUrl')}</div>
            <input
              type="text"
              defaultValue={settings.ai.siliconflow.base_url}
              onBlur={(e) => {
                const v = e.target.value.trim();
                if (v !== settings.ai.siliconflow.base_url) updatePendingAIProvider('siliconflow', { base_url: v });
              }}
              className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-64"
            />
          </div>
          <div className="flex items-center justify-between">
            <div className="text-sm text-zinc-300">{t('settings.embeddingModel')}</div>
            <input
              type="text"
              defaultValue={settings.ai.siliconflow.embedding_model}
              onBlur={(e) => {
                const v = e.target.value.trim();
                if (v !== settings.ai.siliconflow.embedding_model) updatePendingAIProvider('siliconflow', { embedding_model: v });
              }}
              className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-48"
            />
          </div>
          <div className="flex items-center justify-between">
            <div className="text-sm text-zinc-300">{t('settings.rerankerModel')}</div>
            <input
              type="text"
              defaultValue={settings.ai.siliconflow.reranker_model}
              onBlur={(e) => {
                const v = e.target.value.trim();
                if (v !== settings.ai.siliconflow.reranker_model) updatePendingAIProvider('siliconflow', { reranker_model: v });
              }}
              className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-48"
            />
          </div>
        </CardContent>
      </Card>

      {/* 数据采集 */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
            <Moon size={18} /> {t('settings.dataCollection')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-zinc-300">{t('settings.diffMonitor')}</div>
              <div className="text-xs text-zinc-500">{t('settings.diffMonitorHint')}</div>
            </div>
            <Switch
              checked={pendingChanges.diff_enabled ?? settings.diff_enabled}
              onCheckedChange={(checked: boolean) => updatePending('diff_enabled', checked)}
            />
          </div>
          
          <div>
            <div className="text-sm text-zinc-300 mb-2">{t('settings.watchPaths')}</div>
            <div className="space-y-1 mb-2 max-h-32 overflow-y-auto">
              {displayWatchPaths.map((path) => (
                <div key={path} className="flex items-center justify-between bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs">
                  <span className="font-mono text-zinc-400 truncate">{path}</span>
                  <button onClick={() => removeWatchPath(path)} className="text-zinc-600 hover:text-rose-400 ml-2"><X size={12} /></button>
                </div>
              ))}
            </div>
            <div className="flex gap-2">
              <input
                type="text"
                placeholder={t('settings.addWatchPath')}
                value={newWatchPath}
                onChange={(e) => setNewWatchPath(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') addWatchPath(); }}
                className="flex-1 bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 font-mono"
              />
              <button onClick={addWatchPath} className="px-2 py-1 bg-zinc-800 hover:bg-zinc-700 rounded text-xs text-zinc-300"><Plus size={12} /></button>
            </div>
          </div>

          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-zinc-300">{t('settings.browserMonitor')}</div>
              <div className="text-xs text-zinc-500">{t('settings.browserMonitorHint')}</div>
            </div>
            <Switch
              checked={pendingChanges.browser_enabled ?? settings.browser_enabled}
              onCheckedChange={(checked: boolean) => updatePending('browser_enabled', checked)}
            />
          </div>
          <div className="flex items-center justify-between">
            <div className="text-sm text-zinc-300">{t('settings.browserHistoryPath')}</div>
            <input
              type="text"
              defaultValue={settings.browser_history_path}
              onBlur={(e) => { if (e.target.value !== settings.browser_history_path) updatePending('browser_history_path', e.target.value); }}
              className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-400 font-mono w-64"
            />
          </div>
        </CardContent>
      </Card>

      {/* 隐私 */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
            <Shield size={18} /> {t('settings.privacy')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-zinc-300">{t('settings.privacyFilter')}</div>
              <div className="text-xs text-zinc-500">{t('settings.privacyFilterHint')}</div>
            </div>
            <Switch
              checked={pendingChanges.privacy_enabled ?? settings.privacy_enabled}
              onCheckedChange={(checked: boolean) => updatePending('privacy_enabled', checked)}
            />
          </div>
          
          <div>
            <div className="text-sm text-zinc-300 mb-2 flex items-center gap-2">
              <Eye size={14} /> {t('settings.privacyRules')}
            </div>
            <div className="space-y-1 mb-2 max-h-32 overflow-y-auto">
              {displayPrivacyPatterns.map((pattern, idx) => (
                <div key={idx} className="flex items-center justify-between bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs">
                  <span className="font-mono text-zinc-400 truncate">{pattern}</span>
                  <button onClick={() => removePrivacyPattern(pattern)} className="text-zinc-600 hover:text-rose-400 ml-2"><X size={12} /></button>
                </div>
              ))}
            </div>
            <div className="flex gap-2">
              <input
                type="text"
                placeholder={t('settings.addRegexRule')}
                value={newPrivacyPattern}
                onChange={(e) => setNewPrivacyPattern(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') addPrivacyPattern(); }}
                className="flex-1 bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 font-mono"
              />
              <button onClick={addPrivacyPattern} className="px-2 py-1 bg-zinc-800 hover:bg-zinc-700 rounded text-xs text-zinc-300"><Plus size={12} /></button>
            </div>
          </div>

          {/* 脱敏预览（P0 验收点：可预览规则效果） */}
          <div className="space-y-2">
            <div className="text-sm text-zinc-300 flex items-center gap-2">
              <Eye size={14} /> {t('settings.privacyPreview')}
            </div>
            <textarea
              value={privacySample}
              onChange={(e) => setPrivacySample(e.target.value)}
              rows={3}
              className="w-full bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-xs text-zinc-300 font-mono"
              placeholder={t('settings.privacyPreviewHint')}
            />
            <div className="text-xs text-zinc-500">{t('settings.privacyPreviewResult')}</div>
            <pre className="w-full bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-xs text-zinc-200 whitespace-pre-wrap break-words">
              {previewPrivacy(privacySample) || t('settings.empty')}
            </pre>
          </div>
        </CardContent>
      </Card>

      {/* 存储 */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
            <Database size={18} /> {t('settings.dataStorage')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="text-sm text-zinc-300">{t('settings.configFile')}</div>
            <span className="text-xs font-mono text-zinc-500 truncate max-w-[250px]">{settings.config_path}</span>
          </div>
          <div className="flex items-center justify-between">
            <div className="text-sm text-zinc-300">{t('settings.databasePath')}</div>
            <span className="text-xs font-mono text-zinc-500 truncate max-w-[250px]">{settings.db_path}</span>
          </div>
        </CardContent>
      </Card>

      {/* 关于 */}
      <Card className="bg-zinc-900 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-base font-medium text-zinc-200 flex items-center gap-2">
            <SettingsIcon size={18} /> {t('settings.about')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-sm text-zinc-400">
            {t('app.name')} {t('app.version')}
            <br />
            <span className="text-zinc-600">{t('settings.buildDate')}: 2024-12-14</span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
