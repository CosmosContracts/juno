package interchaintest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBasicJunoStart is a basic test to assert that spinning up a Juno network with one validator works properly.
func TestBasicJunoStart(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Base setup
	chains := CreateThisBranchChain(t, 1, 0)
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	require.NotNil(t, ic)
	require.NotNil(t, ctx)

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
