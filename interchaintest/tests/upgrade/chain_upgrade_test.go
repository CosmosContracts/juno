package upgrade_test

import (
	"fmt"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

const (
	upgradeName = "v31"
	baseVersion = "v30.0.0"
	ibcPath     = "ab"
)

var baseChain = ibc.DockerImage{
	Repository: e2esuite.JunoRepo,
	Version:    baseVersion,
	UIDGID:     "1025:1025",
}

type UpgradeTestSuite struct {
	*e2esuite.E2ETestSuite

	eRep *testreporter.RelayerExecReporter
}

func upgradeChainSpecs() []*interchaintest.ChainSpec {
	numValidators := 2
	numFullNodes := 1

	newSpec := func(chainID string) *interchaintest.ChainSpec {
		cfg := e2esuite.DefaultConfig
		cfg.ChainID = chainID
		cfg.Images = []ibc.DockerImage{baseChain}

		return &interchaintest.ChainSpec{
			ChainName:     chainID,
			Name:          "juno",
			NumValidators: &numValidators,
			NumFullNodes:  &numFullNodes,
			Version:       baseVersion,
			NoHostMount:   &e2esuite.DefaultNoHostMount,
			ChainConfig:   cfg,
		}
	}

	return []*interchaintest.ChainSpec{
		newSpec("juno-upgrade-1"),
		newSpec("juno-upgrade-2"),
	}
}

func TestUpgradeTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		upgradeChainSpecs(),
		e2esuite.DefaultTxCfg,
		e2esuite.WithChainConstructor(e2esuite.MultipleChainsConstructor),
		e2esuite.WithInterchainConstructor(e2esuite.TwoChainInterchainConstructor),
	)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	t.Cleanup(func() {
		if s.Relayer != nil {
			if err := s.Relayer.StopRelayer(s.Ctx, eRep); err != nil {
				t.Logf("stopping relayer: %v", err)
			}
		}
		if s.Ic != nil {
			_ = s.Ic.Close()
		}
	})

	testSuite := &UpgradeTestSuite{E2ETestSuite: s, eRep: eRep}
	suite.Run(t, testSuite)
}

