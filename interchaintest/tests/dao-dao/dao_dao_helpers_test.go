package daodao_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildCw4DaoInstantiate(t *testing.T) {
	msg, err := buildDaoInstantiate("juno1member", "2", "3", "1")
	require.NoError(t, err)

	var core map[string]any
	require.NoError(t, json.Unmarshal([]byte(msg), &core))
	require.Equal(t, "DAO DAO cw4 lifecycle", core["name"])
	require.NotNil(t, core["voting_module_instantiate_info"])
	require.Len(t, core["proposal_modules_instantiate_info"], 1)
}

func TestCw4ArtifactChecksums(t *testing.T) {
	require.NoError(t, verifyCw4Artifacts(filepath.Join("..", "..", "contracts")))
}

func TestDecodeContractQueryResponse(t *testing.T) {
	t.Run("enveloped scalar", func(t *testing.T) {
		var address string
		require.NoError(t, decodeContractQueryResponse([]byte(`{"data":"juno1contract"}`), &address))
		require.Equal(t, "juno1contract", address)
	})

	t.Run("enveloped object", func(t *testing.T) {
		var response votingPowerResponse
		require.NoError(t, decodeContractQueryResponse([]byte(`{"data":{"power":"7"}}`), &response))
		require.Equal(t, "7", response.Power)
	})

	t.Run("direct response", func(t *testing.T) {
		var response votingPowerResponse
		require.NoError(t, decodeContractQueryResponse([]byte(`{"power":"9"}`), &response))
		require.Equal(t, "9", response.Power)
	})
}
