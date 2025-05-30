package drip_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

var mnemonic = "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"

type DripTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestDripTestSuite(t *testing.T) {
	// Setup new pre determined user (from test_node.sh)

	addr := "juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl"
	newCfg := e2esuite.DefaultConfig
	newCfg.ModifyGenesis = cosmos.ModifyGenesis(append(e2esuite.DefaultGenesisKV, []cosmos.GenesisKV{
		{
			Key:   "app_state.drip.params.allowed_addresses",
			Value: []string{addr},
		},
	}...))

	spec := &interchaintest.ChainSpec{
		ChainName:     "juno-drip",
		Name:          "juno-drip",
		NumValidators: &e2esuite.DefaultNumValidators,
		NumFullNodes:  &e2esuite.DefaultNumFullNodes,
		Version:       "local",
		NoHostMount:   &e2esuite.DefaultNoHostMount,
		ChainConfig:   newCfg,
	}

	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{spec},
		e2esuite.DefaultTxCfg,
	)

	t.Parallel()
	t.Cleanup(func() {
		_ = s.Ic.Close()
	})

	testSuite := &DripTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

// TestDripModule ensures the x/drip module properly distributes tokens from whitelisted accounts.
func (s *DripTestSuite) TestDripMmodule() {
	t := s.T()
	fees := sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, sdkmath.NewInt(50_000)))

	nativeDenom := s.Chain.Config().Denom
	user, err := s.GetAndFundTestUserWithMnemonic("default", mnemonic, 1_000_000_000_000, s.Chain)
	if err != nil {
		t.Fatal(err)
	}

	// New TF token to distributes
	tfDenom := s.CreateTokenFactoryDenom(s.Chain, user, "dripme", fees)
	distributeAmt := sdkmath.NewInt(1_000_000)
	s.MintTokenFactoryDenom(s.Chain, user, distributeAmt.Uint64(), tfDenom, fees)

	// Stake some tokens
	vals, err := s.QueryClients.StakingClient.Validators(s.Ctx, &types.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
		Pagination: &query.PageRequest{
			Limit: 1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	valoper := vals.Validators[0].OperatorAddress

	stakeAmt := int64(100_000_000_000)
	s.StakeTokens(s.Chain, user, valoper, fmt.Sprintf("%d%s", stakeAmt, nativeDenom), fees, false)

	// Drip the TF Tokens to all stakers
	distribute := int64(1_000_000)
	s.DripTokens(s.Chain, user, fmt.Sprintf("%d%s", distribute, tfDenom), fees)

	// Claim staking rewards to capture the drip
	s.ClaimStakingRewards(s.Chain, user, valoper, fees)

	// Check balances has the TF Denom from the claim
	bals, _ := s.Chain.BankQueryAllBalances(s.Ctx, user.FormattedAddress())
	fmt.Println("balances", bals)

	found := false
	for _, bal := range bals {
		if bal.Denom == tfDenom {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("did not find drip token")
	}
}
