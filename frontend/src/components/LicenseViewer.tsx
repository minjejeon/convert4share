import React, { useState, useMemo } from 'react';
import { X, Search, ExternalLink, Code2, Box } from 'lucide-react';
import licensesData from '../assets/licenses.json';

interface License {
    name: string;
    version: string;
    license: string | string[];
    repository?: string;
    text: string;
    source: string;
}

interface LicenseViewerProps {
    onClose: () => void;
}

export function LicenseViewer({ onClose }: LicenseViewerProps) {
    const [search, setSearch] = useState('');
    const [selectedLicense, setSelectedLicense] = useState<License | null>(null);

    const filteredLicenses = useMemo(() => {
        const lowerSearch = search.toLowerCase();
        return (licensesData as License[]).filter(l =>
            l.name.toLowerCase().includes(lowerSearch) ||
            (typeof l.license === 'string' && l.license.toLowerCase().includes(lowerSearch))
        );
    }, [search]);

    // Group by source (Frontend vs Backend) or just list them?
    // A simple alphabetical list is fine, maybe indicating source.

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
            <div className="absolute inset-0 bg-slate-900/60 backdrop-blur-sm transition-opacity" onClick={onClose} />

            <div className="relative bg-white dark:bg-slate-900 w-full max-w-4xl max-h-[90vh] rounded-xl shadow-2xl flex flex-col border border-slate-200 dark:border-slate-800">
                {/* Header */}
                <div className="flex items-center justify-between p-4 border-b border-slate-200 dark:border-slate-800 shrink-0">
                    <div>
                        <h2 className="text-lg font-bold text-slate-900 dark:text-slate-100">Third-Party Licenses</h2>
                        <p className="text-sm text-slate-500 dark:text-slate-400">Open source software used in this application.</p>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-2 text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                <div className="flex flex-1 min-h-0">
                    {/* Sidebar / List */}
                    <div className="w-1/3 border-r border-slate-200 dark:border-slate-800 flex flex-col min-w-[250px]">
                        <div className="p-3 border-b border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/50">
                            <div className="relative">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                                <input
                                    type="text"
                                    placeholder="Search packages..."
                                    value={search}
                                    onChange={(e) => setSearch(e.target.value)}
                                    className="w-full pl-9 pr-3 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:text-slate-200"
                                />
                            </div>
                        </div>
                        <div className="overflow-y-auto flex-1 p-2 space-y-1">
                            {filteredLicenses.map((pkg, idx) => (
                                <button
                                    key={`${pkg.name}-${pkg.version}-${idx}`}
                                    onClick={() => setSelectedLicense(pkg)}
                                    className={`w-full text-left p-3 rounded-lg text-sm transition-colors flex flex-col gap-1 ${
                                        selectedLicense === pkg
                                            ? 'bg-indigo-50 dark:bg-indigo-900/20 ring-1 ring-indigo-200 dark:ring-indigo-800'
                                            : 'hover:bg-slate-50 dark:hover:bg-slate-800/50'
                                    }`}
                                >
                                    <div className="font-medium text-slate-900 dark:text-slate-200 truncate" title={pkg.name}>
                                        {pkg.name}
                                    </div>
                                    <div className="flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400">
                                        <span className="flex items-center gap-1">
                                            {pkg.source === 'backend' ? <Box className="w-3 h-3" /> : <Code2 className="w-3 h-3" />}
                                            {pkg.version}
                                        </span>
                                        <span className="truncate max-w-[80px] bg-slate-100 dark:bg-slate-800 px-1.5 rounded text-[10px] border border-slate-200 dark:border-slate-700">
                                            {Array.isArray(pkg.license) ? pkg.license.join(', ') : pkg.license}
                                        </span>
                                    </div>
                                </button>
                            ))}
                            {filteredLicenses.length === 0 && (
                                <div className="p-4 text-center text-sm text-slate-500">No packages found.</div>
                            )}
                        </div>
                    </div>

                    {/* Content Area */}
                    <div className="flex-1 overflow-y-auto bg-slate-50/50 dark:bg-black/20 p-6">
                        {selectedLicense ? (
                            <div className="space-y-6">
                                <div className="flex items-start justify-between">
                                    <div>
                                        <h3 className="text-xl font-bold text-slate-900 dark:text-slate-100 break-all">
                                            {selectedLicense.name}
                                        </h3>
                                        <p className="text-sm text-slate-500 dark:text-slate-400 mt-1 font-mono">
                                            v{selectedLicense.version} • {selectedLicense.source === 'backend' ? 'Go Module' : 'NPM Package'}
                                        </p>
                                    </div>
                                    {selectedLicense.repository && (
                                        <a
                                            href={selectedLicense.repository}
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            className="flex items-center gap-1.5 text-sm text-indigo-600 dark:text-indigo-400 hover:underline shrink-0"
                                        >
                                            <ExternalLink className="w-4 h-4" />
                                            Repository
                                        </a>
                                    )}
                                </div>

                                <div className="bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-lg overflow-hidden shadow-sm">
                                    <div className="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-800 px-4 py-2 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                                        License Text
                                    </div>
                                    <div className="p-4 overflow-x-auto">
                                        <pre className="text-xs font-mono text-slate-700 dark:text-slate-300 whitespace-pre-wrap break-words leading-relaxed">
                                            {selectedLicense.text || "No license text found."}
                                        </pre>
                                    </div>
                                </div>
                            </div>
                        ) : (
                            <div className="h-full flex flex-col items-center justify-center text-slate-400 dark:text-slate-600 p-8 text-center">
                                <div className="w-16 h-16 rounded-full bg-slate-100 dark:bg-slate-800/50 flex items-center justify-center mb-4">
                                    <Box className="w-8 h-8 opacity-50" />
                                </div>
                                <p className="text-sm">Select a package from the list to view its license details.</p>
                            </div>
                        )}
                    </div>
                </div>

                {/* Footer */}
                <div className="p-3 bg-slate-50 dark:bg-slate-900/50 border-t border-slate-200 dark:border-slate-800 text-right text-xs text-slate-400 dark:text-slate-500 rounded-b-xl">
                    Convert4Share • Generated Licenses
                </div>
            </div>
        </div>
    );
}
