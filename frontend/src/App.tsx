import React, { useEffect, useState } from 'react';
import { EventsOn } from './wailsjs/runtime';
import { ConvertFiles, GetContextMenuStatus, InstallContextMenu, CopyFileToClipboard, GetThumbnail } from './wailsjs/go/main/App';
import { Layout } from './components/Layout';
import { DropZone } from './components/DropZone';
import { FileList, FileItem } from './components/FileList';
import { SettingsView } from './components/Settings';
import { AlertCircle, Loader2 } from 'lucide-react';

interface ProgressData {
    file: string;
    destFile?: string;
    status: 'queued' | 'processing' | 'done' | 'error';
    progress: number;
    error?: string;
}

function App() {
    const [view, setView] = useState<'home' | 'settings'>('home');
    const [files, setFiles] = useState<FileItem[]>([]);
    const [isInstalled, setIsInstalled] = useState<boolean>(true);
    const [isInstalling, setIsInstalling] = useState<boolean>(false);

    const addFile = (path: string) => {
        setFiles(prev => {
            if (prev.some(f => f.path === path)) return prev;
            return [...prev, { id: path, path, status: 'queued', progress: 0 }];
        });

        // Fetch thumbnail
        GetThumbnail(path).then(thumb => {
            setFiles(prev => prev.map(f => f.path === path ? { ...f, thumbnail: thumb } : f));
        }).catch(err => {
            console.error("Failed to load thumbnail for", path, err);
        });
    };

    const handleRemove = (id: string) => {
        setFiles(prev => prev.filter(f => f.id !== id));
    };

    const handleClearCompleted = () => {
        setFiles(prev => prev.filter(f => f.status !== 'done'));
    };

    const handleCopy = (path: string) => {
        CopyFileToClipboard(path).catch(console.error);
    };

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
                    return {
                        ...f,
                        status: data.status,
                        progress: data.progress,
                        error: data.error,
                        destFile: data.destFile
                    };
                }
                return f;
            }));
        });

        return () => {
            cleanupFileAdded();
            cleanupFilesReceived();
            cleanupProgress();
        };
    }, []);

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
    }, [files]);

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
            setTimeout(() => {
                clearInterval(interval);
                GetContextMenuStatus().then(status => {
                    if (status) setIsInstalled(true);
                    setIsInstalling(false);
                });
            }, 10000);
        } catch (e) {
            console.error(e);
            setIsInstalling(false);
        }
    };

    return (
        <Layout currentView={view} onNavigate={setView}>
            {view === 'home' && (
                <div className="max-w-3xl mx-auto pb-12 flex flex-col gap-8">
                     {!isInstalled && (
                        <div className="rounded-xl bg-slate-800/60 p-4 border border-indigo-500/30 flex items-center justify-between shadow-sm backdrop-blur-sm">
                            <div className="flex items-center gap-4">
                                <div className="p-2 bg-indigo-500/10 rounded-lg shrink-0">
                                    <AlertCircle className="h-5 w-5 text-indigo-400" />
                                </div>
                                <div>
                                    <h3 className="text-sm font-medium text-slate-200">Context Menu Integration</h3>
                                    <p className="text-xs text-slate-400 mt-0.5">Install the right-click menu option for quick access.</p>
                                </div>
                            </div>
                            <button
                                onClick={handleInstall}
                                disabled={isInstalling}
                                className={`px-4 py-2 text-xs font-semibold text-white rounded-lg transition-all flex items-center gap-2 shadow-sm ${
                                    isInstalling
                                        ? "bg-indigo-500/50 cursor-wait"
                                        : "bg-indigo-600 hover:bg-indigo-500 hover:shadow-indigo-500/20"
                                }`}
                            >
                                {isInstalling && <Loader2 className="w-3 h-3 animate-spin" />}
                                {isInstalling ? "Installing..." : "Install Integration"}
                            </button>
                        </div>
                    )}
                    <DropZone />
                    <FileList files={files} onRemove={handleRemove} onCopy={handleCopy} onClearCompleted={handleClearCompleted} />
                </div>
            )}
            {view === 'settings' && <SettingsView isInstalled={isInstalled} onStatusChange={setIsInstalled} />}
        </Layout>
    );
}

export default App;
