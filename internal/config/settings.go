// Package config holds settings-related helpers that are independent
// of the Wails App. The Settings struct itself currently lives in the
// main package to keep the existing Wails-generated TypeScript
// bindings stable (the binding namespace tracks the package the type
// is declared in, not type aliases). A later phase will fold the
// struct definition into this package once the frontend is migrated
// off the implicit `main.Settings` reference.
package config

import "github.com/spf13/viper"

// ExcludePatterns reads the exclude patterns list, preferring the
// canonical `excludePatterns` key but falling back to the legacy
// `excludeStringPatterns` key for backward compatibility with existing
// config.yaml files.
func ExcludePatterns() []string {
	patterns := viper.GetStringSlice("excludePatterns")
	if len(patterns) == 0 {
		patterns = viper.GetStringSlice("excludeStringPatterns")
	}
	return patterns
}
