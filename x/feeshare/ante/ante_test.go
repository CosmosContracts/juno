package ante_test

import (
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"
	protov2 "google.golang.org/protobuf/proto"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/CosmosContracts/juno/v18/app"
	ante "github.com/CosmosContracts/juno/v18/x/feeshare/ante"
	feesharekeeper "github.com/CosmosContracts/juno/v18/x/feeshare/keeper"
	feesharetypes "github.com/CosmosContracts/juno/v18/x/feeshare/types"
)

// Define an empty ante handle
var (
	EmptyAnte = func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}
)

type AnteTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	app            *app.App
	bankKeeper     bankkeeper.Keeper
	feeshareKeeper feesharekeeper.Keeper
}

func (s *AnteTestSuite) SetupTest() {
	isCheckTx := false
	s.app = app.Setup(s.T())

	s.ctx = s.app.BaseApp.NewContext(isCheckTx, tmproto.Header{
		ChainID: "testing",
		Height:  10,
		Time:    time.Now().UTC(),
	})

	s.bankKeeper = s.app.AppKeepers.BankKeeper
	s.feeshareKeeper = s.app.AppKeepers.FeeShareKeeper
}

func TestAnteSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

func (s *AnteTestSuite) TestAnteHandle() {
	// Mint coins to FeeCollector to cover fees
	err := s.FundModule(s.ctx, authtypes.FeeCollectorName, sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(1_000_000))))
	s.Require().NoError(err)

	// Create & fund deployer
	_, _, deployer := testdata.KeyTestPubAddr()
	err = s.FundAccount(s.ctx, deployer, sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(100_000_000))))
	s.Require().NoError(err)

	// Create funds receiver account
	_, _, receiver := testdata.KeyTestPubAddr()

	// Address used to mock a contract
	_, _, contractAddr := testdata.KeyTestPubAddr()

	// Register contract with Fee Share
	registerMsg := feesharetypes.FeeShare{
		ContractAddress:   contractAddr.String(),
		DeployerAddress:   deployer.String(),
		WithdrawerAddress: receiver.String(),
	}
	s.feeshareKeeper.SetFeeShare(s.ctx, registerMsg)

	// Create execute msg
	executeMsg := &wasmtypes.MsgExecuteContract{
		Sender:   deployer.String(),
		Contract: contractAddr.String(),
		Msg:      []byte("{}"),
		Funds:    sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(0))),
	}
	tx := NewMockTx(deployer, executeMsg)

	// Run normal msg through ante handle
	ante := ante.NewFeeSharePayoutDecorator(s.bankKeeper, s.feeshareKeeper)
	_, err = ante.AnteHandle(s.ctx, tx, false, EmptyAnte)
	s.Require().NoError(err)

	// Check that the receiver account was paid
	receiverBal := s.bankKeeper.GetBalance(s.ctx, receiver, "ujuno")
	s.Require().Equal(sdk.NewInt(250).Int64(), receiverBal.Amount.Int64())

	// Create & handle authz msg
	authzMsg := authz.NewMsgExec(deployer, []sdk.Msg{executeMsg})
	_, err = ante.AnteHandle(s.ctx, NewMockTx(deployer, &authzMsg), false, EmptyAnte)
	s.Require().NoError(err)

	// Check that the receiver account was paid
	receiverBal = s.bankKeeper.GetBalance(s.ctx, receiver, "ujuno")
	s.Require().Equal(sdk.NewInt(500).Int64(), receiverBal.Amount.Int64())

	// Create & handle authz msg with nested authz msg
	nestedAuthzMsg := authz.NewMsgExec(deployer, []sdk.Msg{&authzMsg})
	_, err = ante.AnteHandle(s.ctx, NewMockTx(deployer, &nestedAuthzMsg), false, EmptyAnte)
	s.Require().NoError(err)

	// Check that the receiver account was paid
	receiverBal = s.bankKeeper.GetBalance(s.ctx, receiver, "ujuno")
	s.Require().Equal(sdk.NewInt(750).Int64(), receiverBal.Amount.Int64())
}

