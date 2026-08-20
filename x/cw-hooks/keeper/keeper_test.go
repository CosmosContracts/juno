package keeper_test

import (
	"embed"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v31/testutil"
	"github.com/CosmosContracts/juno/v31/x/cw-hooks/keeper"
	"github.com/CosmosContracts/juno/v31/x/cw-hooks/types"
)

var _ = embed.FS{}

//go:embed testdata/juno_staking_hooks_example.wasm
var wasmContract []byte

type KeeperTestSuite struct {
	testutil.KeeperTestHelper

	bankKeeper    bankkeeper.Keeper
	stakingKeeper stakingkeeper.Keeper
	wasmKeeper    wasmkeeper.Keeper

	queryClient   types.QueryClient
	msgServer     types.MsgServer
	wasmMsgServer wasmtypes.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	s.bankKeeper = s.App.AppKeepers.BankKeeper
	s.stakingKeeper = *s.App.AppKeepers.StakingKeeper
	s.wasmKeeper = s.App.AppKeepers.WasmKeeper

	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.msgServer = keeper.NewMsgServerImpl(s.App.AppKeepers.CWHooksKeeper)
	s.wasmMsgServer = wasmkeeper.NewMsgServerImpl(&s.wasmKeeper)
}

func (s *KeeperTestSuite) TestContractFailureNonBlocking() {
	s.SetupTest()

	_, _, sender := testdata.KeyTestPubAddr()

	s.FundAcc(sender, sdk.NewCoins(
		sdk.NewCoin("stake", sdkmath.NewInt(1_000_000_000)),
		sdk.NewCoin("ujuno", sdkmath.NewInt(1_000_000_000)),
	))

	failingContract := s.InstantiateContract(sender.String(), "", wasmContract)

	s.Require().NoError(s.registerContract("staking", sender.String(), failingContract))

	var err error
	var info types.ContractInfo
	err = s.App.AppKeepers.CWHooksKeeper.ExecuteMessageOnContracts(s.Ctx, types.StakingPrefixKey, []byte("{}"))
	s.Require().NoError(err)

	info = s.getContractInfo(failingContract)
	s.Require().EqualValues(1, info.FailureCounter)
	s.Require().NotEmpty(info.LatestError)

	vals, err := s.stakingKeeper.GetValidators(s.Ctx, 1)
	s.Require().NoError(err)
	s.Require().NotEmpty(vals)

	_, err = s.stakingKeeper.Delegate(s.Ctx, sender, sdkmath.NewInt(1), stakingtypes.Bonded, vals[0], false)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestUnregisterContractsOnFailure() {
	s.SetupTest()

	_, _, sender := testdata.KeyTestPubAddr()

	s.FundAcc(sender, sdk.NewCoins(
		sdk.NewCoin("stake", sdkmath.NewInt(1_000_000_000)),
		sdk.NewCoin("ujuno", sdkmath.NewInt(1_000_000_000)),
	))

	failingContract := s.InstantiateContract(sender.String(), "", wasmContract)

	s.Require().NoError(s.registerContract("staking", sender.String(), failingContract))

	var err error
	var info types.ContractInfo
	for i := uint64(0); i < 3; i++ {
		err = s.App.AppKeepers.CWHooksKeeper.ExecuteMessageOnContracts(s.Ctx, types.StakingPrefixKey, []byte("{}"))
		s.Require().NoError(err)
	}

	isRegistered, err := s.App.AppKeepers.CWHooksKeeper.IsContractRegistered(s.Ctx, types.StakingPrefixKey, sdk.MustAccAddressFromBech32(failingContract))
	s.Require().NoError(err)
	s.Require().False(isRegistered)

	err = s.App.AppKeepers.CWHooksKeeper.ExecuteMessageOnContracts(s.Ctx, types.StakingPrefixKey, []byte("{}"))
	s.Require().NoError(err)

	s.Require().NoError(s.registerContract("staking", sender.String(), failingContract))

	err = s.App.AppKeepers.CWHooksKeeper.ExecuteMessageOnContracts(s.Ctx, types.StakingPrefixKey, []byte("{}"))
	s.Require().NoError(err)
	info = s.getContractInfo(failingContract)
	s.Require().EqualValues(1, info.FailureCounter)
	s.Require().NotEmpty(info.LatestError)
	isRegistered, err = s.App.AppKeepers.CWHooksKeeper.IsContractRegistered(s.Ctx, types.StakingPrefixKey, sdk.MustAccAddressFromBech32(failingContract))
	s.Require().NoError(err)
	s.Require().True(isRegistered)
}

func (s *KeeperTestSuite) registerContract(module, sender, contractAddr string) error {
	_, err := s.msgServer.RegisterContract(s.Ctx, &types.MsgRegisterContract{
		Module:          module,
		SenderAddress:   sender,
		ContractAddress: contractAddr,
	})
	return err
}

func (s *KeeperTestSuite) unregisterContract(module, sender, contractAddr string) error {
	_, err := s.msgServer.UnregisterContract(s.Ctx, &types.MsgUnregisterContract{
		Module:          module,
		SenderAddress:   sender,
		ContractAddress: contractAddr,
	})
	return err
}

func (s *KeeperTestSuite) getContractInfo(contractAddr string) types.ContractInfo {
	prefix, err := types.ModulePrefixFromModule("staking")
	s.Require().NoError(err)

	info, err := s.App.AppKeepers.CWHooksKeeper.Contracts.Get(
		s.Ctx,
		types.BuildContractPrimaryKey(prefix, sdk.MustAccAddressFromBech32(contractAddr)),
	)
	s.Require().NoError(err)
	return info
}

type contractTestContext struct {
	sender            sdk.AccAddress
	notAuthorized     sdk.AccAddress
	contract          string
	contractWithAdmin string
	dao               string
	daoChild          string
}

func (s *KeeperTestSuite) buildContractTestContext() contractTestContext {
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, notAuthorized := testdata.KeyTestPubAddr()

	funds := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000)))
	s.FundAcc(sender, funds)
	s.FundAcc(notAuthorized, funds)

	contract := s.InstantiateContract(sender.String(), "", wasmContract)
	contractWithAdmin := s.InstantiateContract(notAuthorized.String(), sender.String(), wasmContract)
	dao := s.InstantiateContract(sender.String(), "", wasmContract)
	daoChild := s.InstantiateContract(dao, dao, wasmContract)

	return contractTestContext{
		sender:            sender,
		notAuthorized:     notAuthorized,
		contract:          contract,
		contractWithAdmin: contractWithAdmin,
		dao:               dao,
		daoChild:          daoChild,
	}
}
