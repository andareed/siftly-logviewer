# Siftly Apps How-To Guide

This guide explains how to run and use the Siftly terminal apps, with screenshot placeholders you can replace.

## 1. Build the apps

Run from repo root:

```bash
go build -o ./dist/hostlog ./cmd/hostlog
go build -o ./dist/todaylog ./cmd/todaylog
go build -o ./dist/pluginlog ./cmd/pluginlog
go build -o ./dist/devfmt ./cmd/devfmt
```

[SCREENSHOT: build-complete-terminal]
`Show terminal after successful build commands`

## 2. App quick map

- `hostlog`: Opens host log files (`.csv` or `.json`)
- `todaylog`: Opens today log files (`.csv` or `.json`)
- `pluginlog`: Opens plugin log files (`.log` or `.json`)
- `devfmt`: Device info viewer with subcommands (`list-groups`, `list-categories`, `show`, `export`)

Default preset files in the project root:
- `hostlog-filters.json`
- `todaylog-filters.json`
- `pluginlog-filters.json`
- `devfmt-filters.json`

[SCREENSHOT: app-quick-map]
`Optional screenshot of this section or a terminal help output`

## 3. Start each app

### hostlog

```bash
./dist/hostlog --debug debug.log --input testdata/hostlog.csv
```

or:

```bash
go run ./cmd/hostlog --debug debug.log testdata/hostlog.csv
```

[SCREENSHOT: hostlog-main-view]
`Main table view after loading hostlog data`

### todaylog

```bash
./dist/todaylog --debug debug.log --input <path-to-today-file.csv-or-json>
```

or:

```bash
go run ./cmd/todaylog --debug debug.log <path-to-today-file.csv-or-json>
```

[SCREENSHOT: todaylog-main-view]
`Main table view after loading todaylog data`

### pluginlog

```bash
./dist/pluginlog --debug debug.log --input <path-to-plugin.log>
```

or:

```bash
go run ./cmd/pluginlog --debug debug.log <path-to-plugin.log>
```

[SCREENSHOT: pluginlog-main-view]
`Main table view after loading plugin log data`

### devfmt

Run interactive export directly from a file:

```bash
./dist/devfmt export --input testdata/devinfo.dump.gz --group sw --debug debug.log
```

Or stream input:

```bash
zcat testdata/devinfo.dump.gz | ./dist/devfmt export --input - --group sw --debug debug.log
```

Useful discovery commands:

```bash
./dist/devfmt list-groups --input testdata/devinfo.dump.gz
./dist/devfmt list-categories --input testdata/devinfo.dump.gz
```

[SCREENSHOT: devfmt-export-view]
`Rendered siftly table opened by devfmt export`

## 4. Core keyboard controls (all table viewers)

- `q`: quit
- `j` / `k` or `↓` / `↑`: move row down/up
- `u` / `d` or `PgUp` / `PgDn`: page up/down
- `h` / `l` or `←` / `→`: horizontal scroll
- `g` / `G`: jump to top/bottom
- `:` then line number then `Enter`: jump to line
- `?`: open help dialog

[SCREENSHOT: help-dialog]
`Help modal showing command list`

## 5. Search, filter, mark, comment workflow

### Search

1. Press `/`
2. Type search text
3. Press `Enter`
4. Use `n` / `N` for next/previous match

[SCREENSHOT: search-in-footer]
`Footer command mode while entering a search`

### Filter

1. Press `f`
2. Type regex pattern
3. Press `Enter` to apply
4. Press `F` to toggle the current filter on/off
5. Optional: inside filter mode press `Ctrl+P` to open filter palette presets/history

[SCREENSHOT: filter-command]
`Regex filter entered in command bar`

[SCREENSHOT: filter-palette]
`Filter palette modal with presets/history`

### Mark rows (RAG)

1. Press `m`
2. Press:
   - `r` for red
   - `a` for amber
   - `g` for green
   - `c` to clear mark
3. Press `]` / `[` to jump next/previous marked row
4. Press `M` to toggle "show only marked rows"

[SCREENSHOT: mark-mode]
`Mark mode hint shown; row about to be marked`

[SCREENSHOT: marked-rows]
`Rows with visible red/amber/green markers`

### Comments

1. Press `c`, then `e` to edit comment on selected row
2. Type comment and press `Enter`
3. Press `c`, then `v` to toggle comment drawer

[SCREENSHOT: comment-edit]
`Comment command mode with text input`

[SCREENSHOT: comment-drawer]
`Drawer open showing comment content`

## 6. View/layout operations

### Column visibility

1. Press `v`, then `c`
2. Enter column names or indexes (comma/space separated)
3. Press `Enter`
4. Use `all` to show all columns

[SCREENSHOT: columns-toggle]
`Columns command with selected fields`

### Sort

1. Press `v`, then `s`
2. Enter sort expression:
   - `<column> asc`
   - `<column> desc`
   - `off` (reset sorting)
3. Press `Enter`

Examples:

```text
timestamp desc
3 asc
off
```

[SCREENSHOT: sort-command]
`Sort expression entered in command bar`

### Column order

1. Press `v`, then `o`
2. Enter desired order by names or indexes
3. Press `Enter`

Example:

```text
timestamp, host, message, severity
```

[SCREENSHOT: column-order-command]
`Column order command before apply`

### Reset layout

- Press `v`, then `r` to reset visibility, sort, and order to defaults

[SCREENSHOT: layout-reset-notice]
`Notice banner after layout reset`

## 7. Time window controls

- `t`, then `w`: open time window panel
- `t`, then `b`: set window start from selected row timestamp
- `t`, then `e`: set window end from selected row timestamp
- `t`, then `r`: reset full window range

[SCREENSHOT: time-window-panel]
`Time window panel open`

[SCREENSHOT: time-window-notice]
`Notice after setting start/end from cursor`

## 8. Save and export

- Press `s` to save session (JSON snapshot with marks/comments)
- Press `e` to export output (CSV)
- Reopen a saved session by launching app with the saved `.json` file

Examples:

```bash
./dist/hostlog --input session.json
./dist/pluginlog --input plugin-session.json
```

[SCREENSHOT: save-dialog]
`Save dialog with .json filename`

[SCREENSHOT: export-dialog]
`Export dialog with .csv filename`

## 9. Recommended screenshot checklist

- `[ ]` App launched with data loaded
- `[ ]` Help dialog (`?`)
- `[ ]` Search and filter usage
- `[ ]` Marked rows (R/A/G)
- `[ ]` Comment edit and drawer
- `[ ]` Columns/sort/order commands
- `[ ]` Time window panel
- `[ ]` Save + export dialogs
