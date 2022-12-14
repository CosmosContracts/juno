package keeper

import (
	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func NewOracleProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		if content == nil {
			return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "content must not be empty")
		}

		switch c := content.(type) {
		case *types.AddTrackingPriceHistoryProposal:
			return handleAddTrackingPriceHistoryProposal(ctx, k, *c)
		case *types.AddTrackingPriceHistoryWithAcceptListProposal:
			return handleAddTrackingPriceHistoryWithAcceptListProposal(ctx, k, *c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized wasm proposal content type: %T", c)
		}
	}
}

func handleAddTrackingPriceHistoryProposal(ctx sdk.Context, k Keeper, p types.AddTrackingPriceHistoryProposal) error {

}

func handleAddTrackingPriceHistoryWithAcceptListProposal(ctx sdk.Context, k Keeper, p types.AddTrackingPriceHistoryWithAcceptListProposal) error {

}
