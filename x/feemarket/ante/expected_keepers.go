package ante

import (
	"context"
	"time"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
)

// AccountKeeper defines the contract needed for AccountKeeper related APIs.
// Interface provides support to use non-sdk AccountKeeper for AnteHandler's decorators.
type AccountKeeper interface {
	GetParams(ctx context.Context) (params authtypes.Params)
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	GetModuleAddress(moduleName string) sdk.AccAddress
	AddressCodec() address.Codec
	UnorderedTransactionsEnabled() bool
	RemoveExpiredUnorderedNonces(ctx sdk.Context) error
	TryAddUnorderedNonce(ctx sdk.Context, sender []byte, timestamp time.Time) error
}

// FeeGrantKeeper defines the expected feegrant keeper.
type FeeGrantKeeper interface {
	UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}

// BankKeeper defines the contract needed for supply related APIs.
type BankKeeper interface {
	bankkeeper.Keeper
}

// FeeMarketKeeper defines the expected feemarket keeper.
type FeeMarketKeeper interface {
	GetState(ctx sdk.Context) (feemarkettypes.State, error)
	GetMinGasPrice(ctx sdk.Context, denom string) (sdk.DecCoin, error)
	GetParams(ctx sdk.Context) (feemarkettypes.Params, error)
	SetState(ctx sdk.Context, state feemarkettypes.State) error
	SetParams(ctx sdk.Context, params feemarkettypes.Params) error
	ResolveToDenom(ctx sdk.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error)
}
