# Voting-power probe contract

Minimal, stateless CosmWasm contract used by the v31 interchaintest to exercise all three Juno `x/voting-snapshot` custom queries from inside wasmvm:

- `voting_power_at`
- `total_voting_power_at`
- `voting_power_over_range`

The checked-in optimized artifact is `../voting_power_probe.wasm`.

## Reproducible build

The authoritative build is the source-controlled [`Dockerfile`](Dockerfile). It pins:

- the official `rust:1.85.1-bookworm` linux/amd64 image by its platform-manifest digest (`sha256:bf7d87666c4da6eace19e06d21bc4859c6e2a5c97a21ac273b0e082112753cf0`); and
- Binaryen `version_120` by the SHA-256 of its upstream linux/amd64 release archive (`ddb097af51d1bdb17300d986b0de7d97422f1933dedb4c9eda3510e0bf4076cc`).

The image verifies `rustc` and `wasm-opt` version output, runs the Rust tests, builds with `Cargo.lock`, optimizes the contract, and rejects an artifact whose checksum is not the expected value. BuildKit can export only the resulting Wasm file:

```sh
cd interchaintest/contracts/voting-power-probe
rm -rf target/reproducible
docker buildx build --target artifact --output type=local,dest=target/reproducible .
cmp target/reproducible/voting_power_probe.wasm ../voting_power_probe.wasm
sha256sum target/reproducible/voting_power_probe.wasm ../voting_power_probe.wasm
```

This build is intentionally linux/amd64 because the pinned Binaryen archive is architecture-specific. The tag in the `FROM` line is descriptive; the digest, not the mutable tag, selects the Rust image. `rust-toolchain.toml` also pins direct host Cargo invocations to Rust 1.85.1, but host rebuilds are only equivalent when `wasm-opt --version` identifies version 120:

```sh
cargo test --locked
cargo build --release --locked --target wasm32-unknown-unknown
wasm-opt --disable-bulk-memory -Oz \
  target/wasm32-unknown-unknown/release/voting_power_probe.wasm \
  -o target/voting_power_probe.host.wasm
cmp target/voting_power_probe.host.wasm ../voting_power_probe.wasm
```

Expected SHA-256:

```text
124d427ff478ec1d026f5412b72892db76a961b066989a299f66bcfb13d0b11a  voting_power_probe.wasm
```

`.cargo/config.toml` makes CosmWasm's wasmvm host imports explicit for the pinned Rust toolchain's bundled linker. The setting applies only to the Wasm target.
