package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"
)

func TestRegisterInterfaces(t *testing.T) {
	registry := types.NewInterfaceRegistry()
	RegisterInterfaces(registry)
	require.Equal(t, registry.ListAllInterfaces(), []string([]string{}))
}
