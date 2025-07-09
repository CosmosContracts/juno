package staking

import (
	"context"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// KeeperInterface defines the interface for the keeper
type KeeperInterface interface {
	GetQueryContext() (context.Context, error)
	GetStakingQueryServer() stakingtypes.QueryServer
}
