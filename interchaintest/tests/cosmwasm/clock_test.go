package cosmwasm_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	clocktypes "github.com/CosmosContracts/juno/v30/x/clock/types"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

type CosmWasmTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestCosmWasmTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{e2esuite.DefaultSpec},
		e2esuite.DefaultTxCfg,
	)

	t.Parallel()
	t.Cleanup(func() {
		_ = s.Ic.Close()
	})

	testSuite := &CosmWasmTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

// TestClockModule ensures the clock module auto executes allowed contracts.
func (s *CosmWasmTestSuite) TestClockModule() {
	require := s.Require()

	// Users
	user := s.GetAndFundTestUser("default", 10_000_000_000, s.Chain)
	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(100000)))
	// -- REGULAR GAS CONTRACT --
	// Ensure logic works as expected for a contract that uses less than the gas limit
	// and has a valid sudo message entry point.

	// Setup contract
	_, contractAddr := s.SetupContract(s.Chain, user.KeyName(), "../../contracts/clock_example.wasm", `{}`, false, fees)

	// Ensure config is 0
	res := s.GetClockContractValue(s.Chain, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.Equal(uint32(0), res.Data.Val)

	// Register the contract
	_, err := s.RegisterClockContract(s.Chain, user, contractAddr)
	require.NoError(err)

	// Validate contract is not jailed
	contract := s.GetClockContract(s.Chain, contractAddr)
	require.False(contract.ClockContract.IsJailed)

	// Validate the contract is now auto incrementing from the end blocker
	res = s.GetClockContractValue(s.Chain, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.GreaterOrEqual(res.Data.Val, uint32(1))

	// Unregister the contract & ensure it is removed from the store
	_, err = s.UnregisterClockContract(s.Chain, user, contractAddr)
	require.NoError(err)
	s.ValidateNoClockContract(s.Chain, contractAddr)

	// -- HIGH GAS CONTRACT --
	// Ensure contracts that exceed the gas limit are jailed.

	// Setup contract
	_, contractAddr = s.SetupContract(s.Chain, user.KeyName(), "../../contracts/clock_example_high_gas.wasm", `{}`, false, fees, "--admin", user.FormattedAddress())

	// Ensure config is 0
	res = s.GetClockContractValue(s.Chain, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.Equal(uint32(0), res.Data.Val)

	// Register the contract
	_, err = s.RegisterClockContract(s.Chain, user, contractAddr)
	require.NoError(err)

	// Validate contract is jailed
	contract = s.GetClockContract(s.Chain, contractAddr)
	require.True(contract.ClockContract.IsJailed)

	// -- MIGRATE CONTRACT --
	// Ensure migrations can patch contracts that error or exceed gas limit
	// so they can be unjailed.

	// Migrate the high gas contract to a contract with lower gas usage
	s.MigrateContract(s.Chain, user.KeyName(), contractAddr, "../../contracts/clock_example_migrate.wasm", `{}`, fees)

	// Unjail the contract
	_, err = s.UnjailClockContract(s.Chain, user, contractAddr)
	require.NoError(err)

	// Validate contract is not jailed
	contract = s.GetClockContract(s.Chain, contractAddr)
	require.False(contract.ClockContract.IsJailed)

	s.QueryClients.ClockClient.ClockContract(
		s.Ctx,
		&clocktypes.QueryClockContractRequest{
			ContractAddress: contractAddr,
		},
	)

	// Validate the contract is now auto incrementing from the end blocker
	res = s.GetClockContractValue(s.Chain, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.GreaterOrEqual(res.Data.Val, uint32(1))

	// -- NO SUDO CONTRACT --
	// Ensure contracts that do not have a sudo message entry point are jailed.

	// Setup contract
	_, contractAddr = s.SetupContract(s.Chain, user.KeyName(), "../../contracts/clock_example_no_sudo.wasm", `{}`, false, fees)

	// Ensure config is 0
	res = s.GetClockContractValue(s.Chain, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.Equal(uint32(0), res.Data.Val)

	// Register the contract
	_, err = s.RegisterClockContract(s.Chain, user, contractAddr)
	require.NoError(err)

	// Validate contract is jailed
	contract = s.GetClockContract(s.Chain, contractAddr)
	require.True(contract.ClockContract.IsJailed)

	// Validate contract is not auto incrementing
	res = s.GetClockContractValue(s.Chain, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.Equal(uint32(0), res.Data.Val)
}

func SubmitParamChangeProp(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, gasLimit uint64) string {
	govAcc := "juno10d07y265gmmuvt4z0w9aw880jnsr700jvss730"
	updateParams := []cosmos.ProtoMessage{
		&clocktypes.MsgUpdateParams{
			Authority: govAcc,
			Params: clocktypes.Params{
				ContractGasLimit: gasLimit,
			},
		},
	}

	proposal, err := chain.BuildProposal(updateParams, "Params Update Gas Limit", "params", "ipfs://CID", fmt.Sprintf(`500000000%s`, chain.Config().Denom), sdk.MustBech32ifyAddressBytes("s.Chain", user.Address()), false)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	height, _ := chain.Height(ctx)

	proposalID, err := strconv.ParseUint(txProp.ProposalID, 10, 64)
	require.NoError(t, err, "failed to parse proposal ID")

	err = chain.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, chain, height, height+e2esuite.DefaultHaltHeightDelta, proposalID, govtypes.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	return txProp.ProposalID
}
