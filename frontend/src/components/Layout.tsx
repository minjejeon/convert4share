import React from 'react';
import { Settings, CheckCircle2 } from 'lucide-react';

interface LayoutProps {
    children: React.ReactNode;
    currentView: 'home' | 'settings';
    onNavigate: (view: 'home' | 'settings') => void;
}

export function Layout({ children, currentView, onNavigate }: LayoutProps) {
    return (
        <div className="min-h-screen bg-slate-900 text-slate-100 flex flex-col font-sans">
            <header className="h-14 border-b border-slate-700 flex items-center justify-between px-6 bg-slate-800/50 backdrop-blur-sm sticky top-0 z-10 select-none drag-region">
                <div className="flex items-center gap-2">
                    <CheckCircle2 className="w-6 h-6 text-indigo-400" />
                    <h1 className="text-lg font-bold tracking-tight">Convert4Share</h1>
                </div>
                <nav className="flex gap-1 no-drag-region">
                    <button
                        onClick={() => onNavigate('home')}
                        className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                            currentView === 'home'
                                ? 'bg-indigo-500/20 text-indigo-300'
                                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700/50'
                        }`}
                    >
                        Convert
                    </button>
                    <button
                        onClick={() => onNavigate('settings')}
                        className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                            currentView === 'settings'
                                ? 'bg-indigo-500/20 text-indigo-300'
                                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700/50'
                        }`}
                    >
                        <Settings className="w-4 h-4 inline-block mr-1.5 -mt-0.5" />
                        Settings
                    </button>
                </nav>
            </header>
            <main className="flex-1 p-6 overflow-auto">
                {children}
            </main>
        </div>
    );
}
