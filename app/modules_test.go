package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	votingsnapshottypes "github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
)

func TestOrderMigrationsKeepsVotingSnapshotAfterStaking(t *testing.T) {
	order := orderMigrations([]string{
		votingsnapshottypes.ModuleName,
		"bank",
		stakingtypes.ModuleName,
	})

	stakingIndex, snapshotIndex := -1, -1
	for i, name := range order {
		switch name {
		case stakingtypes.ModuleName:
			stakingIndex = i
		case votingsnapshottypes.ModuleName:
			snapshotIndex = i
		}
	}

	require.NotEqual(t, -1, stakingIndex)
	require.NotEqual(t, -1, snapshotIndex)
	require.Less(t, stakingIndex, snapshotIndex)
}
