## 2024-05-21 - DropZone Accessibility & Contrast
**Learning:** `text-slate-500` on `bg-slate-100` often fails AA contrast requirements (3.8:1 vs 4.5:1). Also, complex `role="button"` components (like DropZones) need `aria-describedby` to communicate their full purpose (e.g., supported file types) to screen readers, as `aria-label` overrides internal text content.
**Action:** Use `text-slate-600` for helper text and link description text with `aria-describedby` for complex interactive areas.
