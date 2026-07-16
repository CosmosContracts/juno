package cosmwasm_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// from x/cw-hooks/keeper/msg_server_test.go -> TestContractExecution
func (s *CosmWasmTestSuite) TestCwHooks() {
	require := s.Require()

	// Users
	user := s.GetAndFundTestUser("default", 10_000_000_000, s.Chain)

	// Upload & init contract payment to another address
	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(1_000_000)))
	_, contractAddr := s.SetupContract(s.Chain, user.KeyName(), "../../contracts/juno_staking_hooks_example.wasm", `{}`, false, fees)

	// register staking contract (to be tested)
	s.RegisterCwHooksStaking(s.Chain, user, contractAddr)
	sc := s.GetCwHooksStakingContracts()
	require.Equal(1, len(sc))
	require.Equal(contractAddr, sc[0])

	// Validate that governance contract is added
	s.RegisterCwHooksGovernance(s.Chain, user, contractAddr)
	gc := s.GetCwHooksGovernanceContracts()
	require.Equal(1, len(gc))
	require.Equal(contractAddr, gc[0])

	// Perform a Staking Action
	vals := s.QueryValidators(s.Chain)
	valoper := vals[0]

	stakeAmt := 1_000_000
	s.StakeTokens(s.Chain, user, valoper.String(), fmt.Sprintf("%d%s", stakeAmt, s.Chain.Config().Denom), fees, false)

	// Query the smart contract to validate it saw the fire-and-forget update
	res := s.GetCwStakingHookLastDelegationChange(s.Chain, contractAddr, user.FormattedAddress())
	resValAddress := sdk.MustValAddressFromBech32(res.Data.ValidatorAddress)
	require.Equal(valoper, resValAddress)
	require.Equal(user.FormattedAddress(), res.Data.DelegatorAddress)
	require.Equal(fmt.Sprintf("%d.000000000000000000", stakeAmt), res.Data.Shares)

	// HIGH GAS TEST
	//
	// A registered hook contract that runs out of gas must NOT be able to
	// take down the staking tx that triggered it: cw-hooks executes each hook
	// under an isolated child gas meter and recovers OOG panics. This section
	// proves the isolation end-to-end — the delegation succeeds and its state
	// changes, while the misbehaving contract is left un-updated and has its
	// failure counter incremented (heading toward eventual removal).

	// Setup a high gas contract
	highGasFees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(1000000)))
	_, highGasContract := s.SetupContract(s.Chain, user.KeyName(), "../../contracts/juno_staking_hooks_high_gas_example.wasm", `{}`, false, highGasFees)

	// Register staking contract
	s.RegisterCwHooksStaking(s.Chain, user, highGasContract)
	sc = s.GetCwHooksStakingContracts()
	require.Equal(2, len(sc))

	// Snapshot the on-chain delegation before staking again so we can prove the
	// staking state actually changed (and not rely solely on the tx succeeding).
	delBefore := s.QueryStakingDelegation(user.FormattedAddress(), valoper.String())

	// Perform a Staking Action. skipTxCheck=false makes ExecTx assert the tx
	// landed with Code==0 — i.e. staking succeeds despite the OOG hook.
	stakeAmt = 1_000_000
	s.StakeTokens(s.Chain, user, valoper.String(), fmt.Sprintf("%d%s", stakeAmt, s.Chain.Config().Denom), fees, false)

	// Staking state must have advanced by exactly the delegated amount.
	delAfter := s.QueryStakingDelegation(user.FormattedAddress(), valoper.String())
	require.Equal(
		delBefore.Balance.Amount.AddRaw(int64(stakeAmt)).String(),
		delAfter.Balance.Amount.String(),
		"delegation balance should increase by the staked amount",
	)

	// The original (well-behaved) contract's last delegation change should now
	// reflect this second delegation — proving healthy hooks still fire.
	goodRes := s.GetCwStakingHookLastDelegationChange(s.Chain, contractAddr, user.FormattedAddress())
	require.Equal(valoper, sdk.MustValAddressFromBech32(goodRes.Data.ValidatorAddress))
	require.Equal(user.FormattedAddress(), goodRes.Data.DelegatorAddress)

	// The high-gas contract should panic and never update its own value.
	res = s.GetCwStakingHookLastDelegationChange(s.Chain, highGasContract, user.FormattedAddress())
	require.Equal("", res.Data.ValidatorAddress)
	require.Equal("", res.Data.DelegatorAddress)
	require.Equal("", res.Data.Shares)

	// ...and its failure counter must have been incremented, recording the
	// isolated OOG failure. It stays registered (default removal threshold 3).
	info := s.QueryCwHooksContractInfo("staking", highGasContract)
	require.Equal(highGasContract, info.ContractAddress)
	require.Positive(info.FailureCounter, "OOG hook failure should bump the contract's failure counter")
	require.NotEmpty(info.LatestError, "OOG hook failure should record the latest error")
}
