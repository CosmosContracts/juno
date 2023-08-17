package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CosmosContracts/juno/v17/x/clock/types"
)

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name     string
		params   types.Params
		expError bool
	}{
		{"default", types.DefaultParams(), false},
		{
			"valid: no contracts",
			types.NewParams([]string(nil)),
			false,
		},
		{
			"invalid: address malformed",
			types.NewParams([]string{"invalid address"}),
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}
