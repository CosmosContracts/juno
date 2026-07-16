# 04 — Per-module migration notes

Per-module breakdown. Each entry: what the module does, what the SDK surface it touches is, where breakage is most likely, what to verify.

## High-risk

### `x/mint` — Juno-specific inflation fork

**What it is:** Fork of SDK's `x/mint` with Juno-specific issuance schedule. Not a thin wrapper — the inflation curve is rewritten and the keeper holds custom params.

**SDK surface:** `StakingKeeper` (queries bonded ratio + total supply), `BankKeeper` (mints to module account), `AccountKeeper` (module account creation). All three signatures can move between SDK minor versions.

**Where breakage shows up first:**
- `keeper.NewKeeper(...)` constructor in `app/keepers/keepers.go`
- `BeginBlocker` — uses staking and bank to compute and mint inflation each block; signature drift here means BeginBlock panics
- Param storage — older `x/params` subspace patterns may have been replaced in v30 already; verify the module reads its own collections

**To verify:**
- Bonded ratio query returns sane values pre/post upgrade
- A delegation/undelegation cycle doesn't perturb mint rate beyond expected per-block drift
- `BeginBlock` does not double-mint on the upgrade height

**Notes:** `readme.md` in `x/mint` calls out that `x/burn` exists *only* because of this fork. Don't try to "simplify" by replacing with stock SDK mint — that's a different chain.

### `x/burn` — paired with `x/mint`

**What it is:** Funnels burned tokens to a fixed module address compatible with the mint fork. See `x/burn/README.md`.

**SDK surface:** `BankKeeper` (transfer + burn). `WasmKeeper` (contracts can route burns through here).

**Where breakage shows up first:**
- Module account permissions in `app/modules.go` — must match `x/mint` accounts
- `BankKeeper.BurnCoins` signature
- Wasm message hook that exposes burn to contracts

**To verify:**
- Burning JUNO from a contract reduces total supply
- The fixed module address is consistent across upgrades — this is part of consensus state

### `x/cw-hooks` — staking lifecycle → contracts

**What it is:** Registers smart contracts to receive staking and validator lifecycle hooks (`AfterDelegationModified`, `BeforeValidatorSlashed`, etc.) as sudo messages. v2 (recently rewritten on this branch).

**SDK surface:** `StakingHooks` interface — Juno implements every method and dispatches to contracts. `GovHooks` similar.

**Where breakage shows up first:**
- New SDK versions occasionally add hook methods. If the implementation doesn't satisfy the full interface, it won't compile.
- Collections schema. If SDK collections API changed, every persisted set needs migration.
- `ContractFailureRemovalThreshold` — set to 3 in the v30 upgrade handler. The constant lives in module params; verify the param schema didn't shift.

**To verify:**
- A registered contract receives every hook event after a delegation transaction
- A contract that errors three times in a row is removed from the registry (the threshold logic)
- ictest-cwhooks (need to restore or rename — see `01-current-state.md`)

### `app/ante/decorators/handle_fees.go` — custom DeductFeeDecorator

**What it is:** Custom fork of SDK's `DeductFeeDecorator`. Coordinates feemarket (base fee), feepay (contract pays for user), feeshare (split fee to contract dev).

**SDK surface:** `BankKeeper.SendCoinsFromAccountToModule`, `BankKeeper.SendCoins`, fee-grant integration if enabled.

**Where breakage shows up first:**
- The SDK's `DeductFeeDecorator` API moves between minor versions — argument list, return type, internal field access
- `feemarket` keeper API for querying current base fee
- `feepay` keeper API for the contract balance lookup

**To verify:**
- A simple bank send pays the correct fee at the current base price
- A user calling a contract registered for feepay pays zero, and the contract's balance decreases by the fee
- A contract registered for feeshare receives its share, and the rest goes to fee distribution

## Medium-risk

### `x/feemarket` — Skip's AIMD EIP-1559

**What it is:** Dynamic base-fee module. Maintained by Skip (originally), ported into Juno on this branch. Keeper holds fee-market params + state (window, learning rate).

**SDK surface:** `appmodule.AppModule`. Reads consensus params for max block gas.

**Where breakage shows up first:**
- Block listener — base fee updates per-block based on gas usage. If `BeginBlock` / `EndBlock` signatures changed, hook breaks.
- Params storage. The v30 upgrade handler resets params; if the schema shifts, the reset itself fails.

### `x/feepay` — gasless UX for contract devs

**What it is:** Contract devs fund a balance; users calling those contracts pay zero gas. Implemented as an ante decorator + keeper-managed balance.

**SDK surface:** Bank, ante chain integration.

**Where breakage shows up first:**
- The ante decorator (see fee pipeline note above)
- Wasm message hook that lets a contract registers itself for feepay

### `x/feeshare` — Juno Prop 51 dev revenue split

**What it is:** Splits a configurable share of a contract's execution fee to its registered withdraw address. Has its own ante decorator at `x/feeshare/ante/`.

**SDK surface:** Bank send, distribution keeper for the non-share portion.

**Where breakage shows up first:**
- Withdraw address validation
- The split math (LegacyDec multiplications — these moved between SDK versions)

### `x/tokenfactory` — Osmosis-derived

**What it is:** Permissionless `factory/{creator}/{subdenom}` denoms with admin mint/burn/force-transfer.

**SDK surface:** Bank (almost everything), with custom logic for admin-controlled supply changes.

**Where breakage shows up first:**
- `BankKeeper.MintCoins`, `BurnCoins`, `SendCoinsFromAccountToAccount` signatures
- Wasmbinding message handlers (six messages exposed to contracts)
- Stargate query allow-list paths under `/osmosis.tokenfactory.v1beta1.*` — these are stable, but verify

## Low-risk

### `x/clock` — sudo per EndBlock

Lightweight. EndBlocker calls registered contracts via `WasmKeeper.Sudo`. If `Sudo` signature is stable (it has been across minor versions), this is fine.

### `x/drip` — fee-pool injection

Allow-listed senders push tokens into the fee_pool to airdrop stakers. Bank send + distribution keeper. Stable surface.

### `x/stream` — token streaming

Block listener + collections-stored streams. Reworked recently to "simple block listener". Stable as long as listener API doesn't shift.

### `x/wrappers/gov`

Wrapper around `x/gov` for migration / compat. Already touches gov keeper directly; whatever broke is already in the v30 commits. Verify gov v1 paths in `acceptedQueries.go` still resolve.

## Cross-cutting: wasmbindings

Files: `wasmbindings/{wasm.go, query_plugin.go, message_plugin.go, queries.go}`.

**Custom queries** (extend `wasmkeeper.QueryPlugins`):
- Bank metadata (proxy to `BankKeeper.GetDenomMetaData`)
- TokenFactory denom queries

**Custom messages** (extend `wasmkeeper.Messengers`):
- TokenFactory: 6 messages (CreateDenom, MintTokens, ChangeAdmin, BurnTokens, SetMetadata, ForceTransfer)

**Risk:** wasmd v0.70's plugin interfaces. Specifically:
- `wasmkeeper.WithMessageHandlerDecorator` and `wasmkeeper.WithQueryPlugins` argument types
- The `Messenger` interface — wasmd has shuffled this between versions

**Verify:** Build `wasmbindings/test/`. It will fail visibly if the plugin interfaces changed.

## Stargate allow-list

`app/keepers/acceptedQueries.go` — 25 paths. None are in deprecation territory in the v0.54 line. Re-verify each as part of phase 4.
