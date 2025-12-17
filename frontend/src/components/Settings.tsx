import React, { useEffect, useState } from 'react';
import { GetSettings, SaveSettings, InstallContextMenu, UninstallContextMenu, GetContextMenuStatus, SelectBinaryDialog, DetectBinaries } from '../wailsjs/go/main/App';
import { main } from '../wailsjs/go/models';
import { Loader2, Save, Monitor, Check, FolderOpen, Search, Sliders, Cpu, Film, Layers, Moon, Sun, Laptop } from 'lucide-react';
import { cn } from '../lib/utils';

interface SettingsViewProps {
    isInstalled: boolean;
    onStatusChange: (status: boolean) => void;
    theme: 'dark' | 'light' | 'system';
    onThemeChange: (theme: 'dark' | 'light' | 'system') => void;
}

export function SettingsView({ isInstalled, onStatusChange, theme, onThemeChange }: SettingsViewProps) {
    const [settings, setSettings] = useState<main.Settings | null>(null);
    const [saving, setSaving] = useState(false);
    const [saved, setSaved] = useState(false);
    const [loading, setLoading] = useState(true);
    const [togglingMenu, setTogglingMenu] = useState(false);

    useEffect(() => {
        GetSettings().then((s) => {
            setSettings(s);
            setLoading(false);
        });
    }, []);

    useEffect(() => {
        if (saved) {
            const timer = setTimeout(() => setSaved(false), 2000);
            return () => clearTimeout(timer);
        }
    }, [saved]);

    const handleToggleMenu = async () => {
        setTogglingMenu(true);
        try {
            if (isInstalled) {
                await UninstallContextMenu();
            } else {
                await InstallContextMenu();
            }

            // Poll for status change
            const targetStatus = !isInstalled;
            const start = Date.now();
            const interval = setInterval(async () => {
                const status = await GetContextMenuStatus();
                if (status === targetStatus) {
                    onStatusChange(status);
                    setTogglingMenu(false);
                    clearInterval(interval);
                }
                if (Date.now() - start > 15000) { // 15s timeout
                     setTogglingMenu(false);
                     clearInterval(interval);
                     onStatusChange(await GetContextMenuStatus());
                }
            }, 1000);

        } catch (e) {
            console.error(e);
            setTogglingMenu(false);
        }
    };

    const handleSave = async () => {
        if (!settings) return;
        setSaving(true);
        try {
            await SaveSettings(settings);
            setSaved(true);
        } finally {
            setSaving(false);
        }
    };

    const handleBrowse = async (field: 'magickBinary' | 'ffmpegBinary') => {
        const path = await SelectBinaryDialog();
        if (path && settings) {
            setSettings({ ...settings, [field]: path });
        }
    };

    const handleDetect = async () => {
        if (!settings) return;
        try {
            const results = await DetectBinaries();
            const newSettings = { ...settings };
            let updated = false;

            if (results['ffmpeg']) {
                newSettings.ffmpegBinary = results['ffmpeg'];
                updated = true;
            }
            if (results['magick']) {
                newSettings.magickBinary = results['magick'];
                updated = true;
            }

            if (updated) {
                setSettings(newSettings);
            }
        } catch (error) {
            console.error("Error detecting binaries:", error);
        }
    };

    if (loading || !settings) {
        return <div className="flex justify-center p-12"><Loader2 className="animate-spin text-slate-500" /></div>;
    }

    return (
        <div className="max-w-2xl mx-auto space-y-8 pb-12">
            <div>
                <h2 className="text-xl font-bold text-slate-900 dark:text-slate-100">Settings</h2>
                <p className="text-slate-500 dark:text-slate-400 mt-1">Configure conversion parameters and system integration.</p>
            </div>

            <div className="space-y-6">
                {/* Integration Card */}
                <div className="bg-white dark:bg-slate-800/40 rounded-xl p-6 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-colors shadow-sm dark:shadow-none">
                     <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 mb-4 flex items-center gap-2">
                        <Monitor className="h-4 w-4 text-indigo-600 dark:text-indigo-400" />
                        Windows Integration
                     </h3>
                     <div className="flex items-center justify-between bg-slate-100 dark:bg-slate-900/50 p-4 rounded-lg border border-slate-200 dark:border-slate-800/50">
                        <div>
                            <p className="text-sm font-medium text-slate-800 dark:text-slate-200">Context Menu</p>
                            <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                                {isInstalled ? "Currently installed. Right-click files to convert." : "Not installed. Install to add to right-click menu."}
                            </p>
                        </div>
                        <button
                            onClick={handleToggleMenu}
                            disabled={togglingMenu}
                            className={cn(
                                "px-4 py-2 text-xs font-semibold rounded-lg transition-all border flex items-center justify-center min-w-[90px] shadow-sm",
                                isInstalled
                                    ? "border-red-500/20 text-red-500 dark:text-red-400 hover:bg-red-500/10 hover:border-red-500/30"
                                    : "border-indigo-500/20 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-500/10 hover:border-indigo-500/30",
                                togglingMenu && "opacity-50 cursor-wait"
                            )}
                        >
                            {togglingMenu ? <Loader2 className="animate-spin h-4 w-4" /> : (isInstalled ? "Uninstall" : "Install")}
                        </button>
                     </div>

                     <div className="flex items-center justify-between bg-slate-100 dark:bg-slate-900/50 p-4 rounded-lg border border-slate-200 dark:border-slate-800/50">
                        <div>
                            <p className="text-sm font-medium text-slate-800 dark:text-slate-200">Theme</p>
                            <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                                Customize the application appearance.
                            </p>
                        </div>
                        <div className="flex bg-slate-200 dark:bg-slate-800 rounded-lg p-1 gap-1">
                            <button
                                onClick={() => onThemeChange('light')}
                                className={cn(
                                    "p-1.5 rounded-md transition-all",
                                    theme === 'light' ? "bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 shadow-sm" : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200"
                                )}
                                title="Light"
                            >
                                <Sun className="w-4 h-4" />
                            </button>
                            <button
                                onClick={() => onThemeChange('dark')}
                                className={cn(
                                    "p-1.5 rounded-md transition-all",
                                    theme === 'dark' ? "bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 shadow-sm" : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200"
                                )}
                                title="Dark"
                            >
                                <Moon className="w-4 h-4" />
                            </button>
                            <button
                                onClick={() => onThemeChange('system')}
                                className={cn(
                                    "p-1.5 rounded-md transition-all",
                                    theme === 'system' ? "bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 shadow-sm" : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200"
                                )}
                                title="System"
                            >
                                <Laptop className="w-4 h-4" />
                            </button>
                        </div>
                     </div>
                </div>

                {/* Video Options Card */}
                <div className="bg-white dark:bg-slate-800/40 rounded-xl p-6 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-colors space-y-6 shadow-sm dark:shadow-none">
                    <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 flex items-center gap-2">
                        <Film className="h-4 w-4 text-purple-600 dark:text-purple-400" />
                        Video Options
                    </h3>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="space-y-2">
                            <label className="text-xs font-medium text-slate-500 dark:text-slate-400 flex items-center gap-1.5">
                                <Cpu className="w-3 h-3" /> Hardware Accelerator
                            </label>
                            <select
                                className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow"
                                value={settings.hardwareAccelerator}
                                onChange={(e) => setSettings({ ...settings, hardwareAccelerator: e.target.value })}
                            >
                                <option value="none">None (CPU - libx264)</option>
                                <option value="nvidia">NVIDIA (CUDA/NVENC)</option>
                                <option value="amd">AMD (AMF)</option>
                            </select>
                        </div>

                        <div className="space-y-2">
                            <label className="text-xs font-medium text-slate-500 dark:text-slate-400 flex items-center gap-1.5">
                                <Layers className="w-3 h-3" /> Quality Preset
                            </label>
                            <select
                                className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow"
                                value={settings.videoQuality || "high"}
                                onChange={(e) => setSettings({ ...settings, videoQuality: e.target.value })}
                            >
                                <option value="high">High (5 Mbps, Quality)</option>
                                <option value="medium">Medium (2.5 Mbps, Balanced)</option>
                                <option value="low">Low (1 Mbps, Fast)</option>
                            </select>
                        </div>

                        <div className="space-y-2">
                            <label className="text-xs font-medium text-slate-500 dark:text-slate-400">Concurrent Jobs</label>
                            <input
                                type="number"
                                min="1"
                                className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow"
                                value={settings.maxFfmpegWorkers || 1}
                                onChange={(e) => setSettings({ ...settings, maxFfmpegWorkers: parseInt(e.target.value) || 1 })}
                            />
                        </div>

                        <div className="space-y-2">
                            <label className="text-xs font-medium text-slate-500 dark:text-slate-400">Max Resolution (Size)</label>
                             <input
                                type="number"
                                className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow"
                                value={settings.maxSize}
                                onChange={(e) => setSettings({ ...settings, maxSize: parseInt(e.target.value) || 0 })}
                                placeholder="0 for original"
                            />
                        </div>
                    </div>

                    <div className="space-y-2 pt-2 border-t border-slate-200 dark:border-slate-700/50">
                        <label className="text-xs font-medium text-slate-500 dark:text-slate-400">Custom FFmpeg Arguments</label>
                         <input
                            type="text"
                            className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 font-mono text-xs transition-shadow"
                            value={settings.ffmpegCustomArgs}
                            onChange={(e) => setSettings({ ...settings, ffmpegCustomArgs: e.target.value })}
                            placeholder="-crf 23 -preset slow"
                        />
                    </div>
                </div>

                {/* External Tools Card */}
                <div className="bg-white dark:bg-slate-800/40 rounded-xl p-6 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-colors shadow-sm dark:shadow-none">
                    <div className="flex items-center justify-between mb-6">
                        <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 flex items-center gap-2">
                            <Sliders className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
                            External Tools
                        </h3>
                        <button
                            onClick={handleDetect}
                            className="text-xs flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-lg border border-slate-300 dark:border-slate-700/50 text-slate-600 dark:text-slate-300 transition-colors shadow-sm"
                        >
                            <Search className="w-3 h-3" /> Auto-Detect
                        </button>
                    </div>
                    <div className="space-y-5">
                        <div className="space-y-2">
                            <label className="text-xs font-medium text-slate-500 dark:text-slate-400">FFmpeg Binary Path</label>
                            <div className="flex gap-2">
                                <input
                                    type="text"
                                    className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-xs px-3 py-2.5 font-mono truncate transition-shadow"
                                    value={settings.ffmpegBinary}
                                    onChange={(e) => setSettings({ ...settings, ffmpegBinary: e.target.value })}
                                />
                                <button
                                    onClick={() => handleBrowse('ffmpegBinary')}
                                    className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-600 transition-colors shadow-sm"
                                    title="Browse..."
                                >
                                    <FolderOpen className="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                        <div className="space-y-2">
                            <label className="text-xs font-medium text-slate-500 dark:text-slate-400">Magick Binary Path</label>
                            <div className="flex gap-2">
                                <input
                                    type="text"
                                    className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-xs px-3 py-2.5 font-mono truncate transition-shadow"
                                    value={settings.magickBinary}
                                    onChange={(e) => setSettings({ ...settings, magickBinary: e.target.value })}
                                />
                                <button
                                    onClick={() => handleBrowse('magickBinary')}
                                    className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-600 transition-colors shadow-sm"
                                    title="Browse..."
                                >
                                    <FolderOpen className="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Paths Card */}
                <div className="bg-white dark:bg-slate-800/40 rounded-xl p-6 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-colors shadow-sm dark:shadow-none">
                     <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 mb-6 flex items-center gap-2">
                        <FolderOpen className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                        Paths & Filters
                     </h3>
                     <div className="space-y-5">
                        <div className="space-y-2">
                            <label className="text-xs font-medium text-slate-500 dark:text-slate-400">Default Destination Directory</label>
                            <input
                                type="text"
                                className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-xs px-3 py-2.5 font-mono transition-shadow"
                                value={settings.defaultDestDir}
                                onChange={(e) => setSettings({ ...settings, defaultDestDir: e.target.value })}
                            />
                        </div>
                        <div className="space-y-2">
                            <label className="text-xs font-medium text-slate-500 dark:text-slate-400">File Collision Behavior</label>
                            <select
                                className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow"
                                value={settings.collisionOption || "rename"}
                                onChange={(e) => setSettings({ ...settings, collisionOption: e.target.value })}
                            >
                                <option value="rename">Rename</option>
                                <option value="overwrite">Overwrite</option>
                                <option value="error">Error</option>
                            </select>
                        </div>
                        <div className="space-y-2">
                            <label className="text-xs font-medium text-slate-500 dark:text-slate-400">Exclude Patterns (comma separated)</label>
                             <input
                                type="text"
                                className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow"
                                value={settings.excludePatterns?.join(', ')}
                                onChange={(e) => setSettings({ ...settings, excludePatterns: e.target.value.split(',').map(s => s.trim()) })}
                                placeholder="e.g. \Pictures\, \DCIM\"
                            />
                        </div>
                     </div>
                </div>

                <div className="flex justify-end pt-4 sticky bottom-0 bg-slate-50/80 dark:bg-slate-950/80 backdrop-blur-sm p-4 -mx-4 -mb-4 border-t border-slate-200 dark:border-white/5">
                    <button
                        onClick={handleSave}
                        disabled={saving || saved}
                        className={cn(
                            "inline-flex items-center px-6 py-2.5 border border-transparent text-sm font-semibold rounded-lg shadow-lg text-white focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-all",
                            saving && "opacity-75 cursor-wait",
                            saved ? "bg-emerald-600 hover:bg-emerald-700 shadow-emerald-500/20" : "bg-indigo-600 hover:bg-indigo-700 shadow-indigo-500/20"
                        )}
                    >
                        {saving ? (
                            <Loader2 className="animate-spin -ml-1 mr-2 h-4 w-4" />
                        ) : saved ? (
                            <Check className="-ml-1 mr-2 h-4 w-4" />
                        ) : (
                            <Save className="-ml-1 mr-2 h-4 w-4" />
                        )}
                        {saved ? "Settings Saved" : "Save Changes"}
                    </button>
                </div>
            </div>
        </div>
    );
}
