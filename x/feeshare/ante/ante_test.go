package ante_test

import (
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"
	protov2 "google.golang.org/protobuf/proto"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/CosmosContracts/juno/v29/testutil"
	ante "github.com/CosmosContracts/juno/v29/x/feeshare/ante"
	feesharekeeper "github.com/CosmosContracts/juno/v29/x/feeshare/keeper"
	feesharetypes "github.com/CosmosContracts/juno/v29/x/feeshare/types"
)

// Define an empty ante handle
var (
	EmptyAnte = func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nil
	}
)

type AnteTestSuite struct {
	testutil.KeeperTestHelper

	bankKeeper     bankkeeper.Keeper
	feeshareKeeper feesharekeeper.Keeper
}

func (s *AnteTestSuite) SetupTest() {
	s.Setup()
	s.bankKeeper = s.App.AppKeepers.BankKeeper
	s.feeshareKeeper = s.App.AppKeepers.FeeShareKeeper
}

func TestAnteSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

func (s *AnteTestSuite) TestAnteHandle() {
	s.SetupTest()
	// Mint coins to FeeCollector to cover fees
	s.FundModuleAcc(authtypes.FeeCollectorName, sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(1_000_000))))

	// Create & fund deployer
	_, _, deployer := testdata.KeyTestPubAddr()
	s.FundAcc(deployer, sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(100_000_000))))

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
	s.feeshareKeeper.SetFeeShare(s.Ctx, registerMsg)

	// Create execute msg
	executeMsg := &wasmtypes.MsgExecuteContract{
		Sender:   deployer.String(),
		Contract: contractAddr.String(),
		Msg:      []byte("{}"),
		Funds:    sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(0))),
	}
	tx := NewMockTx(deployer, executeMsg)

	// Run normal msg through ante handle
	anteDecorator := ante.NewFeeSharePayoutDecorator(s.bankKeeper, s.feeshareKeeper)
	_, err := anteDecorator.AnteHandle(s.Ctx, tx, false, EmptyAnte)
	s.Require().NoError(err)

	// Check that the receiver account was paid
	receiverBal := s.bankKeeper.GetBalance(s.Ctx, receiver, "ujuno")
	s.Require().Equal(sdkmath.NewInt(250).Int64(), receiverBal.Amount.Int64())

	// Create & handle authz msg
	authzMsg := authz.NewMsgExec(deployer, []sdk.Msg{executeMsg})
	_, err = anteDecorator.AnteHandle(s.Ctx, NewMockTx(deployer, &authzMsg), false, EmptyAnte)
	s.Require().NoError(err)

	// Check that the receiver account was paid
	receiverBal = s.bankKeeper.GetBalance(s.Ctx, receiver, "ujuno")
	s.Require().Equal(sdkmath.NewInt(500).Int64(), receiverBal.Amount.Int64())

	// Create & handle authz msg with nested authz msg
	nestedAuthzMsg := authz.NewMsgExec(deployer, []sdk.Msg{&authzMsg})
	_, err = anteDecorator.AnteHandle(s.Ctx, NewMockTx(deployer, &nestedAuthzMsg), false, EmptyAnte)
	s.Require().NoError(err)

	// Check that the receiver account was paid
	receiverBal = s.bankKeeper.GetBalance(s.Ctx, receiver, "ujuno")
	s.Require().Equal(sdkmath.NewInt(750).Int64(), receiverBal.Amount.Int64())
}

func (s *AnteTestSuite) TestFeeLogic() {
	s.SetupTest()
	// We expect all to pass
	feeCoins := sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(500)), sdk.NewCoin("utoken", sdkmath.NewInt(250)))

	testCases := []struct {
		name               string
		incomingFee        sdk.Coins
		govPercent         sdkmath.LegacyDec
		numContracts       int
		expectedFeePayment sdk.Coins
	}{
		{
			"100% fee / 1 contract",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(100, 2),
			1,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(500)), sdk.NewCoin("utoken", sdkmath.NewInt(250))),
		},
		{
			"100% fee / 2 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(100, 2),
			2,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(250)), sdk.NewCoin("utoken", sdkmath.NewInt(125))),
		},
		{
			"100% fee / 10 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(100, 2),
			10,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(50)), sdk.NewCoin("utoken", sdkmath.NewInt(25))),
		},
		{
			"67% fee / 7 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(67, 2),
			7,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(48)), sdk.NewCoin("utoken", sdkmath.NewInt(24))),
		},
		{
			"50% fee / 1 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(50, 2),
			1,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(250)), sdk.NewCoin("utoken", sdkmath.NewInt(125))),
		},
		{
			"50% fee / 2 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(50, 2),
			2,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(125)), sdk.NewCoin("utoken", sdkmath.NewInt(62))),
		},
		{
			"50% fee / 3 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(50, 2),
			3,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(83)), sdk.NewCoin("utoken", sdkmath.NewInt(42))),
		},
		{
			"25% fee / 2 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(25, 2),
			2,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(62)), sdk.NewCoin("utoken", sdkmath.NewInt(31))),
		},
		{
			"15% fee / 3 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(15, 2),
			3,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(25)), sdk.NewCoin("utoken", sdkmath.NewInt(12))),
		},
		{
			"1% fee / 2 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(1, 2),
			2,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(2)), sdk.NewCoin("utoken", sdkmath.NewInt(1))),
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

func (MockTx) GetGas() uint64 {
	return 200000
}

func (MockTx) GetFee() sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(500)))
}

func (tx MockTx) FeePayer() []byte {
	return tx.feePayer
}

func (MockTx) FeeGranter() []byte {
	return nil
}

func (tx MockTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

func (MockTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

func (MockTx) ValidateBasic() error {
	return nil
}
