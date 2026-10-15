#!/bin/sh
set -eu

BINARY="treegen"
MODULE="github.com/callmegautam/treegen"
DIST="dist"
TARGETS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"

VERSION=""
PUBLISH=0

err() {
	echo "release: $*" >&2
	exit 1
}

usage() {
	echo "usage: ./release.sh vX.Y.Z [--publish]"
	echo
	echo "  Builds every release archive into ./${DIST} and writes checksums.txt."
	echo "  --publish also creates the GitHub release and uploads them."
	exit 1
}

for arg in "$@"; do
	case "$arg" in
	--publish) PUBLISH=1 ;;
	-h | --help) usage ;;
	-*) err "unknown flag: $arg" ;;
	*)
		[ -z "$VERSION" ] || err "version given twice"
		VERSION="$arg"
		;;
	esac
done

[ -n "$VERSION" ] || usage

case "$VERSION" in
v[0-9]*.[0-9]*.[0-9]*) ;;
*) err "version must look like v1.2.3, got ${VERSION}" ;;
esac

command -v go >/dev/null 2>&1 || err "go is required"
command -v tar >/dev/null 2>&1 || err "tar is required"
command -v zip >/dev/null 2>&1 || err "zip is required"
command -v sha256sum >/dev/null 2>&1 || err "sha256sum is required"

if [ "$PUBLISH" = "1" ]; then
	command -v gh >/dev/null 2>&1 || err "gh is required to publish"
	[ -z "$(git status --porcelain)" ] || err "working tree is dirty, commit before publishing"
fi

echo "Running tests..."
go vet ./...
go test ./...

rm -rf "$DIST"
mkdir -p "$DIST"

for target in $TARGETS; do
	os=${target%/*}
	arch=${target#*/}

	ext=""
	if [ "$os" = "windows" ]; then
		ext=".exe"
	fi

	stage="${DIST}/stage"
	rm -rf "$stage"
	mkdir -p "$stage"

	echo "Building ${os}/${arch}..."
	CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build \
		-trimpath \
		-ldflags "-s -w -X ${MODULE}/cmd.version=${VERSION}" \
		-o "${stage}/${BINARY}${ext}" .

	cp README.md LICENSE "$stage"

	if [ "$os" = "windows" ]; then
		archive="${BINARY}_${os}_${arch}.zip"
		(cd "$stage" && zip -q "../${archive}" "${BINARY}${ext}" README.md LICENSE)
	else
		archive="${BINARY}_${os}_${arch}.tar.gz"
		tar -czf "${DIST}/${archive}" -C "$stage" "${BINARY}${ext}" README.md LICENSE
	fi

	rm -rf "$stage"
done

(cd "$DIST" && sha256sum ./*.tar.gz ./*.zip | sed 's| \./| |' >checksums.txt)

echo
echo "Artifacts in ${DIST}:"
ls -1 "$DIST"

if [ "$PUBLISH" = "0" ]; then
	echo
	echo "Nothing published. Re-run with --publish to create the GitHub release."
	exit 0
fi

echo
echo "Creating GitHub release ${VERSION}..."
gh release create "$VERSION" "${DIST}"/* --title "$VERSION" --generate-notes
