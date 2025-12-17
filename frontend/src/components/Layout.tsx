import React from 'react';
import { Settings, CheckCircle2 } from 'lucide-react';

interface LayoutProps {
    children: React.ReactNode;
    currentView: 'home' | 'settings';
    onNavigate: (view: 'home' | 'settings') => void;
}

export function Layout({ children, currentView, onNavigate }: LayoutProps) {
    return (
        <div className="h-screen bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 flex flex-col font-sans selection:bg-indigo-500/30 overflow-hidden">
            <header className="h-14 border-b border-slate-200 dark:border-white/5 flex items-center justify-between px-6 bg-white/80 dark:bg-slate-900/60 backdrop-blur-md sticky top-0 z-50 select-none drag-region shrink-0">
                <div className="flex items-center gap-2.5">
                    <div className="bg-indigo-50 dark:bg-indigo-500/10 p-1.5 rounded-lg ring-1 ring-inset ring-indigo-500/20">
                        <CheckCircle2 className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
                    </div>
                    <h1 className="text-sm font-bold tracking-wide text-slate-900 dark:text-slate-200">Convert4Share</h1>
                </div>
                <nav className="flex gap-1 no-drag-region bg-slate-100 dark:bg-slate-800/50 p-1 rounded-lg border border-slate-200 dark:border-white/5">
                    <button
                        onClick={() => onNavigate('home')}
                        className={`px-3 py-1 rounded-md text-xs font-medium transition-all duration-200 ${
                            currentView === 'home'
                                ? 'bg-white dark:bg-slate-700 text-indigo-600 dark:text-white shadow-sm ring-1 ring-slate-900/5 dark:ring-0'
                                : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-200/50 dark:hover:bg-slate-700/50'
                        }`}
                    >
                        Convert
                    </button>
                    <button
                        onClick={() => onNavigate('settings')}
                        className={`px-3 py-1 rounded-md text-xs font-medium transition-all duration-200 ${
                            currentView === 'settings'
                                ? 'bg-white dark:bg-slate-700 text-indigo-600 dark:text-white shadow-sm ring-1 ring-slate-900/5 dark:ring-0'
                                : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-200/50 dark:hover:bg-slate-700/50'
                        }`}
                    >
                        <Settings className="w-3 h-3 inline-block mr-1.5 -mt-0.5" />
                        Settings
                    </button>
                </nav>
            </header>
            <main className="flex-1 overflow-hidden relative">
                {children}
            </main>
        </div>
    );
}
