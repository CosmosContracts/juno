#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
# shellcheck source=scripts/release/lib.sh
. "$ROOT/scripts/release/lib.sh"

: "${VERSION:?VERSION must be the release tag (v31.x.y)}"
: "${COMMIT:?COMMIT must be the full tagged commit}"
: "${ARCH:?ARCH must be amd64 or arm64}"
: "${OUT_DIR:?OUT_DIR is required}"
validate_version "$VERSION"
validate_commit "$COMMIT"
[ "$(git -C "$ROOT" rev-parse HEAD)" = "$COMMIT" ] || {
	echo "checkout does not match requested commit" >&2
	exit 1
}
[ "$(git -C "$ROOT" rev-list -n1 "$VERSION^{commit}")" = "$COMMIT" ] || {
	echo "tag does not resolve to requested commit" >&2
	exit 1
}
: "${SOURCE_DATE_EPOCH:=$(git -C "$ROOT" show -s --format=%ct "$COMMIT")}"
export SOURCE_DATE_EPOCH

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM
build_once() {
	destination=$1
	docker buildx build --platform "linux/$ARCH" --file "$ROOT/release.Dockerfile" \
		--build-arg "VERSION=$VERSION" --build-arg "COMMIT=$COMMIT" \
		--build-arg "SOURCE_DATE_EPOCH=$SOURCE_DATE_EPOCH" \
		--output "type=local,dest=$destination" "$ROOT"
}
build_once "$tmp/first"
# A second independent export is a release gate, not merely a package test.
build_once "$tmp/second"
cmp "$tmp/first/junod" "$tmp/second/junod" || {
	echo "binary rebuild was not byte-identical" >&2
	exit 1
}
binary="$tmp/first/junod"
[ -x "$binary" ]
long=$($binary version --long)
printf '%s\n' "$long" | grep -Fq "version: $VERSION"
printf '%s\n' "$long" | grep -Fq "commit: $COMMIT"
package_binary "$binary" "$ARCH" "$VERSION" "$COMMIT" "$OUT_DIR"