func (s *AnteTestSuite) TestFeeLogic() {
	// We expect all to pass
	feeCoins := sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(500)), sdk.NewCoin("utoken", sdk.NewInt(250)))

	testCases := []struct {
		name               string
		incomingFee        sdk.Coins
		govPercent         sdk.Dec
		numContracts       int
		expectedFeePayment sdk.Coins
	}{
		{
			"100% fee / 1 contract",
			feeCoins,
			sdk.NewDecWithPrec(100, 2),
			1,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(500)), sdk.NewCoin("utoken", sdk.NewInt(250))),
		},
		{
			"100% fee / 2 contracts",
			feeCoins,
			sdk.NewDecWithPrec(100, 2),
			2,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(250)), sdk.NewCoin("utoken", sdk.NewInt(125))),
		},
		{
			"100% fee / 10 contracts",
			feeCoins,
			sdk.NewDecWithPrec(100, 2),
			10,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(50)), sdk.NewCoin("utoken", sdk.NewInt(25))),
		},
		{
			"67% fee / 7 contracts",
			feeCoins,
			sdk.NewDecWithPrec(67, 2),
			7,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(48)), sdk.NewCoin("utoken", sdk.NewInt(24))),
		},
		{
			"50% fee / 1 contracts",
			feeCoins,
			sdk.NewDecWithPrec(50, 2),
			1,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(250)), sdk.NewCoin("utoken", sdk.NewInt(125))),
		},
		{
			"50% fee / 2 contracts",
			feeCoins,
			sdk.NewDecWithPrec(50, 2),
			2,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(125)), sdk.NewCoin("utoken", sdk.NewInt(62))),
		},
		{
			"50% fee / 3 contracts",
			feeCoins,
			sdk.NewDecWithPrec(50, 2),
			3,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(83)), sdk.NewCoin("utoken", sdk.NewInt(42))),
		},
		{
			"25% fee / 2 contracts",
			feeCoins,
			sdk.NewDecWithPrec(25, 2),
			2,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(62)), sdk.NewCoin("utoken", sdk.NewInt(31))),
		},
		{
			"15% fee / 3 contracts",
			feeCoins,
			sdk.NewDecWithPrec(15, 2),
			3,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(25)), sdk.NewCoin("utoken", sdk.NewInt(12))),
		},
		{
			"1% fee / 2 contracts",
			feeCoins,
			sdk.NewDecWithPrec(1, 2),
			2,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(2)), sdk.NewCoin("utoken", sdk.NewInt(1))),
		},
	}

	for _, tc := range testCases {
		coins := ante.FeePayLogic(tc.incomingFee, tc.govPercent, tc.numContracts)

		for _, coin := range coins {
			for _, expectedCoin := range tc.expectedFeePayment {
				if coin.Denom == expectedCoin.Denom {
					s.Require().Equal(expectedCoin.Amount.Int64(), coin.Amount.Int64(), tc.name)
				}
			}
		}
	}
}

func (s *AnteTestSuite) FundAccount(ctx sdk.Context, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := s.bankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
		return err
	}

	return s.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, amounts)
}

func (s *AnteTestSuite) FundModule(ctx sdk.Context, moduleName string, amounts sdk.Coins) error {
	if err := s.bankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
		return err
	}

	return s.bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, moduleName, amounts)
}

type MockTx struct {
	feePayer sdk.AccAddress
	msgs     []sdk.Msg
}

func NewMockTx(feePayer sdk.AccAddress, msgs ...sdk.Msg) MockTx {
	return MockTx{
		feePayer: feePayer,
		msgs:     msgs,
	}
}

func (tx MockTx) GetGas() uint64 {
	return 200000
}

func (tx MockTx) GetFee() sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(500)))
}

func (tx MockTx) FeePayer() sdk.AccAddress {
	return tx.feePayer
}

func (tx MockTx) FeeGranter() sdk.AccAddress {
	return nil
}

func (tx MockTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

func (tx MockTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

func (tx MockTx) ValidateBasic() error {
	return nil
}
