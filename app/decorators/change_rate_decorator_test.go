package decorators_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/CosmosContracts/juno/v18/app"
	"github.com/stretchr/testify/suite"

	decorators "github.com/CosmosContracts/juno/v18/app/decorators"
	appparams "github.com/CosmosContracts/juno/v18/app/params"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	protov2 "google.golang.org/protobuf/proto"
)

// Define an empty ante handle
var (
	EmptyAnte = func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}
)

type AnteTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	app           *app.App
	stakingKeeper *stakingkeeper.Keeper
}

func (s *AnteTestSuite) SetupTest() {
	isCheckTx := false
	s.app = app.Setup(s.T())

	s.ctx = s.app.BaseApp.NewContext(isCheckTx, tmproto.Header{
		ChainID: "testing",
		Height:  10,
		Time:    time.Now().UTC(),
	})

	s.stakingKeeper = s.app.AppKeepers.StakingKeeper
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

func (s *AnteTestSuite) TestAnteCreateValidator() {

	testCases := []struct {
		name          string
		maxChangeRate string
		expPass       bool
	}{
		{
			name:          "success - maxChangeRate < 5%",
			maxChangeRate: "0.01",
			expPass:       true,
		},
		{
			name:          "success - maxChangeRate = 5%",
			maxChangeRate: "0.05",
			expPass:       true,
		},
		{
			name:          "fail - maxChangeRate > 5%",
			maxChangeRate: "0.06",
			expPass:       false,
		},
		{
			name:          "fail - maxChangeRate = 5.1%",
			maxChangeRate: "0.051",
			expPass:       false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		// Create change rate decorator
		ante := decorators.NewChangeRateDecorator(s.stakingKeeper)

		// Create validator params
		_, msg, err := createValidatorMsg(tc.maxChangeRate)
		s.Require().NoError(err)

		// Submit the creation tx
		_, err = ante.AnteHandle(s.ctx, NewMockTx(msg), false, EmptyAnte)

		// Check if the error is expected
		if tc.expPass {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
			s.Require().Contains(err.Error(), "max change rate must not exceed")
		}
	}
}

func (s *AnteTestSuite) TestAnteEditValidator() {

	testCases := []struct {
		name          string
		maxChangeRate string
		expPass       bool
	}{
		{
			name:          "success - maxChangeRate < 5%",
			maxChangeRate: "0.01",
			expPass:       true,
		},
		{
			name:          "success - maxChangeRate = 5%",
			maxChangeRate: "0.05",
			expPass:       true,
		},
		{
			name:          "fail - maxChangeRate > 5%",
			maxChangeRate: "0.06",
			expPass:       false,
		},
		{
			name:          "fail - maxChangeRate = 5.1%",
			maxChangeRate: "0.051",
			expPass:       false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		// Create change rate decorator
		ante := decorators.NewChangeRateDecorator(s.stakingKeeper)

		// Create validator
		valPub, createMsg, err := createValidatorMsg("0.05")
		s.Require().NoError(err)

		// Submit the creation tx
		_, err = ante.AnteHandle(s.ctx, NewMockTx(createMsg), false, EmptyAnte)
		s.Require().NoError(err)

		// Create the validator
		val, err := stakingtypes.NewValidator(
			sdk.ValAddress(valPub.Address()),
			valPub,
			createMsg.Description,
		)
		s.Require().NoError(err)

		// Set the validator
		s.stakingKeeper.SetValidator(s.ctx, val)
		s.Require().NoError(err)

		// Edit validator params
		valAddr := sdk.ValAddress(valPub.Address())
		description := stakingtypes.NewDescription("test_moniker", "", "", "", "")
		newRate := math.LegacyMustNewDecFromStr(tc.maxChangeRate)
		minDelegation := sdk.OneInt()

		// Edit existing validator msg
		editMsg := stakingtypes.NewMsgEditValidator(
			valAddr,
			description,
			&newRate,
			&minDelegation,
		)

		// Submit the edit tx
		_, err = ante.AnteHandle(s.ctx, NewMockTx(editMsg), false, EmptyAnte)

		// Check if the error is expected
		if tc.expPass {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
			s.Require().Contains(err.Error(), "commission rate cannot change by more than")
		}
	}
}

// A helper function for getting a validator create msg
func createValidatorMsg(maxChangeRate string) (cryptotypes.PubKey, *stakingtypes.MsgCreateValidator, error) {
	// Create validator params
	valPub := secp256k1.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(valPub.Address())
	bondDenom := appparams.BondDenom
	selfBond := sdk.NewCoins(sdk.Coin{Amount: sdk.NewInt(100), Denom: bondDenom})
	stakingCoin := sdk.NewCoin(bondDenom, selfBond[0].Amount)
	description := stakingtypes.NewDescription("test_moniker", "", "", "", "")
	commission := stakingtypes.NewCommissionRates(
		math.LegacyMustNewDecFromStr("0.1"),
		math.LegacyMustNewDecFromStr("1"),
		math.LegacyMustNewDecFromStr(maxChangeRate),
	)

	// Creating a Validator
	msg, err := stakingtypes.NewMsgCreateValidator(
		valAddr,
		valPub,
		stakingCoin,
		description,
		commission,
		sdk.OneInt(),
	)

	// Return generated pub address, creation msg, and err
	return valPub, msg, err
}

func setBlockHeader(ctx sdk.Context, height uint64) sdk.Context {
	h := ctx.BlockHeader()
	h.Height = int64(height)
	return ctx.WithBlockHeader(h)
}

type MockTx struct {
	msgs []sdk.Msg
}

func NewMockTx(msgs ...sdk.Msg) MockTx {
	return MockTx{
		msgs: msgs,
	}
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
