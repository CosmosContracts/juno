package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v17/x/feepay/types"
)

// GetParams returns the total set of fees parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	// set to nil? add a ParamsKey which is a 0x00 value. Maybe a bool for enable and disable (set in .proto file genesis.proto (kill switch))
	// store := ctx.KVStore(k.storeKey)
	// bz := store.Get(types.ParamsKey)
	// if bz == nil {
	// 	return p
	// }

	// k.cdc.MustUnmarshal(bz, &p)
	return p
}

// SetParams sets the fees parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	// if err := p.Validate(); err != nil {
	// 	return err
	// }

	// store := ctx.KVStore(k.storeKey)
	// bz := k.cdc.MustMarshal(&p)
	// store.Set(types.ParamsKey, bz)
	return nil
}
