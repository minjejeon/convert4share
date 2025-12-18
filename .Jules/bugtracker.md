# Bug Tracker

| Issue | Judgment | Severity | Status |
| :--- | :--- | :--- | :--- |
| Excessive Logging in `converter.go` | Logging every line of FFmpeg output (including dynamic `stats` updates) generates massive log files and unnecessary I/O, potentially slowing down the application and filling the disk. | Medium | Resolved |
| Magick Command Output Handling | The `Magick` function assigns `os.Stdout`/`os.Stderr` to the command. In a Windows GUI application (detached from console), these file descriptors may be invalid or closed, causing the subprocess to crash or hang when attempting to write output. | High | Resolved |
| Frontend Thumbnail Memory Accumulation | The `files` state in `App.tsx` retains Base64-encoded thumbnails for all completed files. In long sessions, this can lead to significant memory usage (OOM risk). | Low | Pending |
| Startup Race Condition | The `domReady` function uses a fixed `time.Sleep(500ms)` before emitting `files-received`. On slower machines or heavy loads, the React frontend may not be ready to listen, causing dropped files. | Medium | Resolved |
