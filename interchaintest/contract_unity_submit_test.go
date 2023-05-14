package interchaintest

import (
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/stretchr/testify/require"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

// TestJunoSubmitUnityContract test to ensure the store code properly works on the contract
// - https://github.com/CosmosContracts/cw-unity-prop
func TestJunoUnityContractGovSubmit(t *testing.T) {
	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	juno := chains[0].(*cosmos.CosmosChain)

	nativeDenom := juno.Config().Denom

	// Users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", int64(10000_000000), juno, juno)
	user := users[0]
	withdrawUser := users[1]
	withdrawAddr := withdrawUser.Bech32Address(juno.Config().Bech32Prefix)

	// Upload & init unity contract with no admin in test mode
	msg := fmt.Sprintf(`{"native_denom":"%s","withdraw_address":"%s","withdraw_delay_in_days":28}`, nativeDenom, withdrawAddr)
	_, contractAddr := helpers.SetupContract(t, ctx, juno, user.KeyName, "contracts/cw_unity_prop.wasm", msg)
	t.Log("testing Unity contractAddr", contractAddr)

	// send 2JUNO funds to the contract from user
	juno.SendFunds(ctx, user.KeyName, ibc.WalletAmount{Address: contractAddr, Denom: nativeDenom, Amount: 2000000})

	height, err := juno.Height(ctx)
	require.NoError(t, err, "error fetching height")

	msg = fmt.Sprintf(`{"execute_send":{"amount":"1000000","recipient":"%s"}}`, withdrawAddr)
	helpers.StoreContractGovernanceProposal(t, ctx, juno, user, "Prop Title", "description", fmt.Sprintf(`500000000%s`, nativeDenom), contractAddr, "", msg)
	proposalID := "1"

	// poll for proposal
	err = juno.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, juno, height, height+haltHeightDelta, proposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
