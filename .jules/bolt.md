# Bolt's Journal

## 2024-12-03 - [React List Rendering with High Frequency Updates]
**Learning:** In Wails apps where backend events trigger frequent state updates (like file conversion progress), rendering large lists without memoization causes massive UI lag due to O(N) re-renders for every single update.
**Action:** Always extract list items to `React.memo` components when the parent list receives granular updates via Wails events.

## 2025-12-14 - [Regex Recompilation Overhead]
**Learning:** Compiling regexes inside hot paths (e.g. per-file processing loop) adds ~17μs overhead per call. For 1000s of files, this adds up, but primarily it's unnecessary allocation.
**Action:** Move `regexp.MustCompile` to package-level variables or `init()` to compile once.
