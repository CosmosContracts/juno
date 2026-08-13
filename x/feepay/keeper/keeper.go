package keeper

import (
	"context"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	feemarkettypes "github.com/CosmosContracts/juno/v31/x/feemarket/types"
	feepaytypes "github.com/CosmosContracts/juno/v31/x/feepay/types"
)

var (
	StoreKeyContracts    = []byte("contracts")
	StoreKeyContractUses = []byte("contract-uses")
)

// Keeper of this module maintains collections of feeshares for contracts
// registered to receive transaction fees.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService

	bankKeeper    bankkeeper.Keeper
	wasmKeeper    wasmkeeper.Keeper
	accountKeeper authkeeper.AccountKeeper

	feeMarketKeeper FeeMarketKeeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// FeeMarketKeeper provides the denomination backing FeePay balances.
type FeeMarketKeeper interface {
	GetParams(ctx sdk.Context) (feemarkettypes.Params, error)
}

// NewKeeper creates new instances of the fees Keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	ss storetypes.KVStoreService,
	bk bankkeeper.Keeper,
	wk wasmkeeper.Keeper,
	ak authkeeper.AccountKeeper,
	fmk FeeMarketKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:             cdc,
		storeService:    ss,
		bankKeeper:      bk,
		wasmKeeper:      wk,
		accountKeeper:   ak,
		feeMarketKeeper: fmk,
		authority:       authority,
	}
}

func (k Keeper) feeDenom(ctx context.Context) (string, error) {
	params, err := k.feeMarketKeeper.GetParams(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		return "", err
	}

	return params.FeeDenom, nil
}

// GetAuthority returns the x/feeshare module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", fmt.Sprintf("x/%s", feepaytypes.ModuleName))
}
