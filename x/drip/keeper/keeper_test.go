package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/CosmosContracts/juno/v27/app"
	"github.com/CosmosContracts/juno/v27/testutil"
	"github.com/CosmosContracts/juno/v27/x/drip/keeper"
	"github.com/CosmosContracts/juno/v27/x/drip/types"
	minttypes "github.com/CosmosContracts/juno/v27/x/mint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	app     *app.App
	genesis types.GenesisState

	bankKeeper bankkeeper.Keeper

	queryClient   types.QueryClient
	dripMsgServer types.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	s.app = testutil.Setup(isCheckTx, s.T())
	s.ctx = s.app.BaseApp.NewContext(isCheckTx)
	s.genesis = *types.DefaultGenesisState()

	s.bankKeeper = s.app.AppKeepers.BankKeeper

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(s.app.AppKeepers.DripKeeper))
	s.queryClient = types.NewQueryClient(queryHelper)
	s.dripMsgServer = keeper.NewMsgServerImpl(s.app.AppKeepers.DripKeeper)
}

func (s *KeeperTestSuite) FundAccount(ctx context.Context, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := s.bankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
		return err
	}

	return s.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, amounts)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
