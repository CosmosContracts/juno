package types

// Used for the Juno ante handler so we can properly send 50% of fees to dAPP developers via fee share module

import (
	revtypes "github.com/CosmosContracts/juno/v12/x/feeshare/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	acctypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

type FeegrantKeeper interface {
	UseGrantedFees(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) (account acctypes.AccountI)
	GetModuleAddress(name string) sdk.AccAddress
}

type FeeShareKeeper interface {
	GetParams(ctx sdk.Context) revtypes.Params
	GetFeeShare(ctx sdk.Context, contract sdk.Address) (revtypes.FeeShare, bool)
}
