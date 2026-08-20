# Voting-snapshot EndBlock benchmark

This benchmark is the v31 release gate for the rare validator-wide paths used by slashing and bonded-status transitions. It uses deterministic addresses and a mainnet-shaped validator with 1,000, 6,000, and 10,001 delegators. Every operation includes both the staking hook’s validator walk and voting-snapshot’s same-block EndBlock drain.

## Run

From the repository root:

```sh
go test ./x/voting-snapshot/keeper \
  -run '^TestValidatorWalkOverAdvisoryLimitPreservesSameBlockPower$' \
  -count=1 -v

go test ./x/voting-snapshot/keeper \
  -run '^$' \
  -bench '^BenchmarkValidatorDelegatorWalk$' \
  -benchmem -benchtime=3x -count=3
```

The correctness test deliberately exceeds `MaxDelegatorsPerSnapshotWalk`. It asserts that no delegator is truncated, post-slash power is recorded at the current height, total power is updated at that height, and the validator walk, EndBlock delegation walk, and validator lookups each process exactly 10,001 rows.

## v31 RC threshold

On an otherwise idle amd64 release worker, the 10,001-delegator slash and status cases must each satisfy:

- less than **500 ms/op** in all three runs;
- less than **64 MiB/op**;
- exactly **10,001 validator iterations/op** and **10,001 EndBlock iterations/op**;
- the same-block correctness test passes.

A single noisy result must be rerun once on an idle worker. Two consecutive controlled runs above either time or allocation threshold block the RC and require a bounded, regression-tested optimization. Results below the threshold do not justify consensus-path optimization.

## Baseline evidence

Measured from commit `68d1cbd3c6a490bdbf9848ce4d55b8a10c82cda5` plus this benchmark on Linux/amd64, Intel i7-10710U, Go 1.25.10:

| Case | 10,001-delegator observed range | Allocations |
|---|---:|---:|
| Slash hook + EndBlock | 121–256 ms/op | about 47.5 MB/op |
| Status hook + EndBlock | 114–158 ms/op | about 47.5 MB/op |

All observed runs remained below the v31 RC threshold, so no production optimization was introduced.
