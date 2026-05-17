// Package jobs hosts the Wails v3 service that owns the conversion
// queue, job lifecycle, and progress events. Phase 2.1 only provides
// the empty skeleton so the application can boot under v3; Phase 3
// re-implements the v2 methods (AddFiles, ConvertFiles, CancelJob,
// PauseQueue, ResumeQueue) on top of this struct.
package jobs

// Service is the v3 service binding for conversion jobs.
type Service struct{}

// New constructs an empty Service. State and dependencies will be
// wired in Phase 3.
func New() *Service {
	return &Service{}
}
