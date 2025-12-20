## 2024-05-21 - Settings Accessibility
**Learning:** The settings forms were completely lacking programmatic association between labels and inputs (`htmlFor`/`id`). This makes the settings page difficult for screen reader users and fails basic WCAG compliance.
**Action:** Always verify form inputs have associated labels using `htmlFor` and unique `id`s, especially in componentized forms where context might be split.
