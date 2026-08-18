# DAO DAO cw4 interchaintest artifacts

These WebAssembly binaries are checked in so the interchaintest has no network
dependency and never downloads executable bytes at runtime. The test computes
SHA-256 over every required file before storing any contract on chain.

## Provenance

| File | Authoritative tagged release asset | SHA-256 |
| --- | --- | --- |
| `dao_dao_core.wasm` | [`DA0-DA0/dao-contracts` v2.7.0](https://github.com/DA0-DA0/dao-contracts/releases/download/v2.7.0/dao_dao_core.wasm) | `5d078fc9aec04df18c335446eb8df03d24c73ee745f76fd39624d4c5fa768b4c` |
| `dao_proposal_single.wasm` | [`DA0-DA0/dao-contracts` v2.7.0](https://github.com/DA0-DA0/dao-contracts/releases/download/v2.7.0/dao_proposal_single.wasm) | `e38fc5bb1b5e74ef154340567c673492515498b2120e5f15b0c990cd9fa5fe6a` |
| `dao_voting_cw4.wasm` | [`DA0-DA0/dao-contracts` v2.7.0](https://github.com/DA0-DA0/dao-contracts/releases/download/v2.7.0/dao_voting_cw4.wasm) | `d0e6bac4d7c1861f36328e7c0367f863999f999e2ae21df612e301eea5fe90d8` |
| `cw4_group.wasm` | [`CosmWasm/cw-plus` v1.1.2](https://github.com/CosmWasm/cw-plus/releases/download/v1.1.2/cw4_group.wasm) | `dd2216f1114fc68bc4c043701b02e55ce3e5598cdeb616985388215a400db277` |
| `voting_power_probe.wasm` | Source-controlled build in [`voting-power-probe/`](voting-power-probe/) | `124d427ff478ec1d026f5412b72892db76a961b066989a299f66bcfb13d0b11a` |

The DAO DAO values are copied from that tagged GitHub release's
`checksums.txt`. The cw4-group value is copied from the cw-plus v1.1.2
release's `checksums.txt`. Release assets were fetched once during repository
maintenance, compared with those manifests, and then committed. They are not
built ad hoc or fetched by test code.

To audit the checked-in bytes locally:

```sh
cd interchaintest/contracts
sha256sum cw4_group.wasm dao_voting_cw4.wasm dao_proposal_single.wasm dao_dao_core.wasm voting_power_probe.wasm
```
