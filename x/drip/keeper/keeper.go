package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	driptypes "github.com/CosmosContracts/juno/v16/x/drip/types"
)

// Keeper of this module maintains collections of feeshares for contracts
// registered to receive transaction fees.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	bankKeeper    driptypes.BankKeeper
	accountKeeper driptypes.AccountKeeper

	feeCollectorName string
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates new instances of the fees Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	bk driptypes.BankKeeper,
	ak driptypes.AccountKeeper,
	feeCollector string,
	authority string,
) Keeper {
	return Keeper{
		storeKey:         storeKey,
		cdc:              cdc,
		bankKeeper:       bk,
		accountKeeper:    ak,
		feeCollectorName: feeCollector,
		authority:        authority,
	}
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SendCoinsFromAccountToFeeCollector transfers amt to the fee collector account, where it will be catch up by the distribution module at the next block
func (k Keeper) SendCoinsFromAccountToFeeCollector(ctx sdk.Context, senderAddr sdk.AccAddress, amt sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, k.feeCollectorName, amt)
}
