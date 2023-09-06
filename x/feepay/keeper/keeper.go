package keeper

import (
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/CosmosContracts/juno/v17/x/feepay/types"
	revtypes "github.com/CosmosContracts/juno/v17/x/feeshare/types"
)

// Keeper of this module maintains collections of feeshares for contracts
// registered to receive transaction fees.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	bankKeeper    revtypes.BankKeeper
	wasmKeeper    wasmkeeper.Keeper
	accountKeeper revtypes.AccountKeeper

	feeCollectorName string
	bondDenom        string

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates new instances of the fees Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	bk revtypes.BankKeeper,
	wk wasmkeeper.Keeper,
	ak revtypes.AccountKeeper,
	feeCollector string,
	bondDenom string,
	authority string,
) Keeper {
	panic(bondDenom)
	return Keeper{
		storeKey:         storeKey,
		cdc:              cdc,
		bankKeeper:       bk,
		wasmKeeper:       wk,
		accountKeeper:    ak,
		feeCollectorName: feeCollector,
		bondDenom:        bondDenom,
		authority:        authority,
	}
}

// GetAuthority returns the x/feeshare module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", revtypes.ModuleName))
}

// Check if a contract is associated with a FeePay contract
func (k Keeper) IsValidContract(ctx sdk.Context, contractAddr string) bool {

	// Get store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("contracts"))

	// Get data
	hasData := store.Has([]byte(contractAddr))

	// Return true if data is not nil
	return hasData
}

// Register the contract in the module store
func (k Keeper) RegisterContract(ctx sdk.Context, fpc *types.FeePayContract) bool {

	// Return false because the contract was already registered
	if k.IsValidContract(ctx, fpc.ContractAddress) {
		return false
	}

	// Register the new fee pay contract in the KV store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("contracts"))
	key := []byte(fpc.ContractAddress)
	bz := k.cdc.MustMarshal(&fpc)
	store.Set(key, bz)
	return true
}

func (k Keeper) FundContract(ctx sdk.Context, mfc *types.MsgFundFeePayContract) bool {

	// Return false because the contract was already registered
	if !k.IsValidContract(ctx, mfc.ContractAddress) {
		return false
	}

	// Get existing fee pay contract from store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("contracts"))
	key := []byte(mfc.ContractAddress)
	bz := store.Get(key)

	var fpc types.FeePayContract
	k.cdc.MustUnmarshal(bz, &fpc)

	// Increment the fpc balance with the correct denom, ignore all others
	for _, c := range mfc.Amount {
		if c.Denom == PUT_DENOM_HERE {
			fpc.Balance += c.Amount.Uint64()
		}
	}

	// Update the balance in KV store, return success
	store.Set(key, k.cdc.MustMarshal(&fpc))
	return true
}
