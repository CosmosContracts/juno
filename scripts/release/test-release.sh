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
validate_commit "$commit" || fail "valid commit rejected"
if validate_commit deadbeef 2>/dev/null; then fail "short commit accepted"; fi
pass "full commit validation"

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

generate_checksums "$TMP/out"
(cd "$TMP/out" && sha256sum --check SHA256SUMS)
listed=$(wc -l <"$TMP/out/SHA256SUMS" | tr -d ' ')
[ "$listed" = 2 ] || fail "checksum manifest does not cover every binary/archive"
pass "SHA-256 manifest covers release archives"

python3 "$ROOT/scripts/release/metadata.py" --directory "$TMP/out" --version v31.2.3 \
	--commit "$commit" --repository https://github.com/juno-ai-dev/juno
python3 -m json.tool "$TMP/out/SBOM.spdx.json" >/dev/null
python3 -m json.tool "$TMP/out/provenance.intoto.jsonl" >/dev/null
grep -Fq "$first" "$TMP/out/provenance.intoto.jsonl" || fail "archive absent from provenance subjects"
pass "offline SPDX SBOM and in-toto/SLSA provenance generation"

printf 'juno-v31.2.3-linux-amd64.tar.gz\n' >"$TMP/existing/assets.txt"
if refuse_replay "$TMP/existing/assets.txt" 2>/dev/null; then fail "replay accepted existing assets"; fi
: >"$TMP/existing/assets.txt"
refuse_replay "$TMP/existing/assets.txt" || fail "empty release rejected"
pass "release replay refuses existing artifacts"

workflow="$ROOT/.github/workflows/release.yml"
grep -Fq 'types: [edited]' "$workflow" && fail "release edits trigger publication"
grep -Eq 'uses: [^ ]+@v[0-9]' "$workflow" && fail "release action is not SHA-pinned"
grep -Fq 'linux/amd64,linux/arm64' "$workflow" || fail "supported platforms absent"
grep -Fq 'provenance: mode=max' "$workflow" || fail "container provenance absent"
grep -Fq 'sbom: true' "$workflow" || fail "container SBOM absent"
grep -Fq 'imagetools inspect "$IMAGE:$ref"' "$workflow" || fail "container replay guard absent"
pass "workflow immutability, pinning, architectures, SBOM and provenance"

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
