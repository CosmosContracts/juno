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
	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(100000)))
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
	// Setup a high gas contract
	highGasFees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(1000000)))
	_, contractAddr = s.SetupContract(s.Chain, user.KeyName(), "../../contracts/juno_staking_hooks_high_gas_example.wasm", `{}`, false, highGasFees)

	// Register staking contract
	s.RegisterCwHooksStaking(s.Chain, user, contractAddr)
	sc = s.GetCwHooksStakingContracts()
	require.Equal(2, len(sc))

	// Perform a Staking Action
	stakeAmt = 1_000_000
	s.StakeTokens(s.Chain, user, valoper.String(), fmt.Sprintf("%d%s", stakeAmt, s.Chain.Config().Denom), fees, true)

	// Query the smart contract, should panic and not update value
	res = s.GetCwStakingHookLastDelegationChange(s.Chain, contractAddr, user.FormattedAddress())
	require.Equal("", res.Data.ValidatorAddress)
	require.Equal("", res.Data.DelegatorAddress)
	require.Equal("", res.Data.Shares)
}
