import React from 'react';
import { Image, Layers } from 'lucide-react';
import { main } from '../wailsjs/go/models';

interface SettingsImageProps {
    settings: main.Settings;
    onChange: (settings: main.Settings) => void;
}

export function SettingsImage({ settings, onChange }: SettingsImageProps) {
    return (
        <div className="bg-white dark:bg-slate-800/40 rounded-xl p-6 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-colors space-y-6 shadow-sm dark:shadow-none">
            <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 flex items-center gap-2">
                <Image className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
                Image Options (HEIC to JPG)
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-4">
                    <div className="space-y-2">
                        <label htmlFor="image-workers" className="text-xs font-medium text-slate-500 dark:text-slate-400">Concurrent Jobs</label>
                        <input
                            id="image-workers"
                            type="number"
                            min="1"
                            className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow"
                            value={settings.maxMagickWorkers || 5}
                            onChange={(e) => onChange({ ...settings, maxMagickWorkers: parseInt(e.target.value) || 1 })}
                        />
                    </div>

                    <div className="flex items-start gap-3">
                        <div className="flex items-center h-5">
                            <input
                                id="auto-live-photo"
                                type="checkbox"
                                className="h-4 w-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-600 dark:bg-slate-900 dark:border-slate-700"
                                checked={settings.autoLivePhoto}
                                onChange={(e) => onChange({ ...settings, autoLivePhoto: e.target.checked })}
                            />
                        </div>
                        <div className="text-sm">
                            <label htmlFor="auto-live-photo" className="font-medium text-slate-700 dark:text-slate-200">
                                Live Photo Detection
                            </label>
                            <p className="text-xs text-slate-500 dark:text-slate-400">
                                Skip .mov file if a matching .heic exists in the batch.
                            </p>
                        </div>
                    </div>
                </div>

                <div className="space-y-2">
                    <label htmlFor="image-max-size" className="text-xs font-medium text-slate-500 dark:text-slate-400">Max Resolution (Size)</label>
                     <input
                        id="image-max-size"
                        type="number"
                        className="block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow"
                        value={settings.maxImageSize}
                        onChange={(e) => onChange({ ...settings, maxImageSize: parseInt(e.target.value) || 0 })}
                        placeholder="0 for original (default 2560)"
                    />
                    <p className="text-[10px] text-slate-400 dark:text-slate-500">The longer side will be limited to this value. 0 means no resizing.</p>
                </div>
            </div>
        </div>
    );
}
