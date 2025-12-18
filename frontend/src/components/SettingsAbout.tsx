import React from 'react';
import { Info, FileText } from 'lucide-react';
import { LicenseViewer } from './LicenseViewer';

interface SettingsAboutProps {
    onShowLicenses: () => void;
    showLicenses: boolean;
    onCloseLicenses: () => void;
}

export function SettingsAbout({ onShowLicenses, showLicenses, onCloseLicenses }: SettingsAboutProps) {
    return (
        <>
            <div className="bg-white dark:bg-slate-800/40 rounded-xl p-6 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-colors shadow-sm dark:shadow-none flex items-center justify-between">
                <div>
                    <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 flex items-center gap-2">
                        <Info className="h-4 w-4 text-slate-600 dark:text-slate-400" />
                        About & Legal
                    </h3>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                        View third-party software licenses and legal notices.
                    </p>
                </div>
                <button
                    onClick={onShowLicenses}
                    className="px-4 py-2 text-xs font-semibold rounded-lg bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 transition-colors flex items-center gap-2 shadow-sm"
                >
                    <FileText className="w-4 h-4" />
                    View Licenses
                </button>
            </div>
            {showLicenses && <LicenseViewer onClose={onCloseLicenses} />}
        </>
    );
}
