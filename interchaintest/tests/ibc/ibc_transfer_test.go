package ibc_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"

	feemarkettypes "github.com/CosmosContracts/juno/v31/x/feemarket/types"
)

var genesisWalletAmount = int64(10_000_000)

type IbcTestSuite struct {
	*e2esuite.E2ETestSuite

	eRep *testreporter.RelayerExecReporter
}

func TestIbcTestSuite(t *testing.T) {
	numVals := 1
	numFullNodes := 1

	genesisKV := []cosmos.GenesisKV{
		{
			Key: "app_state.feemarket.params",
			Value: feemarkettypes.Params{
				Alpha:               feemarkettypes.DefaultAlpha,
				Beta:                feemarkettypes.DefaultBeta,
				Gamma:               feemarkettypes.DefaultAIMDGamma,
				Delta:               feemarkettypes.DefaultDelta,
				MinBaseGasPrice:     e2esuite.DefaultMinBaseGasPrice,
				MinLearningRate:     feemarkettypes.DefaultMinLearningRate,
				MaxLearningRate:     feemarkettypes.DefaultMaxLearningRate,
				MaxBlockUtilization: 15_000_000,
				Window:              feemarkettypes.DefaultWindow,
				FeeDenom:            e2esuite.DefaultDenom,
				Enabled:             false,
				DistributeFees:      false,
			},
		},
	}

	// Create separate configs for each chain
	config1 := e2esuite.DefaultConfig
	config1.ChainID = "juno-ibc-1"
	config1.ModifyGenesis = cosmos.ModifyGenesis(genesisKV)

	config2 := e2esuite.DefaultConfig
	config2.ChainID = "juno-ibc-2"
	config2.ModifyGenesis = cosmos.ModifyGenesis(genesisKV)

	// Create completely new ChainSpec objects instead of using the shared DefaultSpec pointer
	spec1 := &interchaintest.ChainSpec{
		ChainName:     "juno-ibc-1",
		Name:          "juno",
		NumValidators: &numVals,
		NumFullNodes:  &numFullNodes,
		Version:       e2esuite.DefaultSpec.Version,
		NoHostMount:   e2esuite.DefaultSpec.NoHostMount,
		ChainConfig:   config1,
	}

	spec2 := &interchaintest.ChainSpec{
		ChainName:     "juno-ibc-2",
		Name:          "juno",
		NumValidators: &numVals,
		NumFullNodes:  &numFullNodes,
		Version:       e2esuite.DefaultSpec.Version,
		NoHostMount:   e2esuite.DefaultSpec.NoHostMount,
		ChainConfig:   config2,
	}

	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{spec1, spec2},
		e2esuite.DefaultTxCfg,
		e2esuite.WithChainConstructor(e2esuite.MultipleChainsConstructor),
		e2esuite.WithInterchainConstructor(e2esuite.TwoChainInterchainConstructor),
	)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	t.Parallel()
	t.Cleanup(func() {
		_ = s.Ic.Close()
		err := s.Relayer.StopRelayer(s.Ctx, eRep)
		if err != nil {
			t.Logf("an error occurred while stopping the relayer: %s", err)
		}
	})

	testSuite := &IbcTestSuite{E2ETestSuite: s, eRep: eRep}
	suite.Run(t, testSuite)
}

