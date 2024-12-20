package interchaintest

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	clocktypes "github.com/CosmosContracts/juno/v26/x/clock/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	cosmosproto "github.com/cosmos/gogoproto/proto"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/stretchr/testify/require"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

// TestJunoClock ensures the clock module auto executes allowed contracts.
func TestJunoClock(t *testing.T) {
	t.Parallel()

	cfg := junoConfig

	// Base setup
	chains := CreateChainWithCustomConfig(t, 1, 0, cfg)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	juno := chains[0].(*cosmos.CosmosChain)

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(10_000_000_000), juno, juno)
	user := users[0]

	//
	// -- REGULAR GAS CONTRACT --
	// Ensure logic works as expected for a contract that uses less than the gas limit
	// and has a valid sudo message entry point.
	//

	// Setup contract
	_, contractAddr := helpers.SetupContract(t, ctx, juno, user.KeyName(), "contracts/clock_example.wasm", `{}`)

	// Ensure config is 0
	res := helpers.GetClockContractValue(t, ctx, juno, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.Equal(t, uint32(0), res.Data.Val)

	// Register the contract
	helpers.RegisterClockContract(t, ctx, juno, user, contractAddr)

	// Validate contract is not jailed
	contract := helpers.GetClockContract(t, ctx, juno, contractAddr)
	require.False(t, contract.ClockContract.IsJailed)

	// Validate the contract is now auto incrementing from the end blocker
	res = helpers.GetClockContractValue(t, ctx, juno, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.GreaterOrEqual(t, res.Data.Val, uint32(1))

	// Unregister the contract & ensure it is removed from the store
	helpers.UnregisterClockContract(t, ctx, juno, user, contractAddr)
	helpers.ValidateNoClockContract(t, ctx, juno, contractAddr)

	//
	// -- HIGH GAS CONTRACT --
	// Ensure contracts that exceed the gas limit are jailed.
	//

	// Setup contract
	_, contractAddr = helpers.SetupContract(t, ctx, juno, user.KeyName(), "contracts/clock_example_high_gas.wasm", `{}`, "--admin", user.FormattedAddress())

	// Ensure config is 0
	res = helpers.GetClockContractValue(t, ctx, juno, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.Equal(t, uint32(0), res.Data.Val)

	// Register the contract
	helpers.RegisterClockContract(t, ctx, juno, user, contractAddr)

	// Validate contract is jailed
	contract = helpers.GetClockContract(t, ctx, juno, contractAddr)
	require.True(t, contract.ClockContract.IsJailed)

	//
	// -- MIGRATE CONTRACT --
	// Ensure migrations can patch contracts that error or exceed gas limit
	// so they can be unjailed.
	//

	// Migrate the high gas contract to a contract with lower gas usage
	helpers.MigrateContract(t, ctx, juno, user.KeyName(), contractAddr, "contracts/clock_example_migrate.wasm", `{}`)

	// Unjail the contract
	helpers.UnjailClockContract(t, ctx, juno, user, contractAddr)

	// Validate contract is not jailed
	contract = helpers.GetClockContract(t, ctx, juno, contractAddr)
	require.False(t, contract.ClockContract.IsJailed)

	// Validate the contract is now auto incrementing from the end blocker
	res = helpers.GetClockContractValue(t, ctx, juno, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.GreaterOrEqual(t, res.Data.Val, uint32(1))

	//
	// -- NO SUDO CONTRACT --
	// Ensure contracts that do not have a sudo message entry point are jailed.
	//

	// Setup contract
	_, contractAddr = helpers.SetupContract(t, ctx, juno, user.KeyName(), "contracts/clock_example_no_sudo.wasm", `{}`)

	// Ensure config is 0
	res = helpers.GetClockContractValue(t, ctx, juno, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.Equal(t, uint32(0), res.Data.Val)

	// Register the contract
	helpers.RegisterClockContract(t, ctx, juno, user, contractAddr)

	// Validate contract is jailed
	contract = helpers.GetClockContract(t, ctx, juno, contractAddr)
	require.True(t, contract.ClockContract.IsJailed)

	// Validate contract is not auto incrementing
	res = helpers.GetClockContractValue(t, ctx, juno, contractAddr)
	fmt.Printf("- res: %v\n", res.Data.Val)
	require.Equal(t, uint32(0), res.Data.Val)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}

func SubmitParamChangeProp(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, gasLimit uint64) string {
	govAcc := "juno10d07y265gmmuvt4z0w9aw880jnsr700jvss730"
	updateParams := []cosmosproto.Message{
		&clocktypes.MsgUpdateParams{
			Authority: govAcc,
			Params: clocktypes.Params{
				ContractGasLimit: gasLimit,
			},
		},
	}

	proposal, err := chain.BuildProposal(updateParams, "Params Update Gas Limit", "params", "ipfs://CID", fmt.Sprintf(`500000000%s`, chain.Config().Denom))
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	height, _ := chain.Height(ctx)

	proposalID, err := strconv.ParseInt(txProp.ProposalID, 10, 64)
	require.NoError(t, err, "failed to parse proposal ID")

	err = chain.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, chain, height, height+haltHeightDelta, proposalID, govtypes.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	return txProp.ProposalID
}
