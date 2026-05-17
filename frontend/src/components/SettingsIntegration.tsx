import { Monitor, Sun, Moon, Laptop, Terminal, Info } from 'lucide-react';
import { cn } from '../lib/utils';
import type { Settings } from '@bindings/config/models';

interface SettingsIntegrationProps {
    theme: 'dark' | 'light' | 'system';
    onThemeChange: (theme: 'dark' | 'light' | 'system') => void;
    settings: Settings;
    onChange: (settings: Settings) => void;
}

export function SettingsIntegration({ theme, onThemeChange, settings, onChange }: SettingsIntegrationProps) {
    return (
        <div className="bg-white dark:bg-slate-800/40 rounded-xl p-6 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-colors shadow-sm dark:shadow-none space-y-4">
             <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 mb-4 flex items-center gap-2">
                <Monitor className="h-4 w-4 text-indigo-600 dark:text-indigo-400" />
                Application
             </h3>

             <div className="flex items-start gap-3 bg-slate-100 dark:bg-slate-900/50 p-4 rounded-lg border border-slate-200 dark:border-slate-800/50">
                <Info className="h-4 w-4 mt-0.5 text-slate-500 dark:text-slate-400 shrink-0" />
                <div>
                    <p className="text-sm font-medium text-slate-800 dark:text-slate-200">Context Menu</p>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                        The right-click "Convert with Convert4Share" entry is registered by the installer. Re-run the installer to add or remove it.
                    </p>
                </div>
             </div>

             <div className="flex items-center justify-between bg-slate-100 dark:bg-slate-900/50 p-4 rounded-lg border border-slate-200 dark:border-slate-800/50">
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
                        title="Light Theme"
                        aria-label="Switch to light theme"
                        aria-pressed={theme === 'light'}
                    >
                        <Sun className="w-4 h-4" />
                    </button>
                    <button
                        onClick={() => onThemeChange('dark')}
                        className={cn(
                            "p-1.5 rounded-md transition-all",
                            theme === 'dark' ? "bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 shadow-sm" : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200"
                        )}
                        title="Dark Theme"
                        aria-label="Switch to dark theme"
                        aria-pressed={theme === 'dark'}
                    >
                        <Moon className="w-4 h-4" />
                    </button>
                    <button
                        onClick={() => onThemeChange('system')}
                        className={cn(
                            "p-1.5 rounded-md transition-all",
                            theme === 'system' ? "bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 shadow-sm" : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200"
                        )}
                        title="System Theme"
                        aria-label="Switch to system theme"
                        aria-pressed={theme === 'system'}
                    >
                        <Laptop className="w-4 h-4" />
                    </button>
                </div>
             </div>

             <div className="flex items-center justify-between bg-slate-100 dark:bg-slate-900/50 p-4 rounded-lg border border-slate-200 dark:border-slate-800/50">
                <div className="flex-1">
                    <div className="flex items-center gap-2">
                        <Terminal className="h-3.5 w-3.5 text-slate-600 dark:text-slate-400" />
                        <p className="text-sm font-medium text-slate-800 dark:text-slate-200">Log Level</p>
                    </div>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                        Control logging verbosity for troubleshooting.
                    </p>
                </div>
                <select
                    id="log-level"
                    className="rounded-lg bg-white dark:bg-slate-800 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 text-xs px-3 py-1.5 transition-shadow"
                    value={settings.logLevel || "info"}
                    onChange={(e) => onChange({ ...settings, logLevel: e.target.value })}
                >
                    <option value="debug">Debug</option>
                    <option value="info">Info</option>
                    <option value="warn">Warning</option>
                    <option value="error">Error</option>
                </select>
             </div>
        </div>
    );
}
