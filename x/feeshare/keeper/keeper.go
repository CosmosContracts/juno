package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	revtypes "github.com/CosmosContracts/juno/v12/x/feeshare/types"
)

// Keeper of this module maintains collections of revenues for contracts
// registered to receive transaction fees.
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace

	accountKeeper revtypes.AccountKeeper
	bankKeeper    revtypes.BankKeeper
	// wasmKeeper    revtypes.WasmKeeper

	hooks            revtypes.RevenueHooks
	feeCollectorName string
}

// NewKeeper creates new instances of the fees Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	ak revtypes.AccountKeeper,
	bk revtypes.BankKeeper,
	feeCollector string,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(revtypes.ParamKeyTable())
	}

	return Keeper{
		storeKey:         storeKey,
		cdc:              cdc,
		paramstore:       ps,
		accountKeeper:    ak,
		bankKeeper:       bk,
		hooks:            nil,
		feeCollectorName: feeCollector,
	}
}

// SetHooks set the epoch hooks
func (k *Keeper) SetHooks(rh revtypes.RevenueHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set revenue hooks twice")
	}

	k.hooks = rh

	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", revtypes.ModuleName))
}
