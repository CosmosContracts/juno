package oracle

import (
	"testing"

	oracletypes "github.com/CosmosContracts/juno/v11/x/oracle/types"
	"github.com/stretchr/testify/require"
)

func TestParamCacheIsOutdated(t *testing.T) {
	testCases := map[string]struct {
		paramCache        ParamCache
		currentBlockHeigh int64
		expected          bool
	}{
		"Params Nil": {
			paramCache: ParamCache{
				params:           nil,
				lastUpdatedBlock: 0,
			},
			currentBlockHeigh: 10,
			expected:          true,
		},
		"currentBlockHeigh < cacheOnChainBlockQuantity": {
			paramCache: ParamCache{
				params:           &oracletypes.Params{},
				lastUpdatedBlock: 0,
			},
			currentBlockHeigh: 199,
			expected:          false,
		},
		"currentBlockHeigh < lastUpdatedBlock": {
			paramCache: ParamCache{
				params:           &oracletypes.Params{},
				lastUpdatedBlock: 205,
			},
			currentBlockHeigh: 203,
			expected:          true,
		},
		"Outdated": {
			paramCache: ParamCache{
				params:           &oracletypes.Params{},
				lastUpdatedBlock: 200,
			},
			currentBlockHeigh: 401,
			expected:          true,
		},
		"Limit to keep in cache": {
			paramCache: ParamCache{
				params:           &oracletypes.Params{},
				lastUpdatedBlock: 200,
			},
			currentBlockHeigh: 400,
			expected:          false,
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.paramCache.IsOutdated(tc.currentBlockHeigh))
		})
	}
}
