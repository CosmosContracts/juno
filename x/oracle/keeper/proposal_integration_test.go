package keeper

import (
	"testing"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
)

func TestAddTrackingPriceHistoryProposal(t *testing.T) {
	parentCtx, keepers := CreateTestInput(t, false)
	govKeeper, oracleKeeper := keepers.GovKeeper, keepers.OracleKeeper

	var priceTrackingList types.DenomList
	params := types.DefaultParams()
	params.PriceTrackingList = priceTrackingList
	oracleKeeper.SetParams(parentCtx, params)

	src := types.
}

func TestAddTrackingPriceHistoryWithAcceptListProposal(t *testing.T) {

}

func TestRemoveTrackingPriceHistoryProposal(t *testing.T) {

}
