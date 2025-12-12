# AGENTS.md

This file provides context and guidelines for AI agents (and human developers) working on the **Convert4Share** project.

## Project Overview

**Convert4Share** is a Windows utility for converting `.mov` and `.heic` files to `.mp4` and `.jpg` respectively.
It was originally a CLI tool integrated with Windows Context Menu, but is being migrated to a **Wails** desktop application (Go backend + React/Vite/Tailwind frontend).

## Tech Stack

-   **Backend**: Go (Wails framework)
-   **Frontend**: React, TypeScript, Vite, Tailwind CSS
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
    -   Provide dummy implementations for other OSs (e.g., `//go:build !windows`) to allow cross-platform compilation/checking (even if runtime functionality is limited).

3.  **Command Line Logic**:
    -   The original `cmd/` package (Cobra) is preserved for `install`/`uninstall` logic.
    -   This logic is invoked before Wails starts if arguments are present.

## Development Tasks & Status

-   **Current Task**: Migrating to Wails.
-   **Frontend**: `frontend/` directory created with `vite-react-ts`. Tailwind configured.
-   **Backend**: `app.go` created. `main.go` modified.
-   **Bindings**: Wails binding generation is in progress (fixing type errors).

## Instructions for Agents

1.  **Always Verify**: After editing code, run `go mod tidy` or a build check to ensure no syntax errors.
2.  **Wails Bindings**: If you modify `App` struct methods, you MUST run `wails generate module` to update frontend bindings.
3.  **Cross-Platform Awareness**: Since the environment might be Linux but the target is Windows, ensure `GOOS=windows` checks or build tags are respected.
4.  **UI/UX**: The user requested a UI with progress bars. The frontend should listen to `conversion-progress` events.
5.  **Config**: Configuration is handled by `viper`. `App` struct exposes settings to frontend.
6.  **Cleanup**: Binary artifacts (e.g., `.exe` files, `dist/` folders) MUST be removed before committing.

## Common Issues / Solutions

-   **Wails Generate Error**: Often due to build tags. Ensure `cmd/` files have appropriate `!windows` fallbacks if they import windows-specific packages.
-   **Vite Build**: Ensure `npm install` is run in `frontend/` before building.
