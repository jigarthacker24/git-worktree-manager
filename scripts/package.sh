#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

APP_NAME="git-worktree-manager"
DIST_DIR="${DIST_DIR:-dist}"

usage() {
	cat <<EOF
Usage: $(basename "$0") <target>

Targets:
  local       Package for the current OS (requires fyne CLI)
  linux       Cross-package for Linux amd64 + arm64 (requires fyne-cross + Docker)
  windows     Cross-package for Windows amd64 + arm64
  darwin      Cross-package for macOS amd64 + arm64
  all         Build linux, windows, and darwin packages

Examples:
  $(basename "$0") local
  $(basename "$0") all

Install tools first:
  go install fyne.io/tools/cmd/fyne@latest
  go install github.com/fyne-io/fyne-cross@latest
EOF
}

install_tools() {
	go install fyne.io/tools/cmd/fyne@latest
	go install github.com/fyne-io/fyne-cross@latest
}

package_local() {
	command -v fyne >/dev/null || install_tools
	mkdir -p "$DIST_DIR"
	fyne package -release -name "$APP_NAME"
	shopt -s nullglob
	for artifact in "$(APP_NAME).app" "$APP_NAME"* *.exe *.tar.xz *.zip; do
		[[ -e "$artifact" ]] || continue
		mv "$artifact" "$DIST_DIR/"
	done
	echo "Output: $DIST_DIR/"
}

package_cross() {
	local os="$1"
	command -v fyne-cross >/dev/null || install_tools
	fyne-cross "$os" -arch=amd64,arm64 -name "$APP_NAME" -release
	mkdir -p "$DIST_DIR"
	cp -a fyne-cross/dist/* "$DIST_DIR/"
	echo "Output: $DIST_DIR/"
}

target="${1:-}"
case "$target" in
local) package_local ;;
linux | windows | darwin) package_cross "$target" ;;
all)
	package_cross linux
	package_cross windows
	package_cross darwin
	;;
-h | --help | help) usage ;;
*)
	usage
	exit 1
	;;
esac
