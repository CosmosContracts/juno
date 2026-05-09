# ictest blocker — chain RPC not reachable from host

**Status: blocking A1 + A2.** Needs a focused debugging session.

## Symptom

`make ictest-basic` fails reliably, even with `make local-image` producing a healthy `ghcr.io/cosmoscontracts/juno:local` (396MB, statically linked, junod v29.0.0-99-g... reports SDK v0.53.7 + Comet v0.38.23 in `version --long`).

The chain *runs successfully* inside the container — interchaintest's INFO logs show consensus committing blocks at `height=49` with finalized state. But interchaintest's `Build` step fails with:

```
interchaintest Build returned error: failed to start chains:
failed to start chain juno: All attempts fail:
  #1: post failed: Post "http://0.0.0.0:37505": dial tcp 0.0.0.0:37505: connect: connection refused
  #2: ... (repeated ~40 times)
```

`37505` is the host-side mapped port for the container's `26657` (Tendermint RPC). interchaintest is dialing the RPC to perform a startup health check; the connection is refused.

## What we know

1. The chain process inside the container runs. Blocks commit. Consensus, state, and txindex modules log normally.
2. `interchaintest` does override `config.toml` `rpc.laddr` → `tcp://0.0.0.0:26657` at chain init (verified in the `interchaintest@v10.0.0/chain/cosmos/chain_node.go` source).
3. Our `scripts/init.sh` does NOT override `rpc.laddr` (only sets CORS), so it shouldn't be clobbering the interchaintest setup.
4. The cleanup panics on `s.Ic.Close()` because `Build` returned an error and `Ic` is nil. This masks the underlying error in the test's stack trace; the actual error is now logged via `t.Logf` in `interchaintest/suite/lib.go` (committed as a debug aid).
5. A standalone `junod start` against an empty data dir panics with "validator set is empty after InitGenesis" — that's a different startup path (no gentx) and not the same failure mode interchaintest hits.

## Hypotheses for the focused session

### H1. Docker port-forward race

interchaintest binds the container's 26657 to a random host port (37505 here). If interchaintest starts polling before Docker has wired the forward, "connection refused" is the expected failure. Plausible if our chain takes longer to start than v29 did and pushes the polling window into the not-yet-bound region.

**Test:** add a fixed sleep before the first poll in `chain_node.go`'s start path, or check Docker's port mapping table at the time of failure (`docker port <container>` in another shell).

### H2. RPC server isn't actually starting

cometbft's RPC server starts during `node.Start()`. If our binary errors during `node.Start()` after consensus comes up, blocks would commit but RPC wouldn't bind. The chain logs we captured don't show any "RPC server" / "started RPC" lines, but they were truncated — need a larger log capture.

**Test:** attach `docker logs -f <container>` to a live ictest run; look for RPC binding messages or RPC errors.

### H3. wasmvm v3 / cosmos-sdk v0.53.7 startup race

A new keeper init order or a cgo init might block the RPC server thread for >2 minutes. Less likely because consensus IS up — same goroutine runtime.

**Test:** trace which goroutines are alive inside the container at the time the host port is unreachable. `docker exec <container> ps -ef` and look for `junod` threads.

### H4. interchaintest expects a /status response shape that we don't return

If `Post "/" ` to the RPC isn't the right probe — but this is generic interchaintest code that works for every other Cosmos chain. Less likely.

## Ruled out

- buildx / docker version: buildx 0.20.0 installed, image builds clean
- libwasmvm static linking: Dockerfile fixed (/v2 → /v3), binary is statically linked
- Build timeout too short: bumped 2min → 5min, no improvement (test still fails at 144s, not at the timeout boundary)

## Live debug aid

`interchaintest/suite/lib.go` now logs the Build error before `require.NoError` calls `t.FailNow`. The log line `lib.go:128: interchaintest Build returned error:` makes the actual error visible in test output for the next debugger.

The cleanup-panic in `chain_start_test.go:25` (`_ = s.Ic.Close()` panicking on nil `s.Ic`) is real but not fixed here — that's a defensive-coding fix worth its own commit, but it doesn't block the underlying RPC issue.

## Next-session checklist

1. Repro `make ictest-basic` with `docker logs -f juno-2-val-0-TestBasicTestSuite` running in another terminal. Capture every RPC-related log line.
2. While the test is running, in another shell: `docker exec juno-2-val-0-TestBasicTestSuite cat /var/cosmos-chain/juno-2/config/config.toml | grep -A1 '\[rpc\]'`. Confirm `laddr = "tcp://0.0.0.0:26657"`.
3. Same shell: `docker exec juno-2-val-0-TestBasicTestSuite ss -tlnp` (or `netstat -tln`). Is anything listening on 26657 inside the container?
4. From host: `docker port juno-2-val-0-TestBasicTestSuite`. Is the 26657 → host-port mapping established?
5. If RPC is listening inside the container but unreachable from host: Docker port-forward issue (H1).
6. If RPC is NOT listening inside the container but blocks are committing: cometbft RPC isn't starting (H2). Check cometbft startup logs.

## Workaround under consideration

If the focused session can't quickly resolve, we can:
- Skip ictest from the v30 PR; leave the workflow matrix in place but mark suites `xfail` for v30
- Run a manual smoke (start junod from a script, deploy a contract, call it) on a v30-rc1 testnet
- Block the v30 proposal on getting ictest green in a follow-up PR

This blocks A1 (DAO DAO contract compatibility) and A2 (`ictest-upgrade` against forked mainnet state). The DAO DAO artifacts are ready in `/workspace/dao-contracts/artifacts/` (29 wasm files including all v2.7.0 modules); the test scaffold can be written before ictest is fixed but can't be run.
