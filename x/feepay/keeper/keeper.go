package keeper

import (
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	feepaytypes "github.com/CosmosContracts/juno/v20/x/feepay/types"
	feesharetypes "github.com/CosmosContracts/juno/v20/x/feeshare/types"
)

var (
	StoreKeyContracts    = []byte("contracts")
	StoreKeyContractUses = []byte("contract-uses")
)

// Keeper of this module maintains collections of feeshares for contracts
// registered to receive transaction fees.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	bankKeeper    bankkeeper.Keeper
	wasmKeeper    wasmkeeper.Keeper
	accountKeeper feesharetypes.AccountKeeper

	bondDenom string

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates new instances of the fees Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	bk bankkeeper.Keeper,
	wk wasmkeeper.Keeper,
	ak feesharetypes.AccountKeeper,
	bondDenom string,
	authority string,
) Keeper {
	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		bankKeeper:    bk,
		wasmKeeper:    wk,
		accountKeeper: ak,
		bondDenom:     bondDenom,
		authority:     authority,
	}
}

// GetAuthority returns the x/feeshare module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", feepaytypes.ModuleName))
}
