package keeper

import (
	"strings"

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
		case *types.RemoveTrackingPriceHistoryProposal:
			return handleRemoveTrackingPriceHistoryProposal(ctx, k, *c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized wasm proposal content type: %T", c)
		}
	}
}

func handleAddTrackingPriceHistoryProposal(ctx sdk.Context, k Keeper, p types.AddTrackingPriceHistoryProposal) error {
	// Validate
	if err := p.ValidateBasic(); err != nil {
		return err
	}

	for i := range p.TrackingList {
		p.TrackingList[i].SymbolDenom = strings.ToUpper(p.TrackingList[i].SymbolDenom)
	}
	// Get params
	params := k.GetParams(ctx)
	// Check if not in accept list
	currentAcceptList := params.AcceptList
	_, _, isSubset := isSubSet(currentAcceptList, p.TrackingList)
	if !isSubset {
		return sdkerrors.Wrap(types.ErrUnknownDenom, "Denom not in accept list")
	}
	// Check if not in tracking list then add to tracking list
	currentTrackingList := params.PriceTrackingList
	unSubList, _, isSubset := isSubSet(currentTrackingList, p.TrackingList)
	if !isSubset {
		currentTrackingList = append(currentTrackingList, unSubList...)
	}

	// Set params
	params.PriceTrackingList = currentTrackingList
	k.SetParams(ctx, params)

	return nil
}

func handleAddTrackingPriceHistoryWithAcceptListProposal(ctx sdk.Context, k Keeper, p types.AddTrackingPriceHistoryWithAcceptListProposal) error {
	// Validate
	if err := p.ValidateBasic(); err != nil {
		return err
	}

	for i := range p.TrackingList {
		p.TrackingList[i].SymbolDenom = strings.ToUpper(p.TrackingList[i].SymbolDenom)
	}
	// Get params
	params := k.GetParams(ctx)
	// Check if not in accept list then add to accept list
	currentAcceptList := params.AcceptList
	unSubList, _, isSubset := isSubSet(currentAcceptList, p.TrackingList)
	if !isSubset {
		currentAcceptList = append(currentAcceptList, unSubList...)
	}
	// Check if not in tracking list then add to tracking list
	currentTrackingList := params.PriceTrackingList
	unSubList, _, isSubset = isSubSet(currentTrackingList, p.TrackingList)
	if !isSubset {
		currentTrackingList = append(currentTrackingList, unSubList...)
	}

	// Set params
	params.AcceptList = currentAcceptList
	params.PriceTrackingList = currentTrackingList
	k.SetParams(ctx, params)

	return nil
}

func handleRemoveTrackingPriceHistoryProposal(ctx sdk.Context, k Keeper, p types.RemoveTrackingPriceHistoryProposal) error {
	// Validate
	if err := p.ValidateBasic(); err != nil {
		return err
	}
	for i := range p.RemoveTrackingList {
		p.RemoveTrackingList[i].SymbolDenom = strings.ToUpper(p.RemoveTrackingList[i].SymbolDenom)
	}
	// Get params
	params := k.GetParams(ctx)
	// Check if denom in tracking list then remove from tracking list
	currentTrackingList := params.PriceTrackingList
	_, subList, _ := isSubSet(currentTrackingList, p.RemoveTrackingList)
	if len(subList) > 0 {
		// remove from tracking list and price tracking store
		for _, denom := range subList {
			// remove tracking list
			currentTrackingList = removeDenomFromList(currentTrackingList, denom)
			// remove store
			var keys []uint64
			k.IterateDenomPriceHistory(ctx, denom.SymbolDenom, func(key uint64, _ types.PriceHistoryEntry) bool {
				keys = append(keys, key)
				return false
			})
			for _, key := range keys {
				k.DeleteDenomPriceHistory(ctx, denom.SymbolDenom, key)
			}
		}
	}
	// Set params
	params.PriceTrackingList = currentTrackingList
	k.SetParams(ctx, params)

	return nil
}

func isSubSet(super, sub types.DenomList) (unSubList types.DenomList, subList types.DenomList, isSubSet bool) {
	if len(sub) == 0 {
		return unSubList, subList, true
	}

	var matches int
	for _, o := range sub {
		var isEqual bool
		for _, s := range super {
			s := s
			if o.Equal(&s) {
				matches++
				isEqual = true
				break
			}
		}
		if isEqual {
			subList = append(subList, o)
		} else {
			unSubList = append(unSubList, o)
		}
	}

	return unSubList, subList, matches == len(sub)
}

func removeDenomFromList(denomList types.DenomList, removeDenom types.Denom) types.DenomList {
	var newDenomList types.DenomList

	for _, denom := range denomList {
		if !denom.Equal(&removeDenom) {
			newDenomList = append(newDenomList, denom)
		}
	}

	return newDenomList
}
