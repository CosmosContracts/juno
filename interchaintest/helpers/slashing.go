package helpers

import (
	"context"
	"encoding/json"
	"testing"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/stretchr/testify/require"
)

func GetSlashingParams(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) slashingtypes.Params {
	cmd := []string{
		"junod", "query", "slashing", "params",
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	results := slashingtypes.QueryParamsResponse{}
	err = json.Unmarshal(stdout, &results)
	require.NoError(t, err)

	t.Log(results)

	return results.Params
}
