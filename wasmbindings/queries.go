package wasmbindings

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	types "github.com/CosmosContracts/juno/v30/wasmbindings/types"
	tokenfactorykeeper "github.com/CosmosContracts/juno/v30/x/tokenfactory/keeper"
	votingsnapshotkeeper "github.com/CosmosContracts/juno/v30/x/voting-snapshot/keeper"
)

type QueryPlugin struct {
	bankKeeper           bankkeeper.Keeper
	tokenFactoryKeeper   *tokenfactorykeeper.Keeper
	votingSnapshotKeeper votingsnapshotkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(b bankkeeper.Keeper, tfk *tokenfactorykeeper.Keeper, vsk votingsnapshotkeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		bankKeeper:           b,
		tokenFactoryKeeper:   tfk,
		votingSnapshotKeeper: vsk,
	}
}

// GetDenomAdmin is a query to get denom admin.
func (qp QueryPlugin) GetDenomAdmin(ctx sdk.Context, denom string) (*types.AdminResponse, error) {
	metadata, err := qp.tokenFactoryKeeper.GetAuthorityMetadata(ctx, denom)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "failed to get admin for denom: %s", denom)
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

// GetVotingPowerAt returns the bonded voting power of `address` at `height`,
// excluding LST-held delegations. Resolves to the most recent snapshot
// at-or-before the requested height.
func (qp QueryPlugin) GetVotingPowerAt(ctx sdk.Context, address string, height int64) (*types.VotingPowerResponse, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid voter address: %s", address)
	}
	power, err := qp.votingSnapshotKeeper.VotingPowerAt(ctx, addr, height)
	if err != nil {
		return nil, errorsmod.Wrap(err, "voting power lookup failed")
	}
	return &types.VotingPowerResponse{Power: power.String()}, nil
}

// GetTotalVotingPowerAt returns the total bonded supply at `height`.
func (qp QueryPlugin) GetTotalVotingPowerAt(ctx sdk.Context, height int64) (*types.VotingPowerResponse, error) {
	power, err := qp.votingSnapshotKeeper.TotalVotingPowerAt(ctx, height)
	if err != nil {
		return nil, errorsmod.Wrap(err, "total voting power lookup failed")
	}
	return &types.VotingPowerResponse{Power: power.String()}, nil
}
