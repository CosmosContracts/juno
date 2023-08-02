package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v16/x/cw-modules/types"
)

// Keeper of the cw-modules store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	contractKeeper wasmkeeper.PermissionedKeeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	contractKeeper wasmkeeper.PermissionedKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:            cdc,
		storeKey:       key,
		contractKeeper: contractKeeper,
		authority:      authority,
	}
}

// GetAuthority returns the x/cw-modules module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetParams sets the x/cw-modules module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)

	return nil
}

// GetParams returns the current x/cw-modules module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// TODO: Should we panic out, or continue on and then return the errors(s) at the end? Then handle in the EndBlocker
// I think we should just use continue (panic is testing), or just BeginBlocker.
func (k Keeper) ExecuteAllContractModulesEndBlock(ctx sdk.Context) error {
	message := []byte(types.EndBlockMessage)

	// p := k.GetParams(ctx)

	// for _, addr := range p.ContractAddresses {
	for _, addr := range []string{"juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8"} {
		// convert addr to sdk.AccAddress
		contract, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			// TODO:
			// !important
			continue
			// panic(err)
		}

		_, err = k.contractKeeper.Sudo(ctx, contract, message)
		if err != nil {
			// TODO:
			// !important
			// panic(err)
			continue
		}
	}

	return nil
}
