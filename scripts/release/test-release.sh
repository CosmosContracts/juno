#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
TMP=${TMPDIR:-/tmp}/juno-release-test.$$
trap 'rm -rf "$TMP"' EXIT INT TERM
mkdir -p "$TMP/bin" "$TMP/out" "$TMP/existing"

fail() {
	printf 'not ok - %s\n' "$1" >&2
	exit 1
}
pass() { printf 'ok - %s\n' "$1"; }

# shellcheck source=scripts/release/lib.sh
. "$ROOT/scripts/release/lib.sh"

validate_version v31.2.3 || fail "valid semantic version rejected"
if validate_version 31.2.3 2>/dev/null || validate_version v31.2 2>/dev/null || validate_version 'v31.2.3;id' 2>/dev/null; then
	fail "invalid semantic version accepted"
fi
pass "strict v31 semantic version validation"

commit=68d1cbd3c6a490bdbf9848ce4d55b8a10c82cda5
workflow_sha=4444444444444444444444444444444444444444
validate_commit "$commit" || fail "valid commit rejected"
if validate_commit deadbeef 2>/dev/null; then fail "short commit accepted"; fi
pass "full commit validation"

git init -q "$TMP/tag-repo"
git -C "$TMP/tag-repo" config user.name release-test
git -C "$TMP/tag-repo" config user.email release-test@example.invalid
printf 'release source\n' >"$TMP/tag-repo/source"
git -C "$TMP/tag-repo" add source
git -C "$TMP/tag-repo" commit -qm initial
tag_commit=$(git -C "$TMP/tag-repo" rev-parse HEAD)
git -C "$TMP/tag-repo" branch v31.2.3
if resolve_local_tag_commit "$TMP/tag-repo" v31.2.3 2>/dev/null; then fail "branch accepted as a release tag"; fi
git -C "$TMP/tag-repo" tag -a v31.2.3 -m release
[ "$(resolve_local_tag_commit "$TMP/tag-repo" v31.2.3)" = "$tag_commit" ] || fail "annotated tag did not peel to commit"
pass "release identity requires and peels an exact Git tag"

printf '#!/bin/sh\ncase "$1" in version) printf "v31.2.3\\n";; esac\n' >"$TMP/bin/junod"
chmod +x "$TMP/bin/junod"
SOURCE_DATE_EPOCH=1700000000 package_binary "$TMP/bin/junod" amd64 v31.2.3 "$commit" "$TMP/out"
archive="$TMP/out/juno-v31.2.3-linux-amd64.tar.gz"
[ -f "$archive" ] || fail "archive missing"
tar -xOf "$archive" BUILD-METADATA.json | grep -Fq "\"commit\":\"$commit\"" || fail "commit absent from archive metadata"
tar -xOf "$archive" BUILD-METADATA.json | grep -Fq '"version":"v31.2.3"' || fail "version absent from archive metadata"
first=$(sha256sum "$archive" | cut -d ' ' -f 1)
rm -rf "$TMP/out" && mkdir "$TMP/out"
SOURCE_DATE_EPOCH=1700000000 package_binary "$TMP/bin/junod" amd64 v31.2.3 "$commit" "$TMP/out"
second=$(sha256sum "$TMP/out/juno-v31.2.3-linux-amd64.tar.gz" | cut -d ' ' -f 1)
[ "$first" = "$second" ] || fail "clean package rebuild differs"
pass "byte-identical deterministic archive with embedded metadata"

printf '\tdep\texample.com/dependency\tv1.0.0\th1:first\n' >"$TMP/out/junod-linux-amd64.modules"
printf 'build-base-0.5-r3\ngit-2.49.1-r0\n' >"$TMP/out/build-dependencies-linux-amd64.txt"

generate_checksums "$TMP/out"
(cd "$TMP/out" && sha256sum --check SHA256SUMS)
listed=$(wc -l <"$TMP/out/SHA256SUMS" | tr -d ' ')
[ "$listed" = 4 ] || fail "checksum manifest does not cover every binary/archive/dependency record"
pass "SHA-256 manifest covers binaries, archives, and dependency records"

python3 "$ROOT/scripts/release/metadata.py" --directory "$TMP/out" --version v31.2.3 \
	--commit "$commit" --repository https://github.com/juno-ai-dev/juno --workflow-sha "$workflow_sha"
