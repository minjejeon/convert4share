import React, { useEffect, useState, useRef } from 'react';
import { GetContextMenuStatus, InstallContextMenu } from './wailsjs/go/main/App';
import { Layout } from './components/Layout';
import { DropZone } from './components/DropZone';
import { FileList } from './components/FileList';
import { SettingsView } from './components/Settings';
import { AlertCircle, Loader2, UploadCloud } from 'lucide-react';
import { useTheme } from './hooks/useTheme';
import { useFileQueue } from './hooks/useFileQueue';

function App() {
    const [view, setView] = useState<'home' | 'settings'>('home');
    const [isInstalled, setIsInstalled] = useState<boolean>(true);
    const [isInstalling, setIsInstalling] = useState<boolean>(false);
    const [isDraggingGlobal, setIsDraggingGlobal] = useState(false);
    const { theme, setTheme } = useTheme();
    const { files, addFile, handleRemove, handleClearCompleted, handleCopy, isPaused, pauseQueue, resumeQueue } = useFileQueue();
    const installIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
    const installTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    useEffect(() => {
        GetContextMenuStatus().then(setIsInstalled);

        const handleWindowDragEnter = (e: DragEvent) => {
            if (e.dataTransfer?.types.includes('Files')) {
                setIsDraggingGlobal(true);
            }
        };

        window.addEventListener('dragenter', handleWindowDragEnter);

        return () => {
            window.removeEventListener('dragenter', handleWindowDragEnter);
        };
    }, []);

    const handleInstall = async () => {
        setIsInstalling(true);
        try {
            await InstallContextMenu();
            const interval = setInterval(() => {
                GetContextMenuStatus().then(status => {
                    if (status) {
                        setIsInstalled(true);
                        setIsInstalling(false);
                        clearInterval(interval);
                    }
                });
            }, 1000);
            installIntervalRef.current = interval;
            const timeout = setTimeout(() => {
                clearInterval(interval);
                GetContextMenuStatus().then(status => {
                    if (status) setIsInstalled(true);
                    setIsInstalling(false);
                });
            }, 10000);
            installTimeoutRef.current = timeout;
        } catch (e) {
            console.error(e);
            setIsInstalling(false);
        }
    };

    useEffect(() => {
        return () => {
            if (installIntervalRef.current) clearInterval(installIntervalRef.current);
            if (installTimeoutRef.current) clearTimeout(installTimeoutRef.current);
        };
    }, []);

    const handleGlobalDrop = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDraggingGlobal(false);
        if (e.dataTransfer.files.length > 0) {
            const paths = Array.from(e.dataTransfer.files).map((f) => {
                const file = f as File & { path?: string };
                return file.path || file.name;
            });
            paths.filter(p => p).forEach(addFile);
        }
    };

    return (
        <Layout currentView={view} onNavigate={setView}>
            {isDraggingGlobal && (
                <div
                    className="fixed inset-0 z-[100] bg-indigo-500/10 backdrop-blur-sm border-4 border-indigo-500 border-dashed m-4 rounded-2xl flex items-center justify-center animate-in fade-in duration-200"
                    onDragOver={(e) => e.preventDefault()}
                    onDragLeave={() => setIsDraggingGlobal(false)}
                    onDrop={handleGlobalDrop}
                >
                    <div className="bg-white dark:bg-slate-900 p-8 rounded-full shadow-2xl pointer-events-none transform transition-transform animate-bounce">
                        <UploadCloud className="w-16 h-16 text-indigo-500" />
                    </div>
                    <div className="absolute bottom-1/4 pointer-events-none bg-white/90 dark:bg-slate-900/90 px-6 py-3 rounded-full shadow-lg backdrop-blur-md">
                         <h2 className="text-2xl font-bold text-indigo-600 dark:text-indigo-400">Drop files anywhere</h2>
                    </div>
                </div>
            )}

            {view === 'home' && (
                <div className="max-w-3xl mx-auto h-full flex flex-col gap-6 p-6">
                     {!isInstalled && (
                        <div className="rounded-xl bg-white dark:bg-slate-800/60 p-4 border border-indigo-200 dark:border-indigo-500/30 flex items-center justify-between shadow-sm backdrop-blur-sm shrink-0">
                            <div className="flex items-center gap-4">
                                <div className="p-2 bg-indigo-50 dark:bg-indigo-500/10 rounded-lg shrink-0">
                                    <AlertCircle className="h-5 w-5 text-indigo-600 dark:text-indigo-400" />
                                </div>
                                <div>
                                    <h3 className="text-sm font-medium text-slate-900 dark:text-slate-200">Context Menu Integration</h3>
                                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Install the right-click menu option for quick access.</p>
                                </div>
                            </div>
                            <button
                                onClick={handleInstall}
                                disabled={isInstalling}
                                className={`px-4 py-2 text-xs font-semibold text-white rounded-lg transition-all flex items-center gap-2 shadow-sm ${
                                    isInstalling
                                        ? "bg-indigo-500/50 cursor-wait"
                                        : "bg-indigo-600 hover:bg-indigo-700 hover:shadow-indigo-500/20"
                                }`}
                            >
                                {isInstalling && <Loader2 className="w-3 h-3 animate-spin" />}
                                {isInstalling ? "Installing..." : "Install Integration"}
                            </button>
                        </div>
                    )}
                    <div className="shrink-0">
                        <DropZone
                            onFilesAdded={(paths) => paths.forEach(addFile)}
                            isCompact={files.length > 0}
                        />
                    </div>
                    <div className="flex-1 min-h-0">
                        <FileList
                            files={files}
                            onRemove={handleRemove}
                            onCopy={handleCopy}
                            onClearCompleted={handleClearCompleted}
                            isPaused={isPaused}
                            onPause={pauseQueue}
                            onResume={resumeQueue}
                        />
                    </div>
                </div>
            )}
            {view === 'settings' && <div className="h-full overflow-y-auto px-6"><SettingsView isInstalled={isInstalled} onStatusChange={setIsInstalled} theme={theme} onThemeChange={setTheme} /></div>}
        </Layout>
    );
}

export default App;
