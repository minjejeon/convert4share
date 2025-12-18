## 2024-05-24 - File Path Readability
**Learning:** Displaying full file paths in lists causes truncation of the most important part (the filename) when paths are long.
**Action:** Always separate filename and directory in file lists, prioritizing filename visibility (bold/primary) and moving directory to secondary text.

## 2025-12-16 - Async Action Feedback
**Learning:** Significant actions like "Installation" that take more than 100ms must have explicit loading states to prevent user uncertainty and rage-clicking.
**Action:** Always wrap async handlers with a loading state that disables the trigger button and shows a spinner.

## 2025-12-17 - Action Confirmation
**Learning:** Icon-only buttons for invisible actions (like "Copy to Clipboard") leave users guessing if the action succeeded.
**Action:** Always implement a temporary visual state change (e.g., Checkmark icon, color change) for immediate confirmation of success.

## 2025-12-18 - Drag-and-Drop Accessibility
**Learning:** Pure drag-and-drop zones create invisible barriers for keyboard users and those who prefer standard file browsing.
**Action:** Always pair drag zones with a hidden file input and semantic interactive wrapper (role="button", tabIndex=0) to ensure universal access.
