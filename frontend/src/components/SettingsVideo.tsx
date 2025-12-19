import React from 'react';
import { Film, Cpu, Layers } from 'lucide-react';
import { main } from '../wailsjs/go/models';

interface SettingsVideoProps {
    settings: main.Settings;
    onChange: (settings: main.Settings) => void;
}

export function SettingsVideo({ settings, onChange }: SettingsVideoProps) {
    return (
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
                        onChange={(e) => onChange({ ...settings, hardwareAccelerator: e.target.value })}
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
                        onChange={(e) => onChange({ ...settings, videoQuality: e.target.value })}
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
                        onChange={(e) => onChange({ ...settings, maxFfmpegWorkers: parseInt(e.target.value) || 1 })}
                    />
                </div>

                <div className="space-y-2">
                    <label className="text-xs font-medium text-slate-500 dark:text-slate-400">Max Resolution (Size)</label>
                     <input
                        type="number"
                        className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow"
                        value={settings.maxSize}
                        onChange={(e) => onChange({ ...settings, maxSize: parseInt(e.target.value) || 0 })}
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
                    onChange={(e) => onChange({ ...settings, ffmpegCustomArgs: e.target.value })}
                    placeholder="-crf 23 -preset slow"
                />
            </div>
        </div>
    );
}
