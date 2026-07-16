package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DenomResolver is an interface to convert a given token to the feemarket's base token.
type DenomResolver interface {
	// ConvertToDenom converts deccoin into the equivalent amount of the token denominated in denom.
	ConvertToDenom(ctx sdk.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error)
	// ExtraDenoms returns a list of denoms in addition of `Params.base_denom` it's possible to pay fees with
	ExtraDenoms(ctx sdk.Context) ([]string, error)
}

// TestDenomResolver is a test implementation of the DenomResolver interface.  It returns "feeCoin.Amount baseDenom" for all coins that are not the baseDenom.
// NOTE: DO NOT USE THIS IN PRODUCTION
type TestDenomResolver struct{}

// ConvertToDenom returns "coin.Amount denom" for all coins that are not the denom.
func (*TestDenomResolver) ConvertToDenom(_ sdk.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error) {
	if coin.Denom == denom {
		return coin, nil
	}

	return sdk.NewDecCoinFromDec(denom, coin.Amount), nil
}

func (*TestDenomResolver) ExtraDenoms(_ sdk.Context) ([]string, error) {
	return []string{}, nil
}

// ErrorDenomResolver accepts only the base fee denom and returns an error for
// every other denom. This is the production resolver for v30: fees are payable
// exclusively in the bond denom. Anything more permissive (e.g. the 1:1
// TestDenomResolver above) lets permissionless tokenfactory denoms satisfy
// fees at par with the bond denom — a straight fee bypass. Replace only with
// an oracle-backed resolver that prices denoms for real.
type ErrorDenomResolver struct{}

// ConvertToDenom returns an error for all coins that are not the denom.
func (*ErrorDenomResolver) ConvertToDenom(_ sdk.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error) {
	if coin.Denom == denom {
		return coin, nil
	}

	return sdk.DecCoin{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "error resolving denom")
}

func (*ErrorDenomResolver) ExtraDenoms(_ sdk.Context) ([]string, error) {
	return []string{}, nil
}
