package bindings

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	types "github.com/CosmosContracts/juno/v29/wasmbindings/types"
	tokenfactorykeeper "github.com/CosmosContracts/juno/v29/x/tokenfactory/keeper"
)

type QueryPlugin struct {
	bankKeeper         bankkeeper.Keeper
	tokenFactoryKeeper *tokenfactorykeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(b bankkeeper.Keeper, tfk *tokenfactorykeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		bankKeeper:         b,
		tokenFactoryKeeper: tfk,
	}
}

// GetDenomAdmin is a query to get denom admin.
func (qp QueryPlugin) GetDenomAdmin(ctx sdk.Context, denom string) (*types.AdminResponse, error) {
	metadata, err := qp.tokenFactoryKeeper.GetAuthorityMetadata(ctx, denom)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin for denom: %s", denom)
	}
	return &types.AdminResponse{Admin: metadata.Admin}, nil
}

func (qp QueryPlugin) GetDenomsByCreator(ctx sdk.Context, creator string) (*types.DenomsByCreatorResponse, error) {
	// TODO: validate creator address
	denoms := qp.tokenFactoryKeeper.GetDenomsFromCreator(ctx, creator)
	return &types.DenomsByCreatorResponse{Denoms: denoms}, nil
}

func (qp QueryPlugin) GetMetadata(ctx sdk.Context, denom string) (*types.MetadataResponse, error) {
	metadata, found := qp.bankKeeper.GetDenomMetaData(ctx, denom)
	var parsed *types.Metadata
	if found {
		parsed = SdkMetadataToWasm(metadata)
	}
	return &types.MetadataResponse{Metadata: parsed}, nil
}

func (qp QueryPlugin) GetParams(ctx sdk.Context) (*types.ParamsResponse, error) {
	params := qp.tokenFactoryKeeper.GetParams(ctx)
	return &types.ParamsResponse{
		Params: types.Params{
			DenomCreationFee: ConvertSdkCoinsToWasmCoins(params.DenomCreationFee),
		},
	}, nil
}
