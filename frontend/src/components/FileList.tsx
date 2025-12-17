// @ts-nocheck
import React, { memo, useState } from 'react';
import { FileVideo, FileImage, AlertCircle, CheckCircle2, Loader2, XCircle, Copy, Trash2, Check, Pause, Play } from 'lucide-react';
import { cn } from '../lib/utils';

export interface FileItem {
    id: string; // usually path
    path: string;
    destFile?: string;
    status: 'queued' | 'processing' | 'done' | 'error';
    progress: number;
    error?: string;
    thumbnail?: string;
}

interface FileListProps {
    files: FileItem[];
    onRemove: (id: string) => void;
    onCopy: (path: string) => void;
    onClearCompleted: () => void;
    isPaused?: boolean;
    onPause?: () => void;
    onResume?: () => void;
}

const FileItemRow = memo(({ file, onRemove, onCopy }: { file: FileItem; onRemove: (id: string) => void; onCopy: (path: string) => void }) => {
    const [isCopied, setIsCopied] = useState(false);
    const lastSeparatorIndex = Math.max(file.path.lastIndexOf('/'), file.path.lastIndexOf('\\'));
    const fileName = lastSeparatorIndex >= 0 ? file.path.substring(lastSeparatorIndex + 1) : file.path;
    const dirName = lastSeparatorIndex >= 0 ? file.path.substring(0, lastSeparatorIndex) : '';

    const handleCopy = () => {
        onCopy(file.destFile!);
        setIsCopied(true);
        setTimeout(() => setIsCopied(false), 2000);
    };

    return (
        <div className="px-1 pb-2">
            <div
                className="group flex items-center gap-4 bg-white dark:bg-slate-800/40 hover:bg-slate-50 dark:hover:bg-slate-800/60 rounded-xl p-4 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-all duration-200 h-full shadow-sm dark:shadow-none"
            >
                {/* Thumbnail / Icon */}
                <div className="shrink-0 relative overflow-hidden w-12 h-12 rounded-lg bg-slate-100 dark:bg-slate-900/50 ring-1 ring-slate-900/5 dark:ring-white/5 flex items-center justify-center">
                    {file.thumbnail ? (
                        <img src={file.thumbnail} alt={fileName} className="w-full h-full object-cover" />
                    ) : (
                        fileName.toLowerCase().endsWith('.mov') ? (
                            <FileVideo className="w-6 h-6 text-indigo-500/80 dark:text-indigo-400/80" />
                        ) : (
                            <FileImage className="w-6 h-6 text-purple-500/80 dark:text-purple-400/80" />
                        )
                    )}
                </div>

                {/* Info & Progress */}
                <div className="min-w-0 flex-1 flex flex-col justify-center gap-2">
                    <div className="flex items-center justify-between gap-4">
                        <div className="min-w-0 flex-1">
                            <div className="flex items-baseline gap-2">
                                <h3 className="text-sm font-medium text-slate-700 dark:text-slate-200 truncate" title={fileName}>
                                    {fileName}
                                </h3>
                                {dirName && (
                                    <span className="text-[10px] text-slate-500 truncate max-w-[150px]" title={dirName}>
                                        {dirName}
                                    </span>
                                )}
                            </div>
                        </div>

                        <span className={cn(
                            "text-[10px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded-full shrink-0",
                            file.status === 'done' && "text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-500/10",
                            file.status === 'processing' && "text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-500/10",
                            file.status === 'error' && "text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-500/10",
                            file.status === 'queued' && "text-slate-500 bg-slate-100 dark:bg-slate-700/50",
                        )}>
                            {file.status === 'queued' ? 'Waiting' : file.status}
                        </span>
                    </div>

                    <div className="relative">
                        <div
                            className="h-1.5 w-full bg-slate-100 dark:bg-slate-700/50 rounded-full overflow-hidden"
                            role="progressbar"
                            aria-valuenow={file.progress}
                            aria-valuemin={0}
                            aria-valuemax={100}
                        >
                            <div
                                className={cn(
                                    "h-full transition-all duration-300 ease-out rounded-full",
                                    file.status === 'error' ? "bg-red-500" : "bg-indigo-500",
                                    file.status === 'done' && "bg-emerald-500"
                                )}
                                style={{ width: `${file.progress}%` }}
                            />
                        </div>
                        {file.error && (
                            <div className="mt-2 w-full text-xs text-red-500 dark:text-red-400 flex items-start gap-1.5 animate-in fade-in" title={file.error}>
                                <AlertCircle className="w-3 h-3 mt-0.5 shrink-0" />
                                <span className="break-all line-clamp-3">
                                    {file.error}
                                </span>
                            </div>
                        )}
                    </div>
                </div>

                {/* Actions */}
                <div className="shrink-0 flex items-center gap-1 pl-2 border-l border-slate-200 dark:border-white/5">
                    {file.status === 'done' && file.destFile ? (
                        <button
                            onClick={handleCopy}
                            className={cn(
                                "p-2 rounded-lg transition-colors",
                                isCopied
                                    ? "text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-500/10"
                                    : "text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-slate-100 dark:hover:bg-slate-700/50"
                            )}
                            title={isCopied ? "Copied!" : "Copy File"}
                            aria-label="Copy file to clipboard"
                        >
                            {isCopied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                        </button>
                    ) : (
                        <div className="w-8 h-8" />
                    )}

                    <button
                        onClick={() => onRemove(file.id)}
                        className="p-2 hover:bg-slate-100 dark:hover:bg-slate-700/50 rounded-lg text-slate-400 hover:text-red-500 dark:hover:text-red-400 transition-colors"
                        title="Remove"
                        aria-label="Remove file from queue"
                    >
                        <Trash2 className="w-4 h-4" />
                    </button>

                    <div className="w-8 h-8 flex items-center justify-center text-slate-500 ml-1">
                        {file.status === 'processing' && <Loader2 className="w-4 h-4 animate-spin text-indigo-500 dark:text-indigo-400" />}
                        {file.status === 'done' && <CheckCircle2 className="w-5 h-5 text-emerald-500" />}
                        {file.status === 'error' && <XCircle className="w-5 h-5 text-red-500" />}
                    </div>
                </div>
            </div>
        </div>
    );
});

FileItemRow.displayName = 'FileItemRow';

const Header = ({ title, count, children }: any) => (
    <div className="flex items-center justify-between px-2 pb-2 pt-2">
        <h2 className="text-xs font-bold text-slate-500 uppercase tracking-widest flex items-center gap-2">
            {title} <span className="bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 px-1.5 py-0.5 rounded text-[10px] min-w-[20px] text-center">{count}</span>
        </h2>
        {children}
    </div>
);

export function FileList({ files, onRemove, onCopy, onClearCompleted, isPaused, onPause, onResume }: FileListProps) {
    const activeFiles = files.filter(f => f.status !== 'done');
    const completedFiles = files.filter(f => f.status === 'done');

    if (files.length === 0) return null;

    return (
        <div className="h-full w-full overflow-y-auto custom-scrollbar">
             {activeFiles.length > 0 && (
                <div className="mb-4">
                    <Header title="Queue" count={activeFiles.length}>
                        <button
                            onClick={isPaused ? onResume : onPause}
                            className="flex items-center gap-1.5 text-[10px] font-medium uppercase tracking-wide px-2 py-1 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded text-indigo-600 dark:text-indigo-400 transition-colors"
                        >
                            {isPaused ? <Play className="w-3 h-3" /> : <Pause className="w-3 h-3" />}
                            {isPaused ? "Resume" : "Pause"}
                        </button>
                    </Header>
                    {activeFiles.map(file => (
                        <FileItemRow key={file.id} file={file} onRemove={onRemove} onCopy={onCopy} />
                    ))}
                </div>
             )}

             {completedFiles.length > 0 && (
                <div className="mb-4">
                    <Header title="Completed" count={completedFiles.length}>
                        <button
                            onClick={onClearCompleted}
                            className="text-[10px] font-medium text-slate-500 hover:text-red-500 dark:hover:text-red-400 transition-colors flex items-center gap-1.5 uppercase tracking-wide px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800/50 rounded"
                        >
                            <Trash2 className="w-3 h-3" /> Clear History
                        </button>
                    </Header>
                    {completedFiles.map(file => (
                        <FileItemRow key={file.id} file={file} onRemove={onRemove} onCopy={onCopy} />
                    ))}
                </div>
             )}
        </div>
    );
}