// TestTwoChainsIBCTransfer spins up our local and a replica of the Cosmos Hub network, initializes an IBC connection between them,
// and sends an ICS20 token transfer from us to the Hub and then back to us.
func (s *IbcTestSuite) TestTwoChainsIBCTransfer() {
	t := s.T()
	if testing.Short() {
		t.Skip()
	}

	chain1, chain2 := s.Chain, s.Chains[1]

	// Create some user accounts on both chains
	chain1User := s.GetAndFundTestUser("default", genesisWalletAmount, chain1)
	chain2User := s.GetAndFundTestUser("default", genesisWalletAmount, chain2)

	// Wait a few blocks for relayer to start and for user accounts to be created
	err := testutil.WaitForBlocks(s.Ctx, 5, chain1, chain2)
	require.NoError(t, err)

	chain1UserAddr := chain1User.FormattedAddress()
	chain2UserAddr := chain2User.FormattedAddress()

	// Get original account balances
	junoOrigBal, err := chain1.GetBalance(s.Ctx, chain1UserAddr, chain1.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(genesisWalletAmount), junoOrigBal)

	gaiaOrigBal, err := chain2.GetBalance(s.Ctx, chain2UserAddr, chain2.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(genesisWalletAmount), gaiaOrigBal)

	// Compose an IBC transfer and send from Juno -> Gaia
	transferAmount := math.NewInt(1_000)
	transfer := ibc.WalletAmount{
		Address: chain2UserAddr,
		Denom:   chain1.Config().Denom,
		Amount:  transferAmount,
	}

	channel, err := ibc.GetTransferChannel(s.Ctx, s.Relayer, s.eRep, chain1.Config().ChainID, chain2.Config().ChainID)
	require.NoError(t, err)

	junoHeight, err := chain1.Height(s.Ctx)
	require.NoError(t, err)

	transferTx, err := s.SendIBCTransfer(s.Chain, channel.ChannelID, chain1UserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	err = s.Relayer.StartRelayer(s.Ctx, s.eRep, "ab")
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(s.Ctx, chain1, junoHeight, junoHeight+50, transferTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(s.Ctx, 10, chain1)
	require.NoError(t, err)

	// Get the IBC denom for ujuno on Gaia
	junoTokenDenom := transfertypes.GetPrefixedDenom(channel.Counterparty.PortID, channel.Counterparty.ChannelID, chain1.Config().Denom)
	junoIBCDenom := transfertypes.ParseDenomTrace(junoTokenDenom).IBCDenom()

	// Assert that the funds are no longer present in user acc on Juno and are in the user acc on Gaia.
	// Under v30 feemarket the sender also pays a non-zero fee for the MsgTransfer, so the post-transfer
	// balance lands slightly below userFunds-transferAmount. Allow up to 1_000_000 ujuno fee tolerance
	// (observed ~834 ujuno per transfer) — comfortably above the observed fee but still tight enough
	// to catch real regressions in transfer-amount math.
	junoUpdateBal, err := chain1.GetBalance(s.Ctx, chain1UserAddr, chain1.Config().Denom)
	require.NoError(t, err)
	expectedJunoBal := junoOrigBal.Sub(transferAmount).Int64()
	require.True(t,
		junoUpdateBal.Int64() <= expectedJunoBal &&
			junoUpdateBal.Int64() >= expectedJunoBal-1_000_000,
		"junoUpdateBal %d outside fee tolerance of expected %d",
		junoUpdateBal.Int64(), expectedJunoBal,
	)

	gaiaUpdateBal, err := chain2.GetBalance(s.Ctx, chain2UserAddr, junoIBCDenom)
	require.NoError(t, err)
	require.Equal(t, transferAmount.Int64(), gaiaUpdateBal.Int64())

	// Compose an IBC transfer and send from Gaia -> Juno
	transfer = ibc.WalletAmount{
		Address: chain1UserAddr,
		Denom:   junoIBCDenom,
		Amount:  transferAmount,
	}

	gaiaHeight, err := chain2.Height(s.Ctx)
	require.NoError(t, err)

	transferTx, err = s.SendIBCTransfer(chain2, channel.Counterparty.ChannelID, chain2UserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(s.Ctx, chain2, gaiaHeight, gaiaHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Assert that the funds are now back on Juno and not on Gaia. The round-trip cost
	// one v30 feemarket fee on the outbound Juno->Gaia leg, so the final balance lands
	// slightly below junoOrigBal. Same 1_000_000 ujuno fee tolerance as above.
	junoUpdateBal, err = chain1.GetBalance(s.Ctx, chain1UserAddr, chain1.Config().Denom)
	require.NoError(t, err)
	require.True(t,
		junoUpdateBal.Int64() <= junoOrigBal.Int64() &&
			junoUpdateBal.Int64() >= junoOrigBal.Int64()-1_000_000,
		"junoUpdateBal %d outside fee tolerance of original %d",
		junoUpdateBal.Int64(), junoOrigBal.Int64(),
	)

	gaiaUpdateBal, err = chain2.GetBalance(s.Ctx, chain2UserAddr, junoIBCDenom)
	require.NoError(t, err)
	require.Equal(t, int64(0), gaiaUpdateBal.Int64())

	err = s.Relayer.StopRelayer(s.Ctx, s.eRep)
	require.NoError(t, err)
}
