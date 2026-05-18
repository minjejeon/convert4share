import { useEffect, useState } from 'react';
import { GetSettings, SaveSettings } from '@bindings/services/settings/service';
import { Settings as SettingsModel } from '@bindings/config/models';
import { Loader2, Save, Check, Puzzle, Film, Image as ImageIcon, Wrench, FolderOpen, Info } from 'lucide-react';
import { cn } from '../lib/utils';
import { LicenseViewer } from './LicenseViewer';
import { SettingsIntegration } from './SettingsIntegration';
import { SettingsVideo } from './SettingsVideo';
import { SettingsImage } from './SettingsImage';
import { SettingsTools } from './SettingsTools';
import { SettingsPaths } from './SettingsPaths';
import { SettingsAbout } from './SettingsAbout';

interface SettingsViewProps {
    theme: 'dark' | 'light' | 'system';
    onThemeChange: (theme: 'dark' | 'light' | 'system') => void;
}

type TabKey = 'integration' | 'video' | 'image' | 'tools' | 'paths' | 'about';

const TABS: { key: TabKey; label: string; Icon: typeof Puzzle }[] = [
    { key: 'integration', label: 'Integration', Icon: Puzzle },
    { key: 'video', label: 'Video', Icon: Film },
    { key: 'image', label: 'Image', Icon: ImageIcon },
    { key: 'tools', label: 'Tools', Icon: Wrench },
    { key: 'paths', label: 'Paths', Icon: FolderOpen },
    { key: 'about', label: 'About', Icon: Info },
];

export function SettingsView({ theme, onThemeChange }: SettingsViewProps) {
    const [settings, setSettings] = useState<SettingsModel | null>(null);
    const [saving, setSaving] = useState(false);
    const [saved, setSaved] = useState(false);
    const [loading, setLoading] = useState(true);
    const [showLicenses, setShowLicenses] = useState(false);
    const [activeTab, setActiveTab] = useState<TabKey>('integration');

    useEffect(() => {
        GetSettings().then((s: SettingsModel) => {
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
        <div className="max-w-2xl mx-auto space-y-6 pb-12">
            <div>
                <h2 className="text-xl font-bold text-slate-900 dark:text-slate-100">Settings</h2>
                <p className="text-slate-500 dark:text-slate-400 mt-1">Configure conversion parameters and system integration.</p>
            </div>

            <nav
                role="tablist"
                aria-label="Settings sections"
                className="flex flex-wrap gap-1 p-1 bg-slate-100 dark:bg-slate-800/50 rounded-lg border border-slate-200 dark:border-white/5 sticky top-0 z-10 backdrop-blur-md"
            >
                {TABS.map(({ key, label, Icon }) => {
                    const isActive = activeTab === key;
                    return (
                        <button
                            key={key}
                            role="tab"
                            id={`tab-${key}`}
                            aria-selected={isActive}
                            aria-controls={`panel-${key}`}
                            onClick={() => setActiveTab(key)}
                            className={cn(
                                "flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium transition-all duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500/60",
                                isActive
                                    ? "bg-white dark:bg-slate-700 text-indigo-600 dark:text-white shadow-sm ring-1 ring-slate-900/5 dark:ring-0"
                                    : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-200/50 dark:hover:bg-slate-700/50"
                            )}
                        >
                            <Icon className="w-3.5 h-3.5" aria-hidden="true" />
                            {label}
                        </button>
                    );
                })}
            </nav>

            <div role="tabpanel" id={`panel-${activeTab}`} aria-labelledby={`tab-${activeTab}`} className="space-y-6">
                {activeTab === 'integration' && (
                    <SettingsIntegration
                        theme={theme}
                        onThemeChange={onThemeChange}
                        settings={settings}
                        onChange={setSettings}
                    />
                )}
                {activeTab === 'video' && <SettingsVideo settings={settings} onChange={setSettings} />}
                {activeTab === 'image' && <SettingsImage settings={settings} onChange={setSettings} />}
                {activeTab === 'tools' && <SettingsTools settings={settings} onChange={setSettings} />}
                {activeTab === 'paths' && <SettingsPaths settings={settings} onChange={setSettings} />}
                {activeTab === 'about' && <SettingsAbout onShowLicenses={() => setShowLicenses(true)} />}

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
