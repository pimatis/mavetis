#!/bin/sh
set -eu
umask 022

project="mavetis"
version="${1:-${MAVETIS_VERSION:-dev}}"
output_dir="${MAVETIS_DIST_DIR:-dist}"
build_flags="-trimpath -buildvcs=false"
ldflags="-s -w"

require_tool() {
  name="$1"
  if command -v "$name" >/dev/null 2>&1; then
    return 0
  fi
  echo "$name is required" >&2
  exit 1
}

mkdir -p "$output_dir"
root_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

require_tool go
require_tool tar
require_tool python3

checksum() {
  file="$1"
  name=$(basename "$file")
  if command -v sha256sum >/dev/null 2>&1; then
    sum=$(sha256sum "$file" | awk '{print $1}')
    printf '%s  %s\n' "$sum" "$name" > "$file.sha256"
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    sum=$(shasum -a 256 "$file" | awk '{print $1}')
    printf '%s  %s\n' "$sum" "$name" > "$file.sha256"
    return 0
  fi
  echo "sha256sum or shasum is required" >&2
  exit 1
}

build_unix() {
  os="$1"
  arch="$2"
  bin_dir=$(mktemp -d)
  trap 'rm -rf "$bin_dir"' EXIT INT TERM
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build $build_flags -ldflags "$ldflags" -o "$bin_dir/$project" "$root_dir"
  archive="$output_dir/${project}_${os}_${arch}.tar.gz"
  tar -C "$bin_dir" -czf "$archive" "$project"
  checksum "$archive"
  rm -rf "$bin_dir"
  trap - EXIT INT TERM
}

build_windows() {
  arch="$1"
  bin_dir=$(mktemp -d)
  trap 'rm -rf "$bin_dir"' EXIT INT TERM
  GOOS="windows" GOARCH="$arch" CGO_ENABLED=0 go build $build_flags -ldflags "$ldflags" -o "$bin_dir/$project.exe" "$root_dir"
  archive="$output_dir/${project}_windows_${arch}.zip"
  BIN_DIR="$bin_dir" OUTPUT_FILE="$archive" PROJECT_NAME="$project" python3 - <<'PY'
import os
import pathlib
import zipfile
root = pathlib.Path(os.environ['BIN_DIR'])
out = pathlib.Path(os.environ['OUTPUT_FILE'])
project = os.environ['PROJECT_NAME']
with zipfile.ZipFile(out, 'w', compression=zipfile.ZIP_DEFLATED) as zf:
    zf.write(root / f'{project}.exe', arcname=f'{project}.exe')
PY
  checksum "$archive"
  rm -rf "$bin_dir"
  trap - EXIT INT TERM
}

rm -f "$output_dir"/*.tar.gz "$output_dir"/*.zip "$output_dir"/*.sha256 "$output_dir"/checksums.txt 2>/dev/null || true
build_unix darwin amd64
build_unix darwin arm64
build_unix linux amd64
build_unix linux arm64
build_windows amd64
build_windows arm64
cat "$output_dir"/*.sha256 > "$output_dir/checksums.txt"

echo "release artifacts written to $output_dir for version $version"
