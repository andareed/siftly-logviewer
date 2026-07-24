# Siftly 0.7 Release Notes

This release consolidates the major Siftly improvements introduced since 0.5, including the 0.6 marking and snapshot work.

## Highlights

- Redesigned the shared `hostlog`, `todaylog`, and `pluginlog` interface for denser, clearer log review.
- Added range selection, multi-row copy and marking, undo/redo, dirty-state tracking, and prompted recovery.
- Added a searchable command palette, improved help, a complete column manager, and a compact row inspector with untruncated values.
- Filters now match complete logical rows, making searches across columns work without special multiline-regex syntax.
- Added clearer contextual status information, operation progress, search position, and active filter, sort, selection, mark, and unsaved states.

## Data And Performance

- Introduced a compact, grep-friendly JSON snapshot format that stores column names once and writes each row on one line.
- Snapshots retain source data and stable row IDs, keeping investigations consistent if backend data later changes.
- Removed duplicate saved headers and fixed reopening `todaylog` JSON; existing snapshot files remain supported.
- Improved large-file loading, filtering, stable-ID generation, memory use, and responsiveness while entering commands.
- Recovery data is now kept in a compact hidden sidecar beside the source and is never restored without prompting.

## Todaylog

- Added prefiltering for very large raw logs.
- Enlarged SVG graph exports to a scalable 1920x1080 canvas with sensible timestamped filenames.
- Improved time-window detection for additional timezone formats.

## Behaviour Changes

- Applications now accept the concise form `hostlog|todaylog|pluginlog [options] FILE`; the old `--input` option remains compatible.
- Save uses `s`, exports use `e`, the column manager opens with `v`, and undo/redo use `u` and `r`.
- Filter history is available with `Ctrl+R` or `Up`; full-source reload is available from the command palette.
- `devfmt` is no longer built or shipped with Siftly.
- Truecolour terminals retain the richest presentation, with an explicit ANSI-256 fallback for constrained remote sessions.
