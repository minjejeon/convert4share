import React from 'react';
import { FileVideo, FileImage, AlertCircle, CheckCircle2, Loader2, XCircle } from 'lucide-react';
import { cn } from '../lib/utils';

export interface FileItem {
    id: string; // usually path
    path: string;
    status: 'queued' | 'processing' | 'done' | 'error';
    progress: number;
    error?: string;
}

interface FileListProps {
    files: FileItem[];
}

export function FileList({ files }: FileListProps) {
    if (files.length === 0) return null;

    return (
        <div className="space-y-3 mt-6">
            <h2 className="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-4 px-1">
                Queue ({files.length})
            </h2>
            {files.map((file) => (
                <div
                    key={file.id}
                    className="bg-slate-800/50 rounded-lg p-4 border border-slate-700/50 flex items-center gap-4 group transition-all"
                >
                    <div className="shrink-0 p-2.5 rounded-md bg-slate-800 ring-1 ring-slate-700">
                        {file.path.toLowerCase().endsWith('.mov') ? (
                            <FileVideo className="w-5 h-5 text-indigo-400" />
                        ) : (
                            <FileImage className="w-5 h-5 text-purple-400" />
                        )}
                    </div>

                    <div className="min-w-0 flex-1">
                        <div className="flex items-center justify-between mb-1.5">
                            <h3 className="text-sm font-medium text-slate-200 truncate pr-4" title={file.path}>
                                {file.path}
                            </h3>
                            <span className={cn(
                                "text-xs font-medium px-2 py-0.5 rounded-full capitalize",
                                file.status === 'done' && "bg-emerald-500/10 text-emerald-400",
                                file.status === 'processing' && "bg-blue-500/10 text-blue-400",
                                file.status === 'error' && "bg-red-500/10 text-red-400",
                                file.status === 'queued' && "bg-slate-700 text-slate-400",
                            )}>
                                {file.status}
                            </span>
                        </div>

                        <div className="h-1.5 w-full bg-slate-700 rounded-full overflow-hidden">
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

                    <div className="shrink-0 text-slate-500">
                        {file.status === 'processing' && <Loader2 className="w-5 h-5 animate-spin text-indigo-400" />}
                        {file.status === 'done' && <CheckCircle2 className="w-5 h-5 text-emerald-400" />}
                        {file.status === 'error' && <XCircle className="w-5 h-5 text-red-400" />}
                    </div>
                </div>
            ))}
        </div>
    );
}
