package keeper_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GenerateSalt generates a random salt, size length/2,  as a HEX encoded string.
func GenerateSalt(length int) (string, error) {
	if length == 0 {
		return "", fmt.Errorf("failed to generate salt: zero length")
	}

	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func (s *IntegrationTestSuite) TestMsgServer_AggregateExchangeRatePrevote() {
	ctx := s.ctx

	exchangeRatesStr := "123.2:ujuno"
	salt, err := GenerateSalt(32)
	s.Require().NoError(err)
	hash := types.GetAggregateVoteHash(salt, exchangeRatesStr, valAddr)

	invalidHash := &types.MsgAggregateExchangeRatePrevote{
		Hash:      "invalid_hash",
		Feeder:    addr.String(),
		Validator: valAddr.String(),
	}
	invalidFeeder := &types.MsgAggregateExchangeRatePrevote{
		Hash:      hash.String(),
		Feeder:    "invalid_feeder",
		Validator: valAddr.String(),
	}
	invalidValidator := &types.MsgAggregateExchangeRatePrevote{
		Hash:      hash.String(),
		Feeder:    addr.String(),
		Validator: "invalid_val",
	}
	validMsg := &types.MsgAggregateExchangeRatePrevote{
		Hash:      hash.String(),
		Feeder:    addr.String(),
		Validator: valAddr.String(),
	}

	_, err = s.msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(ctx), invalidHash)
	s.Require().Error(err)
	_, err = s.msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(ctx), invalidFeeder)
	s.Require().Error(err)
	_, err = s.msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(ctx), invalidValidator)
	s.Require().Error(err)
	_, err = s.msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(ctx), validMsg)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestMsgServer_AggregateExchangeRateVote() {
	ctx := s.ctx

	ratesStr := "ujuno:123.2"
	ratesStrInvalidCoin := "ujuno:123.2,badcoin:234.5"
	salt, err := GenerateSalt(32)
	s.Require().NoError(err)
	hash := types.GetAggregateVoteHash(salt, ratesStr, valAddr)
	hashInvalidRate := types.GetAggregateVoteHash(salt, ratesStrInvalidCoin, valAddr)

	prevoteMsg := &types.MsgAggregateExchangeRatePrevote{
		Hash:      hash.String(),
		Feeder:    addr.String(),
		Validator: valAddr.String(),
	}
	voteMsg := &types.MsgAggregateExchangeRateVote{
		Feeder:        addr.String(),
		Validator:     valAddr.String(),
		Salt:          salt,
		ExchangeRates: ratesStr,
	}
	voteMsgInvalidRate := &types.MsgAggregateExchangeRateVote{
		Feeder:        addr.String(),
		Validator:     valAddr.String(),
		Salt:          salt,
		ExchangeRates: ratesStrInvalidCoin,
	}

	// Flattened acceptList symbols to make checks easier
	acceptList := s.app.OracleKeeper.GetParams(ctx).AcceptList
	var acceptListFlat []string
	for _, v := range acceptList {
		acceptListFlat = append(acceptListFlat, v.SymbolDenom)
	}

	// No existing prevote
	_, err = s.msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(ctx), voteMsg)
	s.Require().EqualError(err, sdkerrors.Wrap(types.ErrNoAggregatePrevote, valAddr.String()).Error())
	_, err = s.msgServer.AggregateExchangeRatePrevote(sdk.WrapSDKContext(ctx), prevoteMsg)
	s.Require().NoError(err)
	// Reveal period mismatch
	_, err = s.msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(ctx), voteMsg)
	s.Require().EqualError(err, types.ErrRevealPeriodMissMatch.Error())

	// Valid
	s.app.OracleKeeper.SetAggregateExchangeRatePrevote(
		ctx,
		valAddr,
		types.NewAggregateExchangeRatePrevote(
			hash, valAddr, 1,
		))
	_, err = s.msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(ctx), voteMsg)
	s.Require().NoError(err)
	vote, err := s.app.OracleKeeper.GetAggregateExchangeRateVote(ctx, valAddr)
	s.Require().Nil(err)
	for _, v := range vote.ExchangeRateTuples {
		s.Require().Contains(acceptListFlat, strings.ToLower(v.Denom))
	}

	// Valid, but with an exchange rate which isn't in AcceptList
	s.app.OracleKeeper.SetAggregateExchangeRatePrevote(
		ctx,
		valAddr,
		types.NewAggregateExchangeRatePrevote(
			hashInvalidRate, valAddr, 1,
		))
	_, err = s.msgServer.AggregateExchangeRateVote(sdk.WrapSDKContext(ctx), voteMsgInvalidRate)
	s.Require().NoError(err)
	vote, err = s.app.OracleKeeper.GetAggregateExchangeRateVote(ctx, valAddr)
	s.Require().NoError(err)
	for _, v := range vote.ExchangeRateTuples {
		s.Require().Contains(acceptListFlat, strings.ToLower(v.Denom))
	}
}

func (s *IntegrationTestSuite) TestMsgServer_DelegateFeedConsent() {
	app, ctx := s.app, s.ctx

	feederAddr := sdk.AccAddress([]byte("addr________________"))
	feederAcc := app.AccountKeeper.NewAccountWithAddress(ctx, feederAddr)
	app.AccountKeeper.SetAccount(ctx, feederAcc)

	_, err := s.msgServer.DelegateFeedConsent(sdk.WrapSDKContext(ctx), &types.MsgDelegateFeedConsent{
		Operator: valAddr.String(),
		Delegate: feederAddr.String(),
	})
	s.Require().NoError(err)
}
