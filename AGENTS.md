# AGENTS.md

This file provides context and guidelines for AI agents (and human developers) working on the **Convert4Share** project.

## Project Overview

**Convert4Share** is a Windows utility for converting `.mov` and `.heic` files to `.mp4` and `.jpg` respectively.
It is a **Wails** desktop application (Go backend + React/Vite/Tailwind frontend).

## Tech Stack

-   **Backend**: Go (Wails framework)
-   **Frontend**: React, TypeScript, Vite, Tailwind CSS (v4)
-   **Package Manager**: `npm` (Enforced. Do not use `pnpm` or `yarn` for frontend dependencies).
-   **OS**: Windows (Target), but development environment might be Linux/macOS (requiring conditional builds).

## Architecture & Coding Guidelines

### 1. Wails Integration
-   `main.go`: Entry point. Checks for CLI args (`install`, `uninstall`). If none, launches Wails `Run()`.
-   **Bindings**: Located in `frontend/src/wailsjs/`. This directory is often gitignored.
    -   **Important**: If you modify `App` struct methods or `models` in Go, you **MUST** run `wails generate module` to update the frontend bindings.
    -   **Verification**: When verifying frontend changes in a browser environment (without the Wails runtime), you must mock these bindings (e.g., inject `window.go.main.App` via Playwright).
-   **Events**: The frontend listens to `conversion-progress` events. Ensure any new long-running tasks emit appropriate events.
-   **Window Management**: To bring the window to front on a second instance launch, use `runtime.WindowUnminimise`, `runtime.WindowShow`, and toggle `runtime.WindowSetAlwaysOnTop`.

### 2. Frontend Development
-   **Tailwind CSS**: Uses **v4** syntax with `darkMode: 'selector'`. To toggle themes, add/remove the `dark` class on the root HTML element.
    -   Styles follow a "Light Mode Default" strategy (`bg-white` vs `dark:bg-slate-800`).
-   **Performance**: List components receiving high-frequency updates (like file progress) **must** use `React.memo` to prevent rendering bottlenecks.
-   **Styling**:
    -   **DropZone**: Centered, `p-10`, flex-col, dashed borders.
    -   **File List**: Filename (primary/bold), Path (secondary/small).
    -   **Buttons**: Interactive buttons (Copy) should use local state for visual feedback (2s delay).

### 3. Backend Logic
-   **Configuration**:
    -   Use `viper` for config.
    -   **Saving**: Use `viper.WriteConfigAs` with an explicit path to ensure creation if missing.
    -   **Defaults**: Use `viper.SetDefault` for critical keys (binary paths). Use `os.UserHomeDir()` for paths.
-   **File Processing**:
    -   **Validation**: Always check `!info.IsDir()` and ensure the file is not the running executable itself.
    -   **Regex**: When parsing FFmpeg output, use `\d+` for the hour component to support >99 hours.
    -   **Concurrency**: Use `sync.WaitGroup` when parsing `stderr` in goroutines.
-   **Video Encoding**:
    -   Supports 'High' (5Mbps), 'Medium' (2.5Mbps), 'Low' (1Mbps) presets.
    -   Flags adapt to hardware (AMD: `quality`/`balanced`/`speed`, NVIDIA: `slow`/`medium`/`fast`).

### 4. Windows Specifics
-   **Context Menu**: Uses `SystemFileAssociations` (Classic) and `OpenWithProgids` (Win11).
-   **Clipboard**: `CopyFileToClipboard` uses PowerShell `Set-Clipboard -AsHtml` or `CF_HDROP`.
    -   **Escaping**: Sanitize paths in PowerShell commands by replacing `'` with `''`.

## Instructions for Agents

1.  **Always Verify**: After editing code, run `go mod tidy` or a build check.
2.  **Verify Frontend**: If touching UI, consider how to verify it (mocking Wails if using standard browser tools).
3.  **Cross-Platform Awareness**: Ensure `GOOS=windows` checks or build tags are respected.
4.  **Dependencies**: Use `npm`. Do not use `pnpm` or `yarn`.
5.  **Code Style**: Avoid verbose comments. Code should be self-documenting.
6.  **Git**: `frontend/dist` is embedded. `go build` requires it to exist.
7.  **Dependency Lock**: Do not change `src/wailsjs/runtime` in `@frontend/package-lock.json`.

## Common Issues / Solutions

-   **Wails Generate Error**: Often due to build tags. Ensure `cmd/` files have appropriate `!windows` fallbacks if they import windows-specific packages.
-   **Viper Configuration**: `viper.WriteConfig` fails if no config file exists; use `WriteConfigAs`.
-   **Path Separators**: Use Unix-style forward slashes (`/`) when mocking paths in frontend tests to avoid escaping issues.
