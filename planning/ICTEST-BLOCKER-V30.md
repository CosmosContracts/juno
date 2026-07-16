# ictest blocker — chain RPC not reachable from host

**Status (updated 2026-07-06): partially unblocked.** `ictest-upgrade`
now runs end-to-end via the `ICTEST_HOST_IP` workaround; the v30 handler
is verified against a real v29→v30 migration. Only a flaky DooD
port-publish race remains (see the 2026-07-06 UPDATE at the bottom).
Original write-up preserved below for history.

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
- cometbft startup race: ✗ (chain commits blocks at height 134+)
- 2-min timeout too short: ✗ (5min equally fails)
- libwasmvm static linking: ✗ (Dockerfile fixed, binary is static)
- Probe-shape mismatch: ✗ (curl from inside container returns full /status JSON instantly)

## What I tried (2026-05-09 session)

1. **Bumped Build timeout 2min → 5min** (committed). No change in
   behavior — failure is structural, not slow-start.
2. **Diagnostic checklist** ran while a chain was live:
   - chain process healthy, height=49+, RPC log "Starting RPC HTTP
     server on [::]:26657" present
   - `:::26657` LISTEN inside container; `0.0.0.0:host-port` mapping
     bound on the OUTER host
   - direct `curl 0.0.0.0:host-port` and `curl <container-ip>:26657`
     both fail from inside SafeClaude — different network namespaces
3. **Built a Python TCP forwarder + watcher** (`/tmp/forward.py` +
   `/tmp/ictest-bridge.sh`) to bind each host-port locally on
   SafeClaude and forward to the chain's container IP. Result:
   the dial advanced from TCP-level "connection refused" to
   HTTP-level "context deadline exceeded" / "connection reset by
   peer" — meaning the bridge was up but the upstream connection
   from SafeClaude → 172.x.x.x couldn't actually deliver data.
4. **Ran `docker network connect` from SafeClaude to attach to the
   per-test interchaintest network.** Docker daemon accepts the call
   (`docker inspect` shows two networks attached) but **the new
   interface doesn't materialize inside SafeClaude's namespace**
   — `/proc/net/dev` shows only `eth0`, `/proc/net/route` has only
   the default-bridge routes. Direct dial of the chain's container
   IP times out.

## Final diagnosis

**SafeClaude's network namespace is sandboxed in a way that
prevents joining additional docker networks at runtime.** The
interface listed by `docker inspect` is registered with the daemon
but never wired into the container's namespace by runc/containerd.
This is an environment limitation, not a v30 bug.

ictest as currently architected requires the test runner and the
chain containers to share a network reachability path. That path
doesn't exist from inside SafeClaude, period.

## Conclusion for the v30 PR

ictest cannot be validated locally from inside SafeClaude. The
remaining options:

1. **Run ictest from the outer host** (CI runners do this naturally;
   the GitHub Actions workflow we already reconciled in Phase 0c is
   the right place for this)
2. **Run ictest from a different container architecture** that has
   privileged network access
3. **Validate v30 via testnet smoke** instead of local ictest before
   the security review and mainnet halt-height proposal

For our purposes (Bucket A items A1 and A2) the practical path is
**(3): testnet smoke**. Once a v30 binary is deployed to a public
testnet (uni-7 successor or a fresh v30-rc1 testnet), Jake/I can
exercise:

- DAO DAO v2.7.0 contract instantiate + propose + vote + execute
  (uses real DAO DAO codebase deployed at known IDs; we have
  the artifacts at `/workspace/dao-contracts/artifacts/`)
- the v30 upgrade handler against a fork of juno-1 mainnet state
  (manually rather than via `ictest-upgrade`)
- wasmbinding `VotingPowerAt` smoke against a small test contract
- IBC transfer from a counterparty testnet for cross-chain check

That validation is qualitatively the same as `ictest-*` but
runs in a real network namespace where the tests actually work.
A1 and A2 then close on testnet smoke evidence rather than CI
green.

## Tooling left in place for next host with proper network access

- `interchaintest/suite/lib.go` keeps the 5-minute Build timeout
  and the `t.Logf(Build error)` debug aid — both useful for any
  future run regardless of where it runs from.
- `/tmp/forward.py` and `/tmp/ictest-bridge.sh` are still in
  place as a reference for the docker-in-docker bridge pattern.
  They work for TCP but not for the dual-namespace upstream-dial
  step under SafeClaude's sandboxing — if a future runtime
  removes that sandbox, they're a one-shell-script away from
  unblocking local ictest.

## Workaround under consideration

If the focused session can't quickly resolve, we can:
- Skip ictest from the v30 PR; leave the workflow matrix in place but mark suites `xfail` for v30
- Run a manual smoke (start junod from a script, deploy a contract, call it) on a v30-rc1 testnet
- Block the v30 proposal on getting ictest green in a follow-up PR

This blocks A1 (DAO DAO contract compatibility) and A2 (`ictest-upgrade` against forked mainnet state). The DAO DAO artifacts are ready in `/workspace/dao-contracts/artifacts/` (29 wasm files including all v2.7.0 modules); the test scaffold can be written before ictest is fixed but can't be run.

## UPDATE 2026-07-06 — partial workaround found; `ictest-upgrade` now runs end-to-end

