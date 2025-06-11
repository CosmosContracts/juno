package ibc_test

import (
	"encoding/json"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
)

type PacketMetadata struct {
	Forward *ForwardMetadata `json:"forward"`
}

type ForwardMetadata struct {
	Receiver       string        `json:"receiver"`
	Port           string        `json:"port"`
	Channel        string        `json:"channel"`
	Timeout        time.Duration `json:"timeout"`
	Retries        *uint8        `json:"retries,omitempty"`
	Next           *string       `json:"next,omitempty"`
	RefundSequence *uint64       `json:"refund_sequence,omitempty"`
}

type PfmTestSuite struct {
	*e2esuite.E2ETestSuite

	eRep *testreporter.RelayerExecReporter
}

func TestPfmTestSuite(t *testing.T) {
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

	config3 := e2esuite.DefaultConfig
	config3.ChainID = "juno-ibc-3"
	config3.ModifyGenesis = cosmos.ModifyGenesis(genesisKV)

	config4 := e2esuite.DefaultConfig
	config4.ChainID = "juno-ibc-4"
	config4.ModifyGenesis = cosmos.ModifyGenesis(genesisKV)

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

	spec3 := &interchaintest.ChainSpec{
		ChainName:     "juno-ibc-3",
		Name:          "juno",
		NumValidators: &numVals,
		NumFullNodes:  &numFullNodes,
		Version:       e2esuite.DefaultSpec.Version,
		NoHostMount:   e2esuite.DefaultSpec.NoHostMount,
		ChainConfig:   config3,
	}

	spec4 := &interchaintest.ChainSpec{
		ChainName:     "juno-ibc-4",
		Name:          "juno",
		NumValidators: &numVals,
		NumFullNodes:  &numFullNodes,
		Version:       e2esuite.DefaultSpec.Version,
		NoHostMount:   e2esuite.DefaultSpec.NoHostMount,
		ChainConfig:   config4,
	}

	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{spec1, spec2, spec3, spec4},
		e2esuite.DefaultTxCfg,
		e2esuite.WithChainConstructor(e2esuite.MultipleChainsConstructor),
		e2esuite.WithInterchainConstructor(e2esuite.FourChainInterchainConstructor),
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

	testSuite := &PfmTestSuite{E2ETestSuite: s, eRep: eRep}
	suite.Run(t, testSuite)
}

