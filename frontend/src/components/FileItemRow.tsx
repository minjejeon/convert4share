import React, { memo, useState } from 'react';
import { FileVideo, FileImage, AlertCircle, CheckCircle2, Loader2, XCircle, Copy, Trash2, Check } from 'lucide-react';
import { cn } from '../lib/utils';

export interface FileItem {
    id: string; // usually path
    path: string;
    destFile?: string;
    status: 'queued' | 'processing' | 'done' | 'error';
    progress: number;
    speed?: string;
    error?: string;
    thumbnail?: string;
    addedAt?: number;
    completedAt?: number;
}

export const FileItemRow = memo(({ file, onRemove, onCopy }: { file: FileItem; onRemove: (id: string) => void; onCopy: (path: string) => void }) => {
    const [isCopied, setIsCopied] = useState(false);
    const [isErrorCopied, setIsErrorCopied] = useState(false);
    const lastSeparatorIndex = Math.max(file.path.lastIndexOf('/'), file.path.lastIndexOf('\\'));
    const fileName = lastSeparatorIndex >= 0 ? file.path.substring(lastSeparatorIndex + 1) : file.path;
    const dirName = lastSeparatorIndex >= 0 ? file.path.substring(0, lastSeparatorIndex) : '';

    const handleCopy = () => {
        onCopy(file.destFile!);
        setIsCopied(true);
        setTimeout(() => setIsCopied(false), 2000);
    };

    const handleCopyError = (errorText: string) => {
        navigator.clipboard.writeText(errorText).then(() => {
            setIsErrorCopied(true);
            setTimeout(() => setIsErrorCopied(false), 2000);
        }).catch(console.error);
    };

    return (
        <div className="px-1 pb-2">
            <div
                className="group flex items-center gap-4 bg-white dark:bg-slate-800/40 hover:bg-slate-50 dark:hover:bg-slate-800/60 rounded-xl p-4 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-all duration-200 h-full shadow-sm dark:shadow-none"
            >
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
                            {file.status === 'processing' && file.speed && <span className="normal-case ml-1 opacity-75">({file.speed})</span>}
                        </span>
                    </div>

                    <div className="relative">
                        <div
                            className="h-1.5 w-full bg-slate-100 dark:bg-slate-700/50 rounded-full overflow-hidden"
                            role="progressbar"
                            aria-valuenow={file.progress}
                            aria-valuemin={0}
                            aria-valuemax={100}
                            aria-label={`Progress: ${Math.round(file.progress)}%`}
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
                            <div className="mt-2 w-full text-xs text-red-500 dark:text-red-400 flex items-start gap-1.5 animate-in fade-in">
                                <AlertCircle className="w-3 h-3 mt-0.5 shrink-0" />
                                <span className="break-all line-clamp-3 flex-1" title={file.error}>
                                    {file.error}
                                </span>
                                <button
                                    onClick={() => handleCopyError(file.error!)}
                                    className="shrink-0 p-1 rounded hover:bg-red-100 dark:hover:bg-red-900/30 text-red-600 dark:text-red-400 transition-colors"
                                    title="Copy Error Log"
                                    aria-label="Copy error log"
                                >
                                    {isErrorCopied ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
                                </button>
                            </div>
                        )}
                    </div>
                </div>

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
