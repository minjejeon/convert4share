// Manual event payload definitions. Wails v3 alpha does not generate
// types for custom event payloads, so these mirror the Go structs that
// EmitEvent/emit() call sites use. Keep these in sync with
// internal/services/jobs/service.go (JobStatus) and the
// `files-dropped` forwarder in cmd/convert4share/main.go.

export type JobStatusName = 'queued' | 'pending' | 'processing' | 'done' | 'error';

export interface JobStatus {
    id: string;
    file: string;
    destFile?: string;
    status: JobStatusName;
    progress: number;
    speed?: string;
    error?: string;
}

export interface FilesDroppedDetails {
    x: number;
    y: number;
    elementId: string;
    classList: string[];
    attributes: Record<string, string>;
}

export interface FilesDroppedPayload {
    files: string[];
    details?: FilesDroppedDetails | null;
}
