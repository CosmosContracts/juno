package keeper

import (
	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Keeper of the globalfee store
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	ss storetypes.KVStoreService,
	authority string,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeService: ss,
		authority:    authority,
	}
}

// GetAuthority returns the x/globalfee module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
