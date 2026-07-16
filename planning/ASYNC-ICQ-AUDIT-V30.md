# async-icq counterparty audit — juno-1 mainnet

**Verdict: clear.** v30 deletes the `interchainquery` store at upgrade
height; **no live ICQ channels** exist on juno-1 mainnet to be
disrupted.

## Method

Queried `juno-1` via Polkachu RPC at app version `v29.1.0`, height
`~37726617` (May 2026):

```bash
curl -s "https://juno-api.polkachu.com/ibc/core/channel/v1/channels?pagination.limit=1000"
```

Filtered the result for any channel whose `port_id` is `icqhost`,
`icqcontroller`, or contains `icq` (case-insensitive).

## Result

- **Total IBC channels on juno-1:** 674
- **Channels with an `icq*` port_id:** 0
- **Pagination:** complete (`next_key: null`); the full channel set
  fits in one page

No counterparty chain has an open ICQ channel against juno-1. No
pending packets to drain. No coordinated channel-close required.

## Implication for v30

`StoreUpgrades.Deleted` in `app/upgrades/v30/constants.go` includes
`"interchainquery"`. At upgrade height the store is purged. Since no
channels exist, no counterparty observes a state break.

If counterparties later try to *open* a new ICQ channel against the
v30 binary, the open will fail at the IBC-router layer (no `icqhost`
port registered). That's the intended behavior until `ibc-apps`
publishes a `/v10` line and we reintroduce async-icq via a v30.x
patch (per `planning/09-deferred-work.md` §B10).

## Re-verification before halt-height

Re-run this audit immediately before submitting the v30 upgrade
proposal to mainnet governance. The window between this audit and the
proposal might let a counterparty open an ICQ channel; if that
happens, hold the proposal until coordinated.

```bash
# Re-verification command (one-liner)
curl -s "https://juno-api.polkachu.com/ibc/core/channel/v1/channels?pagination.limit=1000" \
  | jq '[.channels[] | select(.port_id | test("icq"; "i"))] | length'
```

Expected: `0`. Anything else is a hold-the-proposal signal.
