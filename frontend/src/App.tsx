import { useState } from 'react';
import { Layout } from './components/Layout';
import { DropZone } from './components/DropZone';
import { FileList } from './components/FileList';
import { SettingsView } from './components/Settings';
import { useTheme } from './hooks/useTheme';
import { useFileQueue } from './hooks/useFileQueue';

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
                    <SettingsView theme={theme} onThemeChange={setTheme} />
                </div>
            )}
        </Layout>
    );
}

export default App;
