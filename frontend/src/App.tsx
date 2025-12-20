import React, { useEffect, useState, useCallback, useRef } from 'react';
import { EventsOn, EventsEmit } from './wailsjs/runtime/runtime';
import { ConvertFiles, GetContextMenuStatus, InstallContextMenu, CopyFileToClipboard, GetThumbnail, CancelJob, PauseQueue, ResumeQueue } from './wailsjs/go/main/App';
import { Layout } from './components/Layout';
import { DropZone } from './components/DropZone';
import { FileList, FileItem } from './components/FileList';
import { SettingsView } from './components/Settings';
import { AlertCircle, Loader2, UploadCloud } from 'lucide-react';
import { useTheme } from './hooks/useTheme';

interface ProgressData {
    file: string;
    destFile?: string;
    status: 'queued' | 'processing' | 'done' | 'error';
    progress: number;
    speed?: string;
    error?: string;
}

function App() {
    const [view, setView] = useState<'home' | 'settings'>('home');
    const [files, setFiles] = useState<FileItem[]>([]);
    const [isInstalled, setIsInstalled] = useState<boolean>(true);
    const [isInstalling, setIsInstalling] = useState<boolean>(false);
    const [isDraggingGlobal, setIsDraggingGlobal] = useState(false);
    const [isPaused, setIsPaused] = useState<boolean>(false);
    const { theme, setTheme } = useTheme();
    const filesRef = useRef(files);
    filesRef.current = files;
    const installIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
    const installTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    const addFile = (path: string) => {
        // Prevent redundant thumbnail requests by checking against the current files ref
        if (filesRef.current.some(f => f.path === path)) return;

        setFiles(prev => {
            if (prev.some(f => f.path === path)) return prev;
            return [...prev, { id: path, path, status: 'queued', progress: 0, addedAt: Date.now() }];
        });

        GetThumbnail(path).then(thumb => {
            setFiles(prev => prev.map(f => f.path === path ? { ...f, thumbnail: thumb } : f));
        }).catch(err => {
            console.error("Failed to load thumbnail for", path, err);
        });
    };

    const handleRemove = useCallback((id: string) => {
        CancelJob(id);
        setFiles(prev => prev.filter(f => f.id !== id));
    }, []);

    const handleClearCompleted = useCallback(() => {
        setFiles(prev => prev.filter(f => f.status !== 'done'));
    }, []);

    const handleCopy = useCallback((path: string) => {
        CopyFileToClipboard(path).catch(console.error);
    }, []);

    useEffect(() => {
        GetContextMenuStatus().then(setIsInstalled);

        const cleanupFileAdded = EventsOn("file-added", (path: string) => {
            console.log("File added:", path);
            addFile(path);
        });

        const cleanupFilesReceived = EventsOn("files-received", (paths: string[]) => {
             console.log("Files received:", paths);
             paths.forEach(addFile);
        });

        const cleanupProgress = EventsOn("conversion-progress", (data: ProgressData) => {
            setFiles(prev => prev.map(f => {
                if (f.id === data.file) {
                    const now = Date.now();
                    const isDone = data.status === 'done';
                    return {
                        ...f,
                        status: data.status,
                        progress: data.progress,
                        speed: data.speed,
                        error: data.error,
                        destFile: data.destFile,
                        completedAt: isDone && !f.completedAt ? now : f.completedAt
                    };
                }
                return f;
            }));
        });

        const handleWindowDragEnter = (e: DragEvent) => {
            if (e.dataTransfer?.types.includes('Files')) {
                setIsDraggingGlobal(true);
            }
        };

        window.addEventListener('dragenter', handleWindowDragEnter);
        const cleanupPaused = EventsOn("queue-paused", () => setIsPaused(true));
        const cleanupResumed = EventsOn("queue-resumed", () => setIsPaused(false));

        EventsEmit("frontend-ready");

        return () => {
            cleanupFileAdded();
            cleanupFilesReceived();
            cleanupProgress();
            window.removeEventListener('dragenter', handleWindowDragEnter);
            cleanupPaused();
            cleanupResumed();
        };
    }, []);

    const queuedCount = files.filter(f => f.status === 'queued').length;
    useEffect(() => {
        const queued = files.filter(f => f.status === 'queued');
        if (queued.length > 0) {
            const timeout = setTimeout(() => {
                const queuedPaths = queued.map(f => f.path);
                setFiles(prev => prev.map(f => queuedPaths.includes(f.path) ? { ...f, status: 'processing', progress: 0 } : f));

                ConvertFiles(queuedPaths);
            }, 100);
            return () => clearTimeout(timeout);
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [queuedCount]);

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
                            onPause={PauseQueue}
                            onResume={ResumeQueue}
                        />
                    </div>
                </div>
            )}
            {view === 'settings' && <div className="h-full overflow-y-auto px-6"><SettingsView isInstalled={isInstalled} onStatusChange={setIsInstalled} theme={theme} onThemeChange={setTheme} /></div>}
        </Layout>
    );
}

export default App;
