#!/bin/sh
set -eu
umask 022

owner_repo="pimatis/mavetis"
project="mavetis"
version="${MAVETIS_VERSION:-latest}"
install_dir="${MAVETIS_INSTALL_DIR:-}"

require_dir() {
  path="$1"
  if [ -d "$path" ]; then
    return 0
  fi
  mkdir -p "$path"
}

require_tool() {
  name="$1"
  if command -v "$name" >/dev/null 2>&1; then
    return 0
  fi
  echo "$name is required" >&2
  exit 1
}

fetch() {
  url="$1"
  out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl --proto '=https' --tlsv1.2 -fsSL "$url" -o "$out"
    return 0
  fi
  if command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$out"
    return 0
  fi
  echo "curl or wget is required" >&2
  exit 1
}

verify_checksum() {
  directory="$1"
  file="$2"
  checksum="$3"
  if command -v sha256sum >/dev/null 2>&1; then
    (cd "$directory" && sha256sum -c "$checksum")
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    expected=$(cut -d ' ' -f 1 "$directory/$checksum")
    actual=$(shasum -a 256 "$directory/$file" | cut -d ' ' -f 1)
    if [ "$expected" = "$actual" ]; then
      return 0
    fi
    echo "checksum verification failed" >&2
    exit 1
  fi
  echo "sha256sum or shasum is required" >&2
  exit 1
}

place_binary() {
  from="$1"
  to="$2"
  if command -v install >/dev/null 2>&1; then
    install -m 0755 "$from" "$to"
    return 0
  fi
  cp "$from" "$to"
  chmod 0755 "$to"
}

uname_s=$(uname -s 2>/dev/null | tr '[:upper:]' '[:lower:]')
uname_m=$(uname -m 2>/dev/null | tr '[:upper:]' '[:lower:]')

if [ "$uname_s" = "darwin" ]; then
  os="darwin"
fi
if [ "$uname_s" = "linux" ]; then
  os="linux"
fi
if [ "${os:-}" = "" ]; then
  echo "unsupported operating system: $uname_s" >&2
  exit 1
fi

arch="$uname_m"
if [ "$arch" = "x86_64" ]; then
  arch="amd64"
fi
if [ "$arch" = "aarch64" ] || [ "$arch" = "arm64" ]; then
  arch="arm64"
fi
if [ "$arch" != "amd64" ] && [ "$arch" != "arm64" ]; then
  echo "unsupported architecture: $uname_m" >&2
  exit 1
fi

if [ "$install_dir" = "" ]; then
  if [ -w "/usr/local/bin" ]; then
    install_dir="/usr/local/bin"
  fi
  if [ "$install_dir" = "" ]; then
    install_dir="$HOME/.local/bin"
  fi
fi

require_dir "$install_dir"
require_tool tar

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

archive="${project}_${os}_${arch}.tar.gz"
checksum="${archive}.sha256"
base_url="https://github.com/${owner_repo}/releases"
if [ "$version" = "latest" ]; then
  asset_url="$base_url/latest/download/$archive"
  checksum_url="$base_url/latest/download/$checksum"
fi
if [ "$version" != "latest" ]; then
  clean_version=$(printf '%s' "$version" | sed 's#^v##')
  asset_url="$base_url/download/v${clean_version}/$archive"
  checksum_url="$base_url/download/v${clean_version}/$checksum"
fi

fetch "$asset_url" "$tmp/$archive"
fetch "$checksum_url" "$tmp/$checksum"
verify_checksum "$tmp" "$archive" "$checksum"
tar -xzf "$tmp/$archive" -C "$tmp"
if [ ! -f "$tmp/$project" ]; then
  echo "release archive did not contain $project" >&2
  exit 1
fi
place_binary "$tmp/$project" "$install_dir/$project"

echo "$project installed to $install_dir/$project"
echo "run '$project update --check' to verify future releases"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "add $install_dir to PATH if the command is not found" ;;
esac
