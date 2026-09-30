package helpers_test

import (
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/suite"
	protov2 "google.golang.org/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v31/testutil"
	"github.com/CosmosContracts/juno/v31/x/feepay/helpers"
	"github.com/CosmosContracts/juno/v31/x/feepay/types"
)

type HelpersTestSuite struct {
	testutil.KeeperTestHelper
}

func TestHelpersTestSuite(t *testing.T) {
	suite.Run(t, new(HelpersTestSuite))
}

// mockFeeTx is a minimal sdk.FeeTx for driving IsValidFeePayTransaction.
type mockFeeTx struct {
	fee  sdk.Coins
	msgs []sdk.Msg
}

func (tx mockFeeTx) GetMsgs() []sdk.Msg                 { return tx.msgs }
func (mockFeeTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (mockFeeTx) GetGas() uint64                        { return 200000 }
func (tx mockFeeTx) GetFee() sdk.Coins                  { return tx.fee }
func (mockFeeTx) FeePayer() []byte                      { return nil }
func (mockFeeTx) FeeGranter() []byte                    { return nil }

func (s *HelpersTestSuite) TestIsValidFeePayTransaction() {
	s.Setup()

	keeper := s.App.AppKeepers.FeePayKeeper

	// enable feepay
	s.Require().NoError(keeper.SetParams(s.Ctx, types.Params{EnableFeepay: true}))

	// register a contract (no wasm validation on this genesis-path setter)
	registered := sdk.AccAddress([]byte("feepay_registered_addr")).String()
	keeper.SetFeePayContract(s.Ctx, types.FeePayContract{
		ContractAddress: registered,
		Balance:         1_000_000,
		WalletLimit:     10,
	})

	sender := sdk.AccAddress([]byte("feepay_test_sender_addr")).String()
	execRegistered := &wasmtypes.MsgExecuteContract{Sender: sender, Contract: registered, Msg: []byte("{}")}
	execUnregistered := &wasmtypes.MsgExecuteContract{Sender: sender, Contract: sdk.AccAddress([]byte("unregistered_contract")).String(), Msg: []byte("{}")}

	testCases := []struct {
		name  string
		tx    mockFeeTx
		valid bool
	}{
		{
			name:  "valid: single msg on registered contract, zero fee",
			tx:    mockFeeTx{msgs: []sdk.Msg{execRegistered}},
			valid: true,
		},
		{
			name: "invalid: two messages, even if both are registered contracts",
			// per-tx fee model: only ONE contract is charged/rate-limited per
			// tx, so a multi-message feepay tx is a usage-limit bypass
			tx:    mockFeeTx{msgs: []sdk.Msg{execRegistered, execRegistered}},
			valid: false,
		},
		{
			name:  "invalid: zero messages",
			tx:    mockFeeTx{msgs: []sdk.Msg{}},
			valid: false,
		},
		{
			name:  "invalid: non-zero fee",
			tx:    mockFeeTx{fee: sdk.NewCoins(sdk.NewInt64Coin("ujuno", 1)), msgs: []sdk.Msg{execRegistered}},
			valid: false,
		},
		{
			name:  "invalid: unregistered contract",
			tx:    mockFeeTx{msgs: []sdk.Msg{execUnregistered}},
			valid: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.valid, helpers.IsValidFeePayTransaction(s.Ctx, keeper, tc.tx))
		})
	}

	s.Run("invalid: module disabled", func() {
		s.Require().NoError(keeper.SetParams(s.Ctx, types.Params{EnableFeepay: false}))
		s.Require().False(helpers.IsValidFeePayTransaction(s.Ctx, keeper, mockFeeTx{msgs: []sdk.Msg{execRegistered}}))
	})
}
