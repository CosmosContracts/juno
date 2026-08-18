# Legacy proposer-builder protobuf compatibility

This directory preserves only the generated protobuf types needed to decode historical governance state after the proposer-builder module was removed.

Source: `github.com/skip-mev/pob/x/builder/types` at commit `5fdb53bc1feb` (`v0.0.0-20230906203828-5fdb53bc1feb`). The copied generated files are `tx.pb.go`, `genesis.pb.go`, and `codec.go`.

The package is codec-only: it has no keeper, stores, module manager registration, transaction routing, or runtime behavior. Removing it makes `junod export` and historical governance queries panic when they encounter `/pob.builder.v1.MsgUpdateParams` retained in mainnet state.
