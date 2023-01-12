package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrependJunoIfUnique(t *testing.T) {
	require := require.New(t)
	tcs := []struct {
		in  []string
		out []string
	}{
		// Should prepend "ujuno" to a slice of denoms, unless it is already present.
		{[]string{}, []string{"ujuno"}},
		{[]string{"a"}, []string{"ujuno", "a"}},
		{[]string{"x", "a", "heeeyyy"}, []string{"ujuno", "x", "a", "heeeyyy"}},
		{[]string{"x", "a", "ujuno"}, []string{"x", "a", "ujuno"}},
	}
	for i, tc := range tcs {
		require.Equal(tc.out, prependJunoIfUnique(tc.in), i)
	}
}
