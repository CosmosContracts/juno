package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper defines the expected interface contract defined by the x/staking
// module.
type StakingKeeper interface {
	Validator(ctx sdk.Context, address sdk.ValAddress) stakingtypes.ValidatorI
	GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator
	TotalBondedTokens(sdk.Context) sdk.Int
	Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec) sdk.Int
	Jail(sdk.Context, sdk.ConsAddress)
	ValidatorsPowerStoreIterator(ctx sdk.Context) sdk.Iterator
	MaxValidators(sdk.Context) uint32
	PowerReduction(ctx sdk.Context) (res sdk.Int)
}

// DistributionKeeper defines the expected interface contract defined by the
// x/distribution module.
type DistributionKeeper interface {
	AllocateTokensToValidator(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins)
	GetValidatorOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins
}

// AccountKeeper defines the expected interface contract defined by the x/auth
// module.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI

	// only used for simulation
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

// BankKeeper defines the expected interface contract defined by the x/bank
// module.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
}
