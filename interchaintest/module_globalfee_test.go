package interchaintest

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
	globalfeetypes "github.com/CosmosContracts/juno/v28/x/globalfee/types"
)

// TestJunoGlobalFee
func TestJunoGlobalFee(t *testing.T) {
	t.Parallel()

	cfg := junoConfig
	cfg.GasPrices = "0.003ujuno" // this is used in the faucet cmd, must match initial globalfee
	cfg.GasAdjustment = 2.5

	// 0.002500000000000000
	coin := sdk.NewDecCoinFromDec(cfg.Denom, sdkmath.LegacyNewDecWithPrec(3, 3))
	cfg.ModifyGenesis = cosmos.ModifyGenesis(append(defaultGenesisKV, []cosmos.GenesisKV{
		{
			Key:   "app_state.globalfee.params.minimum_gas_prices",
			Value: sdk.DecCoins{coin},
		},
	}...))

	// Base setup
	chains := CreateChainWithCustomConfig(t, 1, 0, cfg)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// Chains
	juno := chains[0].(*cosmos.CosmosChain)

	nativeDenom := juno.Config().Denom

	// Users
	initFunds := sdkmath.NewInt(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", initFunds, juno, juno)
	sender := users[0]
	receiver := users[1].FormattedAddress()

	// fail: send 1 token to the receiver, no fee provided.
	std := bankSendWithFees(t, ctx, juno, sender, receiver, "1"+nativeDenom, "0"+nativeDenom, 200000)
	require.Contains(t, std, "no fees were specified")

	// fail: not enough fees
	std = bankSendWithFees(t, ctx, juno, sender, receiver, "1"+nativeDenom, "1"+nativeDenom, 200000)
	require.Contains(t, std, "insufficient fees")

	// fail: wrong fee token
	std = bankSendWithFees(t, ctx, juno, sender, receiver, "1"+nativeDenom, "1NOTATOKEN", 200000)
	require.Contains(t, std, "fee denom is not accepted")

	// success: send with enough fee (200k gas * 0.003 = 600)
	std = bankSendWithFees(t, ctx, juno, sender, receiver, "2"+nativeDenom, "600"+nativeDenom, 200000)
	require.Regexp(t, regexp.MustCompile(`raw_log:\s*""`), std)
	require.Regexp(t, regexp.MustCompile(`code:\s*0`), std)

	afterBal, err := juno.GetBalance(ctx, receiver, nativeDenom)
	require.NoError(t, err)
	require.Equal(t, initFunds.Add(sdkmath.NewInt(2)), afterBal)

	// param change proposal (lower fee), then validate it still works
	propID := submitGlobalFeeParamChangeProposal(t, ctx, juno, sender)
	propIDUint64, err := strconv.ParseUint(propID, 10, 64)
	require.NoError(t, err, "error converting propID to int64")
	helpers.ValidatorVote(t, ctx, juno, propIDUint64, 25)

	// success: validate the new value is in effect (200k gas * 0.005 = 200ujuno)
	std = bankSendWithFees(t, ctx, juno, sender, receiver, "3"+nativeDenom, "1000"+nativeDenom, 200000)
	require.Regexp(t, regexp.MustCompile(`raw_log:\s*""`), std)
	require.Regexp(t, regexp.MustCompile(`code:\s*0`), std)

	afterBal, err = juno.GetBalance(ctx, receiver, nativeDenom)
	require.NoError(t, err)
	require.Equal(t, initFunds.Add(sdkmath.NewInt(2)).Add(sdkmath.NewInt(3)), afterBal)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}

func bankSendWithFees(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, from ibc.Wallet, toAddr, coins, feeCoin string, gasAmt int64) string {
	cmd := []string{
		"junod", "tx", "bank", "send", from.KeyName(), toAddr, coins,
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--gas", fmt.Sprintf("%d", gasAmt),
		"--fees", feeCoin,
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	t.Log(string(stdout))

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}

	return string(stdout)
}

func submitGlobalFeeParamChangeProposal(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet) string {
	upgradeMsg := []cosmos.ProtoMessage{
		&globalfeetypes.MsgUpdateParams{
			Authority: "juno10d07y265gmmuvt4z0w9aw880jnsr700jvss730",
			Params: globalfeetypes.Params{
				MinimumGasPrices: sdk.DecCoins{
					// 0.005ujuno
					sdk.NewDecCoinFromDec(chain.Config().Denom, sdkmath.LegacyNewDecWithPrec(5, 3)),
				},
			},
		},
	}

	proposal, err := chain.BuildProposal(upgradeMsg, "New Global Fee", "Summary desc", "ipfs://CID", fmt.Sprintf(`500000000%s`, chain.Config().Denom), sdk.MustBech32ifyAddressBytes("juno", user.Address()), false)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	return txProp.ProposalID
}
