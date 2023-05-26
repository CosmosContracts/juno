package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CosmosContracts/juno/v16/x/tokenfactory/types"
)

func TestDeconstructDenom(t *testing.T) {
	// Note: this seems to be used in osmosis to add some more checks (only 20 or 32 byte addresses),
	// which is good, but not required for these tests as they make code less reuable
	// appparams.SetAddressPrefixes()

	for _, tc := range []struct {
		desc             string
		denom            string
		expectedSubdenom string
		err              error
	}{
		{
			desc:  "empty is invalid",
			denom: "",
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "normal",
			denom:            "factory/juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym/bitcoin",
			expectedSubdenom: "bitcoin",
		},
		{
			desc:             "multiple slashes in subdenom",
			denom:            "factory/juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym/bitcoin/1",
			expectedSubdenom: "bitcoin/1",
		},
		{
			desc:             "no subdenom",
			denom:            "factory/juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym/",
			expectedSubdenom: "",
		},
		{
			desc:  "incorrect prefix",
			denom: "ibc/juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym/bitcoin",
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "subdenom of only slashes",
			denom:            "factory/juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym/////",
			expectedSubdenom: "////",
		},
		{
			desc:  "too long name",
			denom: "factory/juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym/adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			err:   types.ErrInvalidDenom,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			expectedCreator := "juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym"
			creator, subdenom, err := types.DeconstructDenom(tc.denom)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, expectedCreator, creator)
				require.Equal(t, tc.expectedSubdenom, subdenom)
			}
		})
	}
}

func TestGetTokenDenom(t *testing.T) {
	// appparams.SetAddressPrefixes()
	for _, tc := range []struct {
		desc     string
		creator  string
		subdenom string
		valid    bool
	}{
		{
			desc:     "normal",
			creator:  "juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym",
			subdenom: "bitcoin",
			valid:    true,
		},
		{
			desc:     "multiple slashes in subdenom",
			creator:  "juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym",
			subdenom: "bitcoin/1",
			valid:    true,
		},
		{
			desc:     "no subdenom",
			creator:  "juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym",
			subdenom: "",
			valid:    true,
		},
		{
			desc:     "subdenom of only slashes",
			creator:  "juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym",
			subdenom: "/////",
			valid:    true,
		},
		{
			desc:     "too long name",
			creator:  "juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym",
			subdenom: "adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			valid:    false,
		},
		{
			desc:     "subdenom is exactly max length",
			creator:  "juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym",
			subdenom: "bitcoinfsadfsdfeadfsafwefsefsefsdfsdafasefsf",
			valid:    true,
		},
		{
			desc:     "creator is exactly max length",
			creator:  "juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfymjhgjhgkhjklhkjhkjhgjhgjgjghelu",
			subdenom: "bitcoin",
			valid:    true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := types.GetTokenDenom(tc.creator, tc.subdenom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
