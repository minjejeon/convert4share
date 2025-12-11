import React, { useEffect, useState } from 'react';
import { GetSettings, SaveSettings } from '../wailsjs/go/main/App';
import { main } from '../wailsjs/go/models';
import { Loader2, Save } from 'lucide-react';
import { cn } from '../lib/utils';

export function SettingsView() {
    const [settings, setSettings] = useState<main.Settings | null>(null);
    const [saving, setSaving] = useState(false);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        GetSettings().then((s) => {
            setSettings(s);
            setLoading(false);
        });
    }, []);

    const handleSave = async () => {
        if (!settings) return;
        setSaving(true);
        try {
            await SaveSettings(settings);
            // Optionally show toast
        } finally {
            setSaving(false);
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
                        disabled={saving}
                        className={cn(
                            "inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-all",
                            saving && "opacity-75 cursor-wait"
                        )}
                    >
                        {saving ? <Loader2 className="animate-spin -ml-1 mr-2 h-4 w-4"/> : <Save className="-ml-1 mr-2 h-4 w-4" />}
                        Save Changes
                    </button>
                </div>
            </div>
        </div>
    );
}
