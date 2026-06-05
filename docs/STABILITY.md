# Stability baseline

| Item | Value |
|------|--------|
| Commit | `5268dc3` |
| Git tag | `stable-baseline` |
| Date noted | 2026-06-05 |

## Status

Most stable version in current testing. Known issues on this baseline:

- ~~**~13 px shrink** on right-click menu~~ — fixed by removing `w.Resize` guards; selection refresh deferred until menu dismiss.
- ~~**Edge resize while menu open**~~ — same fix; do **not** re-add guards or Win32 pinning.

## Revert

```bash
git checkout stable-baseline -- .
go build -o dev/RelayPane-dev.exe ./cmd/relaypane
```

## Window size invariant

Size may change **only** from:

1. Dragging window **edges**
2. **Double-click** the title bar (maximize / restore)

Right-click, context menus, list refresh, overlays, and programmatic `w.Resize()` must **not** change window size.

---

## Post-mortem: what we tried (do not repeat)

### Timeline (newest relevant commits)

| Commit | What we did | Outcome |
|--------|-------------|---------|
| `5739697` | `WM_NCHITTEST` edge resize + title double-click maximize + `WS_THICKFRAME` | **Good** — user-approved resize/drag |
| `00cde35` | Blank **list row** for empty-area right-click + cross-pane copy | **Bad** — right-click shrinks window to minimum |
| `e9c2c5e` / `c88b42e` | Remove blank row; use layout pad / underlay instead | Shrinking reduced; blank-area menu kept |
| `5db1076` | `paneList*Layout.MinSize` must not sum all file row heights | **Good** — fixed startup window taller than screen |
| `34519be` | `restoreWindowSizeIfShrunk` + `guardWindowSizeWhileMenuOpen` (`w.Resize` poll) | **Bad** — ~13 px jitter on open; fights edge resize while menu visible |
| `ec5fb39` | Win32 `WM_SIZE` hook + `SetWindowPos` pin during menu | **Bad** — window **jumps** (“逃跑”) |
| `a2296e8` | Revert `WM_SIZE` pin | Jump gone; jitter remains |
| `5268dc3` | Remove fixed 48 px strip below list; stack underlay only | **Current baseline** — best overall; jitter + menu+resize bug remain |
| `ae91775` / `b9c07cc` | Remove wndproc / use `SetForegroundWindow` for z-order | Helps “falls behind” but user prefers `5268dc3` frame |

### Root cause: why right-click touches window size

1. `widget.NewPopUpMenu` steals focus → `widget.List.FocusLost()` → `RefreshItem()` → **full list renderer refresh** → layout/`MinSize` churn → window shrinks (sometimes to GLFW minimum).
2. `w.Resize()` guards (`restoreWindowSizeIfShrunk`) masked this as ~13 px jitter and broke edge resize while the menu was open.

**File context menu:** standard Fyne `widget.NewPopUpMenu` (`context_menu.go`). Dismiss any open overlay before show (singleton). Defer list `RefreshItem` until menu closes. No `w.Resize()` guards.

### Root cause: menu open + edge resize

`guardWindowSizeWhileMenuOpen` polls every 50 ms for up to 3 s. Whenever `canvas.Size() < saved - 2`, it calls `w.Resize(saved)`.

- User drags an edge smaller → guard snaps back → “cannot resize”.
- User drags an edge → Fyne reports transient small size during drag → guard restores → fight → sometimes collapses to minimum.
- `ec5fb39`’s `WM_SIZE` pin had the same class of problem at the Win32 layer.

### Things that must stay avoided

| Approach | Why it failed |
|----------|----------------|
| Blank `widget.List` row + `SetItemHeight(blankID, list.Size().Height)` | `list.Size()` during right-click → height goes to 0 → window minimum |
| `MinSize()` = `rowCount × rowHeight` for scrollable list | Window min height = entire directory listing |
| `restoreWindowSizeIfShrunk` / `guardWindowSizeWhileMenuOpen` | `w.Resize()` jitter; blocks resize during menu |
| `WM_SIZE` + `SetWindowPos` size pin | Window position jump |
| `w.Resize()` anywhere on menu open path | Violates invariant; visible flicker |
| `scaledEdgeBorder()` / any Fyne call inside wndproc | Wrong thread → focus/z-order bugs (“沉后面”) |
| `FindWindowW` for HWND | Wrong window → random resize/drag |
| Fixed bottom blank strip (48 px) in layout | Extra layout churn; removed in `5268dc3` |

### Safe patterns (keep)

- **Blank-area right-click**: `pane_list_area.go` — list + transparent `paneListUnderlay` stack; underlay `MinSize()` returns `(0,0)`.
- **List `MinSize`**: only the list widget’s own minimum, never full content height.
- **Borderless frame**: `window_frame_windows.go` — per-window HWND, `WM_NCHITTEST`, `WS_THICKFRAME`, `ReleaseCapture` + `PostMessage` for edges; title drag in `drag_region.go`.
- **Quiet selection before menu**: `selectRowQuiet` avoids full `list.Refresh()` but still calls `RefreshItem` — candidate to defer until menu closes.

### Correct fix (implemented)

1. **No** `PopUpMenu` on the main window for file-browser context menu.
2. **No** `w.Resize()` guards or Win32 size pinning.
3. In-tree `paneFloatingMenu`; defer row `RefreshItem` until menu dismiss.

Do **not** regress: edge resize, title double-click maximize, drag cursor, conflict dialog, blank underlay right-click.
