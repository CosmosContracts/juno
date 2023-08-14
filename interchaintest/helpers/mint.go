package helpers

import (
	"context"
	"encoding/json"
	"testing"

	minttypes "github.com/CosmosContracts/juno/v17/x/mint/types"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/stretchr/testify/require"
)

func GetMintParams(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) minttypes.Params {
	cmd := []string{
		"junod", "query", "mint", "params",
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	results := minttypes.QueryParamsResponse{}
	err = json.Unmarshal(stdout, &results)
	require.NoError(t, err)

	t.Log(results)

	return results.Params
}
