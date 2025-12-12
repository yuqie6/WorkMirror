import React, { useEffect, useMemo, useState } from 'react';
// @ts-ignore
import { GetSettings, SaveSettings } from "../../../wailsjs/go/main/App";

export interface SettingsDTO {
    config_path: string;

    deepseek_api_key_set: boolean;
    deepseek_base_url: string;
    deepseek_model: string;

    siliconflow_api_key_set: boolean;
    siliconflow_base_url: string;
    siliconflow_embedding_model: string;
    siliconflow_reranker_model: string;

    db_path: string;
    diff_watch_paths: string[];
    browser_history_path: string;
}

export interface SaveSettingsRequestDTO {
    deepseek_api_key?: string;
    deepseek_base_url?: string;
    deepseek_model?: string;

    siliconflow_api_key?: string;
    siliconflow_base_url?: string;
    siliconflow_embedding_model?: string;
    siliconflow_reranker_model?: string;

    db_path?: string;
    diff_watch_paths?: string[];
    browser_history_path?: string;
}

const normalizeLines = (value: string): string[] => {
    return value
        .split('\n')
        .map(line => line.trim())
        .filter(Boolean);
};

const SettingsView: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);

    const [settings, setSettings] = useState<SettingsDTO | null>(null);

    const [deepSeekKey, setDeepSeekKey] = useState('');
    const [deepSeekBaseURL, setDeepSeekBaseURL] = useState('');
    const [deepSeekModel, setDeepSeekModel] = useState('');

    const [siliconFlowKey, setSiliconFlowKey] = useState('');
    const [siliconFlowBaseURL, setSiliconFlowBaseURL] = useState('');
    const [siliconFlowEmbeddingModel, setSiliconFlowEmbeddingModel] = useState('');
    const [siliconFlowRerankerModel, setSiliconFlowRerankerModel] = useState('');

    const [dbPath, setDBPath] = useState('');
    const [diffWatchPathsText, setDiffWatchPathsText] = useState('');
    const [browserHistoryPath, setBrowserHistoryPath] = useState('');

    const load = async () => {
        setLoading(true);
        setError(null);
        setSuccess(null);
        try {
            const data: SettingsDTO = await GetSettings();
            setSettings(data);
            setDeepSeekBaseURL(data.deepseek_base_url || '');
            setDeepSeekModel(data.deepseek_model || '');
            setSiliconFlowBaseURL(data.siliconflow_base_url || '');
            setSiliconFlowEmbeddingModel(data.siliconflow_embedding_model || '');
            setSiliconFlowRerankerModel(data.siliconflow_reranker_model || '');
            setDBPath(data.db_path || '');
            setDiffWatchPathsText((data.diff_watch_paths || []).join('\n'));
            setBrowserHistoryPath(data.browser_history_path || '');
            setDeepSeekKey('');
            setSiliconFlowKey('');
        } catch (e: any) {
            setError(e?.message || '加载设置失败');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        void load();
    }, []);

    const previewWatchPaths = useMemo(() => normalizeLines(diffWatchPathsText), [diffWatchPathsText]);

    const save = async () => {
        setSaving(true);
        setError(null);
        setSuccess(null);
        try {
            const req: SaveSettingsRequestDTO = {
                deepseek_api_key: deepSeekKey.trim() || undefined,
                deepseek_base_url: deepSeekBaseURL.trim() || undefined,
                deepseek_model: deepSeekModel.trim() || undefined,

                siliconflow_api_key: siliconFlowKey.trim() || undefined,
                siliconflow_base_url: siliconFlowBaseURL.trim() || undefined,
                siliconflow_embedding_model: siliconFlowEmbeddingModel.trim() || undefined,
                siliconflow_reranker_model: siliconFlowRerankerModel.trim() || undefined,

                db_path: dbPath.trim() || undefined,
                diff_watch_paths: previewWatchPaths,
                browser_history_path: browserHistoryPath.trim() || undefined,
            };
            await SaveSettings(req);
            setSuccess('保存成功，已热重载配置');
            await load();
        } catch (e: any) {
            setError(e?.message || '保存失败');
        } finally {
            setSaving(false);
        }
    };

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[50vh] gap-6 animate-fade-in">
                <div className="w-12 h-12 border-2 border-gray-200 border-t-accent-gold rounded-full animate-spin"></div>
                <p className="text-gray-400 text-sm">加载中...</p>
            </div>
        );
    }

    return (
        <div className="space-y-8 animate-slide-up pb-12">
            <header className="space-y-2">
                <h1 className="text-3xl font-bold text-gray-900">设置</h1>
                <p className="text-gray-500 text-sm">配置 API、存储与采集目录（配置写入本地文件）。</p>
            </header>

            {(error || success) && (
                <div className={`card ${error ? 'bg-red-50 border border-red-200' : 'bg-emerald-50 border border-emerald-200'}`}>
                    <div className={`text-sm ${error ? 'text-red-700' : 'text-emerald-700'}`}>
                        {error || success}
                    </div>
                </div>
            )}

            <div className="card">
                <div className="flex items-start justify-between gap-6">
                    <div>
                        <h3 className="text-sm font-semibold text-gray-900 mb-1">配置文件</h3>
                        <p className="text-xs text-gray-500 break-all">{settings?.config_path || '未知'}</p>
                    </div>
                    <button className="pill hover:pill-active transition-colors" onClick={() => void load()}>
                        刷新
                    </button>
                </div>
            </div>

            <div className="grid grid-cols-12 gap-6">
                <div className="col-span-12">
                    <div className="card">
                        <h3 className="text-sm font-semibold text-gray-900 mb-4">AI（DeepSeek）</h3>
                        <div className="grid grid-cols-12 gap-4">
                            <div className="col-span-12">
                                <label className="text-xs text-gray-500">API Key（{settings?.deepseek_api_key_set ? '已设置' : '未设置'}）</label>
                                <input
                                    className="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm outline-none focus:border-amber-300"
                                    value={deepSeekKey}
                                    onChange={(e) => setDeepSeekKey(e.target.value)}
                                    placeholder="留空表示不修改"
                                    type="password"
                                />
                            </div>
                            <div className="col-span-8">
                                <label className="text-xs text-gray-500">Base URL</label>
                                <input
                                    className="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm outline-none focus:border-amber-300"
                                    value={deepSeekBaseURL}
                                    onChange={(e) => setDeepSeekBaseURL(e.target.value)}
                                    placeholder="https://api.deepseek.com"
                                />
                            </div>
                            <div className="col-span-4">
                                <label className="text-xs text-gray-500">Model</label>
                                <input
                                    className="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm outline-none focus:border-amber-300"
                                    value={deepSeekModel}
                                    onChange={(e) => setDeepSeekModel(e.target.value)}
                                    placeholder="deepseek-chat"
                                />
                            </div>
                        </div>
                    </div>
                </div>

                <div className="col-span-12">
                    <div className="card">
                        <h3 className="text-sm font-semibold text-gray-900 mb-4">AI（SiliconFlow）</h3>
                        <div className="grid grid-cols-12 gap-4">
                            <div className="col-span-12">
                                <label className="text-xs text-gray-500">API Key（{settings?.siliconflow_api_key_set ? '已设置' : '未设置'}）</label>
                                <input
                                    className="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm outline-none focus:border-amber-300"
                                    value={siliconFlowKey}
                                    onChange={(e) => setSiliconFlowKey(e.target.value)}
                                    placeholder="留空表示不修改"
                                    type="password"
                                />
                            </div>
                            <div className="col-span-12">
                                <label className="text-xs text-gray-500">Base URL</label>
                                <input
                                    className="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm outline-none focus:border-amber-300"
                                    value={siliconFlowBaseURL}
                                    onChange={(e) => setSiliconFlowBaseURL(e.target.value)}
                                    placeholder="https://api.siliconflow.cn/v1"
                                />
                            </div>
                            <div className="col-span-6">
                                <label className="text-xs text-gray-500">Embedding Model</label>
                                <input
                                    className="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm outline-none focus:border-amber-300"
                                    value={siliconFlowEmbeddingModel}
                                    onChange={(e) => setSiliconFlowEmbeddingModel(e.target.value)}
                                    placeholder="BAAI/bge-large-zh-v1.5"
                                />
                            </div>
                            <div className="col-span-6">
                                <label className="text-xs text-gray-500">Reranker Model</label>
                                <input
                                    className="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm outline-none focus:border-amber-300"
                                    value={siliconFlowRerankerModel}
                                    onChange={(e) => setSiliconFlowRerankerModel(e.target.value)}
                                    placeholder="BAAI/bge-reranker-v2-m3"
                                />
                            </div>
                        </div>
                    </div>
                </div>

                <div className="col-span-12">
                    <div className="card">
                        <h3 className="text-sm font-semibold text-gray-900 mb-4">存储与目录</h3>
                        <div className="grid grid-cols-12 gap-4">
                            <div className="col-span-12">
                                <label className="text-xs text-gray-500">SQLite DB 路径</label>
                                <input
                                    className="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm outline-none focus:border-amber-300"
                                    value={dbPath}
                                    onChange={(e) => setDBPath(e.target.value)}
                                    placeholder="./data/mirror.db"
                                />
                            </div>
                            <div className="col-span-12">
                                <label className="text-xs text-gray-500">Diff 采集目录（每行一个路径）</label>
                                <textarea
                                    className="mt-1 w-full min-h-[120px] rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm outline-none focus:border-amber-300"
                                    value={diffWatchPathsText}
                                    onChange={(e) => setDiffWatchPathsText(e.target.value)}
                                    placeholder="/path/to/project\n/another/path"
                                />
                                <div className="mt-2 flex flex-wrap gap-2">
                                    {previewWatchPaths.slice(0, 6).map((p) => (
                                        <span key={p} className="pill">{p}</span>
                                    ))}
                                    {previewWatchPaths.length > 6 && <span className="pill">+{previewWatchPaths.length - 6}</span>}
                                </div>
                            </div>
                            <div className="col-span-12">
                                <label className="text-xs text-gray-500">浏览器 History 路径（可选）</label>
                                <input
                                    className="mt-1 w-full rounded-xl border border-gray-200 bg-white px-4 py-2 text-sm outline-none focus:border-amber-300"
                                    value={browserHistoryPath}
                                    onChange={(e) => setBrowserHistoryPath(e.target.value)}
                                    placeholder="Chrome History 文件路径"
                                />
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <div className="flex items-center gap-3">
                <button className="btn-gold disabled:opacity-60" onClick={() => void save()} disabled={saving}>
                    {saving ? '保存中...' : '保存设置'}
                </button>
                <span className="text-xs text-gray-400">API Key 留空不会覆盖；保存会触发热重载。</span>
            </div>
        </div>
    );
};

export default SettingsView;