// TestPacketForwardMiddlewareRouter ensures the PFM module is set up properly and works as expected.
func (s *PfmTestSuite) TestPacketForwardMiddlewareRouter() {
	t := s.T()
	if testing.Short() {
		t.Skip()
	}

	err := s.Relayer.StartRelayer(s.Ctx, s.eRep, "ab", "bc", "cd")
	require.NoError(t, err)

	userFunds := sdkmath.NewInt(10_000_000_000)
	users := s.GetAndFundTestUserOnAllChains(t.Name(), userFunds.Int64(), s.Chain, s.Chains[1], s.Chains[2], s.Chains[3])

	abChan, err := ibc.GetTransferChannel(s.Ctx, s.Relayer, s.eRep, s.Chain.Config().ChainID, s.Chains[1].Config().ChainID)
	require.NoError(t, err)
	baChan := abChan.Counterparty

	cbChan, err := ibc.GetTransferChannel(s.Ctx, s.Relayer, s.eRep, s.Chains[2].Config().ChainID, s.Chains[1].Config().ChainID)
	require.NoError(t, err)
	bcChan := cbChan.Counterparty

	dcChan, err := ibc.GetTransferChannel(s.Ctx, s.Relayer, s.eRep, s.Chains[3].Config().ChainID, s.Chains[2].Config().ChainID)
	require.NoError(t, err)
	cdChan := dcChan.Counterparty

	// Get original account balances
	userA, userB, userC, userD := users[0], users[1], users[2], users[3]
	transferAmount := sdkmath.NewInt(100_000)

	// Compose the prefixed denoms and ibc denom for asserting balances
	firstHopDenom := transfertypes.GetPrefixedDenom(baChan.PortID, baChan.ChannelID, s.Chain.Config().Denom)
	secondHopDenom := transfertypes.GetPrefixedDenom(cbChan.PortID, cbChan.ChannelID, firstHopDenom)
	thirdHopDenom := transfertypes.GetPrefixedDenom(dcChan.PortID, dcChan.ChannelID, secondHopDenom)

	firstHopDenomTrace := transfertypes.ParseDenomTrace(firstHopDenom)
	secondHopDenomTrace := transfertypes.ParseDenomTrace(secondHopDenom)
	thirdHopDenomTrace := transfertypes.ParseDenomTrace(thirdHopDenom)

	firstHopIBCDenom := firstHopDenomTrace.IBCDenom()
	secondHopIBCDenom := secondHopDenomTrace.IBCDenom()
	thirdHopIBCDenom := thirdHopDenomTrace.IBCDenom()

	firstHopEscrowAccount := sdk.MustBech32ifyAddressBytes(s.Chain.Config().Bech32Prefix, transfertypes.GetEscrowAddress(abChan.PortID, abChan.ChannelID))
	secondHopEscrowAccount := sdk.MustBech32ifyAddressBytes(s.Chains[1].Config().Bech32Prefix, transfertypes.GetEscrowAddress(bcChan.PortID, bcChan.ChannelID))
	thirdHopEscrowAccount := sdk.MustBech32ifyAddressBytes(s.Chains[2].Config().Bech32Prefix, transfertypes.GetEscrowAddress(cdChan.PortID, abChan.ChannelID))

	t.Run("multi-hop a->b->c->d", func(t *testing.T) {
		// Send packet from Chain A->Chain B->Chain C->Chain D

		transfer := ibc.WalletAmount{
			Address: userB.FormattedAddress(),
			Denom:   s.Chain.Config().Denom,
			Amount:  transferAmount,
		}

		secondHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userD.FormattedAddress(),
				Channel:  cdChan.ChannelID,
				Port:     cdChan.PortID,
			},
		}
		nextBz, err := json.Marshal(secondHopMetadata)
		require.NoError(t, err)
		next := string(nextBz)

		firstHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userC.FormattedAddress(),
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
				Next:     &next,
			},
		}

		memo, err := json.Marshal(firstHopMetadata)
		require.NoError(t, err)

		chainAHeight, err := s.Chain.Height(s.Ctx)
		require.NoError(t, err)

		transferTx, err := s.SendIBCTransfer(s.Chain, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(s.Ctx, s.Chain, chainAHeight, chainAHeight+35, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(s.Ctx, 5, s.Chain)
		require.NoError(t, err)

		chainABalance, err := s.Chain.GetBalance(s.Ctx, userA.FormattedAddress(), s.Chain.Config().Denom)
		require.NoError(t, err)

		chainBBalance, err := s.Chains[1].GetBalance(s.Ctx, userB.FormattedAddress(), firstHopIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := s.Chains[2].GetBalance(s.Ctx, userC.FormattedAddress(), secondHopIBCDenom)
		require.NoError(t, err)

		chainDBalance, err := s.Chains[3].GetBalance(s.Ctx, userD.FormattedAddress(), thirdHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, userFunds.Sub(transferAmount), chainABalance)
		require.Equal(t, sdkmath.NewInt(0), chainBBalance)
		require.Equal(t, sdkmath.NewInt(0), chainCBalance)
		require.Equal(t, transferAmount.Int64(), chainDBalance.Int64())

		firstHopEscrowBalance, err := s.Chain.GetBalance(s.Ctx, firstHopEscrowAccount, s.Chain.Config().Denom)
		require.NoError(t, err)

		secondHopEscrowBalance, err := s.Chains[1].GetBalance(s.Ctx, secondHopEscrowAccount, firstHopIBCDenom)
		require.NoError(t, err)

		thirdHopEscrowBalance, err := s.Chains[2].GetBalance(s.Ctx, thirdHopEscrowAccount, secondHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, transferAmount.Int64(), firstHopEscrowBalance.Int64())
		require.Equal(t, transferAmount.Int64(), secondHopEscrowBalance.Int64())
		require.Equal(t, transferAmount.Int64(), thirdHopEscrowBalance.Int64())
	})

	err = s.Relayer.StopRelayer(s.Ctx, s.eRep)
	require.NoError(t, err)
}
