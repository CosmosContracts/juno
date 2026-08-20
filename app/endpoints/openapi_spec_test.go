package endpoints_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CosmosContracts/juno/v31/app/endpoints"
)

func TestGetOpenAPIEndpointsIncludesAccountPath(t *testing.T) {
	endpointsList, err := endpoints.GetOpenAPIEndpoints()
	require.NoError(t, err)
	require.NotEmpty(t, endpointsList)

	var found bool
	for _, ep := range endpointsList {
		if ep.Path == "/cosmos/auth/v1beta1/accounts/{address}" {
			found = true
			require.Equal(t, "auth_account", ep.OperationID)
			require.Equal(t, "QueryAccountResponse", ep.ResponseSchema)
			require.NotEmpty(t, ep.Parameters)
			break
		}
	}
	require.True(t, found, "expected auth account endpoint in OpenAPI spec")
}
