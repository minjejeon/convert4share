import React, { useEffect, useState } from 'react';
import { GetSettings, SaveSettings } from '../wailsjs/go/main/App';
import { main } from '../wailsjs/go/models';
import { Loader2, Save, Check } from 'lucide-react';
import { cn } from '../lib/utils';
import { LicenseViewer } from './LicenseViewer';
import { SettingsIntegration } from './SettingsIntegration';
import { SettingsVideo } from './SettingsVideo';
import { SettingsTools } from './SettingsTools';
import { SettingsPaths } from './SettingsPaths';
import { SettingsAbout } from './SettingsAbout';

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
    const [showLicenses, setShowLicenses] = useState(false);

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
                <SettingsIntegration
                    isInstalled={isInstalled}
                    onStatusChange={onStatusChange}
                    theme={theme}
                    onThemeChange={onThemeChange}
                />

                <SettingsVideo
                    settings={settings}
                    onChange={setSettings}
                />

                <SettingsTools
                    settings={settings}
                    onChange={setSettings}
                />

                <SettingsPaths
                    settings={settings}
                    onChange={setSettings}
                />

                <SettingsAbout
                    onShowLicenses={() => setShowLicenses(true)}
                />

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
            {showLicenses && <LicenseViewer onClose={() => setShowLicenses(false)} />}
        </div>
    );
}