func (s *UpgradeTestSuite) TestV31ChainUpgrade() {
	t := s.T()
	require := s.Require()
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	juno, counterparty := s.Chain, s.Chains[1]
	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(1_000_000)))
	user := s.GetAndFundTestUser(t.Name()+"-upgrade", 10_000_000_000, juno)
	feeRecipient := s.GetAndFundTestUser(t.Name()+"-fee-recipient", 10_000_000, juno)
	ibcRecipient := s.GetAndFundTestUser(t.Name()+"-ibc-recipient", 10_000_000, counterparty)

	// Build created the clients, connection and transfer channel while both
	// chains still run v30. Start the relayer before the upgrade so the
	// post-upgrade packet probes continuity of that existing IBC path.
	channel, err := ibc.GetTransferChannel(s.Ctx, s.Relayer, s.eRep, juno.Config().ChainID, counterparty.Config().ChainID)
	require.NoError(err)
	require.NoError(s.Relayer.StartRelayer(s.Ctx, s.eRep, ibcPath))
	require.NoError(testutil.WaitForBlocks(s.Ctx, 2, juno, counterparty))

	// Instantiate and execute an ordinary Wasm contract before the upgrade.
	// Its code, instance and state are queried and mutated again on v31.
	_, wasmContract := s.SetupContract(juno, user.KeyName(), "../../contracts/cw_template.wasm", `{"count":0}`, false, fees)
	_, err = s.ExecuteMsgWithFeeReturn(juno, user, wasmContract, "", `{"increment":{}}`, false, fees)
	require.NoError(err)
	require.Equal(int64(1), s.queryWasmCount(wasmContract))

	// Register a cw-hooks staking contract and prove it receives an event on v30.
	_, hookContract := s.SetupContract(juno, user.KeyName(), "../../contracts/juno_staking_hooks_example.wasm", `{}`, false, fees)
	s.RegisterCwHooksStaking(juno, user, hookContract)
	require.Contains(s.GetCwHooksStakingContracts(), hookContract)

	vals := s.QueryValidators(juno)
	require.NotEmpty(vals, "expected at least one validator")
	valoper := vals[0]
	initialStakeAmt := int64(1_000_000)
	s.StakeTokens(juno, user, valoper.String(), fmt.Sprintf("%d%s", initialStakeAmt, s.Denom), fees, false)
	s.requireDelegationHook(hookContract, user.FormattedAddress(), valoper, initialStakeAmt)

	// Submit the exact v31 software-upgrade name to an exact v30.0.0 base.
	height, err := juno.Height(s.Ctx)
	require.NoError(err, "error fetching height before submit upgrade proposal")
	haltHeight := height + e2esuite.DefaultHaltHeightDelta
	proposalID := s.SubmitSoftwareUpgradeProposal(juno, user, upgradeName, haltHeight, e2esuite.DefaultAuthority)
	proposalIDInt, err := strconv.ParseUint(proposalID, 10, 64)
	require.NoError(err, "failed to parse proposal ID")
	s.ValidatorVoting(juno, proposalIDInt, height, haltHeight)

	// The candidate remains the image selected by the normal ictest CI/local gate.
	repo, version := e2esuite.GetDockerImageInfo()
	s.UpgradeNodes(juno, s.DockerClient, haltHeight, repo, version)

	// Explicit fee execution: a real post-upgrade bank send must execute, credit
	// the recipient exactly, and debit the sender by more than the sent amount.
	feeProbeAmount := math.NewInt(123_456)
	senderBefore, err := juno.GetBalance(s.Ctx, user.FormattedAddress(), s.Denom)
	require.NoError(err)
	recipientBefore, err := juno.GetBalance(s.Ctx, feeRecipient.FormattedAddress(), s.Denom)
	require.NoError(err)
	_, err = s.SendCoins(juno, user.KeyName(), user.FormattedAddress(), feeRecipient.FormattedAddress(), sdk.NewCoins(sdk.NewCoin(s.Denom, feeProbeAmount)), fees)
	require.NoError(err)
	senderAfter, err := juno.GetBalance(s.Ctx, user.FormattedAddress(), s.Denom)
	require.NoError(err)
	recipientAfter, err := juno.GetBalance(s.Ctx, feeRecipient.FormattedAddress(), s.Denom)
	require.NoError(err)
	require.Equal(recipientBefore.Add(feeProbeAmount), recipientAfter, "post-upgrade fee probe must deliver its bank send")
	require.True(senderBefore.Sub(senderAfter).GT(feeProbeAmount), "sender must pay a non-zero fee in addition to the transfer")

	// Feemarket query probes complement the paid execution above.
	feemarketParams := s.QueryFeemarketParams()
	require.True(feemarketParams.Enabled)
	require.True(feemarketParams.MinBaseGasPrice.IsPositive())
	require.Positive(feemarketParams.MaxBlockUtilization)
	require.True(s.QueryFeemarketState().BaseGasPrice.IsPositive())
	gasPrice := s.QueryFeemarketGasPrice(s.Denom)
	require.Equal(s.Denom, gasPrice.Denom)
	require.True(gasPrice.Amount.IsPositive())

	// Wasm code/instance/state continuity: query pre-upgrade state, execute with
	// an explicit fee on v31, then query the incremented value.
	require.Equal(int64(1), s.queryWasmCount(wasmContract))
	_, err = s.ExecuteMsgWithFeeReturn(juno, user, wasmContract, "", `{"increment":{}}`, false, fees)
	require.NoError(err)
	require.Equal(int64(2), s.queryWasmCount(wasmContract))

	// cw-hooks registration and event delivery must both survive. A fresh v31
	// delegation event has cumulative shares distinct from the v30 event.
	require.Contains(s.GetCwHooksStakingContracts(), hookContract)
	cwHooksParams := s.QueryCwHooksParams()
	require.Positive(cwHooksParams.ContractGasLimit)
	additionalStakeAmt := int64(500_000)
	s.StakeTokens(juno, user, valoper.String(), fmt.Sprintf("%d%s", additionalStakeAmt, s.Denom), fees, false)
	s.requireDelegationHook(hookContract, user.FormattedAddress(), valoper, initialStakeAmt+additionalStakeAmt)

	// Query voting-snapshot at a post-upgrade height after the fresh delegation.
	vsParams := s.QueryVotingSnapshotParams()
	require.Positive(vsParams.PruneInterval)
	snapHeight, err := juno.Height(s.Ctx)
	require.NoError(err)
	powerStr := s.QueryVotingPowerAt(user.FormattedAddress(), snapHeight)
	power, ok := math.NewIntFromString(powerStr)
	require.True(ok, "voting power %q should parse as an integer", powerStr)
	require.True(power.IsPositive())
	delegation := s.QueryStakingDelegation(user.FormattedAddress(), valoper.String())
	require.True(power.LTE(delegation.Balance.Amount))
	totalStr := s.QueryTotalVotingPowerAt(snapHeight)
	total, ok := math.NewIntFromString(totalStr)
	require.True(ok, "total voting power %q should parse as an integer", totalStr)
	require.True(total.GTE(power))

	// Send a real ICS-20 packet over the channel created before the upgrade,
	// wait for its acknowledgement, and verify the counterparty voucher balance.
	transferAmount := math.NewInt(77_777)
	ibcDenom := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(
		channel.Counterparty.PortID,
		channel.Counterparty.ChannelID,
		juno.Config().Denom,
	)).IBCDenom()
	ibcBefore, err := counterparty.GetBalance(s.Ctx, ibcRecipient.FormattedAddress(), ibcDenom)
	require.NoError(err)
	transferHeight, err := juno.Height(s.Ctx)
	require.NoError(err)
	transferTx, err := s.SendIBCTransfer(juno, channel.ChannelID, user.KeyName(), ibc.WalletAmount{
		Address: ibcRecipient.FormattedAddress(),
		Denom:   juno.Config().Denom,
		Amount:  transferAmount,
	}, ibc.TransferOptions{})
	require.NoError(err)
	_, err = testutil.PollForAck(s.Ctx, juno, transferHeight, transferHeight+50, transferTx.Packet)
	require.NoError(err, "post-upgrade IBC transfer was not acknowledged")
	ibcAfter, err := counterparty.GetBalance(s.Ctx, ibcRecipient.FormattedAddress(), ibcDenom)
	require.NoError(err)
	require.Equal(ibcBefore.Add(transferAmount), ibcAfter, "counterparty did not receive the post-upgrade IBC voucher")
}

func (s *UpgradeTestSuite) queryWasmCount(contract string) int64 {
	s.T().Helper()
	var res e2esuite.GetCountResponse
	err := s.SmartQueryString(s.Chain, contract, `{"get_count":{}}`, &res)
	s.Require().NoError(err)
	s.Require().NotNil(res.Data)
	return res.Data.Count
}

func (s *UpgradeTestSuite) requireDelegationHook(contract, delegator string, validator sdk.ValAddress, shares int64) {
	s.T().Helper()
	state := s.GetCwStakingHookLastDelegationChange(s.Chain, contract, delegator)
	s.Require().NotNil(state.Data, "cw-hooks contract did not record the delegation event")
	s.Require().Equal(delegator, state.Data.DelegatorAddress)
	s.Require().Equal(validator, sdk.MustValAddressFromBech32(state.Data.ValidatorAddress))
	s.Require().Equal(fmt.Sprintf("%d.000000000000000000", shares), state.Data.Shares)
}
