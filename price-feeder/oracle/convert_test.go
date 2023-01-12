package oracle

import (
	"testing"
	"time"

	"github.com/CosmosContracts/juno/price-feeder/oracle/provider"
	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

var (
	atomPrice  = sdk.MustNewDecFromStr("29.93")
	atomVolume = sdk.MustNewDecFromStr("894123.00")
	usdtPrice  = sdk.MustNewDecFromStr("0.98")
	usdtVolume = sdk.MustNewDecFromStr("894123.00")

	atomPair = types.CurrencyPair{
		Base:  "ATOM",
		Quote: "USDT",
	}
	usdtPair = types.CurrencyPair{
		Base:  "USDT",
		Quote: "USD",
	}
)

func TestGetUSDBasedProviders(t *testing.T) {
	providerPairs := make(map[provider.Name][]types.CurrencyPair, 3)
	providerPairs[provider.ProviderCoinbase] = []types.CurrencyPair{
		{
			Base:  "FOO",
			Quote: "USD",
		},
	}
	providerPairs[provider.ProviderHuobi] = []types.CurrencyPair{
		{
			Base:  "FOO",
			Quote: "USD",
		},
	}
	providerPairs[provider.ProviderKraken] = []types.CurrencyPair{
		{
			Base:  "FOO",
			Quote: "USDT",
		},
	}
	providerPairs[provider.ProviderBinance] = []types.CurrencyPair{
		{
			Base:  "USDT",
			Quote: "USD",
		},
	}

	pairs, err := getUSDBasedProviders("FOO", providerPairs)
	require.NoError(t, err)
	expectedPairs := map[provider.Name]struct{}{
		provider.ProviderCoinbase: {},
		provider.ProviderHuobi:    {},
	}
	require.Equal(t, pairs, expectedPairs)

	pairs, err = getUSDBasedProviders("USDT", providerPairs)
	require.NoError(t, err)
	expectedPairs = map[provider.Name]struct{}{
		provider.ProviderBinance: {},
	}
	require.Equal(t, pairs, expectedPairs)

	_, err = getUSDBasedProviders("BAR", providerPairs)
	require.Error(t, err)
}

func TestConvertCandlesToUSD(t *testing.T) {
	providerCandles := make(provider.AggregatedProviderCandles, 2)

	binanceCandles := map[string][]types.CandlePrice{
		"ATOM": {{
			Price:     atomPrice,
			Volume:    atomVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[provider.ProviderBinance] = binanceCandles

	krakenCandles := map[string][]types.CandlePrice{
		"USDT": {{
			Price:     usdtPrice,
			Volume:    usdtVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[provider.ProviderKraken] = krakenCandles

	providerPairs := map[provider.Name][]types.CurrencyPair{
		provider.ProviderBinance: {atomPair},
		provider.ProviderKraken:  {usdtPair},
	}

	convertedCandles, err := convertCandlesToUSD(
		zerolog.Nop(),
		providerCandles,
		providerPairs,
		make(map[string]sdk.Dec),
	)
	require.NoError(t, err)

	require.Equal(
		t,
		atomPrice.Mul(usdtPrice),
		convertedCandles[provider.ProviderBinance]["ATOM"][0].Price,
	)
}

func TestConvertCandlesToUSDFiltering(t *testing.T) {
	providerCandles := make(provider.AggregatedProviderCandles, 2)

	binanceCandles := map[string][]types.CandlePrice{
		"ATOM": {{
			Price:     atomPrice,
			Volume:    atomVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[provider.ProviderBinance] = binanceCandles

	krakenCandles := map[string][]types.CandlePrice{
		"USDT": {{
			Price:     usdtPrice,
			Volume:    usdtVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[provider.ProviderKraken] = krakenCandles

	gateCandles := map[string][]types.CandlePrice{
		"USDT": {{
			Price:     usdtPrice,
			Volume:    usdtVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[provider.ProviderGate] = gateCandles

	okxCandles := map[string][]types.CandlePrice{
		"USDT": {{
			Price:     sdk.MustNewDecFromStr("100.0"),
			Volume:    usdtVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		}},
	}
	providerCandles[provider.ProviderOkx] = okxCandles

	providerPairs := map[provider.Name][]types.CurrencyPair{
		provider.ProviderBinance: {atomPair},
		provider.ProviderKraken:  {usdtPair},
		provider.ProviderGate:    {usdtPair},
		provider.ProviderOkx:     {usdtPair},
	}

	convertedCandles, err := convertCandlesToUSD(
		zerolog.Nop(),
		providerCandles,
		providerPairs,
		make(map[string]sdk.Dec),
	)
	require.NoError(t, err)

	require.Equal(
		t,
		atomPrice.Mul(usdtPrice),
		convertedCandles[provider.ProviderBinance]["ATOM"][0].Price,
	)
}

func TestConvertTickersToUSD(t *testing.T) {
	providerPrices := make(provider.AggregatedProviderPrices, 2)

	binanceTickers := map[string]types.TickerPrice{
		"ATOM": {
			Price:  atomPrice,
			Volume: atomVolume,
		},
	}
	providerPrices[provider.ProviderBinance] = binanceTickers

	krakenTicker := map[string]types.TickerPrice{
		"USDT": {
			Price:  usdtPrice,
			Volume: usdtVolume,
		},
	}
	providerPrices[provider.ProviderKraken] = krakenTicker

	providerPairs := map[provider.Name][]types.CurrencyPair{
		provider.ProviderBinance: {atomPair},
		provider.ProviderKraken:  {usdtPair},
	}

	convertedTickers, err := convertTickersToUSD(
		zerolog.Nop(),
		providerPrices,
		providerPairs,
		make(map[string]sdk.Dec),
	)
	require.NoError(t, err)

	require.Equal(
		t,
		atomPrice.Mul(usdtPrice),
		convertedTickers[provider.ProviderBinance]["ATOM"].Price,
	)
}

func TestConvertTickersToUSDFiltering(t *testing.T) {
	providerPrices := make(provider.AggregatedProviderPrices, 2)

	binanceTickers := map[string]types.TickerPrice{
		"ATOM": {
			Price:  atomPrice,
			Volume: atomVolume,
		},
	}
	providerPrices[provider.ProviderBinance] = binanceTickers

	krakenTicker := map[string]types.TickerPrice{
		"USDT": {
			Price:  usdtPrice,
			Volume: usdtVolume,
		},
	}
	providerPrices[provider.ProviderKraken] = krakenTicker

	gateTicker := map[string]types.TickerPrice{
		"USDT": krakenTicker["USDT"],
	}
	providerPrices[provider.ProviderGate] = gateTicker

	huobiTicker := map[string]types.TickerPrice{
		"USDT": {
			Price:  sdk.MustNewDecFromStr("10000"),
			Volume: usdtVolume,
		},
	}
	providerPrices[provider.ProviderHuobi] = huobiTicker

	providerPairs := map[provider.Name][]types.CurrencyPair{
		provider.ProviderBinance: {atomPair},
		provider.ProviderKraken:  {usdtPair},
		provider.ProviderGate:    {usdtPair},
		provider.ProviderHuobi:   {usdtPair},
	}

	covertedDeviation, err := convertTickersToUSD(
		zerolog.Nop(),
		providerPrices,
		providerPairs,
		make(map[string]sdk.Dec),
	)
	require.NoError(t, err)

	require.Equal(
		t,
		atomPrice.Mul(usdtPrice),
		covertedDeviation[provider.ProviderBinance]["ATOM"].Price,
	)
}
