# Git Worktree Manager

A minimal desktop app built with [Fyne](https://fyne.io/) for managing git worktrees.

Repository: [github.com/jigarthacker24/git-worktree-manager](https://github.com/jigarthacker24/git-worktree-manager)

## Features

- Open a cloned repository by path (with recent repos)
- List worktrees: branch, folder name, path
- Pin up to 3 worktrees per repository
- Copy branch name or worktree path
- Open a worktree in Cursor IDE
- Create worktrees from an existing branch (searchable) or a new branch
- Remove worktrees (with confirmation and optional force)

## Requirements

### Run from source

- Go 1.22+
- `git` in your PATH
- C compiler (for Fyne; Xcode CLI tools on macOS)

### Run packaged app

- `git` in your PATH
- On macOS: Cursor at `/Applications/Cursor.app` or `cursor` in PATH

## Install (after publishing to GitHub)

```bash
go install fyne.io/tools/cmd/fyne@latest
fyne install github.com/jigarthacker24/git-worktree-manager@latest
```

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

The workflow in `.github/workflows/release.yml` uploads platform packages to the GitHub release.

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
