#!/bin/sh
# Cross-compile the play-sound hook for the three supported platforms.
# Output binaries land in scripts/ next to the other hook binaries and are
# committed to the repo, so end users need no Go toolchain — only the launcher.
#
# Run from anywhere:  sh scripts/play-sound-src/build.sh
set -eu

SRC_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
OUT_DIR=$(dirname -- "$SRC_DIR")   # scripts/

LDFLAGS="-s -w"

build() {
	os=$1
	arch=$2
	out=$3
	echo "building $out ($os/$arch)"
	( cd "$SRC_DIR" && GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 \
		go build -trimpath -ldflags "$LDFLAGS" -o "$OUT_DIR/$out" . )
}

build darwin  arm64 play-sound-darwin-arm64
build linux   amd64 play-sound-linux-amd64
build windows amd64 play-sound-windows-amd64.exe

echo "done -> $OUT_DIR"
