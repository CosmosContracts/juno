package suite

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"golang.org/x/sync/errgroup"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	buildertypes "github.com/skip-mev/pob/x/builder/types"
)

// SimulateTx simulates the provided messages, and checks whether the provided failure condition is met
func (s *E2ETestSuite) SimulateTx(user cosmos.User, height uint64, expectFail bool, msgs ...sdk.Msg) uint64 {
	// create tx factory + Client Context
	txf, err := s.Bc.GetFactory(s.Ctx, user)
	s.Require().NoError(err)

	cc, err := s.Bc.GetClientContext(s.Ctx, user)
	s.Require().NoError(err)

	txf, err = txf.Prepare(cc)
	s.Require().NoError(err)

	// set timeout height
	if height != 0 {
		txf = txf.WithTimeoutHeight(height)
	}

	// get gas for tx
	_, gas, err := tx.CalculateGas(cc, txf, msgs...)
	s.Require().Equal(err != nil, expectFail)

	return gas
}

func (s *E2ETestSuite) SendCoinsMultiBroadcast(sender, receiver ibc.Wallet, amt, fees sdk.Coins, gas int64, numMsg int) (*coretypes.ResultBroadcastTxCommit, error) {
	msgs := make([]sdk.Msg, numMsg)
	for i := range numMsg {
		msgs[i] = &banktypes.MsgSend{
			FromAddress: sender.FormattedAddress(),
			ToAddress:   receiver.FormattedAddress(),
			Amount:      amt,
		}
	}

	tx := s.CreateTx(s.Chain, sender, fees.String(), gas, false, msgs...)

	// get an rpc endpoint for the chain
	c := s.Chain.Nodes()[0].Client
	return c.BroadcastTxCommit(s.Ctx, tx)
}

func (s *E2ETestSuite) SendCoinsMultiBroadcastAsync(sender, receiver ibc.Wallet, amt, fees sdk.Coins,
	gas int64, numMsg int, bumpSequence bool,
) (*coretypes.ResultBroadcastTx, error) {
	msgs := make([]sdk.Msg, numMsg)
	for i := range numMsg {
		msgs[i] = &banktypes.MsgSend{
			FromAddress: sender.FormattedAddress(),
			ToAddress:   receiver.FormattedAddress(),
			Amount:      amt,
		}
	}

	tx := s.CreateTx(s.Chain, sender, fees.String(), gas, bumpSequence, msgs...)

	// get an rpc endpoint for the chain
	c := s.Chain.Nodes()[0].Client
	return c.BroadcastTxAsync(s.Ctx, tx)
}

// SendCoins creates a executes a SendCoins message and broadcasts the transaction.
func (s *E2ETestSuite) SendCoins(chain *cosmos.CosmosChain, keyName, sender, receiver string, amt, fees sdk.Coins) (string, error) {
	resp, err := s.ExecTx(
		chain,
		keyName,
		false,
		false,
		"bank",
		"send",
		sender,
		receiver,
		amt.String(),
		"--fees",
		fees.String(),
		"--gas",
		// strconv.FormatInt(gas, 10),
		"auto",
	)

	return resp, err
}

// GetAndFundTestUserWithMnemonic restores a user using the given mnemonic
// and funds it with the native chain denom.
// The caller should wait for some blocks to complete before the funds will be accessible.
func (s *E2ETestSuite) GetAndFundTestUserWithMnemonic(
	keyNamePrefix, mnemonic string,
	amount int64,
	chain ibc.Chain,
) (ibc.Wallet, error) {
	chainCfg := chain.Config()
	keyName := fmt.Sprintf("%s-%s-%s", keyNamePrefix, chainCfg.ChainID, AlphaString(3))
	user, err := chain.BuildWallet(s.Ctx, keyName, mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to get source user wallet: %w", err)
	}

	s.FundUser(chain, amount, user)
	return user, nil
}

func (s *E2ETestSuite) FundUser(chain ibc.Chain, amount int64, user ibc.Wallet) {
	chainCfg := chain.Config()

	_, err := s.SendCoins(
		chain.(*cosmos.CosmosChain),
		interchaintest.FaucetAccountKeyName,
		interchaintest.FaucetAccountKeyName,
		user.FormattedAddress(),
		sdk.NewCoins(sdk.NewCoin(chainCfg.Denom, math.NewInt(amount))),
		sdk.NewCoins(sdk.NewCoin(chainCfg.Denom, math.NewInt(1_000_000))),
	)
	s.Require().NoError(err, "failed to get funds from faucet")
}

// GetAndFundTestUser generates and funds a chain user with the native chain denom.
// The caller should wait for some blocks to complete before the funds will be accessible.
func (s *E2ETestSuite) GetAndFundTestUser(
	keyNamePrefix string,
	amount int64,
	chain ibc.Chain,
) ibc.Wallet {
	t := s.T()
	t.Helper()
	user, err := s.GetAndFundTestUserWithMnemonic(keyNamePrefix, "", amount, chain)
	s.Require().NoError(err)

	return user
}

