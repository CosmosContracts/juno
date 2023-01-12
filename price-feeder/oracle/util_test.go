package oracle_test

import (
	"testing"
	"time"

	"github.com/CosmosContracts/juno/price-feeder/oracle"
	"github.com/CosmosContracts/juno/price-feeder/oracle/provider"
	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestComputeVWAP(t *testing.T) {
	testCases := map[string]struct {
		prices   map[provider.Name]map[string]types.TickerPrice
		expected map[string]sdk.Dec
	}{
		"empty prices": {
			prices:   make(map[provider.Name]map[string]types.TickerPrice),
			expected: make(map[string]sdk.Dec),
		},
		"nil prices": {
			prices:   nil,
			expected: make(map[string]sdk.Dec),
		},
		"valid prices": {
			prices: map[provider.Name]map[string]types.TickerPrice{
				provider.ProviderBinance: {
					"ATOM": types.TickerPrice{
						Price:  sdk.MustNewDecFromStr("28.21000000"),
						Volume: sdk.MustNewDecFromStr("2749102.78000000"),
					},
					"UMEE": types.TickerPrice{
						Price:  sdk.MustNewDecFromStr("1.13000000"),
						Volume: sdk.MustNewDecFromStr("249102.38000000"),
					},
					"LUNA": types.TickerPrice{
						Price:  sdk.MustNewDecFromStr("64.87000000"),
						Volume: sdk.MustNewDecFromStr("7854934.69000000"),
					},
				},
				provider.ProviderKraken: {
					"ATOM": types.TickerPrice{
						Price:  sdk.MustNewDecFromStr("28.268700"),
						Volume: sdk.MustNewDecFromStr("178277.53314385"),
					},
					"LUNA": types.TickerPrice{
						Price:  sdk.MustNewDecFromStr("64.87853000"),
						Volume: sdk.MustNewDecFromStr("458917.46353577"),
					},
				},
				"FOO": {
					"ATOM": types.TickerPrice{
						Price:  sdk.MustNewDecFromStr("28.168700"),
						Volume: sdk.MustNewDecFromStr("4749102.53314385"),
					},
				},
			},
			expected: map[string]sdk.Dec{
				"ATOM": sdk.MustNewDecFromStr("28.185812745610043621"),
				"UMEE": sdk.MustNewDecFromStr("1.13000000"),
				"LUNA": sdk.MustNewDecFromStr("64.870470848638112395"),
			},
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			vwap := oracle.ComputeVWAP(tc.prices)
			require.Len(t, vwap, len(tc.expected))

			for k, v := range tc.expected {
				require.Equalf(t, v, vwap[k], "unexpected VWAP for %s", k)
			}
		})
	}
}

