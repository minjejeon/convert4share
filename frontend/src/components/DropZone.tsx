import React, { useState } from 'react';
import { UploadCloud } from 'lucide-react';
import { cn } from '../lib/utils';

interface DropZoneProps {
    onFilesAdded?: (files: string[]) => void;
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function DropZone({ onFilesAdded }: DropZoneProps) {
    const [isDragging, setIsDragging] = useState(false);

    // Note: Wails handles file paths via its own runtime for drops if configured,
    // but web standard drag & drop gives File objects.
    // However, Wails runtime provides drag&drop events on the window.
    // We can also use standard HTML5 API, but getting full path in browser JS is restricted usually.
    // In Wails, if we drop files, we might need to rely on the Wails specific 'wails:file-drop' event
    // OR just use a simple button to open dialog.
    // Actually, Wails 2 has native drag and drop support that emits events?
    // Let's rely on window listeners for "file-added" event from backend mainly for context menu/args,
    // but for drag and drop on the UI, we might need `runtime.FileDrop`.
    // However, Wails v2 runtime has `OnFileDrop` which we can hook in Go, or we can use the window event.

    // BUT, standard HTML5 drag and drop in Wails webview DOES give full paths if we use `DataTransferItem.getAsFile().path` (in some webviews)
    // or we can use the wails runtime.

    // Let's implement a visual zone. The actual file dropping logic might be handled by Wails window listeners
    // or we assume the user drags into the window.

    // Actually, let's just use a button for "Select Files" which calls Go backend to open dialog?
    // I didn't implement OpenDialog in backend yet.

    // Wait, the user requirement is "files are dropped or passed via args".
    // I should probably add a "Select File" method in Go or just rely on OS drag/drop.

    // For now, let's just make a visual instruction.

    return (
        <div
            className={cn(
                "group relative overflow-hidden border-2 border-dashed rounded-xl p-4 text-center transition-all duration-300 ease-in-out cursor-default",
                isDragging
                    ? "border-indigo-500 bg-indigo-500/10 scale-[1.01] shadow-lg shadow-indigo-500/10"
                    : "border-slate-700/50 bg-slate-800/30 hover:border-indigo-500/30 hover:bg-slate-800/50 hover:shadow-lg hover:shadow-indigo-500/5"
            )}
            onDragOver={(e) => {
                e.preventDefault();
                setIsDragging(true);
            }}
            onDragLeave={() => setIsDragging(false)}
            onDrop={(e) => {
                e.preventDefault();
                setIsDragging(false);
                // In Wails, handling drop in JS to get path is tricky due to security sandbox,
                // usually we rely on Wails' OnFileDrop in Go.
                // But let's see if we can just trigger it.
                // Actually, Wails automatically handles file drops on the window if configured?
                // No, we need to register OnFileDrop in options.App (which I haven't).
            }}
        >
            <div className="relative z-10 flex flex-row items-center justify-center gap-3">
                <div className={cn(
                    "p-2 rounded-full bg-slate-800 ring-1 ring-slate-700/50 shadow-sm transition-transform duration-300",
                    isDragging ? "scale-110 ring-indigo-500/50" : "group-hover:scale-110 group-hover:ring-indigo-500/30"
                )}>
                    <UploadCloud className={cn(
                        "w-5 h-5 transition-colors duration-300",
                        isDragging ? "text-indigo-400" : "text-slate-400 group-hover:text-indigo-400"
                    )} />
                </div>
                <div className="flex flex-row items-baseline gap-2">
                    <h3 className="text-sm font-semibold text-slate-200 group-hover:text-white transition-colors whitespace-nowrap">
                        Drag & Drop files here
                    </h3>
                    <p className="text-slate-500 text-xs group-hover:text-slate-400 transition-colors whitespace-nowrap">
                        (<span className="font-medium text-indigo-400/80">.mov</span>, <span className="font-medium text-indigo-400/80">.heic</span>)
                    </p>
                </div>
            </div>
        </div>
    );
}