// GetAndFundTestUserOnAllChains generates and funds users wallets on all chains with the native chain denom.
// The caller should wait for some blocks to complete before the funds will be accessible.
func (s *E2ETestSuite) GetAndFundTestUserOnAllChains(
	keyNamePrefix string,
	amount int64,
	chains ...ibc.Chain,
) []ibc.Wallet {
	t := s.T()
	t.Helper()

	users := make([]ibc.Wallet, len(chains))
	var eg errgroup.Group
	for i, chain := range chains {
		eg.Go(func() error {
			user, err := s.GetAndFundTestUserWithMnemonic(keyNamePrefix, "", amount, chain)
			if err != nil {
				return err
			}
			users[i] = user
			return nil
		})
	}
	require.NoError(t, eg.Wait())

	chainHeights := make([]testutil.ChainHeighter, len(chains))
	for i := range chains {
		chainHeights[i] = chains[i]
	}
	return users
}

// ExecTx executes a cli command on a node, waits a block and queries the Tx to verify it was included on chain.
func (s *E2ETestSuite) ExecTx(chain *cosmos.CosmosChain, keyName string, blocking bool, skipTxCheck bool, command ...string) (string, error) {
	node := chain.Validators[0]

	resp, err := node.ExecTx(s.Ctx, keyName, command...)
	if skipTxCheck {
		return resp, nil
	}

	s.Require().NoError(err)

	if !blocking {
		return resp, nil
	}

	height, err := chain.Height(context.Background())
	s.Require().NoError(err)
	s.WaitForHeight(chain, height+1)

	stdout, stderr, err := chain.FullNodes[0].ExecQuery(s.Ctx, "tx", resp, "--type", "hash")
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

// SubmitSoftwareUpgradeProposal submits a software upgrade proposal to the chain.
func (s *E2ETestSuite) SubmitSoftwareUpgradeProposal(chain *cosmos.CosmosChain, user ibc.Wallet, upgradeName string, haltHeight int64, authority string) string {
	t := s.T()
	upgradeMsg := []cosmos.ProtoMessage{
		&upgradetypes.MsgSoftwareUpgrade{
			// gov module account
			Authority: authority,
			Plan: upgradetypes.Plan{
				Name:   upgradeName,
				Height: int64(haltHeight),
			},
		},
	}

	proposal, err := chain.BuildProposal(
		upgradeMsg,
		"Chain Upgrade 1",
		"Summary desc",
		"ipfs://CID",
		fmt.Sprintf(`500000000%s`, chain.Config().Denom),
		sdk.MustBech32ifyAddressBytes("juno", user.Address()),
		false)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(s.Ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	return txProp.ProposalID
}

// DO NOT USE, only used for the gov fix test, not compatible with Juno v28+
func (s *E2ETestSuite) SubmitBuilderParamsUpdate(chain *cosmos.CosmosChain, user ibc.Wallet, authority string) string {
	t := s.T()
	// juno10d07y265gmmuvt4z0w9aw880jnsr700jvss730
	govModule := sdk.MustAccAddressFromBech32(authority)

	updateParamsMsg := []cosmos.ProtoMessage{
		&buildertypes.MsgUpdateParams{
			Authority: authority,
			Params: buildertypes.Params{
				FrontRunningProtection: true,
				ProposerFee:            sdkmath.LegacyMustNewDecFromStr("1"),
				ReserveFee:             sdk.NewCoin("ujuno", sdkmath.NewInt(1)),
				MinBidIncrement:        sdk.NewCoin("ujuno", sdkmath.NewInt(1000)),
				MaxBundleSize:          100,
				EscrowAccountAddress:   govModule.Bytes(),
			},
		},
	}

	proposal, err := chain.BuildProposal(
		updateParamsMsg,
		"Update Builder Params",
		"Summary desc",
		"ipfs://CID",
		fmt.Sprintf(`500000000%s`, chain.Config().Denom),
		sdk.MustBech32ifyAddressBytes("juno", user.Address()),
		false)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(s.Ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	return txProp.ProposalID
}

func (s *E2ETestSuite) VoteOnProp(chain *cosmos.CosmosChain, proposalID uint64, height int64) {
	err := chain.VoteOnProposalAllValidators(s.Ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(s.T(), err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(s.Ctx, chain, height, height+20, proposalID, govv1beta1types.StatusPassed)
	require.NoError(s.T(), err, "proposal status did not change to passed in expected number of blocks")

	_, timeoutCtxCancel := context.WithTimeout(s.Ctx, time.Second*45)
	defer timeoutCtxCancel()
}
