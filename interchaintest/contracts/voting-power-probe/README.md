# Voting-power probe contract

Minimal, stateless CosmWasm contract used by the v31 interchaintest to exercise all three Juno `x/voting-snapshot` custom queries from inside wasmvm:

- `voting_power_at`
- `total_voting_power_at`
- `voting_power_over_range`

The checked-in optimized artifact is `../voting_power_probe.wasm`.

## Reproducible build

Requirements: the Rust toolchain pinned by the repository environment, target `wasm32-unknown-unknown`, and Binaryen `wasm-opt`.

```sh
cargo test --locked
cargo build --release --locked --target wasm32-unknown-unknown
wasm-opt --enable-bulk-memory -Oz \
  target/wasm32-unknown-unknown/release/voting_power_probe.wasm \
  -o ../voting_power_probe.wasm
sha256sum ../voting_power_probe.wasm
```

Expected SHA-256:

```text
a074bc275b8b38eebd79db1603645ae71759a679385da2ccd82622f39f1f57af  voting_power_probe.wasm
```

`.cargo/config.toml` makes CosmWasm's wasmvm host imports explicit for Rust 1.96's stricter bundled linker. The setting applies only to the Wasm target.
