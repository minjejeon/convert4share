import { useState, useCallback, useRef, useEffect } from 'react';
import { EventsOn, EventsEmit } from '../wailsjs/runtime/runtime';
import { ConvertFiles, GetThumbnail, CancelJob, PauseQueue, ResumeQueue, CopyFileToClipboard } from '../wailsjs/go/main/App';
import { FileItem } from '../components/FileItemRow';

interface ProgressData {
    file: string;
    destFile?: string;
    status: 'queued' | 'pending' | 'processing' | 'done' | 'error';
    progress: number;
    speed?: string;
    error?: string;
}

export function useFileQueue() {
    const [files, setFiles] = useState<FileItem[]>([]);
    const [isPaused, setIsPaused] = useState<boolean>(false);
    const filesRef = useRef(files);
    filesRef.current = files;

    const addFile = useCallback((path: string) => {
        // Prevent duplicates
        if (filesRef.current.some(f => f.path === path)) return;

        setFiles(prev => {
            if (prev.some(f => f.path === path)) return prev;
            return [...prev, { id: path, path, status: 'queued', progress: 0, addedAt: Date.now() }];
        });
    }, []);

    const [visiblePaths, setVisiblePaths] = useState<string[]>([]);
    const [fetchingPath, setFetchingPath] = useState<string | null>(null);

    const trackVisibility = useCallback((path: string, isVisible: boolean) => {
        setVisiblePaths(prev => {
            if (isVisible) {
                if (prev.includes(path)) return prev;
                return [...prev, path];
            } else {
                return prev.filter(p => p !== path);
            }
        });
    }, []);

    // Process thumbnail queue
    useEffect(() => {
        if (fetchingPath) return;

        // Find the first visible path that doesn't have a thumbnail yet
        const nextPath = visiblePaths.find(path => {
            const file = filesRef.current.find(f => f.path === path);
            return file && !file.thumbnail;
        });

        if (nextPath) {
            setFetchingPath(nextPath);
            GetThumbnail(nextPath).then(thumb => {
                setFiles(prev => prev.map(f => f.path === nextPath ? { ...f, thumbnail: thumb } : f));
            }).catch(err => {
                console.error("Failed to load thumbnail for", nextPath, err);
            }).finally(() => {
                setFetchingPath(null);
            });
        }
    }, [visiblePaths, fetchingPath]);

    const handleRemove = useCallback((id: string) => {
        CancelJob(id);
        setFiles(prev => prev.filter(f => f.id !== id));
        setVisiblePaths(prev => prev.filter(p => p !== id));
    }, []);

    const handleRetry = useCallback((id: string) => {
        setFiles(prev => prev.map(f => 
            f.id === id ? { ...f, status: 'queued', progress: 0, error: undefined, speed: undefined } : f
        ));
    }, []);

    const handleClearCompleted = useCallback(() => {
        setFiles(prev => {
            const completedIds = prev.filter(f => f.status === 'done').map(f => f.id);
            setVisiblePaths(v => v.filter(p => !completedIds.includes(p)));
            return prev.filter(f => f.status !== 'done');
        });
    }, []);

    const handleCopy = useCallback((path: string) => {
        CopyFileToClipboard(path).catch(console.error);
    }, []);

    useEffect(() => {
        const cleanupFileAdded = EventsOn("file-added", (path: string) => {
            addFile(path);
        });

        const cleanupFilesReceived = EventsOn("files-received", (paths: string[]) => {
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

        const cleanupPaused = EventsOn("queue-paused", () => setIsPaused(true));
        const cleanupResumed = EventsOn("queue-resumed", () => setIsPaused(false));

        EventsEmit("frontend-ready");

        return () => {
            cleanupFileAdded();
            cleanupFilesReceived();
            cleanupProgress();
            cleanupPaused();
            cleanupResumed();
        };
    }, [addFile]);

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

    return {
        files,
        addFile,
        trackVisibility,
        handleRemove,
        handleRetry,
        handleClearCompleted,
        handleCopy,
        isPaused,
        pauseQueue: PauseQueue,
        resumeQueue: ResumeQueue
    };
}
