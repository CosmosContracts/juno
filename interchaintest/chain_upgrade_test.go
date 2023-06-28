package interchaintest

import (
	"context"
	"fmt"
	"testing"
	"time"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const (
	haltHeightDelta    = uint64(10) // will propose upgrade this many blocks in the future
	blocksAfterUpgrade = uint64(7)
)

func TestBasicJunoUpgrade(t *testing.T) {
	repo, version := GetDockerImageInfo()
	startVersion := "v15.0.0"
	upgradeName := "v16"
	CosmosChainUpgradeTest(t, "juno", startVersion, version, repo, upgradeName)
}

func CosmosChainUpgradeTest(t *testing.T, chainName, initialVersion, upgradeBranchVersion, upgradeRepo, upgradeName string) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	t.Log(chainName, initialVersion, upgradeBranchVersion, upgradeRepo, upgradeName)

	// v45 genesis params
	genesisKVs := []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.voting_params.voting_period",
			Value: VotingPeriod,
		},
		{
			Key:   "app_state.gov.deposit_params.max_deposit_period",
			Value: MaxDepositPeriod,
		},
		{
			Key:   "app_state.gov.deposit_params.min_deposit.0.denom",
			Value: Denom,
		},
	}

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:      chainName,
			ChainName: chainName,
			Version:   initialVersion,
			ChainConfig: ibc.ChainConfig{
				Images: []ibc.DockerImage{
					{
						Repository: JunoE2ERepo,
						Version:    initialVersion,
						UidGid:     JunoImage.UidGid,
					},
				},
				GasPrices:     fmt.Sprintf("0%s", Denom),
				ModifyGenesis: cosmos.ModifyGenesis(genesisKVs),
			},
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chain := chains[0].(*cosmos.CosmosChain)

	ic := interchaintest.NewInterchain().
		AddChain(chain)

	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	err = ic.Build(ctx, nil, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = ic.Close()
	})

	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	chainUser := users[0]

	// create a tokenfactory denom before upgrade (invalid genesis for hard forking due to x/bank validation)
	emptyFullDenom := helpers.CreateTokenFactoryDenom(t, ctx, chain, chainUser, "empty", "")

	mintedDenom := helpers.CreateTokenFactoryDenom(t, ctx, chain, chainUser, "minted", "")
	helpers.MintToTokenFactoryDenom(t, ctx, chain, chainUser, chainUser, 100, mintedDenom)

	mintedAndModified := helpers.CreateTokenFactoryDenom(t, ctx, chain, chainUser, "mandm", "")
	helpers.MintToTokenFactoryDenom(t, ctx, chain, chainUser, chainUser, 100, mintedAndModified)

	ticker, desc, exponent := "TICKER", "desc", "6"
	helpers.UpdateTokenFactoryMetadata(t, ctx, chain, chainUser, mintedAndModified, ticker, desc, exponent)

	// Validate pre upgrade denoms do not have the proper metadata
	// metadata:<denom_units:<denom:"factory/juno1hql0qadnznq8skf5q2psqmwj4thl2ajnvr3qrx/empty" > base:"factory/juno1hql0qadnznq8skf5q2psqmwj4thl2ajnvr3qrx/empty" >
	res := helpers.GetTokenFactoryDenomMetadata(t, ctx, chain, emptyFullDenom)
	require.Equal(t, res.DenomUnits[0].Denom, emptyFullDenom)
	require.Equal(t, res.Base, emptyFullDenom)
	require.Empty(t, res.Description)
	require.Empty(t, res.Display)
	require.Empty(t, res.Name)
	require.Empty(t, res.Symbol)

	res = helpers.GetTokenFactoryDenomMetadata(t, ctx, chain, mintedDenom)
	require.Equal(t, res.DenomUnits[0].Denom, mintedDenom)
	require.Equal(t, res.Base, mintedDenom)
	require.Empty(t, res.Description)
	require.Empty(t, res.Display)
	require.Empty(t, res.Name)
	require.Empty(t, res.Symbol)

	// Denom data should be as modified above
	modifiedRes := helpers.GetTokenFactoryDenomMetadata(t, ctx, chain, mintedAndModified)
	require.Equal(t, modifiedRes.DenomUnits[0].Denom, mintedAndModified)
	require.Equal(t, modifiedRes.Base, mintedAndModified)
	require.Equal(t, modifiedRes.Name, mintedAndModified)
	require.Equal(t, modifiedRes.Symbol, ticker)
	require.NotEmpty(t, modifiedRes.Description)

	// upgrade
	height, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := height + haltHeightDelta

	proposal := cosmos.SoftwareUpgradeProposal{
		Deposit:     "500000000" + chain.Config().Denom, // greater than min deposit
		Title:       "Chain Upgrade 1",
		Name:        upgradeName,
		Description: "First chain software upgrade",
		Height:      haltHeight,
	}

	upgradeTx, err := chain.UpgradeProposal(ctx, chainUser.KeyName(), proposal)
	require.NoError(t, err, "error submitting software upgrade proposal tx")

	err = chain.VoteOnProposalAllValidators(ctx, upgradeTx.ProposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, chain, height, height+haltHeightDelta, upgradeTx.ProposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	// this should timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height), chain)

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")

	// make sure that chain is halted
	require.Equal(t, haltHeight, height, "height is not equal to halt height")

	// bring down nodes to prepare for upgrade
	t.Log("stopping node(s)")
	err = chain.StopAllNodes(ctx)
	require.NoError(t, err, "error stopping node(s)")

	// upgrade version on all nodes
	t.Log("upgrading node(s)")
	chain.UpgradeVersion(ctx, client, upgradeRepo, upgradeBranchVersion)

	// start all nodes back up.
	// validators reach consensus on first block after upgrade height
	// and chain block production resumes.
	t.Log("starting node(s)")
	err = chain.StartAllNodes(ctx)
	require.NoError(t, err, "error starting upgraded node(s)")

	timeoutCtx, timeoutCtxCancel = context.WithTimeout(ctx, time.Second*60)
	defer timeoutCtxCancel()

	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), chain)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	height, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height after upgrade")

	require.GreaterOrEqual(t, height, haltHeight+blocksAfterUpgrade, "height did not increment enough after upgrade")

	// Check that the tokenfactory denom's properly migrated
	postRes := helpers.GetTokenFactoryDenomMetadata(t, ctx, chain, emptyFullDenom)
	require.Equal(t, postRes.DenomUnits[0].Denom, emptyFullDenom)
	require.Equal(t, postRes.Base, emptyFullDenom)
	require.Equal(t, postRes.Display, emptyFullDenom)
	require.Equal(t, postRes.Name, emptyFullDenom)
	require.Equal(t, postRes.Symbol, emptyFullDenom)

	postRes = helpers.GetTokenFactoryDenomMetadata(t, ctx, chain, mintedDenom)
	require.Equal(t, postRes.DenomUnits[0].Denom, mintedDenom)
	require.Equal(t, postRes.Base, mintedDenom)
	require.Equal(t, postRes.Display, mintedDenom)
	require.Equal(t, postRes.Name, mintedDenom)
	require.Equal(t, postRes.Symbol, mintedDenom)

	// since we already set it, the should remain the same.
	postModified := helpers.GetTokenFactoryDenomMetadata(t, ctx, chain, mintedAndModified)
	require.Equal(t, postModified, modifiedRes)

	// Ensure after the upgrade, the denoms are properly set with the Denom Metadata.
	// (Due to migrating hardcoded, we have to set a fee after the upgrade).
	afterUpgrade := helpers.CreateTokenFactoryDenom(t, ctx, chain, chainUser, "post", "250000"+Denom)
	newRes := helpers.GetTokenFactoryDenomMetadata(t, ctx, chain, afterUpgrade)
	require.Equal(t, newRes.Display, afterUpgrade)
	require.Equal(t, newRes.Name, afterUpgrade)
	require.Equal(t, newRes.Symbol, afterUpgrade)
}
