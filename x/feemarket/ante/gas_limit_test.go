package ante_test

import (
	"math"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	protov2 "google.golang.org/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/CosmosContracts/juno/v31/app/ante/decorators"
	feepaytypes "github.com/CosmosContracts/juno/v31/x/feepay/types"
)

type gasLimitFeeTx struct {
	gas        uint64
	fee        sdk.Coins
	feePayer   sdk.AccAddress
	feeGranter sdk.AccAddress
	msgs       []sdk.Msg
}

func (tx gasLimitFeeTx) GetMsgs() []sdk.Msg                 { return tx.msgs }
func (gasLimitFeeTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (tx gasLimitFeeTx) GetGas() uint64                     { return tx.gas }
func (tx gasLimitFeeTx) GetFee() sdk.Coins                  { return tx.fee }
func (tx gasLimitFeeTx) FeePayer() []byte                   { return tx.feePayer }
func (tx gasLimitFeeTx) FeeGranter() []byte                 { return tx.feeGranter }

func (s *AnteTestSuite) gasLimitDecorator() decorators.DeductFeeDecorator {
	return decorators.NewDeductFeeDecorator(
		s.App.AppKeepers.FeePayKeeper,
		*s.App.AppKeepers.FeeMarketKeeper,
		s.App.AppKeepers.AccountKeeper,
		s.App.AppKeepers.BankKeeper,
		s.App.AppKeepers.FeeGrantKeeper,
		"stake",
		nil,
		authante.NewDeductFeeDecorator(
			s.App.AppKeepers.AccountKeeper,
			s.App.AppKeepers.BankKeeper,
			s.App.AppKeepers.FeeGrantKeeper,
			nil,
		),
	)
}

func (s *AnteTestSuite) TestRejectGasLimitAboveMaxInt64() {
	payer := s.fullAccs[0].Account.GetAddress()
	granter := s.fullAccs[1].Account.GetAddress()
	contract := sdk.AccAddress([]byte("gas_limit_feepay_contract")).String()
	s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, feepaytypes.FeePayContract{
		ContractAddress: contract,
		Balance:         math.MaxUint64,
		WalletLimit:     10,
	})

	ordinaryMsg := NewTestMsg(payer)
	ordinaryFee := sdk.NewCoins(sdk.NewInt64Coin("stake", 1))
	feePayMsg := &wasmtypes.MsgExecuteContract{
		Sender:   payer.String(),
		Contract: contract,
		Msg:      []byte("{}"),
	}

	tests := []struct {
		name       string
		fee        sdk.Coins
		feeGranter sdk.AccAddress
		msgs       []sdk.Msg
		simulate   bool
	}{
		{name: "ordinary fee", fee: ordinaryFee, msgs: []sdk.Msg{ordinaryMsg}},
		{name: "ordinary fee simulation", fee: ordinaryFee, msgs: []sdk.Msg{ordinaryMsg}, simulate: true},
		{name: "feegrant", fee: ordinaryFee, feeGranter: granter, msgs: []sdk.Msg{ordinaryMsg}},
		{name: "feegrant simulation", fee: ordinaryFee, feeGranter: granter, msgs: []sdk.Msg{ordinaryMsg}, simulate: true},
		{name: "FeePay", msgs: []sdk.Msg{feePayMsg}},
		{name: "FeePay simulation", msgs: []sdk.Msg{feePayMsg}, simulate: true},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			nextCalled := false
			tx := gasLimitFeeTx{
				gas:        uint64(math.MaxInt64) + 1,
				fee:        tc.fee,
				feePayer:   payer,
				feeGranter: tc.feeGranter,
				msgs:       tc.msgs,
			}

			_, err := s.gasLimitDecorator().AnteHandle(s.Ctx, tx, tc.simulate, func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
				nextCalled = true
				return ctx, nil
			})

			s.Require().ErrorIs(err, sdkerrors.ErrInvalidGasLimit)
			s.Require().ErrorContains(err, "exceeds maximum")
			s.Require().False(nextCalled)
		})
	}
}

func (s *AnteTestSuite) TestMaxInt64GasLimitIsAllowedInSimulation() {
	payer := s.fullAccs[0].Account.GetAddress()
	nextCalled := false
	tx := gasLimitFeeTx{
		gas:      uint64(math.MaxInt64),
		feePayer: payer,
		msgs:     []sdk.Msg{NewTestMsg(payer)},
	}

	_, err := s.gasLimitDecorator().AnteHandle(s.Ctx, tx, true, func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		nextCalled = true
		return ctx, nil
	})

	s.Require().NoError(err)
	s.Require().True(nextCalled)
}
