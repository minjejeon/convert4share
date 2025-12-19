import React from 'react';
import { FolderOpen } from 'lucide-react';
import { main } from '../wailsjs/go/models';

interface SettingsPathsProps {
    settings: main.Settings;
    onChange: (settings: main.Settings) => void;
}

export function SettingsPaths({ settings, onChange }: SettingsPathsProps) {
    return (
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
                        onChange={(e) => onChange({ ...settings, defaultDestDir: e.target.value })}
                    />
                </div>
                <div className="space-y-2">
                    <label className="text-xs font-medium text-slate-500 dark:text-slate-400">File Collision Behavior</label>
                    <select
                        className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow"
                        value={settings.collisionOption || "rename"}
                        onChange={(e) => onChange({ ...settings, collisionOption: e.target.value })}
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
                        onChange={(e) => onChange({ ...settings, excludePatterns: e.target.value.split(',').map(s => s.trim()) })}
                        placeholder="e.g. \Pictures\, \DCIM\"
                    />
                </div>
             </div>
        </div>
    );
}
