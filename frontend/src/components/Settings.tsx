import React, { useEffect, useState } from 'react';
import { GetSettings, SaveSettings, InstallContextMenu, UninstallContextMenu, GetContextMenuStatus, SelectBinaryDialog, DetectBinaries } from '../wailsjs/go/main/App';
import { main } from '../wailsjs/go/models';
import { Loader2, Save, Monitor, Check, FolderOpen, Search } from 'lucide-react';
import { cn } from '../lib/utils';

interface SettingsViewProps {
    isInstalled: boolean;
    onStatusChange: (status: boolean) => void;
}

export function SettingsView({ isInstalled, onStatusChange }: SettingsViewProps) {
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
        <div className="max-w-2xl mx-auto space-y-8">
            <div>
                <h2 className="text-xl font-bold text-slate-100">Settings</h2>
                <p className="text-slate-400 mt-1">Configure conversion parameters and paths.</p>
            </div>

            <div className="space-y-6">
                <div className="pb-6 border-b border-slate-700/50">
                     <h3 className="text-sm font-semibold text-slate-200 mb-3 flex items-center gap-2">
                        <Monitor className="h-4 w-4" />
                        Windows Integration
                     </h3>
                     <div className="flex items-center justify-between bg-slate-800/50 p-4 rounded-lg border border-slate-700">
                        <div>
                            <p className="text-sm font-medium text-slate-200">Context Menu</p>
                            <p className="text-xs text-slate-400 mt-1">
                                {isInstalled ? "Currently installed. Right-click files to convert." : "Not installed. Install to add to right-click menu."}
                            </p>
                        </div>
                        <button
                            onClick={handleToggleMenu}
                            disabled={togglingMenu}
                            className={cn(
                                "px-3 py-1.5 text-xs font-medium rounded-md transition-colors border flex items-center justify-center min-w-[80px]",
                                isInstalled
                                    ? "border-red-500/30 text-red-400 hover:bg-red-500/10"
                                    : "border-blue-500/30 text-blue-400 hover:bg-blue-500/10",
                                togglingMenu && "opacity-50 cursor-wait"
                            )}
                        >
                            {togglingMenu ? <Loader2 className="animate-spin h-4 w-4" /> : (isInstalled ? "Uninstall" : "Install")}
                        </button>
                     </div>
                </div>

                <div className="space-y-4">
                    <label className="block">
                        <span className="text-sm font-medium text-slate-300">Hardware Accelerator</span>
                        <select
                            className="mt-1.5 block w-full rounded-md bg-slate-800 border-slate-700 text-slate-200 focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2"
                            value={settings.hardwareAccelerator}
                            onChange={(e) => setSettings({ ...settings, hardwareAccelerator: e.target.value })}
                        >
                            <option value="none">None (CPU - libx264)</option>
                            <option value="nvidia">NVIDIA (CUDA/NVENC)</option>
                            <option value="amd">AMD (AMF)</option>
                        </select>
                    </label>

                    <label className="block">
                        <span className="text-sm font-medium text-slate-300">Video Quality Preset</span>
                        <select
                            className="mt-1.5 block w-full rounded-md bg-slate-800 border-slate-700 text-slate-200 focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2"
                            value={settings.videoQuality || "high"}
                            onChange={(e) => setSettings({ ...settings, videoQuality: e.target.value })}
                        >
                            <option value="high">High (5 Mbps, Slow/Quality)</option>
                            <option value="medium">Medium (2.5 Mbps, Balanced)</option>
                            <option value="low">Low (1 Mbps, Fast)</option>
                        </select>
                    </label>

                    <label className="block">
                        <span className="text-sm font-medium text-slate-300">Max Resolution (Size)</span>
                         <input
                            type="number"
                            className="mt-1.5 block w-full rounded-md bg-slate-800 border-slate-700 text-slate-200 focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2"
                            value={settings.maxSize}
                            onChange={(e) => setSettings({ ...settings, maxSize: parseInt(e.target.value) || 0 })}
                        />
                         <span className="text-xs text-slate-500">Maximum width/height in pixels. Aspect ratio preserved.</span>
                    </label>

                    <label className="block">
                        <span className="text-sm font-medium text-slate-300">Custom FFmpeg Args</span>
                         <input
                            type="text"
                            className="mt-1.5 block w-full rounded-md bg-slate-800 border-slate-700 text-slate-200 focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2"
                            value={settings.ffmpegCustomArgs}
                            onChange={(e) => setSettings({ ...settings, ffmpegCustomArgs: e.target.value })}
                            placeholder="-crf 23 -preset slow"
                        />
                    </label>
                </div>

                <div className="pt-4 border-t border-slate-700/50">
                    <div className="flex items-center justify-between mb-3">
                        <h3 className="text-sm font-semibold text-slate-200">External Tools</h3>
                        <button
                            onClick={handleDetect}
                            className="text-xs flex items-center gap-1.5 px-2 py-1 bg-slate-800 hover:bg-slate-700 rounded border border-slate-700 text-slate-300 transition-colors"
                        >
                            <Search className="w-3 h-3" /> Auto-Detect
                        </button>
                    </div>
                    <div className="space-y-4 mb-6">
                        <label className="block">
                            <span className="text-sm font-medium text-slate-300">FFmpeg Binary</span>
                            <div className="mt-1.5 flex gap-2">
                                <input
                                    type="text"
                                    className="block w-full rounded-md bg-slate-800 border-slate-700 text-slate-200 focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2"
                                    value={settings.ffmpegBinary}
                                    onChange={(e) => setSettings({ ...settings, ffmpegBinary: e.target.value })}
                                />
                                <button
                                    onClick={() => handleBrowse('ffmpegBinary')}
                                    className="px-3 py-2 bg-slate-700 hover:bg-slate-600 rounded-md text-slate-200 border border-slate-600 transition-colors"
                                    title="Browse..."
                                >
                                    <FolderOpen className="w-4 h-4" />
                                </button>
                            </div>
                        </label>
                        <label className="block">
                            <span className="text-sm font-medium text-slate-300">Magick Binary</span>
                            <div className="mt-1.5 flex gap-2">
                                <input
                                    type="text"
                                    className="block w-full rounded-md bg-slate-800 border-slate-700 text-slate-200 focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2"
                                    value={settings.magickBinary}
                                    onChange={(e) => setSettings({ ...settings, magickBinary: e.target.value })}
                                />
                                <button
                                    onClick={() => handleBrowse('magickBinary')}
                                    className="px-3 py-2 bg-slate-700 hover:bg-slate-600 rounded-md text-slate-200 border border-slate-600 transition-colors"
                                    title="Browse..."
                                >
                                    <FolderOpen className="w-4 h-4" />
                                </button>
                            </div>
                        </label>
                    </div>

                     <h3 className="text-sm font-semibold text-slate-200 mb-3">Paths</h3>
                     <div className="space-y-4">
                        <label className="block">
                            <span className="text-sm font-medium text-slate-300">Default Destination</span>
                            <input
                                type="text"
                                className="mt-1.5 block w-full rounded-md bg-slate-800 border-slate-700 text-slate-200 focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2"
                                value={settings.defaultDestDir}
                                onChange={(e) => setSettings({ ...settings, defaultDestDir: e.target.value })}
                            />
                        </label>
                        <label className="block">
                            <span className="text-sm font-medium text-slate-300">Exclude Patterns (Comma separated)</span>
                             <input
                                type="text"
                                className="mt-1.5 block w-full rounded-md bg-slate-800 border-slate-700 text-slate-200 focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2"
                                value={settings.excludePatterns?.join(', ')}
                                onChange={(e) => setSettings({ ...settings, excludePatterns: e.target.value.split(',').map(s => s.trim()) })}
                            />
                        </label>
                     </div>
                </div>

                <div className="flex justify-end pt-4">
                    <button
                        onClick={handleSave}
                        disabled={saving || saved}
                        className={cn(
                            "inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-all",
                            saving && "opacity-75 cursor-wait",
                            saved ? "bg-emerald-600 hover:bg-emerald-700" : "bg-indigo-600 hover:bg-indigo-700"
                        )}
                    >
                        {saving ? (
                            <Loader2 className="animate-spin -ml-1 mr-2 h-4 w-4" />
                        ) : saved ? (
                            <Check className="-ml-1 mr-2 h-4 w-4" />
                        ) : (
                            <Save className="-ml-1 mr-2 h-4 w-4" />
                        )}
                        {saved ? "Saved!" : "Save Changes"}
                    </button>
                </div>
            </div>
        </div>
    );
}
