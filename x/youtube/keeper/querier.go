package keeper

import (
	"context"

	"github.com/CosmosContracts/juno/v19/x/youtube/types"
)

var _ types.QueryServer = &Querier{}

type Querier struct {
	keeper Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{
		keeper: k,
	}
}

// YoutubeContracts implements types.QueryServer.
func (Querier) YoutubeContracts(context.Context, *types.QueryYoutubemetadata) (*types.QueryYoutubemetadataResponse, error) {
	panic("unimplemented")
}