func TestComputeTVWAP(t *testing.T) {
	testCases := map[string]struct {
		candles  provider.AggregatedProviderCandles
		expected map[string]sdk.Dec
	}{
		"empty prices": {
			candles:  make(provider.AggregatedProviderCandles),
			expected: make(map[string]sdk.Dec),
		},
		"nil prices": {
			candles:  nil,
			expected: make(map[string]sdk.Dec),
		},
		"valid prices": {
			candles: map[provider.Name]map[string][]types.CandlePrice{
				provider.ProviderBinance: {
					"ATOM": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("25.09183"),
							Volume:    sdk.MustNewDecFromStr("98444.123455"),
							TimeStamp: provider.PastUnixTime(1 * time.Minute),
						},
					},
				},
				provider.ProviderKraken: {
					"ATOM": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("28.268700"),
							Volume:    sdk.MustNewDecFromStr("178277.53314385"),
							TimeStamp: provider.PastUnixTime(2 * time.Minute),
						},
					},
					"UMEE": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("1.13000000"),
							Volume:    sdk.MustNewDecFromStr("178277.53314385"),
							TimeStamp: provider.PastUnixTime(2 * time.Minute),
						},
					},
					"LUNA": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("64.87853000"),
							Volume:    sdk.MustNewDecFromStr("458917.46353577"),
							TimeStamp: provider.PastUnixTime(1 * time.Minute),
						},
					},
				},
				"FOO": {
					"ATOM": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("28.168700"),
							Volume:    sdk.MustNewDecFromStr("4749102.53314385"),
							TimeStamp: provider.PastUnixTime(130 * time.Second),
						},
					},
				},
			},
			expected: map[string]sdk.Dec{
				"ATOM": sdk.MustNewDecFromStr("28.045149332478338614"),
				"UMEE": sdk.MustNewDecFromStr("1.13000000"),
				"LUNA": sdk.MustNewDecFromStr("64.878530000000000000"),
			},
		},
		"one expired price": {
			candles: map[provider.Name]map[string][]types.CandlePrice{
				provider.ProviderBinance: {
					"ATOM": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("25.09183"),
							Volume:    sdk.MustNewDecFromStr("98444.123455"),
							TimeStamp: provider.PastUnixTime(1 * time.Minute),
						},
					},
				},
				provider.ProviderKraken: {
					"ATOM": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("28.268700"),
							Volume:    sdk.MustNewDecFromStr("178277.53314385"),
							TimeStamp: provider.PastUnixTime(2 * time.Minute),
						},
					},
					"UMEE": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("1.13000000"),
							Volume:    sdk.MustNewDecFromStr("178277.53314385"),
							TimeStamp: provider.PastUnixTime(2 * time.Minute),
						},
					},
					"LUNA": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("64.87853000"),
							Volume:    sdk.MustNewDecFromStr("458917.46353577"),
							TimeStamp: provider.PastUnixTime(1 * time.Minute),
						},
					},
				},
				"FOO": {
					"ATOM": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("28.168700"),
							Volume:    sdk.MustNewDecFromStr("4749102.53314385"),
							TimeStamp: provider.PastUnixTime(5 * time.Minute),
						},
					},
				},
			},
			expected: map[string]sdk.Dec{
				"ATOM": sdk.MustNewDecFromStr("26.601468076898424151"),
				"UMEE": sdk.MustNewDecFromStr("1.13000000"),
				"LUNA": sdk.MustNewDecFromStr("64.878530000000000000"),
			},
		},
		"all expired prices": {
			candles: map[provider.Name]map[string][]types.CandlePrice{
				provider.ProviderBinance: {
					"ATOM": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("25.09183"),
							Volume:    sdk.MustNewDecFromStr("98444.123455"),
							TimeStamp: provider.PastUnixTime(5 * time.Minute),
						},
					},
				},
				provider.ProviderKraken: {
					"ATOM": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("28.268700"),
							Volume:    sdk.MustNewDecFromStr("178277.53314385"),
							TimeStamp: provider.PastUnixTime(5 * time.Minute),
						},
					},
					"UMEE": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("1.13000000"),
							Volume:    sdk.MustNewDecFromStr("178277.53314385"),
							TimeStamp: provider.PastUnixTime(5 * time.Minute),
						},
					},
					"LUNA": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("64.87853000"),
							Volume:    sdk.MustNewDecFromStr("458917.46353577"),
							TimeStamp: provider.PastUnixTime(5 * time.Minute),
						},
					},
				},
				"FOO": {
					"ATOM": []types.CandlePrice{
						{
							Price:     sdk.MustNewDecFromStr("28.168700"),
							Volume:    sdk.MustNewDecFromStr("4749102.53314385"),
							TimeStamp: provider.PastUnixTime(5 * time.Minute),
						},
					},
				},
			},
			expected: map[string]sdk.Dec{},
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			vwap, err := oracle.ComputeTVWAP(tc.candles)
			require.NoError(t, err)
			require.Len(t, vwap, len(tc.expected))

			for k, v := range tc.expected {
				require.Equalf(t, v, vwap[k], "unexpected VWAP for %s", k)
			}
		})
	}
}

