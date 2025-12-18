import React from 'react';
import { Loader2, Monitor, Moon, Sun, Laptop } from 'lucide-react';
import { cn } from '../lib/utils';

interface SettingsIntegrationProps {
    isInstalled: boolean;
    onStatusChange: (status: boolean) => void;
    theme: 'dark' | 'light' | 'system';
    onThemeChange: (theme: 'dark' | 'light' | 'system') => void;
    togglingMenu: boolean;
    onToggleMenu: () => void;
}

export function SettingsIntegration({
    isInstalled,
    theme,
    onThemeChange,
    togglingMenu,
    onToggleMenu
}: SettingsIntegrationProps) {
    return (
        <div className="bg-white dark:bg-slate-800/40 rounded-xl p-6 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-colors shadow-sm dark:shadow-none">
            <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 mb-4 flex items-center gap-2">
                <Monitor className="h-4 w-4 text-indigo-600 dark:text-indigo-400" />
                Windows Integration
            </h3>
            <div className="flex items-center justify-between bg-slate-100 dark:bg-slate-900/50 p-4 rounded-lg border border-slate-200 dark:border-slate-800/50">
                <div>
                    <p className="text-sm font-medium text-slate-800 dark:text-slate-200">Context Menu</p>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                        {isInstalled ? "Currently installed. Right-click files to convert." : "Not installed. Install to add to right-click menu."}
                    </p>
                </div>
                <button
                    onClick={onToggleMenu}
                    disabled={togglingMenu}
                    className={cn(
                        "px-4 py-2 text-xs font-semibold rounded-lg transition-all border flex items-center justify-center min-w-[90px] shadow-sm",
                        isInstalled
                            ? "border-red-500/20 text-red-500 dark:text-red-400 hover:bg-red-500/10 hover:border-red-500/30"
                            : "border-indigo-500/20 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-500/10 hover:border-indigo-500/30",
                        togglingMenu && "opacity-50 cursor-wait"
                    )}
                >
                    {togglingMenu ? <Loader2 className="animate-spin h-4 w-4" /> : (isInstalled ? "Uninstall" : "Install")}
                </button>
            </div>

            <div className="flex items-center justify-between bg-slate-100 dark:bg-slate-900/50 p-4 rounded-lg border border-slate-200 dark:border-slate-800/50 mt-4">
                <div>
                    <p className="text-sm font-medium text-slate-800 dark:text-slate-200">Theme</p>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                        Customize the application appearance.
                    </p>
                </div>
                <div className="flex bg-slate-200 dark:bg-slate-800 rounded-lg p-1 gap-1">
                    <button
                        onClick={() => onThemeChange('light')}
                        className={cn(
                            "p-1.5 rounded-md transition-all",
                            theme === 'light' ? "bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 shadow-sm" : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200"
                        )}
                        title="Light"
                    >
                        <Sun className="w-4 h-4" />
                    </button>
                    <button
                        onClick={() => onThemeChange('dark')}
                        className={cn(
                            "p-1.5 rounded-md transition-all",
                            theme === 'dark' ? "bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 shadow-sm" : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200"
                        )}
                        title="Dark"
                    >
                        <Moon className="w-4 h-4" />
                    </button>
                    <button
                        onClick={() => onThemeChange('system')}
                        className={cn(
                            "p-1.5 rounded-md transition-all",
                            theme === 'system' ? "bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 shadow-sm" : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200"
                        )}
                        title="System"
                    >
                        <Laptop className="w-4 h-4" />
                    </button>
                </div>
            </div>
        </div>
    );
}
