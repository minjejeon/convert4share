// Package settings hosts the Wails v3 service that exposes user
// settings (GetSettings / SaveSettings) backed by viper. Phase 2.1
// only provides the empty skeleton; Phase 3 wires it to
// internal/config.
package settings

// Service is the v3 service binding for application settings.
type Service struct{}

// New constructs an empty Service. Configuration access will be wired
// in Phase 3.
func New() *Service {
	return &Service{}
}
