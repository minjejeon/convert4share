// Package tools hosts the Wails v3 service for ancillary helpers
// (file/binary dialogs, thumbnail generation, clipboard copy, winget
// install, binary detection). Phase 2.1 only provides the empty
// skeleton; Phase 3 reimplements the v2 methods on top of this
// struct.
package tools

// Service is the v3 service binding for tool helpers.
type Service struct{}

// New constructs an empty Service. Dependencies will be wired in
// Phase 3.
func New() *Service {
	return &Service{}
}
