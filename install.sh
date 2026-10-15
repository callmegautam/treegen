#!/bin/sh
set -eu

REPO="callmegautam/treegen"
BINARY="treegen"

VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-}"

err() {
	echo "install: $*" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || err "$1 is required but not installed"
}

detect_os() {
	os=$(uname -s | tr '[:upper:]' '[:lower:]')
	case "$os" in
	linux) echo linux ;;
	darwin) echo darwin ;;
	msys* | mingw* | cygwin*) err "on Windows use the PowerShell installer, see the README" ;;
	*) err "unsupported operating system: $os" ;;
	esac
}

detect_arch() {
	arch=$(uname -m)
	case "$arch" in
	x86_64 | amd64) echo amd64 ;;
	aarch64 | arm64) echo arm64 ;;
	*) err "unsupported architecture: $arch" ;;
	esac
}

pick_install_dir() {
	if [ -n "$INSTALL_DIR" ]; then
		echo "$INSTALL_DIR"
		return
	fi
	for dir in /usr/local/bin "$HOME/.local/bin"; do
		if [ -d "$dir" ] && [ -w "$dir" ]; then
			echo "$dir"
			return
		fi
	done
	if [ -w /usr/local ] || [ "$(id -u)" = "0" ]; then
		echo /usr/local/bin
		return
	fi
	echo "$HOME/.local/bin"
}

download() {
	_url=$1
	_out=$2
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$_url" -o "$_out"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO "$_out" "$_url"
	else
		err "curl or wget is required"
	fi
}

verify() {
	_file=$1
	_sums=$2
	_name=$(basename "$_file")

	if command -v sha256sum >/dev/null 2>&1; then
		_actual=$(sha256sum "$_file" | cut -d' ' -f1)
	elif command -v shasum >/dev/null 2>&1; then
		_actual=$(shasum -a 256 "$_file" | cut -d' ' -f1)
	else
		echo "install: no sha256 tool found, skipping checksum verification" >&2
		return
	fi

	_expected=$(grep "[ *]${_name}\$" "$_sums" | cut -d' ' -f1)
	[ -n "$_expected" ] || err "no checksum listed for ${_name}"
	[ "$_actual" = "$_expected" ] || err "checksum mismatch for ${_name}"
	echo "Checksum verified."
}

main() {
	need uname
	need tar
	need mktemp

	os=$(detect_os)
	arch=$(detect_arch)
	archive="${BINARY}_${os}_${arch}.tar.gz"

	if [ -n "${BASE_URL:-}" ]; then
		base="$BASE_URL"
	elif [ "$VERSION" = "latest" ]; then
		base="https://github.com/${REPO}/releases/latest/download"
	else
		base="https://github.com/${REPO}/releases/download/${VERSION}"
	fi

	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT INT TERM

	echo "Downloading ${BINARY} (${os}/${arch})..."
	download "${base}/${archive}" "${tmp}/${archive}" ||
		err "could not download ${base}/${archive}"

	if download "${base}/checksums.txt" "${tmp}/checksums.txt" 2>/dev/null; then
		verify "${tmp}/${archive}" "${tmp}/checksums.txt"
	else
		echo "install: checksums.txt not available, skipping verification" >&2
	fi

	tar -xzf "${tmp}/${archive}" -C "$tmp"
	[ -f "${tmp}/${BINARY}" ] || err "archive did not contain ${BINARY}"
	chmod +x "${tmp}/${BINARY}"

	dir=$(pick_install_dir)
	mkdir -p "$dir"

	if [ -w "$dir" ]; then
		mv "${tmp}/${BINARY}" "${dir}/${BINARY}"
	elif command -v sudo >/dev/null 2>&1; then
		echo "Installing to ${dir}, this needs sudo."
		sudo mv "${tmp}/${BINARY}" "${dir}/${BINARY}"
	else
		err "${dir} is not writable and sudo is not available, set INSTALL_DIR to somewhere you can write"
	fi

	echo "Installed ${BINARY} to ${dir}/${BINARY}"

	case ":${PATH}:" in
	*":${dir}:"*) ;;
	*) echo "Note: ${dir} is not on your PATH, add it to your shell profile." ;;
	esac

	"${dir}/${BINARY}" --version 2>/dev/null || true
}

main "$@"
