import { lazy, Suspense, useState } from 'react';
import { Loader2 } from 'lucide-react';
import { Layout } from './components/Layout';
import { DropZone } from './components/DropZone';
import { FileList } from './components/FileList';
import { useTheme } from './hooks/useTheme';
import { useFileQueue } from './hooks/useFileQueue';

// Lazy-load the settings screen so the home view ships with a smaller
// initial bundle. Settings pulls in 6 form sub-components, the license
// viewer, and licenses.json — none of which the home view needs.
const SettingsView = lazy(() =>
    import('./components/Settings').then((m) => ({ default: m.SettingsView })),
);

function SettingsFallback() {
    return (
        <div className="flex justify-center p-12">
            <Loader2 className="animate-spin text-slate-500" aria-label="Loading settings" />
        </div>
    );
}

function App() {
    const [view, setView] = useState<'home' | 'settings'>('home');
    const { theme, setTheme } = useTheme();
    const {
        files,
        addFile,
        trackVisibility,
        handleRemove,
        handleRetry,
        handleClearCompleted,
        handleCopy,
        isPaused,
        pauseQueue,
        resumeQueue,
    } = useFileQueue();

    return (
        <Layout currentView={view} onNavigate={setView}>
            {view === 'home' && (
                <div className="max-w-3xl mx-auto h-full flex flex-col gap-6 p-6">
                    <div className="shrink-0">
                        <DropZone
                            onFilesAdded={(paths) => paths.forEach(addFile)}
                            isCompact={files.length > 0}
                        />
                    </div>
                    <div className="flex-1 min-h-0">
                        <FileList
                            files={files}
                            onRemove={handleRemove}
                            onRetry={handleRetry}
                            onCopy={handleCopy}
                            onClearCompleted={handleClearCompleted}
                            trackVisibility={trackVisibility}
                            isPaused={isPaused}
                            onPause={pauseQueue}
                            onResume={resumeQueue}
                        />
                    </div>
                </div>
            )}
            {view === 'settings' && (
                <div className="h-full overflow-y-auto px-6">
                    <Suspense fallback={<SettingsFallback />}>
                        <SettingsView theme={theme} onThemeChange={setTheme} />
                    </Suspense>
                </div>
            )}
        </Layout>
    );
}

export default App;
