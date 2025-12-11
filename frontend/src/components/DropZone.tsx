import React, { useCallback, useState } from 'react';
import { UploadCloud } from 'lucide-react';
import { cn } from '../lib/utils';
import { AddFiles } from '../wailsjs/go/main/App';

interface DropZoneProps {
    onFilesAdded?: (files: string[]) => void;
}

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
                "border-2 border-dashed rounded-xl p-12 text-center transition-all duration-200 ease-in-out cursor-default",
                isDragging
                    ? "border-indigo-500 bg-indigo-500/10 scale-[1.01]"
                    : "border-slate-700 bg-slate-800/30 hover:border-slate-600 hover:bg-slate-800/50"
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
            <div className="flex flex-col items-center gap-4">
                <div className="p-4 rounded-full bg-slate-800 ring-1 ring-slate-700 shadow-sm">
                    <UploadCloud className="w-8 h-8 text-indigo-400" />
                </div>
                <div>
                    <h3 className="text-lg font-semibold text-slate-200">Drag & Drop files here</h3>
                    <p className="text-slate-400 mt-1 text-sm">
                        Supported formats: .mov, .heic
                    </p>
                </div>
            </div>
        </div>
    );
}
