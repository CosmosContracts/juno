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
		},
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

}

func TestRemoveTrackingPriceHistoryProposal(t *testing.T) {

}
