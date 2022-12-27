package keeper

import (
	"testing"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	"github.com/stretchr/testify/require"
)

func TestAddTrackingPriceHistoryProposal(t *testing.T) {
	ctx, keepers := CreateTestInput(t, false)
	govKeeper, oracleKeeper := keepers.GovKeeper, keepers.OracleKeeper

	var priceTrackingList types.DenomList
	params := types.DefaultParams()
	params.PriceTrackingList = priceTrackingList
	oracleKeeper.SetParams(ctx, params)

	params = oracleKeeper.GetParams(ctx)
	require.Equal(t, 0, len(params.PriceTrackingList))

	trackingList := types.DenomList{
		{
			BaseDenom:   types.JunoDenom,
			SymbolDenom: types.JunoSymbol,
			Exponent:    types.JunoExponent,
		}, // Already in AcceptList (Default params)
	}

	src := types.AddTrackingPriceHistoryProposalFixture(func(p *types.AddTrackingPriceHistoryProposal) {
		p.TrackingList = trackingList
	})

	// submit proposal
	submitedProposal, err := govKeeper.SubmitProposal(ctx, src)
	require.NoError(t, err)
	// execute proposal
	handler := govKeeper.Router().GetRoute(submitedProposal.ProposalRoute())
	err = handler(ctx, submitedProposal.GetContent())
	require.NoError(t, err)

	params = oracleKeeper.GetParams(ctx)
	require.Equal(t, 1, len(params.PriceTrackingList))
	require.Equal(t, params.PriceTrackingList, trackingList)
}

func TestAddTrackingPriceHistoryWithAcceptListProposal(t *testing.T) {
	ctx, keepers := CreateTestInput(t, false)
	govKeeper, oracleKeeper := keepers.GovKeeper, keepers.OracleKeeper

	var emptyDenomList types.DenomList
	params := types.DefaultParams()
	params.AcceptList = emptyDenomList
	params.PriceTrackingList = emptyDenomList
	oracleKeeper.SetParams(ctx, params)

	params = oracleKeeper.GetParams(ctx)
	require.Equal(t, 0, len(params.AcceptList))
	require.Equal(t, 0, len(params.PriceTrackingList))

	trackingList := types.DenomList{
		{
			BaseDenom:   types.JunoDenom,
			SymbolDenom: types.JunoSymbol,
			Exponent:    types.JunoExponent,
		},
		{
			BaseDenom:   types.AtomDenom,
			SymbolDenom: types.AtomSymbol,
			Exponent:    types.AtomExponent,
		},
	}

	src := types.AddTrackingPriceHistoryWithAcceptListProposalFixture(func(p *types.AddTrackingPriceHistoryWithAcceptListProposal) {
		p.TrackingList = trackingList
	})

	submittedProposal, err := govKeeper.SubmitProposal(ctx, src)
	require.NoError(t, err)

	// execute proposal
	handler := govKeeper.Router().GetRoute(submittedProposal.ProposalRoute())
	err = handler(ctx, submittedProposal.GetContent())
	require.NoError(t, err)

	params = oracleKeeper.GetParams(ctx)
	require.Equal(t, params.AcceptList, trackingList)
	require.Equal(t, params.PriceTrackingList, trackingList)
}

func TestRemoveTrackingPriceHistoryProposal(t *testing.T) {
	ctx, keepers := CreateTestInput(t, false)
	govKeeper, oracleKeeper := keepers.GovKeeper, keepers.OracleKeeper

	params := oracleKeeper.GetParams(ctx)
	require.Equal(t, 2, len(params.AcceptList))
	require.Equal(t, 2, len(params.PriceTrackingList))

	trackingList := types.DenomList{
		{
			BaseDenom:   types.JunoDenom,
			SymbolDenom: types.JunoSymbol,
			Exponent:    types.JunoExponent,
		},
		{
			BaseDenom:   types.AtomDenom,
			SymbolDenom: types.AtomSymbol,
			Exponent:    types.AtomExponent,
		},
	}

	src := types.RemoveTrackingPriceHistoryProposalFixture(func(p *types.RemoveTrackingPriceHistoryProposal) {
		p.RemoveTrackingList = types.DenomList{trackingList[0]}
	})

	submittedProposal, err := govKeeper.SubmitProposal(ctx, src)
	require.NoError(t, err)
	handler := govKeeper.Router().GetRoute(submittedProposal.ProposalRoute())
	err = handler(ctx, submittedProposal.GetContent())
	require.NoError(t, err)

	params = oracleKeeper.GetParams(ctx)
	require.Equal(t, params.AcceptList, trackingList)
	require.Equal(t, params.PriceTrackingList, types.DenomList{trackingList[1]})
}
