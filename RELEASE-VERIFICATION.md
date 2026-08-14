# Verify and install a v31 release

Release artifacts are immutable: a replay aborts if the GitHub release already
exists. `SHA256SUMS` covers every attached binary, archive, SBOM, provenance,
container-digest report, and this guide. The container report records both the
manifest-list digest and its `linux/amd64` and `linux/arm64` child digests.

Set the release and architecture, download the files, and verify the selected
binary and archive **before** installing:

```sh
VERSION=v31.0.0
ARCH=amd64 # use arm64 on 64-bit ARM Linux
mkdir "juno-$VERSION" && cd "juno-$VERSION"
gh release download "$VERSION" --repo juno-ai-dev/juno
grep -E " (junod-linux-$ARCH|juno-$VERSION-linux-$ARCH.tar.gz)$" SHA256SUMS | sha256sum --check
tar --extract --gzip --file "juno-$VERSION-linux-$ARCH.tar.gz"
./"junod-linux-$ARCH" version --long
install -m 0755 "junod-linux-$ARCH" "$HOME/.local/bin/junod"
"$HOME/.local/bin/junod" version --long
```

The two `version --long` outputs must report the requested semantic version and
the full commit from `provenance.intoto.jsonl`. Independently inspect metadata:

```sh
jq -r '.predicate.buildDefinition.externalParameters' provenance.intoto.jsonl
jq . container-digests.json
gh attestation verify "junod-linux-$ARCH" --repo juno-ai-dev/juno
docker buildx imagetools inspect "$(jq -r .image container-digests.json)@$(jq -r .manifest_digest container-digests.json)"
```

Rebuilding uses the digest-pinned Go 1.25.10/Alpine 3.22 builder in
`release.Dockerfile`, `SOURCE_DATE_EPOCH` from the tagged commit, trimmed paths,
and an empty Go build ID. CI builds and compares each binary twice and creates
deterministic tar/gzip archives. Alpine repositories and GitHub runner/QEMU are
external services; their identities are recorded or pinned where available,
but their availability is not reproducible offline. Payload packaging, SBOM,
provenance, strict input validation, and replay refusal are tested offline by
`scripts/release/test-release.sh`.