
## 2024-12-03 - [React List Rendering with High Frequency Updates]
**Learning:** In Wails apps where backend events trigger frequent state updates (like file conversion progress), rendering large lists without memoization causes massive UI lag due to O(N) re-renders for every single update.
**Action:** Always extract list items to `React.memo` components when the parent list receives granular updates via Wails events.
