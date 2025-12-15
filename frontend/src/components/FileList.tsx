import React, { memo } from 'react';
import { FileVideo, FileImage, AlertCircle, CheckCircle2, Loader2, XCircle, Copy, Trash2 } from 'lucide-react';
import { cn } from '../lib/utils';

export interface FileItem {
    id: string; // usually path
    path: string;
    destFile?: string;
    status: 'queued' | 'processing' | 'done' | 'error';
    progress: number;
    error?: string;
}

interface FileListProps {
    files: FileItem[];
    onRemove: (id: string) => void;
    onCopy: (path: string) => void;
}

// Optimized: Extract FileItemRow and wrap with React.memo to prevent
// re-rendering of all items when only one item's progress updates.
const FileItemRow = memo(({ file, onRemove, onCopy }: { file: FileItem; onRemove: (id: string) => void; onCopy: (path: string) => void }) => {
    // Extract filename and directory for better UX
    // We handle both Windows (\) and Unix (/) separators
    const lastSeparatorIndex = Math.max(file.path.lastIndexOf('/'), file.path.lastIndexOf('\\'));
    const fileName = lastSeparatorIndex >= 0 ? file.path.substring(lastSeparatorIndex + 1) : file.path;
    const dirName = lastSeparatorIndex >= 0 ? file.path.substring(0, lastSeparatorIndex) : '';

    return (
        <div
            className="bg-slate-800/50 rounded-lg p-4 border border-slate-700/50 flex items-center gap-4 group transition-all"
        >
            <div className="shrink-0 p-2.5 rounded-md bg-slate-800 ring-1 ring-slate-700">
                {fileName.toLowerCase().endsWith('.mov') ? (
                    <FileVideo className="w-5 h-5 text-indigo-400" />
                ) : (
                    <FileImage className="w-5 h-5 text-purple-400" />
                )}
            </div>

            <div className="min-w-0 flex-1">
                <div className="flex items-center justify-between mb-1.5 gap-4">
                    <div className="min-w-0 flex-1 flex flex-col">
                        <h3 className="text-sm font-medium text-slate-200 truncate" title={fileName}>
                            {fileName}
                        </h3>
                        {dirName && (
                             <p className="text-xs text-slate-500 truncate" title={dirName}>
                                {dirName}
                             </p>
                        )}
                    </div>

                    <span className={cn(
                        "text-xs font-medium px-2 py-0.5 rounded-full capitalize shrink-0 self-start mt-0.5",
                        file.status === 'done' && "bg-emerald-500/10 text-emerald-400",
                        file.status === 'processing' && "bg-blue-500/10 text-blue-400",
                        file.status === 'error' && "bg-red-500/10 text-red-400",
                        file.status === 'queued' && "bg-slate-700 text-slate-400",
                    )}>
                        {file.status === 'queued' ? 'Waiting' : file.status}
                    </span>
                </div>

                <div
                    className="h-2 w-full bg-slate-700 rounded-full overflow-hidden"
                    role="progressbar"
                    aria-valuenow={file.progress}
                    aria-valuemin={0}
                    aria-valuemax={100}
                    aria-label={`Progress for ${fileName}`}
                    aria-valuetext={`${file.status}: ${Math.round(file.progress)}%`}
                >
                    <div
                        className={cn(
                            "h-full transition-all duration-300 ease-out",
                            file.status === 'error' ? "bg-red-500" : "bg-indigo-500",
                            file.status === 'done' && "bg-emerald-500"
                        )}
                        style={{ width: `${file.progress}%` }}
                    />
                </div>

                {file.error && (
                    <p className="text-xs text-red-400 mt-1.5 flex items-center gap-1.5">
                        <AlertCircle className="w-3 h-3" />
                        {file.error}
                    </p>
                )}
            </div>

            <div className="shrink-0 flex items-center gap-2">
                {file.status === 'done' && file.destFile && (
                    <button
                        onClick={() => onCopy(file.destFile!)}
                        className="p-1.5 hover:bg-slate-700 rounded-full text-slate-400 hover:text-indigo-400 transition-colors"
                        title="Copy File"
                    >
                        <Copy className="w-4 h-4" />
                    </button>
                )}

                <button
                    onClick={() => onRemove(file.id)}
                    className="p-1.5 hover:bg-slate-700 rounded-full text-slate-400 hover:text-red-400 transition-colors"
                    title="Remove from queue"
                >
                    <Trash2 className="w-4 h-4" />
                </button>

                <div className="w-5 h-5 flex items-center justify-center text-slate-500">
                    {file.status === 'processing' && <Loader2 className="w-5 h-5 animate-spin text-indigo-400" />}
                    {file.status === 'done' && <CheckCircle2 className="w-5 h-5 text-emerald-400" />}
                    {file.status === 'error' && <XCircle className="w-5 h-5 text-red-400" />}
                </div>
            </div>
        </div>
    );
});

FileItemRow.displayName = 'FileItemRow';

export function FileList({ files, onRemove, onCopy }: FileListProps) {
    if (files.length === 0) return null;

    const activeFiles = files.filter(f => f.status !== 'done');
    const completedFiles = files.filter(f => f.status === 'done');

    return (
        <div className="space-y-6 mt-6">
            {activeFiles.length > 0 && (
                <div className="space-y-3">
                     <h2 className="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-4 px-1">
                        Queue ({activeFiles.length})
                    </h2>
                    {activeFiles.map(file => (
                        <FileItemRow key={file.id} file={file} onRemove={onRemove} onCopy={onCopy} />
                    ))}
                </div>
            )}

            {completedFiles.length > 0 && (
                <div className="space-y-3">
                     <h2 className="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-4 px-1">
                        Completed ({completedFiles.length})
                    </h2>
                    {completedFiles.map(file => (
                        <FileItemRow key={file.id} file={file} onRemove={onRemove} onCopy={onCopy} />
                    ))}
                </div>
            )}
        </div>
    );
}
