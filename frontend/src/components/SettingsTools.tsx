import React from 'react';
import { Sliders, Search, FolderOpen, Download, Loader2 } from 'lucide-react';
import { main } from '../wailsjs/go/models';

interface SettingsToolsProps {
    settings: main.Settings;
    onSettingsChange: (settings: main.Settings) => void;
    onBrowse: (field: 'magickBinary' | 'ffmpegBinary') => void;
    onInstall: (tool: string) => void;
    onDetect: () => void;
    installing: string | null;
}

export function SettingsTools({
    settings,
    onSettingsChange,
    onBrowse,
    onInstall,
    onDetect,
    installing
}: SettingsToolsProps) {
    return (
        <div className="bg-white dark:bg-slate-800/40 rounded-xl p-6 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-colors shadow-sm dark:shadow-none">
            <div className="flex items-center justify-between mb-6">
                <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 flex items-center gap-2">
                    <Sliders className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
                    External Tools
                </h3>
                <button
                    onClick={onDetect}
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
                            onChange={(e) => onSettingsChange({ ...settings, ffmpegBinary: e.target.value } as main.Settings)}
                        />
                        <button
                            onClick={() => onBrowse('ffmpegBinary')}
                            className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-600 transition-colors shadow-sm"
                            title="Browse..."
                        >
                            <FolderOpen className="w-4 h-4" />
                        </button>
                        <button
                            onClick={() => onInstall('ffmpeg')}
                            disabled={!!installing}
                            className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-600 transition-colors shadow-sm"
                            title="Install via WinGet"
                        >
                            {installing === 'ffmpeg' ? <Loader2 className="w-4 h-4 animate-spin" /> : <Download className="w-4 h-4" />}
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
                            onChange={(e) => onSettingsChange({ ...settings, magickBinary: e.target.value } as main.Settings)}
                        />
                        <button
                            onClick={() => onBrowse('magickBinary')}
                            className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-600 transition-colors shadow-sm"
                            title="Browse..."
                        >
                            <FolderOpen className="w-4 h-4" />
                        </button>
                        <button
                            onClick={() => onInstall('magick')}
                            disabled={!!installing}
                            className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-600 transition-colors shadow-sm"
                            title="Install via WinGet"
                        >
                            {installing === 'magick' ? <Loader2 className="w-4 h-4 animate-spin" /> : <Download className="w-4 h-4" />}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}
