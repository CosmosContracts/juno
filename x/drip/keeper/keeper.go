package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	revtypes "github.com/CosmosContracts/juno/v16/x/drip/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Keeper of this module maintains collections of feeshares for contracts
// registered to receive transaction fees.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	bankKeeper    revtypes.BankKeeper
	accountKeeper revtypes.AccountKeeper

	feeCollectorName string
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates new instances of the fees Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	bk revtypes.BankKeeper,
	ak revtypes.AccountKeeper,
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
	if senderAddr.Empty() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "senderAddr address cannot be empty")
	}

	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, k.feeCollectorName, amt)
}