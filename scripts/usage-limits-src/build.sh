#!/bin/sh
# Cross-compile the usage-limits hook for the three supported platforms.
# Output binaries land in scripts/, one per OS, and are committed to the repo, so
# end users need no Go toolchain — only the sh launcher in .claude/hooks/.
#
# Run from anywhere:  sh scripts/usage-limits-src/build.sh
set -eu

SRC_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

# Where the binaries land. Defaults to the folder holding this source directory (scripts/), which is
# where most projects keep them; pass a directory to override, since a few keep them elsewhere
# (scripts/claude-code/, .claude/bin/) and the launcher looks in all three.
if [ "$#" -ge 1 ]; then
	# Created if missing: the folder a project wants its binaries in may not exist yet, which is
	# exactly the case when one is being moved to a different layout.
	mkdir -p -- "$1"
	OUT_DIR=$(CDPATH= cd -- "$1" && pwd)
else
	OUT_DIR=$(dirname -- "$SRC_DIR")
fi

LDFLAGS="-s -w"

# -buildvcs=false, for two reasons. It keeps the binary reproducible: with stamping on, the same
# source built in two checkouts differs, and every commit changes the stamp, so a committed binary
# would show a diff on every rebuild. And it builds at all in a project whose git state Go cannot
# read, which otherwise fails the build outright.

build() {
	os=$1
	arch=$2
	out=$3
	echo "building $out ($os/$arch)"
	( cd "$SRC_DIR" && GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 \
		go build -trimpath -buildvcs=false -ldflags "$LDFLAGS" -o "$OUT_DIR/$out" . )
}

build darwin  arm64 usage-limits-darwin-arm64
build linux   amd64 usage-limits-linux-amd64
build windows amd64 usage-limits-windows-amd64.exe

echo "done -> $OUT_DIR"
