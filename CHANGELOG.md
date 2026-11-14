# Changelog

## 0.2.0

- Lasso handwriting into a persistent task-review queue, adapted from
  [jdkruzr's reTaskable](https://github.com/jdkruzr/reTaskable) (GPL-3.0-only).
- Counted import button to the left of Tasks; automatically advance after Save
  or Discard, show position/total, and preserve unfinished captures on Cancel.
- Durable duplicate prevention and dismissal without changing source notebooks.
- Optional home-screen meeting-note calendar icon launcher, with a persistent
  Settings toggle restoring the native meeting-note action when disabled.
- Color accents on known Paper Pro/Move models, per-calendar/list color selection,
  soft event/task fills, and the chosen color logo. Grayscale/unknown devices retain
  the monochrome interface. Colors are local display preferences.
- Separate static ARMv7 and aarch64 builds, architecture-checked installers, and
  release archives with corresponding source and checksums.
- ARMv7 supports RM1 (user-confirmed) and RM2 (tested on-device). aarch64 and physical
  color-panel behavior are unverified on hardware; off-device checks pass.
- Updated setup, privacy, attribution and development docs with actual fictional
  QML screenshots and retained design explorations.
