package feemarket_test

import (
	"context"
	"fmt"
	"sync"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/interchaintest/v10/ibc"

	feemarketypes "github.com/CosmosContracts/juno/v31/x/feemarket/types"
	streamtypes "github.com/CosmosContracts/juno/v31/x/stream/types"
)

// monitorGasPrice continuously monitors gas price changes
func (s *FeemarketTestSuite) monitorGasPrice(
	updates chan<- sdk.DecCoin,
	stop <-chan struct{},
	done chan<- struct{},
) {
	defer close(done)

	// establish a cancellable context so we can stop the stream cleanly
	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()

	// cancel the stream when stop is signaled
	go func() {
		<-stop
		cancel()
	}()

	// subscribe to the feemarket GasPrice stream for the suite denom
	stream, err := s.QueryClients.StreamClient.Stream(ctx, &streamtypes.StreamDynamicRequest{
		Module: "feemarket",
		Method: "GasPrice",
		Params: map[string]string{
			"denom": s.Denom,
		},
	})
	if err != nil {
		s.T().Logf("failed to open gas price stream: %v", err)
		return
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			// stream closed or context canceled
			return
		}

		if resp == nil || resp.Result == nil {
			continue
		}

		var out feemarketypes.GasPriceResponse
		if err := out.Unmarshal(resp.Result.Value); err != nil {
			s.T().Logf("failed to decode gas price stream response: %v", err)
			continue
		}

		select {
		case updates <- out.Price:
		case <-stop:
			return
		}
	}
}

// monitorFeemarketState streams the feemarket State and logs per-block details
// including current-slot utilization, average utilization, and learning rate.
func (s *FeemarketTestSuite) monitorFeemarketState(stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)

	// fetch params once to compute average utilization with correct MaxBlockUtilization
	params := s.QueryFeemarketParams()

	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()

	go func() {
		<-stop
		cancel()
	}()

	stream, err := s.QueryClients.StreamClient.Stream(ctx, &streamtypes.StreamDynamicRequest{
		Module: "feemarket",
		Method: "State",
		Params: map[string]string{},
	})
	if err != nil {
		s.T().Logf("failed to open feemarket state stream: %v", err)
		return
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			return
		}
		if resp == nil || resp.Result == nil {
			continue
		}

		var out feemarketypes.StateResponse
		if err := out.Unmarshal(resp.Result.Value); err != nil {
			s.T().Logf("failed to decode state stream response: %v", err)
			continue
		}

		st := out.State
		idx := int(st.Index)
		// window[idx] is the next slot (already zeroed in IncrementHeight),
		// so the last block's utilization sits at the previous index.
		prev := 0
		if l := len(st.Window); l > 0 {
			prev = (idx - 1 + l) % l
		}
		cur := uint64(0)
		if prev >= 0 && prev < len(st.Window) {
			cur = st.Window[prev]
		}

		avg := st.GetAverageUtilization(params)
		lr := st.LearningRate

		s.T().Logf("feemarket state: slot_util=%d avg_util=%s lr=%s", cur, avg.String(), lr.String())
	}
}

func (s *FeemarketTestSuite) createNetworkCongestion(users []ibc.Wallet) []error {
	var allErrors []error

	// Deploy the staking hooks high-gas contract once
	deployer := s.GetAndFundTestUser("deployer", 200000000000, s.Chain)
	setupFees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(1_000_000)))
	codeID := s.StoreContract(s.Chain, deployer.KeyName(), "../../contracts/juno_staking_hooks_high_gas_example.wasm", setupFees)

	numContracts := 20
	for i := range numContracts {
		contractAddr, err := s.InstantiateContract(s.Chain, deployer.KeyName(), codeID, `{}`, setupFees, true, false)
		if err != nil {
			return []error{fmt.Errorf("failed to instantiate hook contract %d: %w", i, err)}
		}
		s.RegisterCwHooksStaking(s.Chain, deployer, contractAddr)
	}
	s.T().Logf("Registered %d staking-hooks high-gas contracts", numContracts)

	// Choose a validator to delegate to
	vals := s.QueryValidators(s.Chain)
	if len(vals) == 0 {
		s.T().Logf("no validators found; aborting load")
		return []error{fmt.Errorf("no validators found")}
	}
	valoper := vals[0].String()

	rounds := 20
	s.T().Logf("Congesting network (staking-async) for %d rounds with %d users", rounds, len(users))

	// Prepare per-account next sequence
	nextSeq := make(map[string]uint64, len(users))
	for _, u := range users {
		addr := u.FormattedAddress()
		nextSeq[addr] = 0
	}

	for r := range rounds {
		currentGasPrice := s.QueryFeemarketGasPrice(s.Denom)
		gasBudget := int64(1_000_000)
		feeAmount := currentGasPrice.Amount.Mul(math.LegacyNewDec(gasBudget)).Mul(math.LegacyNewDec(3))
		fees := sdk.NewCoins(sdk.NewCoin(currentGasPrice.Denom, feeAmount.TruncateInt()))

		var wg sync.WaitGroup
		var roundErrors []error
		var errorsMu sync.Mutex
		var seqMu sync.Mutex

		for i := range users {
			sender := users[i]
			wg.Add(1)
			go func(idx int, from ibc.Wallet) {
				defer wg.Done()
				addr := from.FormattedAddress()
				seqMu.Lock()
				seq := nextSeq[addr]
				nextSeq[addr] = seq + 1
				seqMu.Unlock()
				memo := fmt.Sprintf("stake-u%03d-r%02d-s%06d", idx, r, seq)
				amount := sdk.NewInt64Coin(s.Denom, 100_000)
				_, err := s.StakeTokensAsync(from, valoper, amount, fees, gasBudget, memo, seq)
				if err != nil {
					errorsMu.Lock()
					roundErrors = append(roundErrors, fmt.Errorf("round %d user %s: %v", r, addr, err))
					errorsMu.Unlock()
					return
				}
			}(i, sender)
		}

		wg.Wait()

		if len(roundErrors) > 0 {
			s.T().Logf("round %d had %d tx errors", r, len(roundErrors))
			allErrors = append(allErrors, roundErrors...)
		}
	}

	// Settle and allow inclusion of pending txs. The congestion loop fires
	// 20 rounds × 20 users = 400 high-gas staking txs (~1M gas each); at the
	// 25M-gas/block ceiling the mempool needs ≥16 blocks just to drain, plus
	// a few more for feemarket gas-price decay so the next subtest's
	// faucet sends don't race with the backlog. h+6 was too tight — query
	// tx for the first faucet send in TestSendTxFailures was returning "tx
	// not found" because the tx sat in mempool past the 2-block ExecTx wait.
	h, _ := s.Chain.Height(s.Ctx)
	s.WaitForHeight(s.Chain, h+30)

	return allErrors
}
