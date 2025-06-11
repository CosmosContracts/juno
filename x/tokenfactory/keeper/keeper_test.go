package keeper_test

import (
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CosmosContracts/juno/v30/testutil"
	"github.com/CosmosContracts/juno/v30/x/tokenfactory/keeper"
	"github.com/CosmosContracts/juno/v30/x/tokenfactory/types"
)

type KeeperTestSuite struct {
	testutil.KeeperTestHelper

	contractKeeper wasmtypes.ContractOpsKeeper

	queryClient   types.QueryClient
	msgServer     types.MsgServer
	bankMsgServer banktypes.MsgServer

	// defaultDenom is on the suite, as it depends on the creator test address.
	defaultDenom string
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	// Fund every TestAcc with two denoms, one of which is the denom creation fee
	fundAccsAmount := sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(100000000)), sdk.NewCoin("usecond", sdkmath.NewInt(100000000)))
	for _, acc := range s.TestAccs {
		s.FundAcc(acc, fundAccsAmount)
	}

	s.contractKeeper = s.App.AppKeepers.ContractKeeper

	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.msgServer = keeper.NewMsgServerImpl(s.App.AppKeepers.TokenFactoryKeeper)
	s.bankMsgServer = bankkeeper.NewMsgServerImpl(s.App.AppKeepers.BankKeeper)
}

func (s *KeeperTestSuite) CreateDefaultDenom() {
	res, _ := s.msgServer.CreateDenom(s.Ctx, &types.MsgCreateDenom{Sender: s.TestAccs[0].String(), Subdenom: "bitcoin"})
	s.defaultDenom = res.GetNewTokenDenom()
}

func (s *KeeperTestSuite) TestCreateModuleAccount() {
	// setup new next account number
	nextAccountNumber := s.App.AppKeepers.AccountKeeper.NextAccountNumber(s.Ctx)

	// remove module account
	tokenfactoryModuleAccount := s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.App.AppKeepers.AccountKeeper.RemoveAccount(s.Ctx, tokenfactoryModuleAccount)

	// ensure module account was removed
	s.Ctx = s.App.NewContextLegacy(false, cmtproto.Header{})
	tokenfactoryModuleAccount = s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().Nil(tokenfactoryModuleAccount)

	// create module account
	s.App.AppKeepers.TokenFactoryKeeper.CreateModuleAccount(s.Ctx)

	// check that the module account is now initialized
	tokenfactoryModuleAccount = s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().NotNil(tokenfactoryModuleAccount)

	// check that the account number of the module account is now initialized correctly
	s.Require().Equal(nextAccountNumber+1, tokenfactoryModuleAccount.GetAccountNumber())
}
