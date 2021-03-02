package keeper

import (
	"github.com/cosmoscontracts/juno/x/juno/types"
)

var _ types.QueryServer = Keeper{}
