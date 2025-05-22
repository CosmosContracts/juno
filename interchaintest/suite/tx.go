package suite

import (
	"context"
	"fmt"
	"strconv"

	"cosmossdk.io/math"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

// SimulateTx simulates the provided messages, and checks whether the provided failure condition is met
func (s *E2ETestSuite) SimulateTx(ctx context.Context, user cosmos.User, height uint64, expectFail bool, msgs ...sdk.Msg) {
	// create tx factory + Client Context
	txf, err := s.Bc.GetFactory(ctx, user)
	s.Require().NoError(err)

	cc, err := s.Bc.GetClientContext(ctx, user)
	s.Require().NoError(err)

	txf, err = txf.Prepare(cc)
	s.Require().NoError(err)

	// set timeout height
	if height != 0 {
		txf = txf.WithTimeoutHeight(height)
	}

	// get gas for tx
	_, _, err = tx.CalculateGas(cc, txf, msgs...)
	s.Require().Equal(err != nil, expectFail)
}

func (s *E2ETestSuite) SendCoinsMultiBroadcast(ctx context.Context, sender, receiver ibc.Wallet, amt, fees sdk.Coins, gas int64, numMsg int) (*coretypes.ResultBroadcastTxCommit, error) {
	msgs := make([]sdk.Msg, numMsg)
	for i := 0; i < numMsg; i++ {
		msgs[i] = &banktypes.MsgSend{
			FromAddress: sender.FormattedAddress(),
			ToAddress:   receiver.FormattedAddress(),
			Amount:      amt,
		}
	}

	tx := s.CreateTx(s.Chain, sender, fees.String(), gas, false, msgs...)

	// get an rpc endpoint for the chain
	c := s.Chain.Nodes()[0].Client
	return c.BroadcastTxCommit(ctx, tx)
}

func (s *E2ETestSuite) SendCoinsMultiBroadcastAsync(ctx context.Context, sender, receiver ibc.Wallet, amt, fees sdk.Coins,
	gas int64, numMsg int, bumpSequence bool,
) (*coretypes.ResultBroadcastTx, error) {
	msgs := make([]sdk.Msg, numMsg)
	for i := 0; i < numMsg; i++ {
		msgs[i] = &banktypes.MsgSend{
			FromAddress: sender.FormattedAddress(),
			ToAddress:   receiver.FormattedAddress(),
			Amount:      amt,
		}
	}

	tx := s.CreateTx(s.Chain, sender, fees.String(), gas, bumpSequence, msgs...)

	// get an rpc endpoint for the chain
	c := s.Chain.Nodes()[0].Client
	return c.BroadcastTxAsync(ctx, tx)
}

// SendCoins creates a executes a SendCoins message and broadcasts the transaction.
func (s *E2ETestSuite) SendCoins(ctx context.Context, keyName, sender, receiver string, amt, fees sdk.Coins, gas int64) (string, error) {
	resp, err := s.ExecTx(
		ctx,
		s.Chain,
		keyName,
		false,
		"bank",
		"send",
		sender,
		receiver,
		amt.String(),
		"--fees",
		fees.String(),
		"--gas",
		strconv.FormatInt(gas, 10),
	)

	return resp, err
}

// GetAndFundTestUserWithMnemonic restores a user using the given mnemonic
// and funds it with the native chain denom.
// The caller should wait for some blocks to complete before the funds will be accessible.
func (s *E2ETestSuite) GetAndFundTestUserWithMnemonic(
	ctx context.Context,
	keyNamePrefix, mnemonic string,
	amount int64,
	chain ibc.Chain,
) (ibc.Wallet, error) {
	chainCfg := chain.Config()
	keyName := fmt.Sprintf("%s-%s", keyNamePrefix, chainCfg.ChainID)
	user, err := chain.BuildWallet(ctx, keyName, mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to get source user wallet: %w", err)
	}

	s.FundUser(ctx, chain, amount, user)
	return user, nil
}

func (s *E2ETestSuite) FundUser(ctx context.Context, chain ibc.Chain, amount int64, user ibc.Wallet) {
	chainCfg := chain.Config()

	_, err := s.SendCoins(
		ctx,
		interchaintest.FaucetAccountKeyName,
		interchaintest.FaucetAccountKeyName,
		user.FormattedAddress(),
		sdk.NewCoins(sdk.NewCoin(chainCfg.Denom, math.NewInt(amount))),
		sdk.NewCoins(sdk.NewCoin(chainCfg.Denom, math.NewInt(1000000000000))),
		1000000,
	)
	s.Require().NoError(err, "failed to get funds from faucet")
}

// GetAndFundTestUser generates and funds a chain user with the native chain denom.
// The caller should wait for some blocks to complete before the funds will be accessible.
func (s *E2ETestSuite) GetAndFundTestUser(
	ctx context.Context,
	keyNamePrefix string,
	amount int64,
	chain ibc.Chain,
) ibc.Wallet {
	user, err := s.GetAndFundTestUserWithMnemonic(ctx, keyNamePrefix, "", amount, chain)
	s.Require().NoError(err)

	return user
}

// ExecTx executes a cli command on a node, waits a block and queries the Tx to verify it was included on chain.
func (s *E2ETestSuite) ExecTx(ctx context.Context, chain *cosmos.CosmosChain, keyName string, blocking bool, command ...string) (string, error) {
	node := chain.Validators[0]

	resp, err := node.ExecTx(ctx, keyName, command...)
	s.Require().NoError(err)

	if !blocking {
		return resp, nil
	}

	height, err := chain.Height(context.Background())
	s.Require().NoError(err)
	s.WaitForHeight(chain, height+1)

	stdout, stderr, err := chain.FullNodes[0].ExecQuery(ctx, "tx", resp, "--type", "hash")
	s.Require().NoError(err)
	s.Require().Nil(stderr)

	return string(stdout), nil
}

// CreateTx creates a new transaction to be signed by the given user, including a provided set of messages
func (s *E2ETestSuite) CreateTx(chain *cosmos.CosmosChain, user cosmos.User, fee string, gas int64,
	bumpSequence bool, msgs ...sdk.Msg,
) []byte {
	bc := cosmos.NewBroadcaster(s.T(), chain)

	ctx := context.Background()
	// create tx factory + Client Context
	txf, err := bc.GetFactory(ctx, user)
	s.Require().NoError(err)

	cc, err := bc.GetClientContext(ctx, user)
	s.Require().NoError(err)

	txf = txf.WithSimulateAndExecute(true)

	txf, err = txf.Prepare(cc)
	s.Require().NoError(err)

	// get gas for tx
	txf = txf.WithGas(uint64(gas))
	txf = txf.WithGasAdjustment(0)
	txf = txf.WithGasPrices("")
	txf = txf.WithFees(fee)

	// update sequence number
	txf = txf.WithSequence(txf.Sequence())
	if bumpSequence {
		txf = txf.WithSequence(txf.Sequence() + 1)
	}

	// sign the tx
	txBuilder, err := txf.BuildUnsignedTx(msgs...)
	s.Require().NoError(err)
	s.Require().NoError(tx.Sign(cc.CmdContext, txf, cc.GetFromName(), txBuilder, true))

	// encode and return
	bz, err := cc.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)
	return bz
}