func TestStandardDeviation(t *testing.T) {
	type deviation struct {
		mean      sdk.Dec
		deviation sdk.Dec
	}
	testCases := map[string]struct {
		prices   map[provider.Name]map[string]sdk.Dec
		expected map[string]deviation
	}{
		"empty prices": {
			prices:   make(map[provider.Name]map[string]sdk.Dec),
			expected: map[string]deviation{},
		},
		"nil prices": {
			prices:   nil,
			expected: map[string]deviation{},
		},
		"not enough prices": {
			prices: map[provider.Name]map[string]sdk.Dec{
				provider.ProviderBinance: {
					"ATOM": sdk.MustNewDecFromStr("28.21000000"),
					"UMEE": sdk.MustNewDecFromStr("1.13000000"),
					"LUNA": sdk.MustNewDecFromStr("64.87000000"),
				},
				provider.ProviderKraken: {
					"ATOM": sdk.MustNewDecFromStr("28.23000000"),
					"UMEE": sdk.MustNewDecFromStr("1.13050000"),
					"LUNA": sdk.MustNewDecFromStr("64.85000000"),
				},
			},
			expected: map[string]deviation{},
		},
		"some prices": {
			prices: map[provider.Name]map[string]sdk.Dec{
				provider.ProviderBinance: {
					"ATOM": sdk.MustNewDecFromStr("28.21000000"),
					"UMEE": sdk.MustNewDecFromStr("1.13000000"),
					"LUNA": sdk.MustNewDecFromStr("64.87000000"),
				},
				provider.ProviderKraken: {
					"ATOM": sdk.MustNewDecFromStr("28.23000000"),
					"UMEE": sdk.MustNewDecFromStr("1.13050000"),
				},
				provider.ProviderOsmosis: {
					"ATOM": sdk.MustNewDecFromStr("28.40000000"),
					"UMEE": sdk.MustNewDecFromStr("1.14000000"),
					"LUNA": sdk.MustNewDecFromStr("64.10000000"),
				},
			},
			expected: map[string]deviation{
				"ATOM": {
					mean:      sdk.MustNewDecFromStr("28.28"),
					deviation: sdk.MustNewDecFromStr("0.085244745683629475"),
				},
				"UMEE": {
					mean:      sdk.MustNewDecFromStr("1.1335"),
					deviation: sdk.MustNewDecFromStr("0.004600724580614015"),
				},
			},
		},

		"non empty prices": {
			prices: map[provider.Name]map[string]sdk.Dec{
				provider.ProviderBinance: {
					"ATOM": sdk.MustNewDecFromStr("28.21000000"),
					"UMEE": sdk.MustNewDecFromStr("1.13000000"),
					"LUNA": sdk.MustNewDecFromStr("64.87000000"),
				},
				provider.ProviderKraken: {
					"ATOM": sdk.MustNewDecFromStr("28.23000000"),
					"UMEE": sdk.MustNewDecFromStr("1.13050000"),
					"LUNA": sdk.MustNewDecFromStr("64.85000000"),
				},
				provider.ProviderOsmosis: {
					"ATOM": sdk.MustNewDecFromStr("28.40000000"),
					"UMEE": sdk.MustNewDecFromStr("1.14000000"),
					"LUNA": sdk.MustNewDecFromStr("64.10000000"),
				},
			},
			expected: map[string]deviation{
				"ATOM": {
					mean:      sdk.MustNewDecFromStr("28.28"),
					deviation: sdk.MustNewDecFromStr("0.085244745683629475"),
				},
				"UMEE": {
					mean:      sdk.MustNewDecFromStr("1.1335"),
					deviation: sdk.MustNewDecFromStr("0.004600724580614015"),
				},
				"LUNA": {
					mean:      sdk.MustNewDecFromStr("64.606666666666666666"),
					deviation: sdk.MustNewDecFromStr("0.358360464089193609"),
				},
			},
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			deviation, mean, err := oracle.StandardDeviation(tc.prices)
			require.NoError(t, err)
			require.Len(t, deviation, len(tc.expected))
			require.Len(t, mean, len(tc.expected))

			for k, v := range tc.expected {
				require.Equalf(t, v.deviation, deviation[k], "unexpected deviation for %s", k)
				require.Equalf(t, v.mean, mean[k], "unexpected mean for %s", k)
			}
		})
	}
}
