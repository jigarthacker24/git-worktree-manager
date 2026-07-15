# Git Worktree Manager

A minimal desktop app built with [Fyne](https://fyne.io/) for managing git worktrees.

Repository: [github.com/jigarthacker24/git-worktree-manager](https://github.com/jigarthacker24/git-worktree-manager)

## Features

- Open a cloned repository by path (with recent repos)
- List worktrees: directory, branch, path
- Pin up to 3 worktrees per repository
- Copy branch name or worktree path
- Open a worktree in VS Code, Cursor, or Claude Code (icons disabled when not installed)
- Create worktrees from an existing branch (searchable) or a new branch
- Remove worktrees (with confirmation and optional force)

## Requirements

### Run from source

- Go 1.22+
- `git` in your PATH
- C compiler (for Fyne; Xcode CLI tools on macOS)

### Run packaged app

- `git` in your PATH
- Optional IDEs (install any you want to use; unavailable IDEs show as disabled icons):
  - **VS Code:** `code` in PATH or `/Applications/Visual Studio Code.app` (macOS)
  - **Cursor:** `cursor` in PATH or `/Applications/Cursor.app` (macOS)
  - **Claude Code:** `claude` CLI and/or `claude-cli://` URL handler (installed by the [native installer](https://claude.ai/install)). The Claude desktop chat app alone does not include Code on the free plan (Pro/Max required for desktop Code).

## Install

### macOS (build from source via Fyne)

```bash
go install fyne.io/tools/cmd/fyne@latest
fyne install github.com/jigarthacker24/git-worktree-manager@latest
```

### Linux (pre-built binary — recommended)

`fyne install` builds from source on Linux and can fail on Ubuntu with a `_FORTIFY_SOURCE` CGO error. Use the pre-built release instead:

```bash
VERSION=v1.0.0
ARCH=amd64   # use arm64 on Apple Silicon Linux / aarch64 machines

curl -LO "https://github.com/jigarthacker24/git-worktree-manager/releases/download/${VERSION}/git-worktree-manager-${VERSION}-linux-${ARCH}.tar.xz"
sudo tar -xJf "git-worktree-manager-${VERSION}-linux-${ARCH}.tar.xz" -C /
```

Then launch **Git Worktree Manager** from your app menu, or run:

```bash
git-worktree-manager
```

See [Releases](https://github.com/jigarthacker24/git-worktree-manager/releases) for all versions and architectures.

## Run from source

```bash
git clone https://github.com/jigarthacker24/git-worktree-manager.git
cd git-worktree-manager
go run .
```

## Build binary (local)

```bash
go build -o git-worktree-manager .
```

## Package for distribution

This project uses [Fyne packaging](https://docs.fyne.io/started/packaging.html) and optional [fyne-cross](https://github.com/fyne-io/fyne-cross) for cross-platform builds.

### Install packaging tools

```bash
make install-tools
# or
go install fyne.io/tools/cmd/fyne@latest
go install github.com/fyne-io/fyne-cross@latest
```

### Package for your current OS

```bash
make package-local
# or
./scripts/package.sh local
```

Output goes to `dist/`.

### Cross-package (Linux, Windows, macOS)

Requires **Docker** for `fyne-cross`.

```bash
# All platforms
make package-all

# Or one platform at a time
make package-linux
make package-windows
make package-darwin

# Or via script
./scripts/package.sh all
./scripts/package.sh linux
```

### Deliverables by platform

| Platform | Typical output |
|----------|----------------|
| **macOS** | `git-worktree-manager.app` (`.app` bundle) |
| **Windows** | `git-worktree-manager.exe` |
| **Linux** | `git-worktree-manager` binary + `.tar.xz` archive |

Artifacts are collected under `dist/` (and `fyne-cross/dist/` during cross builds).

### Release tags (CI)

Push a version tag to build packages via GitHub Actions:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow in `.github/workflows/release.yml` uploads Linux `.tar.xz` packages to the GitHub release.

### Linux install troubleshooting (source builds)

If you build from source with `fyne install` or `go run .`, you may need:

```bash
sudo apt install gcc libgl1-mesa-dev xorg-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev
CGO_CFLAGS="-U_FORTIFY_SOURCE" fyne install github.com/jigarthacker24/git-worktree-manager@latest
```

## List on [apps.fyne.io](https://apps.fyne.io)

After the repo is public and `fyne install github.com/jigarthacker24/git-worktree-manager@latest` works, submit the app at [developer.fyne.io/submit](https://developer.fyne.io/submit).

## App metadata

Packaging metadata lives in `FyneApp.toml` (name, ID, version, icon).

## Project layout

```
main.go                 # Fyne UI
FyneApp.toml            # Package metadata
Icon.png                # App icon
internal/gitops/        # git worktree commands
internal/ide/           # Open in Cursor
internal/ui/            # Icons, hints, window maximize
scripts/package.sh      # Packaging helper
Makefile                # build / package targets
```
