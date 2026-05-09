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

## Diagnosis run on 2026-05-09 — root cause identified

Ran the diagnostic checklist against a live container. Findings:

1. **Chain container starts fine.** `juno-2-val-0-TestBasicTestSuite`
   on docker network `interchaintest-jpsprqaw`, IP `172.217.0.2`.
2. **RPC IS listening inside the container.** `netstat -tln` shows
   `:::26657` in LISTEN state alongside `:::1317`, `:::9090`, `:::26656`.
3. **Docker port mapping IS established.** `docker port` shows
   `26657/tcp -> 0.0.0.0:41517` (and similar for the other ports).
4. **Container IP is unreachable from SafeClaude.** Direct
   `curl 172.217.0.2:26657` from inside SafeClaude times out.
5. **Host-mapped port is unreachable from SafeClaude.** Direct
   `curl 0.0.0.0:41517` (or `localhost:41517`) returns
   `Connection refused` — the host port lives on the OUTER host's
   network namespace, not SafeClaude's.

**Root cause: docker network namespace isolation.**

- The juno-1 chain container is on docker network `interchaintest-jpsprqaw` (172.217.0.0/16)
- The SafeClaude container is on docker network `bridge` (172.17.0.0/16)
- Different networks, isolated by default
- Host-port mappings (`0.0.0.0:41517`) bind on the OUTER physical host's network stack, which SafeClaude can't see — SafeClaude's `localhost` is its own loopback, not the outer host's

interchaintest's chain readiness check polls `tcp://0.0.0.0:<host-port>`
via `tn.hostRPCPort` (set from `containerLifecycle.GetHostPorts` —
see chain_node.go:1240). This is correct for a test runner running on
the same host as the chain containers, but breaks for docker-in-docker
setups where the test runner is itself in a container.

## Fix path

The interchaintest framework was designed for test runners that share
network namespace with the spawned chain containers. SafeClaude doesn't.

**Two-part fix:**

1. **Fork interchaintest** with a one-line patch to use the chain
   container's *internal* hostname (`tn.HostName():26657`) instead of
   the host-mapped port (`tn.hostRPCPort`) when an env var like
   `INTERCHAINTEST_USE_CONTAINER_HOSTNAME=1` is set. Add a `replace`
   directive in `interchaintest/go.mod` pointing to the fork:
   ```
   replace github.com/cosmos/interchaintest/v10 => github.com/juno-ai-dev/interchaintest/v10 v10.0.0-...
   ```
2. **Connect SafeClaude to the interchaintest network at test setup**
   via a hook in `interchaintest/suite/lib.go`. The interchaintest
   network name is dynamic per run; the hook needs to:
   - Wait for the chain container to come up
   - Identify its docker network via `docker inspect <container>`
   - Run `docker network connect <network> <safeclaude-container>` from
     inside SafeClaude (it has docker.sock mounted)
   - At test teardown, run the equivalent `disconnect`
   With those two pieces, SafeClaude can resolve the chain's hostname
   via Docker DNS on the shared network, and the patched
   chain_node.go polls successfully.

**Alternative fix (simpler, less clean):**

Add a TCP forwarder (`socat` or a tiny Go proxy) inside SafeClaude
that mirrors the dynamic host-port mapping locally:

```bash
# pseudo: for each host-port mapping, forward localhost:<port> → container-ip:<port>
docker port <chain> | while read line; do
    container_port=$(echo $line | cut -d/ -f1)
    host_port=$(echo $line | cut -d: -f2)
    socat TCP-LISTEN:$host_port,bind=0.0.0.0,fork TCP:$container_ip:$container_port &
done
```

This works without forking interchaintest, but is per-run setup.

## Workaround for v30 timeline

If neither fix is in place by the time we want to ship v30:
- **Run ictest from outside SafeClaude** — on the actual host machine
  where the docker daemon lives. CI does this naturally.
- **Skip ictest from PR review surface** — keep workflow matrix in
  place; mark v30 PR as ictest-pending and validate via a v30-rc1
  testnet smoke test instead.
- **Manual smoke** — start junod from a script outside SafeClaude,
  deploy a contract, exercise it, capture logs.

## Confirmed not the issue

- Docker port forward isn't set up: ✗ (mapping exists)
- RPC server isn't binding inside container: ✗ (listening on :::26657)
- cometbft startup race: ✗ (chain commits blocks at height 49+)
- 2-min timeout too short: ✗ (5min equally fails, at ~144s, not at deadline)
- libwasmvm static linking: ✗ (Dockerfile fixed, binary is static)
- Probe-shape mismatch: ✗ (we now know the dial fails at TCP level, not HTTP)

## Workaround under consideration

If the focused session can't quickly resolve, we can:
- Skip ictest from the v30 PR; leave the workflow matrix in place but mark suites `xfail` for v30
- Run a manual smoke (start junod from a script, deploy a contract, call it) on a v30-rc1 testnet
- Block the v30 proposal on getting ictest green in a follow-up PR

This blocks A1 (DAO DAO contract compatibility) and A2 (`ictest-upgrade` against forked mainnet state). The DAO DAO artifacts are ready in `/workspace/dao-contracts/artifacts/` (29 wasm files including all v2.7.0 modules); the test scaffold can be written before ictest is fixed but can't be run.
