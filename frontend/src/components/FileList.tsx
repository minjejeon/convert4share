import React, { useState } from 'react';
import { Trash2, Pause, Play, ArrowDown, ArrowUp } from 'lucide-react';
import { FileItemRow, FileItem } from './FileItemRow';

interface FileListProps {
    files: FileItem[];
    onRemove: (id: string) => void;
    onCopy: (path: string) => void;
    onClearCompleted: () => void;
    isPaused?: boolean;
    onPause?: () => void;
    onResume?: () => void;
}

const Header = ({ title, count, children }: { title: string; count: number; children: React.ReactNode }) => (
    <div className="flex items-center justify-between px-2 pb-2 pt-2">
        <h2 className="text-xs font-bold text-slate-500 uppercase tracking-widest flex items-center gap-2">
            {title} <span className="bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 px-1.5 py-0.5 rounded text-[10px] min-w-[20px] text-center">{count}</span>
        </h2>
        {children}
    </div>
);

export function FileList({ files, onRemove, onCopy, onClearCompleted, isPaused, onPause, onResume }: FileListProps) {
    const activeFiles = files.filter(f => f.status !== 'done');
    const [sortField, setSortField] = useState<'name' | 'added' | 'completed'>('completed');
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');

    const completedFiles = files.filter(f => f.status === 'done').sort((a, b) => {
        let cmp = 0;
        switch (sortField) {
            case 'name':
                cmp = a.path.localeCompare(b.path);
                break;
            case 'added':
                cmp = (a.addedAt || 0) - (b.addedAt || 0);
                break;
            case 'completed':
                cmp = (a.completedAt || 0) - (b.completedAt || 0);
                break;
        }
        return sortDirection === 'asc' ? cmp : -cmp;
    });

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
                        <div className="flex items-center gap-2">
                             <div className="flex items-center bg-slate-100 dark:bg-slate-800 rounded-lg p-0.5 border border-slate-200 dark:border-slate-700/50">
                                <select
                                    value={sortField}
                                    onChange={(e) => setSortField(e.target.value as 'name' | 'added' | 'completed')}
                                    className="bg-transparent text-[10px] font-medium uppercase tracking-wide text-slate-600 dark:text-slate-300 px-2 py-0.5 outline-none border-none cursor-pointer hover:bg-white/50 dark:hover:bg-black/20 rounded"
                                    aria-label="Sort files by"
                                >
                                    <option className="bg-slate-50 dark:bg-slate-900 text-slate-700 dark:text-slate-200" value="completed">Completed Time</option>
                                    <option className="bg-slate-50 dark:bg-slate-900 text-slate-700 dark:text-slate-200" value="added">Created Time</option>
                                    <option className="bg-slate-50 dark:bg-slate-900 text-slate-700 dark:text-slate-200" value="name">Name</option>
                                </select>
                                <button
                                    onClick={() => setSortDirection(prev => prev === 'asc' ? 'desc' : 'asc')}
                                    className="p-1 hover:bg-white dark:hover:bg-slate-700 rounded text-slate-500 hover:text-indigo-500 transition-colors"
                                    title={sortDirection === 'asc' ? "Ascending" : "Descending"}
                                    aria-label={sortDirection === 'asc' ? "Sort ascending" : "Sort descending"}
                                >
                                    {sortDirection === 'asc' ? <ArrowUp className="w-3 h-3" /> : <ArrowDown className="w-3 h-3" />}
                                </button>
                             </div>

                            <button
                                onClick={onClearCompleted}
                                className="text-[10px] font-medium text-slate-500 hover:text-red-500 dark:hover:text-red-400 transition-colors flex items-center gap-1.5 uppercase tracking-wide px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800/50 rounded"
                                aria-label="Clear completed files history"
                            >
                                <Trash2 className="w-3 h-3" aria-hidden="true" /> Clear History
                            </button>
                        </div>
                    </Header>
                    {completedFiles.map(file => (
                        <FileItemRow key={file.id} file={file} onRemove={onRemove} onCopy={onCopy} />
                    ))}
                </div>
             )}
        </div>
    );
}
