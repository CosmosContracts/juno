package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CosmosContracts/juno/v18/x/clock/types"
)

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name     string
		params   types.Params
		expError bool
	}{
		{"default", types.DefaultParams(), false},
		{
			"valid: no contracts, enough gas",
			types.NewParams(100_000),
			false,
		},
		{
			"invalid: address malformed",
			types.NewParams(100_000),
			true,
		},
		{
			"invalid: not enough gas",
			types.NewParams(1),
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