The May-9 diagnosis was too pessimistic. We do NOT need to join the
per-test interchaintest network. Published host ports bind on the
OUTER host's `0.0.0.0`, and the docker-bridge gateway `172.17.0.1`
(reachable from this container) routes into that namespace. So
pointing interchaintest's host-port resolution at the gateway IP
instead of the literal `0.0.0.0` makes the chain reachable.

**Workaround (local only — not committed):**

1. A one-line patch to `dockerutil.GetHostPort` in a local fork of
   `interchaintest/v10`, gated on `ICTEST_HOST_IP`: when the published
   `HostIP` is `0.0.0.0`/empty, substitute the env value. Wired via a
   git-excluded `interchaintest/go.work` `replace` (fork lives in the
   session scratchpad). Run with `ICTEST_HOST_IP=172.17.0.1`.
2. Verified reachability directly: `curl 172.17.0.1:<published-port>`
   into a throwaway container returns immediately (vs. the May-9
   `0.0.0.0`/container-IP timeouts).

With that, `make ictest-upgrade` (v29.0.0 → v30 `:local`):

- **Builds and starts the pre-upgrade chain** (health check passes).
- **Runs every pre-upgrade step** — deploys the cw-hooks staking
  contract, registers it, delegates, asserts hook state.
- **Submits + votes the software-upgrade proposal, halts, recreates
  every node container on the v30 image, and resumes block
  production.** Confirmed on the live v30 node via `docker exec`:
  height advanced past the upgrade height, `catching_up=false`, image
  `ghcr.io/cosmoscontracts/juno:local`. Handler logs show
  `applying upgrade "v30"`, `v30: successfully configured x/feemarket`,
  `v30: successfully set x/cw-hooks params`,
  `v30: successfully backfilled x/voting-snapshot from staking state`.
  On-chain queries confirm the migrated state:
  `feemarket max_block_utilization=30000000` (read from consensus
  `block.max_gas`, not the 25M fallback), `enabled=true`,
  `min_base_gas_price=0.075`; `cw-hooks contract_failure_removal_threshold=3`.

**Two real test-harness bugs fixed while getting there (committed to the branch):**

- `interchaintest/suite/{suite,upgrade}.go`: the suite's `s.GrpcClient`
  was dialed once in `SetupSuite` and never refreshed. `UpgradeVersion`
  recreates every node container → new host gRPC port → every
  post-upgrade gRPC query hit a dead port (`connection refused`). Added
  `RefreshGRPCClients()` and call it at the end of `UpgradeNodes`.
- Strengthened the upgrade assertions so a silent feemarket fallback
  can't pass: consensus `block.max_gas` is seeded to 30M (≠ the
  handler's 25M fallback) and the test asserts the post-upgrade
  `MaxBlockUtilization == 30M`.

**CORRECTION (2026-07-07): this was NOT DooD-specific — it reproduces on
CI (GitHub Actions, host netns) and is now fixed.** The earlier note
blamed a docker-in-docker port-reservation race; that was wrong. CI runs
`ictest-upgrade` on the daemon's own netns and still failed 2/2 with the
same `Post "http:": http: no Host in request URL` at `StartAllNodes`.

Root cause (confirmed by instrumenting a live run): interchaintest's
`ChainNode.StartContainer` reads the container's mapped host port with a
single `ContainerInspect` immediately after `docker start`, and caches
`""` if the daemon hasn't finished *publishing* the port yet — it never
re-reads. On a warm post-upgrade recreation the publish lag routinely
exceeds interchaintest's ~1s buffer, so the RPC client is built as
`tcp://` and every subsequent call fails. A port-watch during the failure
proved the containers are fine: **all three nodes reach `running exit=0`
with `26657/tcp` published and stable** — the framework just read the
mapping too early. (This also confirms, independently, that the v30
binary starts cleanly post-upgrade.)

Retrying the whole recreation does not help — under load the lag
consistently loses the read — so the fix does not depend on winning the
race:

- `interchaintest/suite/upgrade.go` — after `StartAllNodes` (let it run
  to completion so containers fully settle), **unconditionally rebuild
  every node's Tendermint client** from the live mapping via
  `ChainNode.GetHostAddress` (which re-inspects the running container) +
  `NewClient`. Idempotent: a no-op refresh on the happy path, a repair on
  the racy path.
- `interchaintest/suite/suite.go` — `RefreshGRPCClients` now reads the
  gRPC port the same way (`GetFullNode().GetHostAddress("9090/tcp")`,
  polled) instead of the value interchaintest cached at container start,
  which can likewise be `""` after a recreation.

Validated locally: with the rebuild running, the suite drives the whole
post-upgrade flow — `WaitForBlocks`, `Height`, and the feemarket +
cw-hooks gRPC assertions all pass against the rebuilt clients. The only
remaining local failure is `unknown service juno.votingsnapshot.v1.Query`,
an artifact of the stale local `:local` image (buildx is unavailable in
this container, so it can't be rebuilt from current source); CI builds the
image from source and has that service.

**Bottom line for v30:** the upgrade *handler* is verified against a real
v29→v30 store migration (halt/resume + on-chain migrated state), and the
ictest harness race that blocked CI is fixed in-repo. CI is the right
place to confirm green end-to-end, since it builds a current-source image
and reliably exercises the recreation race the fix targets.
