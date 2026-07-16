package ante_test

import (
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"
	protov2 "google.golang.org/protobuf/proto"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/CosmosContracts/juno/v30/testutil"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	ante "github.com/CosmosContracts/juno/v30/x/feeshare/ante"
	feesharekeeper "github.com/CosmosContracts/juno/v30/x/feeshare/keeper"
	feesharetypes "github.com/CosmosContracts/juno/v30/x/feeshare/types"
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
	// Mint coins to the feemarket fee collector. The v30 DeductFeeDecorator
	// escrows tx fees into feemarkettypes.FeeCollectorName, and the feeshare
	// ante payout reads from that same module account (the post-handler
	// drains it later). The legacy authtypes.FeeCollectorName is empty at
	// ante time under v30 and would fail with "spendable balance 0ujuno".
	s.FundModuleAcc(feemarkettypes.FeeCollectorName, sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(1_000_000))))

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

// TestAnteHandleMultipleWithdrawers asserts that with several feeshare
// recipients in one tx the aggregate payout equals the dev pool exactly and
// never overdraws the escrowed balance (the pre-v30 per-recipient round-up
// could exceed the pool and revert the tx).
func (s *AnteTestSuite) TestAnteHandleMultipleWithdrawers() {
	s.SetupTest()

	// Escrow LESS than the dev pool would claim: fee is 500ujuno, developer
	// share 50% → pool 250. Only 100 is escrowed, so the payout must clamp.
	s.FundModuleAcc(feemarkettypes.FeeCollectorName, sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(100))))

	_, _, deployer := testdata.KeyTestPubAddr()

	numContracts := 3
	receivers := make([]sdk.AccAddress, numContracts)
	msgs := make([]sdk.Msg, numContracts)
	for i := range numContracts {
		_, _, receiver := testdata.KeyTestPubAddr()
		_, _, contractAddr := testdata.KeyTestPubAddr()
		receivers[i] = receiver

		s.feeshareKeeper.SetFeeShare(s.Ctx, feesharetypes.FeeShare{
			ContractAddress:   contractAddr.String(),
			DeployerAddress:   deployer.String(),
			WithdrawerAddress: receiver.String(),
		})

		msgs[i] = &wasmtypes.MsgExecuteContract{
			Sender:   deployer.String(),
			Contract: contractAddr.String(),
			Msg:      []byte("{}"),
		}
	}

	anteDecorator := ante.NewFeeSharePayoutDecorator(s.bankKeeper, s.feeshareKeeper)
	_, err := anteDecorator.AnteHandle(s.Ctx, NewMockTx(deployer, msgs...), false, EmptyAnte)
	s.Require().NoError(err)

	// aggregate paid == clamped pool (100), never more than escrowed
	total := sdkmath.ZeroInt()
	for _, receiver := range receivers {
		total = total.Add(s.bankKeeper.GetBalance(s.Ctx, receiver, "ujuno").Amount)
	}
	s.Require().Equal(sdkmath.NewInt(100).String(), total.String())

	// last recipient absorbed the remainder: 33 / 33 / 34
	s.Require().Equal(int64(33), s.bankKeeper.GetBalance(s.Ctx, receivers[0], "ujuno").Amount.Int64())
	s.Require().Equal(int64(33), s.bankKeeper.GetBalance(s.Ctx, receivers[1], "ujuno").Amount.Int64())
	s.Require().Equal(int64(34), s.bankKeeper.GetBalance(s.Ctx, receivers[2], "ujuno").Amount.Int64())

	// module account fully drained but not overdrawn
	collectorAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(feemarkettypes.FeeCollectorName)
	s.Require().True(s.bankKeeper.GetBalance(s.Ctx, collectorAddr, "ujuno").IsZero())
}

func (s *AnteTestSuite) TestFeeLogic() {
	s.SetupTest()
	// We expect all to pass
	feeCoins := sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(500)), sdk.NewCoin("utoken", sdkmath.NewInt(250)))

	testCases := []struct {
		name         string
		incomingFee  sdk.Coins
		govPercent   sdkmath.LegacyDec
		numContracts int
		expectedPool sdk.Coins
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
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(500)), sdk.NewCoin("utoken", sdkmath.NewInt(250))),
		},
		{
			"100% fee / 10 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(100, 2),
			10,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(500)), sdk.NewCoin("utoken", sdkmath.NewInt(250))),
		},
		{
			"67% fee / 7 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(67, 2),
			7,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(335)), sdk.NewCoin("utoken", sdkmath.NewInt(168))),
		},
		{
			"50% fee / 1 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(50, 2),
			1,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(250)), sdk.NewCoin("utoken", sdkmath.NewInt(125))),
		},
		{
			"50% fee / 3 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(50, 2),
			3,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(250)), sdk.NewCoin("utoken", sdkmath.NewInt(125))),
		},
		{
			// 1% of 250utoken = 2.5 → banker's rounding → 2
			"1% fee / 2 contracts",
			feeCoins,
			sdkmath.LegacyNewDecWithPrec(1, 2),
			2,
			sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(5)), sdk.NewCoin("utoken", sdkmath.NewInt(2))),
		},
	}

	for _, tc := range testCases {
		pool := ante.CalculateFeeSharePool(tc.incomingFee, tc.govPercent)
		s.Require().Equal(tc.expectedPool.String(), pool.String(), tc.name)

		splits := ante.SplitFeeSharePool(pool, tc.numContracts)
		s.Require().Len(splits, tc.numContracts, tc.name)

		// The aggregate paid out must equal the pool EXACTLY — the pre-v30
		// per-recipient RoundInt could overshoot the pool and overdraw the
		// escrow.
		var total sdk.Coins
		for _, split := range splits {
			total = total.Add(split...)
		}
		s.Require().Equal(pool.String(), total.String(), tc.name)

		// Every recipient except the last receives the truncated even share;
		// the last additionally absorbs the remainder.
		for _, c := range pool {
			share := c.Amount.QuoRaw(int64(tc.numContracts))
			for i := range tc.numContracts - 1 {
				s.Require().Equal(share.String(), splits[i].AmountOf(c.Denom).String(), tc.name)
			}
			last := splits[tc.numContracts-1].AmountOf(c.Denom)
			s.Require().Equal(c.Amount.String(), share.MulRaw(int64(tc.numContracts-1)).Add(last).String(), tc.name)
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
