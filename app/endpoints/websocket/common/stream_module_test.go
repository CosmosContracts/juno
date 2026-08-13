package common

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CosmosContracts/juno/v31/app/endpoints"
	"github.com/CosmosContracts/juno/v31/x/stream/types/encoding"
)

func TestDescriptorResolverPrefersModuleHint(t *testing.T) {
	responseName := "QueryAccountResponse"
	resolver := newDescriptorResolverFromList([]*encoding.MethodDescriptor{
		{Module: "bank", StreamName: "account", ResponseType: "cosmos.bank.v1beta1." + responseName, Version: "v1beta1"},
		{Module: "auth", StreamName: "account", ResponseType: "cosmos.auth.v1beta1." + responseName, Version: "v1beta1"},
	})
	endpoint := endpoints.OpenAPIEndpoint{
		Path:           "/cosmos/auth/v1beta1/accounts/{address}",
		OperationID:    "auth_account",
		ResponseSchema: responseName,
	}

	desc := resolver.Resolve(endpoint)
	require.NotNil(t, desc)
	require.Equal(t, "auth", desc.Module)
}

func TestDescriptorResolverFallsBackToStreamName(t *testing.T) {
	responseName := "LatestBlockResponse"
	resolver := newDescriptorResolverFromList([]*encoding.MethodDescriptor{
		{Module: "base", StreamName: "block", ResponseType: "cosmos.base.v1." + responseName, Version: "v1beta1"},
		{Module: "base", StreamName: "latest_block", ResponseType: "cosmos.base.v1." + responseName, Version: "v1"},
	})
	endpoint := endpoints.OpenAPIEndpoint{
		Path:           "/cosmos/base/tendermint/v1beta1/blocks/latest",
		OperationID:    "base_latest_block",
		ResponseSchema: responseName,
	}

	desc := resolver.Resolve(endpoint)
	require.NotNil(t, desc)
	require.Equal(t, "latest_block", desc.StreamName)
}