python3 -m json.tool "$TMP/out/SBOM.spdx.json" >/dev/null
python3 -m json.tool "$TMP/out/provenance.intoto.jsonl" >/dev/null
grep -Fq "$first" "$TMP/out/provenance.intoto.jsonl" || fail "archive absent from provenance subjects"
grep -Fq 'example.com/dependency' "$TMP/out/SBOM.spdx.json" || fail "Go dependency absent from SBOM"
grep -Fq 'build-base-0.5-r3' "$TMP/out/provenance.intoto.jsonl" || fail "builder package absent from provenance"
grep -Fq "release.yml@$workflow_sha" "$TMP/out/provenance.intoto.jsonl" || fail "trusted workflow SHA absent from builder identity"
sbom_first=$(sha256sum "$TMP/out/SBOM.spdx.json" | cut -d ' ' -f 1)
provenance_first=$(sha256sum "$TMP/out/provenance.intoto.jsonl" | cut -d ' ' -f 1)
printf '	dep	example.com/dependency	v1.0.1	h1:second\n' >"$TMP/out/junod-linux-amd64.modules"
python3 "$ROOT/scripts/release/metadata.py" --directory "$TMP/out" --version v31.2.3 \
	--commit "$commit" --repository https://github.com/juno-ai-dev/juno --workflow-sha "$workflow_sha"
[ "$sbom_first" != "$(sha256sum "$TMP/out/SBOM.spdx.json" | cut -d ' ' -f 1)" ] || fail "SBOM ignored dependency version change"
[ "$provenance_first" != "$(sha256sum "$TMP/out/provenance.intoto.jsonl" | cut -d ' ' -f 1)" ] || fail "provenance ignored dependency version change"
pass "dependency-bearing SPDX and SLSA metadata changes with dependencies"

require_absent_http_status 404 fixture || fail "404 absence rejected"
if require_absent_http_status 200 fixture 2>/dev/null; then fail "existing identity accepted"; fi
if require_absent_http_status 500 fixture 2>/dev/null; then fail "registry/server failure accepted as absence"; fi
pass "existence guards distinguish absence and fail closed"

workflow="$ROOT/.github/workflows/release.yml"
grep -Fq 'types: [edited]' "$workflow" && fail "release edits trigger publication"
grep -E '^[[:space:]]*- uses:' "$workflow" | grep -Ev '@[0-9a-f]{40}([[:space:]]|$)' >/dev/null && fail "release action is not pinned to a full SHA"
grep -Fq 'linux/amd64,linux/arm64' "$workflow" || fail "supported platforms absent"
grep -Fq 'provenance: mode=max' "$workflow" || fail "container provenance absent"
grep -Fq 'sbom: true' "$workflow" || fail "container SBOM absent"
grep -Fq 'existence guard failed closed' "$ROOT/scripts/release/lib.sh" || fail "fail-closed replay guard absent"
grep -Fq 'release-publication-${{ github.repository }}' "$workflow" || fail "global release serialization absent"
grep -Fq 'ref: ${{ needs.guard.outputs.commit }}' "$workflow" || fail "validated commit checkout absent"
grep -Fq -- '--no-cache --pull' "$ROOT/scripts/release/build.sh" || fail "clean BuildKit rebuild guard absent"
grep -Fq 'APK_BUILD_BASE=0.5-r3' "$ROOT/release.Dockerfile" || fail "builder packages are not version pinned"
pass "workflow identity, replay, pinning, architecture, SBOM and provenance guards"

# Exercise the operator checksum/extract/version/install commands from the guide
# against the deterministic fixture (network-only gh download is intentionally skipped).
(
	cd "$TMP/out" && VERSION=v31.2.3 ARCH=amd64 HOME="$TMP/home" sh -eu <<'COMMANDS'
mkdir -p "$HOME/.local/bin"
grep -E " (junod-linux-$ARCH|juno-$VERSION-linux-$ARCH.tar.gz)$" SHA256SUMS | sha256sum --check
tar --extract --gzip --file "juno-$VERSION-linux-$ARCH.tar.gz"
./"junod-linux-$ARCH" version --long
install -m 0755 "junod-linux-$ARCH" "$HOME/.local/bin/junod"
"$HOME/.local/bin/junod" version --long
COMMANDS
)
pass "validator checksum and installation commands run verbatim"
