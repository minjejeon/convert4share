import React, { useState } from 'react';
import { UploadCloud } from 'lucide-react';
import { cn } from '../lib/utils';

interface DropZoneProps {
    onFilesAdded?: (files: string[]) => void;
    isCompact?: boolean;
}

export function DropZone({ onFilesAdded, isCompact = false }: DropZoneProps) {
    const [isDragging, setIsDragging] = useState(false);

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(false);

        if (onFilesAdded && e.dataTransfer.files.length > 0) {
            const paths = Array.from(e.dataTransfer.files).map((f: any) => {
                return f.path ? f.path : f.name;
            });
            const validPaths = paths.filter(p => typeof p === 'string' && p.length > 0);
            if (validPaths.length > 0) {
                onFilesAdded(validPaths);
            }
        }
    };

    if (isCompact) {
        return (
            <div
                className={cn(
                    "group relative overflow-hidden border border-dashed rounded-xl px-4 py-3 transition-all duration-300 ease-out cursor-default",
                    isDragging
                        ? "border-indigo-500/50 bg-indigo-50 dark:bg-indigo-500/5 scale-[1.01] shadow-lg shadow-indigo-500/10"
                        : "border-slate-300/60 dark:border-slate-700/40 bg-slate-100/50 dark:bg-slate-800/20 hover:border-indigo-500/20 hover:bg-slate-200/50 dark:hover:bg-slate-800/40"
                )}
                onDragOver={(e) => {
                    e.preventDefault();
                    setIsDragging(true);
                }}
                onDragLeave={() => setIsDragging(false)}
                onDrop={handleDrop}
            >
                <div className="relative z-10 flex items-center gap-4">
                    <div className={cn(
                        "shrink-0 p-2.5 rounded-full bg-white dark:bg-slate-800/50 ring-1 ring-slate-200 dark:ring-slate-700/50 shadow-sm transition-transform duration-300",
                        isDragging ? "scale-110 ring-indigo-500/50" : "group-hover:scale-105 group-hover:ring-indigo-500/20"
                    )}>
                        <UploadCloud className={cn(
                            "w-5 h-5 transition-colors duration-300",
                            isDragging ? "text-indigo-600 dark:text-indigo-400" : "text-slate-400 group-hover:text-indigo-500 dark:group-hover:text-indigo-300"
                        )} />
                    </div>
                    <div className="flex flex-col sm:flex-row sm:items-baseline sm:gap-2">
                        <h3 className="text-sm font-medium text-slate-700 dark:text-slate-200 group-hover:text-slate-900 dark:group-hover:text-white transition-colors">
                            Drag & Drop files here
                        </h3>
                        <p className="text-xs text-slate-500 group-hover:text-slate-600 dark:group-hover:text-slate-400 transition-colors">
                            Support for <span className="font-medium text-indigo-600/80 dark:text-indigo-400/80">.mov</span> and <span className="font-medium text-indigo-600/80 dark:text-indigo-400/80">.heic</span>
                        </p>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div
            className={cn(
                "group relative overflow-hidden border border-dashed rounded-xl p-10 text-center transition-all duration-300 ease-out cursor-default",
                isDragging
                    ? "border-indigo-500/50 bg-indigo-50 dark:bg-indigo-500/5 scale-[1.01] shadow-lg shadow-indigo-500/10"
                    : "border-slate-300/60 dark:border-slate-700/40 bg-slate-100/50 dark:bg-slate-800/20 hover:border-indigo-500/20 hover:bg-slate-200/50 dark:hover:bg-slate-800/40"
            )}
            onDragOver={(e) => {
                e.preventDefault();
                setIsDragging(true);
            }}
            onDragLeave={() => setIsDragging(false)}
            onDrop={handleDrop}
        >
            <div className="relative z-10 flex flex-col items-center justify-center gap-4">
                <div className={cn(
                    "p-4 rounded-full bg-white dark:bg-slate-800/50 ring-1 ring-slate-200 dark:ring-slate-700/50 shadow-sm transition-transform duration-300",
                    isDragging ? "scale-110 ring-indigo-500/50" : "group-hover:scale-105 group-hover:ring-indigo-500/20"
                )}>
                    <UploadCloud className={cn(
                        "w-8 h-8 transition-colors duration-300",
                        isDragging ? "text-indigo-600 dark:text-indigo-400" : "text-slate-400 group-hover:text-indigo-500 dark:group-hover:text-indigo-300"
                    )} />
                </div>
                <div className="space-y-1">
                    <h3 className="text-base font-medium text-slate-700 dark:text-slate-200 group-hover:text-slate-900 dark:group-hover:text-white transition-colors">
                        Drag & Drop files here
                    </h3>
                    <p className="text-slate-500 text-sm group-hover:text-slate-600 dark:group-hover:text-slate-400 transition-colors">
                        Support for <span className="font-medium text-indigo-600/80 dark:text-indigo-400/80">.mov</span> and <span className="font-medium text-indigo-600/80 dark:text-indigo-400/80">.heic</span>
                    </p>
                </div>
            </div>
        </div>
    );
}
