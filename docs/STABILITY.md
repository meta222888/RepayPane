# Stability baseline

| Item | Value |
|------|--------|
| Commit | `5268dc3` |
| Git tag | `stable-baseline` |
| Date noted | 2026-06-05 |

## Status

Most stable version in current testing. Known issues on this baseline:

1. **~13 px shrink** when opening the file-browser right-click menu (one frame).
2. **Edge resize while menu is open** — resize blocked or window snaps to minimum.

Both are side effects of the current “fix” in `browser.go` (`restoreWindowSizeIfShrunk` / `guardWindowSizeWhileMenuOpen`). Do **not** add more `w.Resize()` or Win32 pinning to fight them.

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

1. `widget.NewPopUpMenu` adds a canvas overlay → Fyne runs `EnsureMinSize` on next layout pass.
2. `selectRowQuiet` + `list.RefreshItem` before the menu can change reported `MinSize` for a frame.
3. Fyne may shrink the GLFW canvas by a few pixels (~13 px observed) to satisfy the new minimum.

This is a **layout side effect**, not intentional resize. Fighting it with `w.Resize(saved)` makes the shrink **visible** (jitter) and **blocks** user edge resize while the guard loop runs.

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

### Correct fix direction (next work)

1. **Delete** `restoreWindowSizeIfShrunk` and `guardWindowSizeWhileMenuOpen` — do not replace with Win32 pinning.
2. **Defer** `RefreshItem` / list refresh until the popup overlay is gone (watch `Overlays().List()`).
3. Show `PopUpMenu` without any `w.Resize()` before or after.
4. If ~13 px shrink remains after (1–3), treat it as a Fyne overlay/layout issue; fix at layout level (e.g. stable `MinSize` on pane chrome), not resize fights.

Do **not** regress: edge resize, title double-click maximize, drag cursor, conflict dialog, blank underlay right-click.
