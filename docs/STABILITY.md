# Stability baseline

| Item | Value |
|------|--------|
| Commit | `5268dc3` |
| Git tag | `stable-baseline` |
| Date noted | 2026-06-05 |

## Status

Most stable version in current testing. **Small jitter once** when opening the file-browser right-click menu.

## Revert

```bash
git checkout stable-baseline -- .
go build -o dev/RelayPane-dev.exe ./cmd/relaypane
```

## Window size rule

Size changes only via **edge drag-resize** or **title bar double-click**. Nothing else should resize the window.

## Why right-click seemed to change size

Right-click should not resize the window. The jitter is from Fyne layout plus our guards in `browser.go` that call `w.Resize()` when they detect a few pixels of drift after the menu overlay opens. Removing those guards (without replacing them with Win32 pinning) is the correct long-term fix.
