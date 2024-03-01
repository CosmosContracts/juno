package decorators_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	protov2 "google.golang.org/protobuf/proto"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v19/app"
	decorators "github.com/CosmosContracts/juno/v19/app/decorators"
	appparams "github.com/CosmosContracts/juno/v19/app/params"
)

// Define an empty ante handle
var (
	EmptyAnte = func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
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

// Test the change rate decorator with standard create msgs,
// authz create messages, and inline authz create messages
func (s *AnteTestSuite) TestAnteCreateValidator() {
	// Grantee used for authz msgs
	grantee := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// Loop through all possible change rates
	for i := 0; i <= 100; i++ {

		// Calculate change rate
		maxChangeRate := getChangeRate(i)

		// Create change rate decorator
		ante := decorators.NewChangeRateDecorator(s.stakingKeeper)

		// Create validator params
		_, msg, err := createValidatorMsg(maxChangeRate)
		s.Require().NoError(err)

		// Submit the creation tx
		_, err = ante.AnteHandle(s.ctx, NewMockTx(msg), false, EmptyAnte)
		validateCreateMsg(s, err, i)

		// Submit the creation tx with authz
		authzMsg := authz.NewMsgExec(grantee, []sdk.Msg{msg})
		_, err = ante.AnteHandle(s.ctx, NewMockTx(&authzMsg), false, EmptyAnte)
		validateCreateMsg(s, err, i)

		// Submit the creation tx with inline authz
		inlineAuthzMsg := authz.NewMsgExec(grantee, []sdk.Msg{&authzMsg})
		_, err = ante.AnteHandle(s.ctx, NewMockTx(&inlineAuthzMsg), false, EmptyAnte)
		validateCreateMsg(s, err, i)
	}
}

// Test the change rate decorator with standard edit msgs,
// authz edit messages, and inline authz edit messages
func (s *AnteTestSuite) TestAnteEditValidator() {
	// Grantee used for authz msgs
	grantee := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// Loop through all possible change rates
	for i := 0; i <= 100; i++ {

		// Calculate change rate
		maxChangeRate := getChangeRate(i)

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
		newRate := math.LegacyMustNewDecFromStr(maxChangeRate)
		minDelegation := sdk.OneInt()

		// Edit existing validator msg
		editMsg := stakingtypes.NewMsgEditValidator(
			valAddr,
			createMsg.Description,
			&newRate,
			&minDelegation,
		)

		// Submit the edit tx
		_, err = ante.AnteHandle(s.ctx, NewMockTx(editMsg), false, EmptyAnte)
		validateEditMsg(s, err, i)

		// Submit the edit tx with authz
		authzMsg := authz.NewMsgExec(grantee, []sdk.Msg{editMsg})
		_, err = ante.AnteHandle(s.ctx, NewMockTx(&authzMsg), false, EmptyAnte)
		validateEditMsg(s, err, i)

		// Submit the edit tx with inline authz
		inlineAuthzMsg := authz.NewMsgExec(grantee, []sdk.Msg{&authzMsg})
		_, err = ante.AnteHandle(s.ctx, NewMockTx(&inlineAuthzMsg), false, EmptyAnte)
		validateEditMsg(s, err, i)
	}
}

// Convert an integer to a percentage, formatted as a string
// Example: 5 -> "0.05", 10 -> "0.1"
func getChangeRate(i int) string {
	if i >= 100 {
		return "1.00"
	}

	return fmt.Sprintf("0.%02d", i)
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

// Validate the create msg err is expected
func validateCreateMsg(s *AnteTestSuite, err error, i int) {
	if i <= 5 {
		s.Require().NoError(err)
	} else {
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "max change rate must not exceed")
	}
}

// Validate the edit msg err is expected
func validateEditMsg(s *AnteTestSuite, err error, i int) {
	if i <= 5 {
		s.Require().NoError(err)
	} else {
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "commission rate cannot change by more than")
	}
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
