import React, { useEffect, useState } from 'react';
import { EventsOn } from './wailsjs/runtime';
import { ConvertFiles, GetContextMenuStatus, InstallContextMenu } from './wailsjs/go/main/App';
import { Layout } from './components/Layout';
import { DropZone } from './components/DropZone';
import { FileList, FileItem } from './components/FileList';
import { SettingsView } from './components/Settings';
import { AlertCircle } from 'lucide-react';

interface ProgressData {
    file: string;
    status: 'queued' | 'processing' | 'done' | 'error';
    progress: number;
    error?: string;
}

function App() {
    const [view, setView] = useState<'home' | 'settings'>('home');
    const [files, setFiles] = useState<FileItem[]>([]);
    const [isInstalled, setIsInstalled] = useState<boolean>(true);

    const addFile = (path: string) => {
        setFiles(prev => {
            if (prev.some(f => f.path === path)) return prev;
            return [...prev, { id: path, path, status: 'queued', progress: 0 }];
        });
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
                    return { ...f, status: data.status, progress: data.progress, error: data.error };
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
        try {
            await InstallContextMenu();
            const interval = setInterval(() => {
                GetContextMenuStatus().then(status => {
                    if (status) {
                        setIsInstalled(true);
                        clearInterval(interval);
                    }
                });
            }, 1000);
            setTimeout(() => clearInterval(interval), 10000);
        } catch (e) {
            console.error(e);
        }
    };

    return (
        <Layout currentView={view} onNavigate={setView}>
            {view === 'home' && (
                <div className="max-w-3xl mx-auto pb-12">
                     {!isInstalled && (
                        <div className="mb-6 rounded-md bg-blue-500/10 p-4 border border-blue-500/20 flex items-center justify-between">
                            <div className="flex items-center gap-3">
                                <AlertCircle className="h-5 w-5 text-blue-400" />
                                <div>
                                    <h3 className="text-sm font-medium text-blue-100">Context Menu Integration</h3>
                                    <p className="text-sm text-blue-300/80">Install the right-click menu option for quick access.</p>
                                </div>
                            </div>
                            <button
                                onClick={handleInstall}
                                className="px-3 py-1.5 text-xs font-medium bg-blue-500 hover:bg-blue-600 text-white rounded-md transition-colors"
                            >
                                Install Now
                            </button>
                        </div>
                    )}
                    <DropZone />
                    <FileList files={files} />
                </div>
            )}
            {view === 'settings' && <SettingsView isInstalled={isInstalled} onStatusChange={setIsInstalled} />}
        </Layout>
    );
}

export default App;
