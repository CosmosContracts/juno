package keeper

import (
	"github.com/CosmosContracts/Juno/x/juno/types"
)

var _ types.QueryServer = Keeper{}
