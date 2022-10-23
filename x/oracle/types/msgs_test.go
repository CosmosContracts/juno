package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgFeederDelegation(t *testing.T) {
	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
		sdk.AccAddress([]byte("addr2_______________")),
	}

	msgInvalidOperatorAddr := "invalid operator address (empty address string is not allowed): invalid address"
	msgInvalidDelegatorAddr := "invalid delegate address (empty address string is not allowed): invalid address"

	tests := []struct {
		delegator        sdk.ValAddress
		delegate         sdk.AccAddress
		expectPass       bool
		expectedErrorMsg string
	}{
		{sdk.ValAddress(addrs[0]), addrs[1], true, "test should pass"},
		{sdk.ValAddress{}, addrs[1], false, msgInvalidOperatorAddr},
		{sdk.ValAddress(addrs[0]), sdk.AccAddress{}, false, msgInvalidDelegatorAddr},
		{nil, nil, false, msgInvalidOperatorAddr},
	}

	for i, tc := range tests {
		msg := NewMsgDelegateFeedConsent(tc.delegator, tc.delegate)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.ErrorContainsf(t, msg.ValidateBasic(), tc.expectedErrorMsg, "test: %v", i)
		}
	}
}

func TestMsgAggregateExchangeRatePrevote(t *testing.T) {
	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
	}

	exchangeRates := sdk.DecCoins{sdk.NewDecCoinFromDec(UmeeDenom, sdk.OneDec()), sdk.NewDecCoinFromDec(UmeeDenom, sdk.NewDecWithPrec(32121, 1))}
	bz := GetAggregateVoteHash("1", exchangeRates.String(), sdk.ValAddress(addrs[0]))
	msgInvalidHashLength := "invalid hash length; should equal 20"
	msgInvalidFeederAddr := "invalid feeder address (empty address string is not allowed): invalid address"
	msgInvalidOperatorAddr := "invalid operator address (empty address string is not allowed): invalid address"

	tests := []struct {
		hash             AggregateVoteHash
		exchangeRates    sdk.DecCoins
		feeder           sdk.AccAddress
		validator        sdk.AccAddress
		expectPass       bool
		expectedErrorMsg string
	}{
		{bz, exchangeRates, addrs[0], addrs[0], true, "test should pass"},
		{bz[1:], exchangeRates, addrs[0], addrs[0], false, msgInvalidHashLength},
		{[]byte("0\x01"), exchangeRates, addrs[0], addrs[0], false, msgInvalidHashLength},
		{AggregateVoteHash{}, exchangeRates, addrs[0], addrs[0], false, msgInvalidHashLength},
		{bz, exchangeRates, sdk.AccAddress{}, addrs[0], false, msgInvalidFeederAddr},
		{bz, exchangeRates, addrs[0], sdk.AccAddress{}, false, msgInvalidOperatorAddr},
	}

	for i, tc := range tests {
		msg := NewMsgAggregateExchangeRatePrevote(tc.hash, tc.feeder, sdk.ValAddress(tc.validator))
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.ErrorContainsf(t, msg.ValidateBasic(), tc.expectedErrorMsg, "test: %v", i)
		}
	}
}

