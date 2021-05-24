package keeper

import (
	"github.com/CosmosContracts/juno/x/juno/types"
)

var _ types.QueryServer = Keeper{}
