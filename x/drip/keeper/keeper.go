package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	revtypes "github.com/CosmosContracts/juno/v15/x/drip/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Keeper of this module maintains collections of feeshares for contracts
// registered to receive transaction fees.
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace

	bankKeeper    revtypes.BankKeeper
	accountKeeper revtypes.AccountKeeper

	feeCollectorName string
}

// NewKeeper creates new instances of the fees Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	bk revtypes.BankKeeper,
	ak revtypes.AccountKeeper,
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
		bankKeeper:       bk,
		accountKeeper:    ak,
		feeCollectorName: feeCollector,
	}
}

// SendCoinsFromAccountToFeeCollector transfers amt to the fee collector account, where it will be catch up by the distribution module at the next block
func (k Keeper) SendCoinsFromAccountToFeeCollector(ctx sdk.Context, senderAddr sdk.AccAddress, amt sdk.Coins) error {
	if senderAddr.Empty() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "senderAddr address cannot be empty")
	}

	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, k.feeCollectorName, amt)
}