func TestMsgAggregateExchangeRateVote(t *testing.T) {
	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
	}

	invalidExchangeRates := "a,b"
	exchangeRates := "foo:1.0,bar:1232.123"
	zeroExchangeRates := "foo:0.0,bar:1232.132"
	negativeExchangeRates := "foo:-1234.5,bar:1232.132"
	overFlowMsgExchangeRates := StringWithCharset(4097, "56432")
	overFlowExchangeRates := "foo:100000000000000000000000000000000000000000000000000000000000000000000000000000.01,bar:1232.132"
	validSalt := "0cf33fb528b388660c3a42c3f3250e983395290b75fef255050fb5bc48a6025f"
	saltWithColon := "0cf33fb528b388660c3a42c3f3250e983395290b75fef255050fb5bc48a6025:"
	msgInvalidSalt := "invalid salt length; must be 64"
	msgInvalidOverflowValue := "out of range; bitLen:"
	msgInvalidHexString := "salt must be a valid hex string: invalid salt format"
	msgInvalidUnknownRequest := "must provide at least one oracle exchange rate: unknown request"
	msgInvalidFeederAddr := "invalid feeder address (empty address string is not allowed): invalid address"
	msgInvalidOperatorAddr := "invalid operator address (empty address string is not allowed): invalid address"
	msgInvalidOraclePrice := "failed to parse exchange rates string cause: invalid oracle price: invalid coins"
	msgInvalidOverflowExceedCharacter := "exchange rates string can not exceed 4096 characters: invalid request"
	msgInvalidExchangeRates := "failed to parse exchange rates string cause: invalid exchange rate a: invalid coins"

	tests := []struct {
		feeder           sdk.AccAddress
		validator        sdk.AccAddress
		salt             string
		exchangeRates    string
		expectPass       bool
		expectedErrorMsg string
	}{
		{addrs[0], addrs[0], validSalt, exchangeRates, true, "test should pass"},
		{addrs[0], addrs[0], validSalt, invalidExchangeRates, false, msgInvalidExchangeRates},
		{addrs[0], addrs[0], validSalt, zeroExchangeRates, false, msgInvalidOraclePrice},
		{addrs[0], addrs[0], validSalt, negativeExchangeRates, false, msgInvalidOraclePrice},
		{addrs[0], addrs[0], validSalt, overFlowMsgExchangeRates, false, msgInvalidOverflowExceedCharacter},
		{addrs[0], addrs[0], validSalt, overFlowExchangeRates, false, msgInvalidOverflowValue},
		{sdk.AccAddress{}, sdk.AccAddress{}, validSalt, exchangeRates, false, msgInvalidFeederAddr},
		{addrs[0], sdk.AccAddress{}, validSalt, exchangeRates, false, msgInvalidOperatorAddr},
		{addrs[0], addrs[0], "", exchangeRates, false, msgInvalidSalt},
		{addrs[0], addrs[0], validSalt, "", false, msgInvalidUnknownRequest},
		{addrs[0], addrs[0], saltWithColon, exchangeRates, false, msgInvalidHexString},
	}

	for i, tc := range tests {
		msg := NewMsgAggregateExchangeRateVote(tc.salt, tc.exchangeRates, tc.feeder, sdk.ValAddress(tc.validator))
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.ErrorContainsf(t, msg.ValidateBasic(), tc.expectedErrorMsg, "test: %v", i)
		}
	}
}

func TestNewMsgAggregateExchangeRatePrevote(t *testing.T) {
	vals := GenerateRandomValAddr(2)
	feederAddr := sdk.AccAddress(vals[1])

	exchangeRates := sdk.DecCoins{sdk.NewDecCoinFromDec(UmeeDenom, sdk.OneDec()), sdk.NewDecCoinFromDec(UmeeDenom, sdk.NewDecWithPrec(32121, 1))}
	bz := GetAggregateVoteHash("1", exchangeRates.String(), sdk.ValAddress(vals[0]))

	aggregateExchangeRatePreVote := NewMsgAggregateExchangeRatePrevote(
		bz,
		feederAddr,
		vals[0],
	)

	require.NotNil(t, aggregateExchangeRatePreVote.GetSignBytes())
	require.Equal(t, aggregateExchangeRatePreVote.GetSigners(), []sdk.AccAddress{feederAddr})
}

func TestNewMsgAggregateExchangeRateVote(t *testing.T) {
	vals := GenerateRandomValAddr(2)
	feederAddr := sdk.AccAddress(vals[1])

	aggregateExchangeRateVote := NewMsgAggregateExchangeRateVote(
		"salt",
		"0.1",
		feederAddr,
		vals[0],
	)

	require.NotNil(t, aggregateExchangeRateVote.GetSignBytes())
	require.Equal(t, aggregateExchangeRateVote.GetSigners(), []sdk.AccAddress{feederAddr})
}

func TestMsgDelegateFeedConsent(t *testing.T) {
	vals := GenerateRandomValAddr(2)
	msgFeedConsent := NewMsgDelegateFeedConsent(vals[0], sdk.AccAddress(vals[1]))

	require.NotNil(t, msgFeedConsent.GetSignBytes())
	require.Equal(t, msgFeedConsent.GetSigners(), []sdk.AccAddress{sdk.AccAddress(vals[0])})
}
