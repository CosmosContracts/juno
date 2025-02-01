package keeper_test

import (
	"context"
	"crypto/sha256"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"

	_ "embed"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v27/app"
	"github.com/CosmosContracts/juno/v27/testutil"
	"github.com/CosmosContracts/juno/v27/x/clock/keeper"
	"github.com/CosmosContracts/juno/v27/x/clock/types"
	minttypes "github.com/CosmosContracts/juno/v27/x/mint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.App

	queryClient    types.QueryClient
	clockMsgServer types.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	s.app = testutil.Setup(isCheckTx, s.T())
	s.ctx = s.app.BaseApp.NewContext(false)

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(s.app.AppKeepers.ClockKeeper))
	s.queryClient = types.NewQueryClient(queryHelper)
	s.clockMsgServer = keeper.NewMsgServerImpl(s.app.AppKeepers.ClockKeeper)
}

func (s *KeeperTestSuite) FundAccount(ctx context.Context, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := s.app.AppKeepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
		return err
	}

	return s.app.AppKeepers.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, amounts)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

//go:embed testdata/clock_example.wasm
var wasmContract []byte

func (s *KeeperTestSuite) StoreCode() {
	_, _, sender := testdata.KeyTestPubAddr()
	msg := wasmtypes.MsgStoreCodeFixture(func(m *wasmtypes.MsgStoreCode) {
		m.WASMByteCode = wasmContract
		m.Sender = sender.String()
	})
	rsp, err := s.app.MsgServiceRouter().Handler(msg)(s.ctx, msg)
	s.Require().NoError(err)
	var result wasmtypes.MsgStoreCodeResponse
	s.Require().NoError(s.app.AppCodec().Unmarshal(rsp.Data, &result))
	s.Require().Equal(uint64(1), result.CodeID)
	expHash := sha256.Sum256(wasmContract)
	s.Require().Equal(expHash[:], result.Checksum)
	// and
	info := s.app.AppKeepers.WasmKeeper.GetCodeInfo(s.ctx, 1)
	s.Require().NotNil(info)
	s.Require().Equal(expHash[:], info.CodeHash)
	s.Require().Equal(sender.String(), info.Creator)
	s.Require().Equal(wasmtypes.DefaultParams().InstantiateDefaultPermission.With(sender), info.InstantiateConfig)
}

func (s *KeeperTestSuite) InstantiateContract(sender string, admin string) string {
	msgStoreCode := wasmtypes.MsgStoreCodeFixture(func(m *wasmtypes.MsgStoreCode) {
		m.WASMByteCode = wasmContract
		m.Sender = sender
	})
	_, err := s.app.MsgServiceRouter().Handler(msgStoreCode)(s.ctx, msgStoreCode)
	s.Require().NoError(err)

	msgInstantiate := wasmtypes.MsgInstantiateContractFixture(func(m *wasmtypes.MsgInstantiateContract) {
		m.Sender = sender
		m.Admin = admin
		m.Msg = []byte(`{}`)
	})
	resp, err := s.app.MsgServiceRouter().Handler(msgInstantiate)(s.ctx, msgInstantiate)
	s.Require().NoError(err)
	var result wasmtypes.MsgInstantiateContractResponse
	s.Require().NoError(s.app.AppCodec().Unmarshal(resp.Data, &result))
	contractInfo := s.app.AppKeepers.WasmKeeper.GetContractInfo(s.ctx, sdk.MustAccAddressFromBech32(result.Address))
	s.Require().Equal(contractInfo.CodeID, uint64(1))
	s.Require().Equal(contractInfo.Admin, admin)
	s.Require().Equal(contractInfo.Creator, sender)

	return result.Address
}

// Helper method for quickly registering a clock contract
func (s *KeeperTestSuite) RegisterClockContract(senderAddress string, contractAddress string) {
	err := s.app.AppKeepers.ClockKeeper.RegisterContract(s.ctx, senderAddress, contractAddress)
	s.Require().NoError(err)
}

// Helper method for quickly unregistering a clock contract
func (s *KeeperTestSuite) UnregisterClockContract(senderAddress string, contractAddress string) {
	err := s.app.AppKeepers.ClockKeeper.UnregisterContract(s.ctx, senderAddress, contractAddress)
	s.Require().NoError(err)
}

// Helper method for quickly jailing a clock contract
func (s *KeeperTestSuite) JailClockContract(contractAddress string) {
	err := s.app.AppKeepers.ClockKeeper.SetJailStatus(s.ctx, contractAddress, true)
	s.Require().NoError(err)
}

// Helper method for quickly unjailing a clock contract
func (s *KeeperTestSuite) UnjailClockContract(senderAddress string, contractAddress string) {
	err := s.app.AppKeepers.ClockKeeper.SetJailStatusBySender(s.ctx, senderAddress, contractAddress, false)
	s.Require().NoError(err)
}
