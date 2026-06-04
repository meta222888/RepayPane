# RelayPane

A lightweight WinSCP-style SFTP client written in Go.

**RelayPane** = relay files between your machine and remote servers, in a dual-pane view.

## Features

- **Server management** — save, edit, delete SFTP connections (password or private key)
- **Dual-pane browser** — local files on the left, remote on the right
- **Drag & drop** — drop files from Explorer onto the right half to upload; onto the left half to copy locally
- **Upload / Download** — select a file and use the toolbar buttons
- **Remote file editor** — double-click a remote file to edit; **Ctrl+S** saves directly back to the server
- **Large file guard** — files over 2 MB prompt before opening in the editor

## Requirements

- Go 1.22+
- C compiler for Fyne (on Windows: [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or MinGW-w64)

## Build & Run

```powershell
cd d:\work\RelayPane
go mod tidy
go run ./cmd/relaypane
```

Or build an executable:

```powershell
go build -o RelayPane.exe ./cmd/relaypane
```

## Usage

1. Click **Add Server** and fill in host, port (22), username, and password or private key path.
2. Select a server in the list to connect.
3. Browse folders by double-clicking directories.
4. Double-click a remote file to edit it; press **Ctrl+S** to save to the server.
5. Drag files from Windows Explorer onto the window (right side = upload, left side = local copy).

Server profiles are stored in `%USERPROFILE%\.relaypane\servers.json`.

## Project layout

```
cmd/relaypane/     entry point
internal/config/   saved server profiles
internal/remote/   SFTP client
internal/ui/       Fyne GUI
```

## License

MIT
