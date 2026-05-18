import React, { useState } from 'react';
import { Trash2, Pause, Play, ArrowDown, ArrowUp } from 'lucide-react';
import { FileItemRow, FileItem } from './FileItemRow';

interface FileListProps {
    files: FileItem[];
    onRemove: (id: string) => void;
    onRetry: (id: string) => void;
    onCopy: (path: string) => void;
    onClearCompleted: () => void;
    trackVisibility: (path: string, isVisible: boolean) => void;
    isPaused?: boolean;
    onPause?: () => void;
    onResume?: () => void;
}

type SortField = 'name' | 'added' | 'completed';

const SORT_LABEL: Record<SortField, string> = {
    completed: 'Completed',
    added: 'Created',
    name: 'Name',
};

const SORT_ORDER: SortField[] = ['completed', 'added', 'name'];

const Header = ({ title, count, children }: { title: string; count: number; children: React.ReactNode }) => (
    <div className="flex items-center justify-between px-2 pb-2 pt-2">
        <h2 className="text-xs font-bold text-slate-500 uppercase tracking-widest flex items-center gap-2">
            {title} <span className="bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 px-1.5 py-0.5 rounded text-[10px] min-w-[20px] text-center">{count}</span>
        </h2>
        {children}
    </div>
);

export function FileList({ files, onRemove, onRetry, onCopy, onClearCompleted, trackVisibility, isPaused, onPause, onResume }: FileListProps) {
    const activeFiles = files.filter(f => f.status !== 'done');
    const [sortField, setSortField] = useState<SortField>('completed');
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');

    const cycleSortField = () => {
        const idx = SORT_ORDER.indexOf(sortField);
        setSortField(SORT_ORDER[(idx + 1) % SORT_ORDER.length]);
    };

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
                            aria-label={isPaused ? "Resume conversion queue" : "Pause conversion queue"}
                            aria-pressed={isPaused}
                        >
                            {isPaused ? <Play className="w-3 h-3" /> : <Pause className="w-3 h-3" />}
                            {isPaused ? "Resume" : "Pause"}
                        </button>
                    </Header>
                    {activeFiles.map(file => (
                        <FileItemRow key={file.id} file={file} onRemove={onRemove} onRetry={onRetry} onCopy={onCopy} trackVisibility={trackVisibility} />
                    ))}
                </div>
             )}

             {completedFiles.length > 0 && (
                <div className="mb-4">
                    <Header title="Completed" count={completedFiles.length}>
                        <div className="flex items-center gap-1.5">
                            <div className="inline-flex items-stretch bg-slate-100 dark:bg-slate-800 rounded-md overflow-hidden border border-slate-200/60 dark:border-slate-700/50">
                                <button
                                    onClick={cycleSortField}
                                    className="text-[10px] font-medium uppercase tracking-wide text-slate-600 dark:text-slate-300 px-2 py-1 hover:bg-white/70 dark:hover:bg-black/20 transition-colors"
                                    title="Cycle sort field"
                                    aria-label={`Sort by ${SORT_LABEL[sortField]} — click to change field`}
                                >
                                    {SORT_LABEL[sortField]}
                                </button>
                                <button
                                    onClick={() => setSortDirection(prev => prev === 'asc' ? 'desc' : 'asc')}
                                    className="px-1.5 border-l border-slate-200/70 dark:border-slate-700/60 text-slate-500 hover:text-indigo-500 hover:bg-white/70 dark:hover:bg-black/20 transition-colors"
                                    title={sortDirection === 'asc' ? 'Ascending — click to flip' : 'Descending — click to flip'}
                                    aria-label={sortDirection === 'asc' ? 'Sort ascending' : 'Sort descending'}
                                >
                                    {sortDirection === 'asc' ? <ArrowUp className="w-3 h-3" /> : <ArrowDown className="w-3 h-3" />}
                                </button>
                            </div>

                            <button
                                onClick={onClearCompleted}
                                className="text-[10px] font-medium text-slate-500 hover:text-red-500 dark:hover:text-red-400 transition-colors flex items-center gap-1.5 uppercase tracking-wide px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800/50 rounded"
                                aria-label="Clear completed files history"
                            >
                                <Trash2 className="w-3 h-3" aria-hidden="true" /> Clear
                            </button>
                        </div>
                    </Header>
                    {completedFiles.map(file => (
                        <FileItemRow key={file.id} file={file} onRemove={onRemove} onRetry={onRetry} onCopy={onCopy} trackVisibility={trackVisibility} />
                    ))}
                </div>
             )}
        </div>
    );
}
