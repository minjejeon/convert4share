# Bug Tracker

| ID | Severity | Description | Status | User Decision | Notes |
|----|----------|-------------|--------|---------------|-------|
| BUG-001 | **Critical** | **Deadlock/Panic in `ConvertFiles`**: The semaphore release (`defer func() { <-ffmpegSem }()`) is executed even if the semaphore was not acquired. | **Resolved** | Approved | Fixed in `app.go`. |
| BUG-002 | **Major** | **Uninterruptible Thumbnail Generation**: `GenerateThumbnail` used `exec.Command` without context. | **Resolved** | Approved | Fixed in `converter/converter.go` and `app_tools.go`. |
| BUG-003 | **Minor** | **Redundant Thumbnail Requests**: `App.tsx` called `GetThumbnail` redundantly. | **Resolved** | Approved | Fixed in `frontend/src/App.tsx`. |
| BUG-004 | **Minor** | **Interval Leak in `handleInstall`**: Cleanup missing. | **Resolved** | Approved | Fixed in `frontend/src/App.tsx`. |
| BUG-005 | **Minor** | **Magick Stdin Safety**: `Magick` command missing `Stdin = nil`. | **Resolved** | Approved | Fixed in `converter/converter.go`. |
| BUG-006 | **Info** | **Frontend Memory Accumulation (Base64)**: Storing thumbnails as Base64 strings. | Deferred | **Deferred** | User explicitly instructed to defer this. |
