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
REPOSITORY=CosmosContracts/juno
mkdir "juno-$VERSION" && cd "juno-$VERSION"
gh release download "$VERSION" --repo "$REPOSITORY"
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
gh attestation verify "junod-linux-$ARCH" --repo "$REPOSITORY"
docker buildx imagetools inspect "$(jq -r .image container-digests.json)@$(jq -r .manifest_digest container-digests.json)"
```

Rebuilding uses the digest-pinned Go 1.25.10/Alpine 3.22 builder in
`release.Dockerfile`, `SOURCE_DATE_EPOCH` from the tagged commit, trimmed paths,
and an empty Go build ID. CI builds each binary in two independent cacheless
BuildKit builders, compares the binaries and dependency records, and creates
deterministic tar/gzip archives. Direct Alpine package versions are pinned and
the complete installed package set plus Go's embedded module build information
are attached and represented in SBOM/provenance. Alpine repositories and GitHub
runner/QEMU remain external availability dependencies. Payload packaging,
dependency mutation, metadata, strict tag identity, and replay refusal are
tested offline by `scripts/release/test-release.sh`.