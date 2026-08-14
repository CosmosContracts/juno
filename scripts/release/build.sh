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
	sequence=$2
	builder="juno-release-${ARCH}-$$-$sequence"
	(
		trap 'docker buildx rm --force "$builder" >/dev/null 2>&1 || true' EXIT INT TERM
		docker buildx create --name "$builder" --driver docker-container >/dev/null
		docker buildx build --builder "$builder" --no-cache --pull \
			--platform "linux/$ARCH" --file "$ROOT/release.Dockerfile" \
			--build-arg "VERSION=$VERSION" --build-arg "COMMIT=$COMMIT" \
			--build-arg "SOURCE_DATE_EPOCH=$SOURCE_DATE_EPOCH" \
			--output "type=local,dest=$destination" "$ROOT"
	)
}
build_once "$tmp/first" first
# Separate cacheless BuildKit builders make this a clean rebuild gate rather
# than two exports of one cached compilation.
build_once "$tmp/second" second
for artifact in junod junod.modules build-dependencies.txt; do
	cmp "$tmp/first/$artifact" "$tmp/second/$artifact" || {
		echo "clean rebuild differed: $artifact" >&2
		exit 1
	}
done
binary="$tmp/first/junod"
[ -x "$binary" ]
long=$($binary version --long)
printf '%s\n' "$long" | grep -Fq "version: $VERSION"
printf '%s\n' "$long" | grep -Fq "commit: $COMMIT"
package_binary "$binary" "$ARCH" "$VERSION" "$COMMIT" "$OUT_DIR"
cp "$tmp/first/junod.modules" "$OUT_DIR/junod-linux-$ARCH.modules"
cp "$tmp/first/build-dependencies.txt" "$OUT_DIR/build-dependencies-linux-$ARCH.txt"
