import React, { useEffect, useState } from 'react';
import { GetSettings, SaveSettings, InstallContextMenu, UninstallContextMenu, GetContextMenuStatus, SelectBinaryDialog, DetectBinaries, InstallTool } from '../wailsjs/go/main/App';
import { main } from '../wailsjs/go/models';
import { Loader2, Save, Check } from 'lucide-react';
import { cn } from '../lib/utils';
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
    const [togglingMenu, setTogglingMenu] = useState(false);
    const [showLicenses, setShowLicenses] = useState(false);
    const [installing, setInstalling] = useState<string | null>(null);

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

    const handleInstall = async (tool: string) => {
        setInstalling(tool);
        try {
            await InstallTool(tool);
            await handleDetect();
            setSaved(true);
        } catch (e) {
            console.error(e);
            const message = e instanceof Error ? e.message : String(e);
            alert("Installation failed. Please try installing manually via 'winget install " + (tool === 'ffmpeg' ? 'Gyan.FFmpeg' : 'ImageMagick.ImageMagick') + "' in PowerShell.\n\nError: " + message);
        } finally {
            setInstalling(null);
        }
    };

    const handleBrowse = async (field: 'magickBinary' | 'ffmpegBinary') => {
        const path = await SelectBinaryDialog();
        if (path && settings) {
            setSettings({ ...settings, [field]: path } as main.Settings);
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
                setSettings(newSettings as main.Settings);
            }
        } catch (error) {
            console.error("Error detecting binaries:", error);
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
                    togglingMenu={togglingMenu}
                    onToggleMenu={handleToggleMenu}
                />

                <SettingsVideo
                    settings={settings}
                    onSettingsChange={setSettings}
                />

                <SettingsTools
                    settings={settings}
                    onSettingsChange={setSettings}
                    onBrowse={handleBrowse}
                    onInstall={handleInstall}
                    onDetect={handleDetect}
                    installing={installing}
                />

                <SettingsPaths
                    settings={settings}
                    onSettingsChange={setSettings}
                />

                <SettingsAbout
                    onShowLicenses={() => setShowLicenses(true)}
                    showLicenses={showLicenses}
                    onCloseLicenses={() => setShowLicenses(false)}
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
        </div>
    );
}
