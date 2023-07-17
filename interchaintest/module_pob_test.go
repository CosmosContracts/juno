package interchaintest

// credit: https://github.com/persistenceOne/persistenceCore/blob/main/interchaintest/module_pob_test.go

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

const testWallet = "juno10r39fueph9fq7a6lgswu4zdsg8t3gxlq670lt0"

// TestSkipMevAuction tests that x/builder corretly wired and allows to make auctions to prioritise txns
func TestSkipMevAuction(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// override SDK beck prefixes with chain specific
	helpers.SetConfig()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)
	chain := chains[0].(*cosmos.CosmosChain)
	testDenom := chain.Config().Denom

	require.NotNil(t, ic)
	require.NotNil(t, ctx)

	t.Cleanup(func() {
		_ = ic.Close()
	})

	const userFunds = int64(10_000_000_000)

	chainUserMnemonic := helpers.NewMnemonic()
	chainUser, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, t.Name(), chainUserMnemonic, userFunds, chain)
	require.NoError(t, err)

	chainNode := chain.Nodes()[0]

	paramsStdout, _, err := chainNode.ExecQuery(ctx, "builder", "params")
	require.NoError(t, err)
	t.Log("checking POB params", string(paramsStdout))

	kb, err := helpers.NewKeyringFromMnemonic(cosmos.DefaultEncoding().Codec, chainUser.KeyName(), chainUserMnemonic)
	require.NoError(t, err)

	clientContext := chainNode.CliContext()
	clientContext = clientContext.WithCodec(chain.Config().EncodingConfig.Codec)

	txFactory := helpers.NewTxFactory(clientContext)
	txFactory = txFactory.WithKeybase(kb)
	txFactory = txFactory.WithTxConfig(junoEncoding().TxConfig)

	accountRetriever := authtypes.AccountRetriever{}
	accountNum, currentSeq, err := accountRetriever.GetAccountNumberSequence(clientContext, chainUser.Address())
	require.NoError(t, err)

	txFactory = txFactory.WithAccountNumber(accountNum)
	// the tx that we put on auction will have the next sequence
	txFactory = txFactory.WithSequence(currentSeq + 1)

	txn, err := txFactory.BuildUnsignedTx(&banktypes.MsgSend{
		FromAddress: chainUser.FormattedAddress(),
		ToAddress:   testWallet,
		Amount:      sdk.NewCoins(sdk.NewCoin(testDenom, sdk.NewInt(100))),
	})
	require.NoError(t, err)

	currentHeight, err := chain.Height(ctx)
	require.NoError(t, err)

	// transaction simulation there is possible, but we skip it for now
	txn.SetGasLimit(100000)
	txn.SetTimeoutHeight(currentHeight + 5)

	err = tx.Sign(txFactory, chainUser.KeyName(), txn, true)
	require.NoError(t, err)

	auctionBid := sdk.NewCoin(testDenom, sdk.NewInt(100))
	BuilderAuctionBid(
		t, ctx, chain,
		chainUser,
		chainUser.FormattedAddress(),
		auctionBid,
		currentHeight+5,
		txn.GetTx(),
	)

	recipientBalance, err := chain.GetBalance(ctx, testWallet, testDenom)
	require.NoError(t, err)

	require.Equal(t, int64(100), recipientBalance, "recipient must have balance")

	// TODO: verify that tx is actually prioritized over other tx's in the block
	// The best way to do so it by using a wasm counter contract, but it requires some more orchestration
	// Send three tx's: [low bid, higher bid, normal tx] three times.
}

func BuilderAuctionBid(
	t *testing.T,
	ctx context.Context,
	chain *cosmos.CosmosChain,
	user ibc.Wallet,
	bidder string,
	bid sdk.Coin,
	timeoutHeight uint64,
	transactions ...sdk.Tx,
) {
	txBytes := make([]string, 0, len(transactions))
	for _, tx := range transactions {
		bz, err := chain.Config().EncodingConfig.TxConfig.TxEncoder()(tx)
		if err != nil {
			require.NoError(t, err)
			return
		}

		txBytes = append(txBytes, fmt.Sprintf("%2X", bz))
	}

	//  junod tx builder auction-bid [bidder] [bid] [bundled_tx1,bundled_tx2,...,bundled_txN]
	cmd := append([]string{
		"builder", "auction-bid", bidder, bid.String(),
	}, txBytes...)

	// NOTE: --timeout-height is mandatory
	cmd = append(cmd, fmt.Sprintf("--timeout-height=%d", timeoutHeight))

	chainNode := chain.Nodes()[0]
	txHash, err := chainNode.ExecTx(ctx, user.KeyName(), cmd...)
	require.NoError(t, err)

	stdout, _, err := chainNode.ExecQuery(ctx, "tx", txHash)
	require.NoError(t, err)

	fmt.Println(string(stdout))
}
