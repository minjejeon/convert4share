import React, { useId } from 'react';
import { UploadCloud } from 'lucide-react';
import { cn } from '../lib/utils';
import { SelectFiles } from '@bindings/services/tools/service';

interface DropZoneProps {
    onFilesAdded?: (files: string[]) => void;
    isCompact?: boolean;
}

// Wails v3 captures file drag/drop at the native window level and adds
// `.file-drop-target-active` to every `data-file-drop-target` element
// during a drag. The webview's HTML5 drag events do NOT fire, so all
// active-state styling lives in index.css under that class selector.
const baseClasses = "dropzone-cue group relative overflow-hidden border border-dashed rounded-xl transition-all duration-300 ease-out cursor-pointer focus:outline-none focus:ring-2 focus:ring-indigo-500 border-slate-300/60 dark:border-slate-700/40 bg-slate-100/50 dark:bg-slate-800/20 hover:border-indigo-500/20 hover:bg-slate-200/50 dark:hover:bg-slate-800/40";

export function DropZone({ onFilesAdded, isCompact = false }: DropZoneProps) {
    const descriptionId = useId();

    const handleClick = async () => {
        if (!onFilesAdded) return;
        const paths = await SelectFiles();
        if (paths && paths.length > 0) {
            onFilesAdded(paths);
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            handleClick();
        }
    };

    if (isCompact) {
        return (
            <div
                role="button"
                tabIndex={0}
                aria-label="Upload files"
                aria-describedby={descriptionId}
                data-file-drop-target
                onClick={handleClick}
                onKeyDown={handleKeyDown}
                className={cn(baseClasses, "px-4 py-3")}
            >
                <div className="relative z-10 flex items-center gap-4">
                    <div className="dropzone-icon-wrap shrink-0 p-2.5 rounded-full bg-white dark:bg-slate-800/50 ring-1 ring-slate-200 dark:ring-slate-700/50 shadow-sm transition-all duration-300 group-hover:scale-105 group-hover:ring-indigo-500/20">
                        <UploadCloud className="dropzone-icon w-5 h-5 text-slate-400 group-hover:text-indigo-500 dark:group-hover:text-indigo-300 transition-colors duration-300" />
                    </div>
                    <div className="flex flex-col sm:flex-row sm:items-baseline sm:gap-2">
                        <h3 className="text-sm font-medium text-slate-700 dark:text-slate-200 group-hover:text-slate-900 dark:group-hover:text-white transition-colors">
                            Drag &amp; drop or click to browse
                        </h3>
                        <p id={descriptionId} className="text-xs text-slate-600 group-hover:text-slate-700 dark:group-hover:text-slate-400 transition-colors">
                            Support for <span className="font-medium text-indigo-600/80 dark:text-indigo-400/80">.mov</span> and <span className="font-medium text-indigo-600/80 dark:text-indigo-400/80">.heic</span>
                        </p>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div
            role="button"
            tabIndex={0}
            aria-label="Upload files"
            aria-describedby={descriptionId}
            data-file-drop-target
            onClick={handleClick}
            onKeyDown={handleKeyDown}
            className={cn(baseClasses, "p-10 text-center")}
        >
            <div className="relative z-10 flex flex-col items-center justify-center gap-4">
                <div className="dropzone-icon-wrap p-4 rounded-full bg-white dark:bg-slate-800/50 ring-1 ring-slate-200 dark:ring-slate-700/50 shadow-sm transition-all duration-300 group-hover:scale-105 group-hover:ring-indigo-500/20">
                    <UploadCloud className="dropzone-icon w-8 h-8 text-slate-400 group-hover:text-indigo-500 dark:group-hover:text-indigo-300 transition-colors duration-300" />
                </div>
                <div className="space-y-1">
                    <h3 className="text-base font-medium text-slate-700 dark:text-slate-200 group-hover:text-slate-900 dark:group-hover:text-white transition-colors">
                        Drag &amp; drop files or click to browse
                    </h3>
                    <p id={descriptionId} className="text-slate-600 text-sm group-hover:text-slate-700 dark:group-hover:text-slate-400 transition-colors">
                        Support for <span className="font-medium text-indigo-600/80 dark:text-indigo-400/80">.mov</span> and <span className="font-medium text-indigo-600/80 dark:text-indigo-400/80">.heic</span>
                    </p>
                </div>
            </div>
        </div>
    );
}
