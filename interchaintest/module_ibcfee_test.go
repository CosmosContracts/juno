package interchaintest

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
)

type fee_enabled_channels struct {
	ChannelID string `json:"channel_id"`
	PortID    string `json:"port_id"`
}

func ExecuteIBCFeeRegisterPayee(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, relayer ibc.Wallet, portID, channelID, payoutAddr string) {
	// junod tx ibc-fee register-payee transfer channel-0 cosmos1rsp837a4kvtgp2m4uqzdge0zzu6efqgucm0qdh cosmos153lf4zntqt33a4v0sm5cytrxyqn78q7kz8j8x5

	cmd := []string{"junod", "tx", "ibc-fee", "register-payee", portID, channelID, relayer.FormattedAddress(), payoutAddr,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", relayer.KeyName(),
		"--gas", "500000",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	t.Log(cmd)
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	t.Log(string(stdout))

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}

func GetIBCFeeRegisteredPayoutAddress(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, channelID, address string) *ibcfeetypes.QueryPayeeResponse {
	cmd := []string{"junod", "query", "ibc-fee", "payee", channelID, address,
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	t.Log(string(stdout))

	results := &ibcfeetypes.QueryPayeeResponse{}
	err = json.Unmarshal(stdout, results)
	require.NoError(t, err)

	return results
}

func ExecuteIBCFeePayPacketFee(t *testing.T, ctx context.Context, wallet ibc.Wallet, chain *cosmos.CosmosChain, portID, channelID string, sequence uint64, recvFee, ackFee, timeoutFee string) {
	// junod tx ibc-fee pay-packet-fee transfer channel-0 1 --recv-fee 10stake --ack-fee 10stake --timeout-fee 10stake
	cmd := []string{"junod", "tx", "ibc-fee", "pay-packet-fee", portID, channelID, fmt.Sprintf("%d", sequence),
		"--recv-fee", recvFee, "--ack-fee", ackFee, "--timeout-fee", timeoutFee,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", wallet.KeyName(),
		"--gas", "500000",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	t.Log(cmd)
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	t.Log(string(stdout))

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}

// TestJunoIBCFee test the IBCFee module.
func TestJunoIBCFee(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Create chain factory with Juno and Gaia
	numVals := 1
	numFullNodes := 0

	enabledIBCFee := cosmos.GenesisKV{
		Key: "app_state.feeibc.fee_enabled_channels",
		Value: []fee_enabled_channels{
			{
				ChannelID: "channel-0",
				PortID:    "transfer",
			},
		},
	}

	cfg := junoConfig.Clone()
	cfg.ModifyGenesis = cosmos.ModifyGenesis(append(defaultGenesisKV, enabledIBCFee))

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "juno",
			ChainConfig:   cfg,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "gaia",
			Version:       "v9.1.0",
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	const (
		path = "ibcfee-path"
	)

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	client, network := interchaintest.DockerSetup(t)

	juno, gaia := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	relayerType, relayerName := ibc.CosmosRly, "relay"

	// Get a relayer instance
	rf := interchaintest.NewBuiltinRelayerFactory(
		relayerType,
		zaptest.NewLogger(t),
		interchaintestrelayer.CustomDockerImage("ghcr.io/cosmos/relayer", "latest", "100:1000"),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)

	r := rf.Build(t, client, network)

	ic := interchaintest.NewInterchain().
		AddChain(juno).
		AddChain(gaia).
		AddRelayer(r, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  juno,
			Chain2:  gaia,
			Relayer: r,
			Path:    path,
		})

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
		SkipPathCreation:  false,
	}))

	// Get channel
	channel, err := ibc.GetTransferChannel(ctx, r, eRep, juno.Config().ChainID, gaia.Config().ChainID)
	t.Log(channel)
	require.NoError(t, err)

	const userFunds = int64(1_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, juno, juno, gaia)
	junoPayoutUser := users[0]
	junoSender := users[1]
	gaiaUser := users[2]

	relayerWallet, found := r.GetWallet(cfg.ChainID)
	require.True(t, found)

	// for some reason relayerWallet does not have a keyname. So we load it in here as relayer key
	relayer, _ := juno.BuildWallet(ctx, "myRelayerKey", relayerWallet.Mnemonic())

	ExecuteIBCFeeRegisterPayee(t, ctx, juno, relayer, "transfer", "channel-0", junoPayoutUser.FormattedAddress())

	payeeResp := GetIBCFeeRegisteredPayoutAddress(t, ctx, juno, "channel-0", relayerWallet.FormattedAddress())
	require.Equal(t, payeeResp.PayeeAddress, junoPayoutUser.FormattedAddress())

	const transferAmount = int64(1_000)
	transfer := ibc.WalletAmount{
		Address: gaiaUser.FormattedAddress(),
		Denom:   juno.Config().Denom,
		Amount:  transferAmount,
	}

	// TODO: This logic may break if/when the relayer automatically transfers packets.
	// Currently breaks: &{STATE_INIT ORDER_UNORDERED {transfer } [connection-0] ics20-1 transfer channel-0}
	// State should be OPEN. Seems 
	transferTx, err := juno.SendIBCTransfer(ctx, channel.ChannelID, junoSender.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	require.NotNil(t, transferTx)

	junoHeight, err := juno.Height(ctx)
	require.NoError(t, err)

	// get balance of junoPayoutUser
	bal, _ := juno.GetBalance(ctx, junoPayoutUser.FormattedAddress(), cfg.Denom)
	require.Equal(t, bal, userFunds)

	// Incentivize packet
	ExecuteIBCFeePayPacketFee(t, ctx, junoSender, juno, "transfer", "channel-0", transferTx.Packet.Sequence, "999ujuno", "88ujuno", "7ujuno")

	// we expect a proper packet to be sent.
	r.Flush(ctx, eRep, path, channel.ChannelID)
	_, err = testutil.PollForAck(ctx, juno, junoHeight-5, junoHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// check payoutAddr balance has increased
	bal, _ = juno.GetBalance(ctx, junoPayoutUser.FormattedAddress(), cfg.Denom)
	t.Log("Payout Addr Balance", bal)
	require.True(t, bal > userFunds)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
