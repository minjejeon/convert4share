import { useEffect, useState } from 'react';
import { FolderOpen } from 'lucide-react';
import type { Settings } from '@bindings/config/models';
import { PreviewFileName } from '@bindings/services/settings/service';

interface SettingsPathsProps {
    settings: Settings;
    onChange: (settings: Settings) => void;
}

const PRESETS: { label: string; format: string }[] = [
    { label: '26-07-05 10-22-43 3574', format: '{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}' },
    { label: '2026-07-05_10-22-43', format: '{YYYY}-{MM}-{DD}_{HH}-{mm}-{ss}' },
    { label: '20260705_102243_3574', format: '{YYYY}{MM}{DD}_{HH}{mm}{ss}_{num}' },
    { label: 'IMG_20260705_102243', format: 'IMG_{YYYY}{MM}{DD}_{HH}{mm}{ss}' },
    { label: 'name + date', format: '{name}_{YYYY}{MM}{DD}' },
];

const inputClass =
    'block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow';

export function SettingsPaths({ settings, onChange }: SettingsPathsProps) {
    const captureTime = settings.fileNaming === 'captureTime';
    const [preview, setPreview] = useState('');

    useEffect(() => {
        if (!captureTime) return;
        const p = PreviewFileName(settings.fileNameFormat || '');
        p.then(setPreview);
        return () => {
            p.cancel();
        };
    }, [captureTime, settings.fileNameFormat]);

    return (
        <div className="bg-white dark:bg-slate-800/40 rounded-xl p-6 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-colors shadow-sm dark:shadow-none">
             <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 mb-6 flex items-center gap-2">
                <FolderOpen className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                Paths & Filters
             </h3>
             <div className="space-y-5">
                <div className="space-y-2">
                    <label htmlFor="paths-dest-dir" className="text-xs font-medium text-slate-500 dark:text-slate-400">Default Destination Directory</label>
                    <input
                        id="paths-dest-dir"
                        type="text"
                        className={inputClass.replace('sm:text-sm', 'sm:text-xs') + ' font-mono'}
                        value={settings.defaultDestDir}
                        onChange={(e) => onChange({ ...settings, defaultDestDir: e.target.value })}
                    />
                </div>
                <div className="space-y-2">
                    <label htmlFor="paths-collision" className="text-xs font-medium text-slate-500 dark:text-slate-400">File Collision Behavior</label>
                    <select
                        id="paths-collision"
                        className={inputClass}
                        value={settings.collisionOption || 'rename'}
                        onChange={(e) => onChange({ ...settings, collisionOption: e.target.value })}
                    >
                        <option value="rename">Rename</option>
                        <option value="overwrite">Overwrite</option>
                        <option value="error">Error</option>
                    </select>
                </div>
                <div className="space-y-2">
                    <label htmlFor="paths-naming" className="text-xs font-medium text-slate-500 dark:text-slate-400">File Naming</label>
                    <select
                        id="paths-naming"
                        className={inputClass}
                        value={settings.fileNaming || 'original'}
                        onChange={(e) => onChange({ ...settings, fileNaming: e.target.value })}
                    >
                        <option value="original">Keep original name</option>
                        <option value="captureTime">Capture time based</option>
                    </select>
                </div>
                {captureTime && (
                    <div className="space-y-2 pl-3 border-l-2 border-slate-200 dark:border-slate-700">
                        <label htmlFor="paths-name-preset" className="text-xs font-medium text-slate-500 dark:text-slate-400">Preset</label>
                        <select
                            id="paths-name-preset"
                            className={inputClass}
                            value={settings.fileNameFormat || ''}
                            onChange={(e) => onChange({ ...settings, fileNameFormat: e.target.value })}
                        >
                            {!PRESETS.some((p) => p.format === settings.fileNameFormat) && (
                                <option value={settings.fileNameFormat}>Custom</option>
                            )}
                            {PRESETS.map((p) => (
                                <option key={p.format} value={p.format}>{p.label}</option>
                            ))}
                        </select>
                        <label htmlFor="paths-name-format" className="text-xs font-medium text-slate-500 dark:text-slate-400">Format</label>
                        <input
                            id="paths-name-format"
                            type="text"
                            className={inputClass + ' font-mono'}
                            value={settings.fileNameFormat || ''}
                            onChange={(e) => onChange({ ...settings, fileNameFormat: e.target.value })}
                            placeholder="{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}"
                        />
                        <p className="text-[10px] text-slate-400 dark:text-slate-500">
                            Preview: <span className="font-mono text-slate-600 dark:text-slate-300">{preview || '…'}</span>
                            <span className="ml-2">· Tokens: {'{YYYY} {YY} {MM} {DD} {HH} {mm} {ss} {num} {seq} {name}'}</span>
                        </p>
                    </div>
                )}
                <div className="space-y-2">
                    <label htmlFor="paths-exclude" className="text-xs font-medium text-slate-500 dark:text-slate-400">Exclude Patterns (comma separated)</label>
                     <input
                        id="paths-exclude"
                        type="text"
                        className={inputClass}
                        value={settings.excludePatterns?.join(', ')}
                        onChange={(e) => onChange({ ...settings, excludePatterns: e.target.value.split(',').filter(s => s.trim() !== '').map(s => s.trim()) })}
                        placeholder="e.g. \Pictures\, \DCIM\"
                    />
                </div>
                <div className="space-y-2">
                    <label htmlFor="paths-copy-only" className="text-xs font-medium text-slate-500 dark:text-slate-400">Copy Only Extensions (comma separated)</label>
                     <input
                        id="paths-copy-only"
                        type="text"
                        className={inputClass}
                        value={settings.copyOnlyExtensions?.join(', ')}
                        onChange={(e) => onChange({ ...settings, copyOnlyExtensions: e.target.value.split(',').filter(s => s.trim() !== '').map(s => s.trim()) })}
                        placeholder="e.g. .jpg, .jpeg, .mp4"
                    />
                    <p className="text-[10px] text-slate-400 dark:text-slate-500">Files with these extensions will be copied to the destination without conversion.</p>
                </div>
             </div>
        </div>
    );
}
