package keeper_test

// import (
// 	"testing"

// 	"github.com/cosmos/cosmos-sdk/codec"

// 	"github.com/stretchr/testify/require"

// 	keep "github.com/CosmosContracts/juno/v27/x/mint/keeper"
// 	"github.com/CosmosContracts/juno/v27/x/mint/types"
// 	sdk "github.com/cosmos/cosmos-sdk/types"

// 	abci "github.com/cometbft/cometbft/abci/types"
// )

// func TestNewQuerier(t *testing.T) {
// 	app, ctx := createTestApp(true)
// 	legacyQuerierCdc := codec.NewAminoCodec(app.LegacyAmino())
// 	querier := keep.NewQuerier(app.AppKeepers.MintKeeper, legacyQuerierCdc.LegacyAmino)

// 	query := abci.RequestQuery{
// 		Path: "",
// 		Data: []byte{},
// 	}

// 	_, err := querier(ctx, []string{types.QueryParameters}, query)
// 	require.NoError(t, err)

// 	_, err = querier(ctx, []string{types.QueryInflation}, query)
// 	require.NoError(t, err)

// 	_, err = querier(ctx, []string{types.QueryAnnualProvisions}, query)
// 	require.NoError(t, err)

// 	_, err = querier(ctx, []string{"foo"}, query)
// 	require.Error(t, err)
// }

// func TestQueryParams(t *testing.T) {
// 	app, ctx := createTestApp(true)
// 	legacyQuerierCdc := codec.NewAminoCodec(app.LegacyAmino())
// 	querier := keep.NewQuerier(app.AppKeepers.MintKeeper, legacyQuerierCdc.LegacyAmino)

// 	var params types.Params

// 	res, sdkErr := querier(ctx, []string{types.QueryParameters}, abci.RequestQuery{})
// 	require.NoError(t, sdkErr)

// 	err := app.LegacyAmino().UnmarshalJSON(res, &params)
// 	require.NoError(t, err)

// 	require.Equal(t, app.AppKeepers.MintKeeper.GetParams(ctx), params)
// }

// func TestQueryInflation(t *testing.T) {
// 	app, ctx := createTestApp(true)
// 	legacyQuerierCdc := codec.NewAminoCodec(app.LegacyAmino())
// 	querier := keep.NewQuerier(app.AppKeepers.MintKeeper, legacyQuerierCdc.LegacyAmino)

// 	var inflation sdk.Dec

// 	res, sdkErr := querier(ctx, []string{types.QueryInflation}, abci.RequestQuery{})
// 	require.NoError(t, sdkErr)

// 	err := app.LegacyAmino().UnmarshalJSON(res, &inflation)
// 	require.NoError(t, err)

// 	require.Equal(t, app.AppKeepers.MintKeeper.GetMinter(ctx).Inflation, inflation)
// }

// func TestQueryAnnualProvisions(t *testing.T) {
// 	app, ctx := createTestApp(true)
// 	legacyQuerierCdc := codec.NewAminoCodec(app.LegacyAmino())
// 	querier := keep.NewQuerier(app.AppKeepers.MintKeeper, legacyQuerierCdc.LegacyAmino)

// 	var annualProvisions sdk.Dec

// 	res, sdkErr := querier(ctx, []string{types.QueryAnnualProvisions}, abci.RequestQuery{})
// 	require.NoError(t, sdkErr)

// 	err := app.LegacyAmino().UnmarshalJSON(res, &annualProvisions)
// 	require.NoError(t, err)

// 	require.Equal(t, app.AppKeepers.MintKeeper.GetMinter(ctx).AnnualProvisions, annualProvisions)
// }
