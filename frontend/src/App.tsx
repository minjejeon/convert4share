import React, { useEffect, useState } from 'react';
import { EventsOn } from './wailsjs/runtime';
import { ConvertFiles } from './wailsjs/go/main/App';
import { Layout } from './components/Layout';
import { DropZone } from './components/DropZone';
import { FileList, FileItem } from './components/FileList';
import { SettingsView } from './components/Settings';

function App() {
    const [view, setView] = useState<'home' | 'settings'>('home');
    const [files, setFiles] = useState<FileItem[]>([]);

    useEffect(() => {
        // Listen for new files added (e.g. from context menu or args)
        // The payload 'data' is the file path string
        const stopFileAdded = EventsOn("file-added", (path: string) => {
            console.log("File added:", path);
            addFile(path);
        });

        // Listen for batch of files
        const stopFilesReceived = EventsOn("files-received", (paths: string[]) => {
             console.log("Files received:", paths);
             paths.forEach(addFile);
        });

        // Listen for progress updates
        const stopProgress = EventsOn("conversion-progress", (data: any) => {
            setFiles(prev => prev.map(f => {
                // We use path as ID for now
                if (f.id === data.file) {
                    return { ...f, status: data.status, progress: data.progress, error: data.error };
                }
                return f;
            }));
        });

        return () => {
            // Cleanup events if component unmounts (rare for root App)
             // EventsOff("file-added"); // EventsOff expects event name
             // EventsOff("files-received");
             // EventsOff("conversion-progress");
             // Note: Wails JS runtime EventsOff removes ALL listeners for that event
        };
    }, []);

    const addFile = (path: string) => {
        setFiles(prev => {
            if (prev.some(f => f.path === path)) return prev;
            return [...prev, { id: path, path, status: 'queued', progress: 0 }];
        });
    };

    // Auto-start conversion when files are added?
    // Or add a "Start" button?
    // The requirement implies seamless "convert files" action.
    // The previous CLI tool converted immediately.
    // Let's add an effect to convert 'queued' files.
    // However, if we drop 10 files, we don't want to call ConvertFiles 10 times individually if the backend expects a list.
    // But our backend ConvertFiles takes []string.
    // Let's implement a "Start Conversion" button, OR auto-start with a small debounce.
    // Given the "Send to" nature, it should probably just start.

    // Let's modify: When files are added, if they are not processing, trigger conversion for queued files?
    // Actually, simply adding them to the list and then having a button is safer UI,
    // but context menu usage expects immediate action.

    // Let's make it start immediately for now for "Context Menu" feel.

    useEffect(() => {
        const queued = files.filter(f => f.status === 'queued');
        if (queued.length > 0) {
            // Check if we are already processing something?
            // The backend handles concurrency (semaphores).
            // So we can just fire them off.

            // To avoid rapid firing if multiple files added in loop, we can debounce.
            const timeout = setTimeout(() => {
                const queuedPaths = queued.map(f => f.path);
                // Mark them as processing in UI immediately so we don't re-trigger
                setFiles(prev => prev.map(f => queuedPaths.includes(f.path) ? { ...f, status: 'processing', progress: 0 } : f));

                ConvertFiles(queuedPaths);
            }, 100);
            return () => clearTimeout(timeout);
        }
    }, [files]);

    return (
        <Layout currentView={view} onNavigate={setView}>
            {view === 'home' && (
                <div className="max-w-3xl mx-auto pb-12">
                    <DropZone />
                    <FileList files={files} />
                </div>
            )}
            {view === 'settings' && <SettingsView />}
        </Layout>
    );
}

export default App;
