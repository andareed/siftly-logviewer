# TODO

## 0.7 TUI polish

- [x] Redesign the footer and status bar to render contextual key hints and clearly show active filter, sort, time window, marks-only, range selection, search position, operation and unsaved states.
- [x] Add a persistent row inspector showing every column name and full value, original row number, mark and comment, with field and row copy actions.
- [x] Track dirty state, warn before quitting with unsaved changes, provide undo for annotations and view operations, and maintain a source-adjacent recovery sidecar that is restored only after an explicit launch prompt.
- [ ] Make the graph legend interactive: use `Tab`/`Shift+Tab` to cycle through series and allow the focused series to be included or excluded from both the graph and SVG export, with clear focus and exclusion states.
- [ ] Show determinate progress for long operations where possible, including rows processed, percentage and elapsed time, with `Esc` cancellation while retaining the previous view.
- [x] Add a searchable command palette covering every action and shortcut, and reorganise help into clear command categories.
- [x] Replace text-based column operations with a searchable column manager supporting visibility, ordering, sort direction, frozen columns and auto-fit.
- [ ] Add a simple column/operator/value filter builder while retaining regex as an advanced mode; show removable active-filter components and support saved views.
- [ ] Add optional mouse support for scrolling, row selection, range extension and column-header actions while retaining full keyboard operation.
- [ ] Add adaptive layouts for narrow terminals and explicit truecolour, ANSI-256 and monochrome themes with contrast tests for all interaction states.

### Appearance improvements

- [x] Replace the three-line footer with a compact state bar and contextual action line, giving active modes and important state a clear visual hierarchy.
- [x] Make the table denser and easier to scan with single-line rows by default, stronger headers, reduced emphasis on repeated values, and full content available in the row inspector.
- [x] Centralise colours, emphasis levels, borders and state styles into shared design tokens used consistently by tables, panels, dialogs and notices.
- [x] Size the comment drawer, row inspector, graph and dialogs responsively so they use only the space their content and terminal dimensions allow.
- [x] Update the help presentation, screenshots and documentation to reflect the current interface and use consistent terminology and visual styling.

## Compatibility

- [ ] Add explicit ANSI-256 palette fallbacks and rendering tests for all truecolour UI styles, ensuring cursor and range backgrounds remain distinct over SSH and terminals reporting `xterm-256color`.
