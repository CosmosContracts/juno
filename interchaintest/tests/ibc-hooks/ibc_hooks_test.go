package ibc_test

import (
	"fmt"
	"strings"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

var genesisWalletAmount = int64(10_000_000)

type IbcHooksTestSuite struct {
	*e2esuite.E2ETestSuite

	eRep *testreporter.RelayerExecReporter
}

func TestIbcHooksTestSuite(t *testing.T) {
	numVals := 1
	numFullNodes := 0

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
		}}

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

	testSuite := &IbcHooksTestSuite{E2ETestSuite: s, eRep: eRep}
	suite.Run(t, testSuite)
}

// TestIBCHooks ensures the ibc-hooks middleware from osmosis works.
func (s *IbcHooksTestSuite) TestIBCHooks() {
	t := s.T()
	if testing.Short() {
		t.Skip()
	}

	err := s.Relayer.StartRelayer(s.Ctx, s.eRep, "ab")
	require.NoError(t, err)

	// Create some user accounts on both chains
	junoUser := s.GetAndFundTestUser("default", genesisWalletAmount, s.Chain)
	gaiaUser := s.GetAndFundTestUser("default", genesisWalletAmount, s.Chains[1])

	// Wait a few blocks for relayer to start and for user accounts to be created
	err = testutil.WaitForBlocks(s.Ctx, 5, s.Chain, s.Chains[1])
	require.NoError(t, err)

	junoUserAddr := junoUser.FormattedAddress()

	channel, err := ibc.GetTransferChannel(s.Ctx, s.Relayer, s.eRep, s.Chain.Config().ChainID, s.Chains[1].Config().ChainID)
	require.NoError(t, err)

	fees := sdk.NewCoins(sdk.NewCoin(s.Chain.Config().Denom, math.NewInt(100_000)))
	_, contractAddr := s.SetupContract(s.Chains[1], gaiaUser.KeyName(), "../../contracts/ibchooks_counter.wasm", `{"count":0}`, false, fees)

	// do an ibc transfer through the memo to the other chain.
	transfer := ibc.WalletAmount{
		Address: contractAddr,
		Denom:   s.Chain.Config().Denom,
		Amount:  math.NewInt(1),
	}

	memo := ibc.TransferOptions{
		Memo: fmt.Sprintf(`{"wasm":{"contract":"%s","msg":%s}}`, contractAddr, `{"increment":{}}`),
	}

	// Initial transfer. Account is created by the wasm execute is not so we must do this twice to properly set up
	transferTx, err := s.SendIBCTransfer(s.Chain, channel.ChannelID, junoUser.KeyName(), transfer, memo)
	require.NoError(t, err)
	junoHeight, err := s.Chain.Height(s.Ctx)
	require.NoError(t, err)

	_, err = testutil.PollForAck(s.Ctx, s.Chain, junoHeight-5, junoHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Second time, this will make the counter == 1 since the account is now created.
	transferTx, err = s.Chain.SendIBCTransfer(s.Ctx, channel.ChannelID, junoUser.KeyName(), transfer, memo)
	require.NoError(t, err)
	junoHeight, err = s.Chain.Height(s.Ctx)
	require.NoError(t, err)

	_, err = testutil.PollForAck(s.Ctx, s.Chain, junoHeight-5, junoHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Get the address on the other chain's side
	addr := s.GetIBCHooksUserAddress(s.Chain, channel.ChannelID, junoUserAddr)
	require.NotEmpty(t, addr)

	// Get funds on the receiving chain
	funds := s.GetIBCHookTotalFunds(s.Chains[1], contractAddr, addr)
	require.Equal(t, int(1), len(funds.Data.TotalFunds))

	var ibcDenom string
	for _, coin := range funds.Data.TotalFunds {
		if strings.HasPrefix(coin.Denom, "ibc/") {
			ibcDenom = coin.Denom
			break
		}
	}
	require.NotEmpty(t, ibcDenom)

	// ensure the count also increased to 1 as expected.
	count := s.GetIBCHookCount(s.Chains[1], contractAddr, addr)
	require.Equal(t, int64(1), count.Data.Count)

	err = s.Relayer.StopRelayer(s.Ctx, s.eRep)
	require.NoError(t, err)
}
