#!/bin/sh
set -eu

ROOT=$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)
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
git -C "$TMP/tag-repo" tag v31.2.4
[ "$(resolve_local_tag_commit "$TMP/tag-repo" v31.2.4)" = "$tag_commit" ] || fail "lightweight tag did not resolve to commit"
pass "release identity requires and peels an exact Git tag"

git clone -q --bare "$TMP/tag-repo" "$TMP/tag-remote.git"
git -C "$TMP/tag-repo" remote add origin "$TMP/tag-remote.git"
identity=$(resolve_remote_tag_identity "$TMP/tag-repo" origin v31.2.3)
tag_oid=${identity%% *}
resolved_commit=${identity#* }
[ "$resolved_commit" = "$tag_commit" ] || fail "remote tag did not peel to expected commit"
identity=$(resolve_remote_tag_identity "$TMP/tag-repo" origin v31.2.4)
lightweight_oid=${identity%% *}
resolved_commit=${identity#* }
if [ "$lightweight_oid" != "$tag_commit" ] || [ "$resolved_commit" != "$tag_commit" ]; then
	fail "remote lightweight tag identity was not preserved"
fi
printf 'moved source\n' >>"$TMP/tag-repo/source"
git -C "$TMP/tag-repo" commit -qam moved
git -C "$TMP/tag-repo" tag -fa v31.2.3 -m moved
git -C "$TMP/tag-repo" push -qf origin refs/tags/v31.2.3
if resolve_remote_tag_identity "$TMP/tag-repo" origin v31.2.3 "$tag_oid" "$tag_commit" 2>/dev/null; then
	fail "moved remote tag passed bound identity validation"
fi
pass "remote tag object and peeled commit are consistently bound"

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

printf '\tdep\texample.com/original\tv1.0.0\th1:AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=\n\t=>\texample.com/replacement\tv1.4.0\th1:AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI=\n\tdep\texample.com/a+b\tv1.0.0\th1:AwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwM=\n\t=>\texample.com/shared\tv1.0.0\th1:BAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ=\n\tdep\texample.com/a_b\tv1.0.0\th1:BQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQU=\n\t=>\texample.com/shared\tv1.0.0\th1:BAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ=\n' >"$TMP/out/junod-linux-amd64.modules"
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
python3 - "$TMP/out/provenance.intoto.jsonl" <<'PY' || fail "SLSA dependency digests are not lowercase hexadecimal"
import json, re, sys
statement = json.load(open(sys.argv[1]))
for dependency in statement["predicate"]["buildDefinition"]["resolvedDependencies"]:
    for value in dependency.get("digest", {}).values():
        assert re.fullmatch(r"[0-9a-f]+", value), value
    if dependency.get("name") in {"original Go dependency", "selected replacement for example.com/original@v1.0.0", "selected replacement for example.com/a+b@v1.0.0", "selected replacement for example.com/a_b@v1.0.0"}:
        assert set(dependency.get("digest", {})) == {"dirHash1"}, dependency
PY
grep -Fq "$first" "$TMP/out/provenance.intoto.jsonl" || fail "archive absent from provenance subjects"
grep -Fq 'example.com/original' "$TMP/out/SBOM.spdx.json" || fail "original Go dependency absent from SBOM"
grep -Fq 'example.com/replacement' "$TMP/out/SBOM.spdx.json" || fail "selected Go replacement absent from SBOM"
grep -Fq 'build-base-0.5-r3' "$TMP/out/provenance.intoto.jsonl" || fail "builder package absent from provenance"
grep -Fq "release.yml@$workflow_sha" "$TMP/out/provenance.intoto.jsonl" || fail "trusted workflow SHA absent from builder identity"
sbom_first=$(sha256sum "$TMP/out/SBOM.spdx.json" | cut -d ' ' -f 1)
provenance_first=$(sha256sum "$TMP/out/provenance.intoto.jsonl" | cut -d ' ' -f 1)
printf '	dep	example.com/original	v1.0.0	h1:AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=\n	=>	example.com/replacement	v1.4.1	h1:BgYGBgYGBgYGBgYGBgYGBgYGBgYGBgYGBgYGBgYGBgY=\n	dep	example.com/a+b	v1.0.0	h1:AwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwM=\n	=>	example.com/shared	v1.0.0	h1:BAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ=\n	dep	example.com/a_b	v1.0.0	h1:BQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQU=\n	=>	example.com/shared	v1.0.0	h1:BAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ=\n' >"$TMP/out/junod-linux-amd64.modules"
python3 "$ROOT/scripts/release/metadata.py" --directory "$TMP/out" --version v31.2.3 \
	--commit "$commit" --repository https://github.com/juno-ai-dev/juno --workflow-sha "$workflow_sha"
[ "$sbom_first" != "$(sha256sum "$TMP/out/SBOM.spdx.json" | cut -d ' ' -f 1)" ] || fail "SBOM ignored selected replacement mutation"
[ "$provenance_first" != "$(sha256sum "$TMP/out/provenance.intoto.jsonl" | cut -d ' ' -f 1)" ] || fail "provenance ignored selected replacement mutation"
python3 - "$TMP/out/SBOM.spdx.json" <<'PY' || fail "SPDX identifiers collide"
import json, sys
document = json.load(open(sys.argv[1]))
packages = document["packages"]
ids = [package["SPDXID"] for package in packages]
assert len(ids) == len(set(ids))
assert all(len(identifier.rsplit("-", 1)[-1]) == 16 for identifier in ids)
shared = [package for package in packages if package["name"] == "example.com/shared"]
assert len(shared) == 1
shared_id = shared[0]["SPDXID"]
variants = [item for item in document["relationships"] if item["spdxElementId"] == shared_id and item["relationshipType"] == "VARIANT_OF"]
assert len(variants) == 2
PY
mkdir -p "$TMP/local-replacement"
printf 'fixture\n' >"$TMP/local-replacement/junod-linux-amd64"
printf '\tdep\texample.com/original\tv1.0.0\th1:AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=\n\t=>\t../local-module\t(devel)\n' >"$TMP/local-replacement/junod-linux-amd64.modules"
if python3 "$ROOT/scripts/release/metadata.py" --directory "$TMP/local-replacement" --version v31.2.3 \
	--commit "$commit" --repository https://github.com/juno-ai-dev/juno --workflow-sha "$workflow_sha" 2>/dev/null; then
	fail "local Go replacement produced misleading release metadata"
fi
pass "SPDX and SLSA include mutation-sensitive original/replacement identities and collision-resistant IDs"

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
grep -Fq 'path: source' "$workflow" || fail "container source checkout is not isolated"
grep -Fq 'path: policy' "$workflow" || fail "trusted container policy checkout is absent"
grep -Fq '. policy/scripts/release/lib.sh' "$workflow" || fail "container write job executes tag-controlled policy"
grep -Fq 'context: source' "$workflow" || fail "container build context is not bound to validated source"
grep -Fq 'file: source/release.Dockerfile' "$workflow" || fail "container Dockerfile is not bound to validated source"
grep -Fq 'EVENT_AFTER: ${{ github.event.after }}' "$workflow" || fail "tag push is not bound to event.after"
grep -Fq 'if [ "$EVENT_AFTER" != "$tag_oid" ]; then' "$workflow" || fail "tag push does not bind the exact event ref object"
grep -Fq "printf 'Authorization: Bearer %s' \"\$GH_TOKEN\"" "$workflow" || fail "authorization token is not interpolated"
python3 - "$workflow" <<'PY' || fail "authorization format discards the token argument"
import pathlib, sys
lines = [line for line in pathlib.Path(sys.argv[1]).read_text().splitlines() if "github_auth=$(printf" in line]
assert len(lines) == 1
assert [ord(char) for char in "%s"] == [37, 115]
assert [37, 115] == [ord(char) for char in lines[0][lines[0].index("Bearer ") + 7:][:2]]
PY
grep -Fq 'push-by-digest=true,name-canonical=true,push=true' "$workflow" || fail "digest-only container publication absent"
if grep -Fq '${{ env.IMAGE }}:' "$workflow"; then fail "mutable container tag publication present"; fi
grep -Fq 'resolve_remote_tag_identity' "$workflow" || fail "side-effect source revalidation absent"
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
