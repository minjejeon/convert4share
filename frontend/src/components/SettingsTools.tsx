import React, { useState } from 'react';
import { Sliders, Search, FolderOpen, Download, Loader2 } from 'lucide-react';
import { main } from '../wailsjs/go/models';
import { SelectBinaryDialog, InstallTool, DetectBinaries } from '../wailsjs/go/main/App';

interface SettingsToolsProps {
    settings: main.Settings;
    onChange: (settings: main.Settings) => void;
}

export function SettingsTools({ settings, onChange }: SettingsToolsProps) {
    const [installing, setInstalling] = useState<string | null>(null);

    const handleBrowse = async (field: 'magickBinary' | 'ffmpegBinary') => {
        const path = await SelectBinaryDialog();
        if (path) {
            onChange({ ...settings, [field]: path });
        }
    };

    const handleDetect = async () => {
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
                onChange(newSettings);
            }
        } catch (error) {
            console.error("Error detecting binaries:", error);
        }
    };

    const handleInstall = async (tool: string) => {
        setInstalling(tool);
        try {
            await InstallTool(tool);
            await handleDetect();
            // Note: In the original code, setSaved(true) was called here to indicate save.
            // The parent component handles saving via the Save button, but maybe we want to trigger a save or just update the state.
            // The user still needs to click Save.
            // However, InstallTool in backend calls SaveSettings internally!
            // So we might need to re-fetch settings?
            // Or just update local state to match what's on disk.
            // Since InstallTool saves, we should probably consider the settings 'saved'.
            // But we only update local state here.
            // If we update local state, and the user clicks Save, it saves again. That's fine.
        } catch (e) {
            console.error(e);
            const message = e instanceof Error ? e.message : String(e);
            alert("Installation failed. Please try installing manually via 'winget install " + (tool === 'ffmpeg' ? 'Gyan.FFmpeg' : 'ImageMagick.ImageMagick') + "' in PowerShell.\n\nError: " + message);
        } finally {
            setInstalling(null);
        }
    };

    return (
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
                    <label htmlFor="tools-ffmpeg-path" className="text-xs font-medium text-slate-500 dark:text-slate-400">FFmpeg Binary Path</label>
                    <div className="flex gap-2">
                        <input
                            id="tools-ffmpeg-path"
                            type="text"
                            className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-xs px-3 py-2.5 font-mono truncate transition-shadow"
                            value={settings.ffmpegBinary}
                            onChange={(e) => onChange({ ...settings, ffmpegBinary: e.target.value })}
                        />
                        <button
                            onClick={() => handleBrowse('ffmpegBinary')}
                            className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-600 transition-colors shadow-sm"
                            title="Browse..."
                            aria-label="Browse for FFmpeg binary"
                        >
                            <FolderOpen className="w-4 h-4" />
                        </button>
                        <button
                            onClick={() => handleInstall('ffmpeg')}
                            disabled={!!installing}
                            className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-600 transition-colors shadow-sm"
                            title="Install via WinGet"
                            aria-label="Install FFmpeg via WinGet"
                        >
                            {installing === 'ffmpeg' ? <Loader2 className="w-4 h-4 animate-spin" /> : <Download className="w-4 h-4" />}
                        </button>
                    </div>
                </div>
                <div className="space-y-2">
                    <label htmlFor="tools-magick-path" className="text-xs font-medium text-slate-500 dark:text-slate-400">Magick Binary Path</label>
                    <div className="flex gap-2">
                        <input
                            id="tools-magick-path"
                            type="text"
                            className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-xs px-3 py-2.5 font-mono truncate transition-shadow"
                            value={settings.magickBinary}
                            onChange={(e) => onChange({ ...settings, magickBinary: e.target.value })}
                        />
                        <button
                            onClick={() => handleBrowse('magickBinary')}
                            className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-600 transition-colors shadow-sm"
                            title="Browse..."
                            aria-label="Browse for ImageMagick binary"
                        >
                            <FolderOpen className="w-4 h-4" />
                        </button>
                        <button
                            onClick={() => handleInstall('magick')}
                            disabled={!!installing}
                            className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-600 transition-colors shadow-sm"
                            title="Install via WinGet"
                            aria-label="Install ImageMagick via WinGet"
                        >
                            {installing === 'magick' ? <Loader2 className="w-4 h-4 animate-spin" /> : <Download className="w-4 h-4" />}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}
