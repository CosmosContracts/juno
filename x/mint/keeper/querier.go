package keeper

// TODO: Can not find anything in the docs which says we need this? the Route() was removed and seems this was apart of that.
// If so it does nothing now.

// import (
// 	abci "github.com/cometbft/cometbft/abci/types"

// 	"github.com/CosmosContracts/juno/v27/x/mint/types"
// 	"github.com/cosmos/cosmos-sdk/codec"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
// )

// // NewQuerier returns a minting Querier handler.
// func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
// 	return func(ctx sdk.Context, path []string, _ abci.RequestQuery) ([]byte, error) {
// 		switch path[0] {
// 		case types.QueryParameters:
// 			return queryParams(ctx, k, legacyQuerierCdc)

// 		case types.QueryInflation:
// 			return queryInflation(ctx, k, legacyQuerierCdc)

// 		case types.QueryAnnualProvisions:
// 			return queryAnnualProvisions(ctx, k, legacyQuerierCdc)

// 		default:
// 			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
// 		}
// 	}
// }

// func queryParams(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
// 	params := k.GetParams(ctx)

// 	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
// 	if err != nil {
// 		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
// 	}

// 	return res, nil
// }

// func queryInflation(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
// 	minter := k.GetMinter(ctx)

// 	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, minter.Inflation)
// 	if err != nil {
// 		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
// 	}

// 	return res, nil
// }

// func queryAnnualProvisions(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
// 	minter := k.GetMinter(ctx)

// 	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, minter.AnnualProvisions)
// 	if err != nil {
// 		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
// 	}

// 	return res, nil
// }
