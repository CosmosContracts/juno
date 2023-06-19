package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v16/x/globalfee/types"
)

// Keeper of the globalfee store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	authority string,
) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  key,
		authority: authority,
	}
}

// GetAuthority returns the x/globalfee module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetParams sets the x/globalfee module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)

	return nil
}

// GetParams returns the current x/globalfee module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}
