# AGENTS.md

This file provides context and guidelines for AI agents (and human developers) working on the **Convert4Share** project.

## Project Overview

**Convert4Share** is a Windows utility for converting `.mov` and `.heic` files to `.mp4` and `.jpg` respectively.
It is a **Wails** desktop application (Go backend + React/Vite/Tailwind frontend).

## Tech Stack

-   **Backend**: Go (Wails framework)
-   **Frontend**: React, TypeScript, Vite, Tailwind CSS
-   **Package Manager**: `pnpm` (Enforced. Do not use `npm` or `yarn` for frontend dependencies).
-   **OS**: Windows (Target), but development environment might be Linux/macOS (requiring conditional builds).

## Architecture Guidelines

1.  **Wails Integration**:
    -   `main.go`: Entry point. Checks for CLI args (`install`, `uninstall`). If none, launches Wails `Run()`.
    -   `app.go`: Contains the `App` struct bound to Wails. Exposes methods to Frontend (`ConvertFiles`, `SaveSettings` etc.).
    -   `frontend/`: Contains the standard Vite+React project.

2.  **Platform Specifics**:
    -   The app relies heavily on Windows features (Context Menu Registry, `powershell` for GPU detection).
    -   **Context Menu**: Uses `SystemFileAssociations` for Windows 10/Classic menu and `OpenWithProgids` for Windows 11 Modern menu integration.
    -   Use `//go:build windows` for code that imports `windows` or `golang.org/x/sys/windows`.
    -   Provide dummy implementations for other OSs (e.g., `//go:build !windows`) to allow cross-platform compilation/checking.

3.  **Command Line Logic**:
    -   The original `cmd/` package (Cobra) is preserved for `install`/`uninstall` logic.
    -   This logic is invoked before Wails starts if arguments are present.

## Development Tasks & Status

-   **Status**: Wails migration is complete. Frontend and Backend are integrated.
-   **Frontend**: Located in `frontend/`. Uses `pnpm`.
-   **Bindings**: Generated in `frontend/wailsjs/`. Run `wails generate module` to update if `App` struct changes.

## Instructions for Agents

1.  **Always Verify**: After editing code, run `go mod tidy` or a build check to ensure no syntax errors.
2.  **Wails Bindings**: If you modify `App` struct methods, you MUST run `wails generate module` to update frontend bindings.
3.  **Cross-Platform Awareness**: Ensure `GOOS=windows` checks or build tags are respected.
4.  **Dependency Management**: Use `pnpm` for all frontend dependency tasks.
5.  **UI/UX**: The frontend listens to `conversion-progress` events. Ensure any new long-running tasks emit appropriate events.
6.  **Config**: Configuration is handled by `viper`. `App` struct exposes settings to frontend.
7.  **Cleanup**: Binary artifacts (e.g., `.exe` files, `dist/` folders) MUST be removed before committing.
8.  **Code Comments**: Avoid verbose comments that are unnecessary because the actual code is sufficiently self-explanatory. Code should be self-documenting where possible.

## Common Issues / Solutions

-   **Wails Generate Error**: Often due to build tags. Ensure `cmd/` files have appropriate `!windows` fallbacks if they import windows-specific packages.
-   **Vite Build**: Ensure `pnpm install` is run in `frontend/` before building.
-   **Viper Configuration**: Ensure `viper.SetDefault` is used for all critical configuration keys (especially binary paths) to prevent runtime errors when `config.yaml` is missing. Use `os.UserHomeDir()` for default paths.
