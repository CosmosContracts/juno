package oracle

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/price-feeder/config"
	"github.com/CosmosContracts/juno/price-feeder/oracle/client"
	"github.com/CosmosContracts/juno/price-feeder/oracle/provider"
	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
)

type mockProvider struct {
	prices map[string]types.TickerPrice
}

func (m mockProvider) GetTickerPrices(_ ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	return m.prices, nil
}

func (m mockProvider) GetCandlePrices(_ ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candles := make(map[string][]types.CandlePrice)
	for pair, price := range m.prices {
		candles[pair] = []types.CandlePrice{
			{
				Price:     price.Price,
				TimeStamp: provider.PastUnixTime(1 * time.Minute),
				Volume:    price.Volume,
			},
		}
	}
	return candles, nil
}

func (m mockProvider) SubscribeCurrencyPairs(_ ...types.CurrencyPair) error {
	return nil
}

func (m mockProvider) GetAvailablePairs() (map[string]struct{}, error) {
	return map[string]struct{}{}, nil
}

type failingProvider struct {
	prices map[string]types.TickerPrice
}

func (m failingProvider) GetTickerPrices(_ ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	return nil, fmt.Errorf("unable to get ticker prices")
}

func (m failingProvider) GetCandlePrices(_ ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	return nil, fmt.Errorf("unable to get candle prices")
}

func (m failingProvider) SubscribeCurrencyPairs(_ ...types.CurrencyPair) error {
	return nil
}

func (m failingProvider) GetAvailablePairs() (map[string]struct{}, error) {
	return map[string]struct{}{}, nil
}

type OracleTestSuite struct {
	suite.Suite

	oracle *Oracle
}

// SetupSuite executes once before the suite's tests are executed.
func (ots *OracleTestSuite) SetupSuite() {
	ots.oracle = New(
		zerolog.Nop(),
		client.OracleClient{},
		[]config.CurrencyPair{
			{
				Base:      "JUNO",
				Quote:     "USDT",
				Providers: []provider.Name{provider.ProviderBinance},
			},
			{
				Base:      "JUNO",
				Quote:     "USDC",
				Providers: []provider.Name{provider.ProviderKraken},
			},
			{
				Base:      "XBT",
				Quote:     "USDT",
				Providers: []provider.Name{provider.ProviderOsmosis},
			},
			{
				Base:      "USDC",
				Quote:     "USD",
				Providers: []provider.Name{provider.ProviderHuobi},
			},
			{
				Base:      "USDT",
				Quote:     "USD",
				Providers: []provider.Name{provider.ProviderCoinbase},
			},
		},
		time.Millisecond*100,
		make(map[string]sdk.Dec),
		make(map[provider.Name]provider.Endpoint),
	)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OracleTestSuite))
}

func (ots *OracleTestSuite) TestStop() {
	ots.Eventually(
		func() bool {
			ots.oracle.Stop()
			return true
		},
		5*time.Second,
		time.Second,
	)
}

func (ots *OracleTestSuite) TestGetLastPriceSyncTimestamp() {
	// when no tick() has been invoked, assume zero value
	ots.Require().Equal(time.Time{}, ots.oracle.GetLastPriceSyncTimestamp())
}

func (ots *OracleTestSuite) TestPrices() {
	// initial prices should be empty (not set)
	ots.Require().Empty(ots.oracle.GetPrices())

	// Use a mock provider with exchange rates that are not specified in
	// configuration.
	ots.oracle.priceProviders = map[provider.Name]provider.Provider{
		provider.ProviderBinance: mockProvider{
			prices: map[string]types.TickerPrice{
				"JUNOUSDX": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		provider.ProviderKraken: mockProvider{
			prices: map[string]types.TickerPrice{
				"JUNOUSDX": {
					Price:  sdk.MustNewDecFromStr("3.70"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().Error(ots.oracle.SetPrices(context.TODO()))
	ots.Require().Empty(ots.oracle.GetPrices())

	// use a mock provider without a conversion rate for these stablecoins
	ots.oracle.priceProviders = map[provider.Name]provider.Provider{
		provider.ProviderBinance: mockProvider{
			prices: map[string]types.TickerPrice{
				"JUNOUSDT": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		provider.ProviderKraken: mockProvider{
			prices: map[string]types.TickerPrice{
				"JUNOUSDC": {
					Price:  sdk.MustNewDecFromStr("3.70"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().Error(ots.oracle.SetPrices(context.TODO()))

	prices := ots.oracle.GetPrices()
	ots.Require().Len(prices, 0)

	// use a mock provider to provide prices for the configured exchange pairs
	ots.oracle.priceProviders = map[provider.Name]provider.Provider{
		provider.ProviderBinance: mockProvider{
			prices: map[string]types.TickerPrice{
				"JUNOUSDT": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		provider.ProviderKraken: mockProvider{
			prices: map[string]types.TickerPrice{
				"JUNOUSDC": {
					Price:  sdk.MustNewDecFromStr("3.70"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		provider.ProviderHuobi: mockProvider{
			prices: map[string]types.TickerPrice{
				"USDCUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("2396974.34000000"),
				},
			},
		},
		provider.ProviderCoinbase: mockProvider{
			prices: map[string]types.TickerPrice{
				"USDTUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		provider.ProviderOsmosis: mockProvider{
			prices: map[string]types.TickerPrice{
				"XBTUSDT": {
					Price:  sdk.MustNewDecFromStr("3.717"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().NoError(ots.oracle.SetPrices(context.TODO()))

	prices = ots.oracle.GetPrices()
	ots.Require().Len(prices, 4)
	ots.Require().Equal(sdk.MustNewDecFromStr("3.710916056220858266"), prices["JUNO"])
	ots.Require().Equal(sdk.MustNewDecFromStr("3.717"), prices["XBT"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDC"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDT"])

	// use one working provider and one provider with an incorrect exchange rate
	ots.oracle.priceProviders = map[provider.Name]provider.Provider{
		provider.ProviderBinance: mockProvider{
			prices: map[string]types.TickerPrice{
				"JUNOUSDX": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		provider.ProviderKraken: mockProvider{
			prices: map[string]types.TickerPrice{
				"JUNOUSDC": {
					Price:  sdk.MustNewDecFromStr("3.70"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		provider.ProviderHuobi: mockProvider{
			prices: map[string]types.TickerPrice{
				"USDCUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("2396974.34000000"),
				},
			},
		},
		provider.ProviderCoinbase: mockProvider{
			prices: map[string]types.TickerPrice{
				"USDTUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		provider.ProviderOsmosis: mockProvider{
			prices: map[string]types.TickerPrice{
				"XBTUSDT": {
					Price:  sdk.MustNewDecFromStr("3.717"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().NoError(ots.oracle.SetPrices(context.TODO()))
	prices = ots.oracle.GetPrices()
	ots.Require().Len(prices, 4)
	ots.Require().Equal(sdk.MustNewDecFromStr("3.70"), prices["JUNO"])
	ots.Require().Equal(sdk.MustNewDecFromStr("3.717"), prices["XBT"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDC"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDT"])

	// use one working provider and one provider that fails
	ots.oracle.priceProviders = map[provider.Name]provider.Provider{
		provider.ProviderBinance: failingProvider{
			prices: map[string]types.TickerPrice{
				"JUNOUSDC": {
					Price:  sdk.MustNewDecFromStr("3.72"),
					Volume: sdk.MustNewDecFromStr("2396974.02000000"),
				},
			},
		},
		provider.ProviderKraken: mockProvider{
			prices: map[string]types.TickerPrice{
				"JUNOUSDC": {
					Price:  sdk.MustNewDecFromStr("3.71"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		provider.ProviderHuobi: mockProvider{
			prices: map[string]types.TickerPrice{
				"USDCUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("2396974.34000000"),
				},
			},
		},
		provider.ProviderCoinbase: mockProvider{
			prices: map[string]types.TickerPrice{
				"USDTUSD": {
					Price:  sdk.MustNewDecFromStr("1"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
		provider.ProviderOsmosis: mockProvider{
			prices: map[string]types.TickerPrice{
				"XBTUSDT": {
					Price:  sdk.MustNewDecFromStr("3.717"),
					Volume: sdk.MustNewDecFromStr("1994674.34000000"),
				},
			},
		},
	}

	ots.Require().NoError(ots.oracle.SetPrices(context.TODO()))
	prices = ots.oracle.GetPrices()
	ots.Require().Len(prices, 4)
	ots.Require().Equal(sdk.MustNewDecFromStr("3.71"), prices["JUNO"])
	ots.Require().Equal(sdk.MustNewDecFromStr("3.717"), prices["XBT"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDC"])
	ots.Require().Equal(sdk.MustNewDecFromStr("1"), prices["USDT"])
}

func TestGenerateSalt(t *testing.T) {
	salt, err := GenerateSalt(0)
	require.Error(t, err)
	require.Empty(t, salt)

	salt, err = GenerateSalt(32)
	require.NoError(t, err)
	require.NotEmpty(t, salt)
}

func TestGenerateExchangeRatesString(t *testing.T) {
	testCases := map[string]struct {
		input    map[string]sdk.Dec
		expected string
	}{
		"empty input": {
			input:    make(map[string]sdk.Dec),
			expected: "",
		},
		"single denom": {
			input: map[string]sdk.Dec{
				"JUNO": sdk.MustNewDecFromStr("3.72"),
			},
			expected: "JUNO:3.720000000000000000",
		},
		"multi denom": {
			input: map[string]sdk.Dec{
				"JUNO": sdk.MustNewDecFromStr("3.72"),
				"ATOM": sdk.MustNewDecFromStr("40.13"),
				"OSMO": sdk.MustNewDecFromStr("8.69"),
			},
			expected: "ATOM:40.130000000000000000,OSMO:8.690000000000000000,JUNO:3.720000000000000000",
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			out := GenerateExchangeRatesString(tc.input)
			require.Equal(t, tc.expected, out)
		})
	}
}

func TestSuccessSetProviderTickerPricesAndCandles(t *testing.T) {
	providerPrices := make(provider.AggregatedProviderPrices, 1)
	providerCandles := make(provider.AggregatedProviderCandles, 1)
	pair := types.CurrencyPair{
		Base:  "ATOM",
		Quote: "USDT",
	}

	atomPrice := sdk.MustNewDecFromStr("29.93")
	atomVolume := sdk.MustNewDecFromStr("894123.00")

	prices := make(map[string]types.TickerPrice, 1)
	prices[pair.String()] = types.TickerPrice{
		Price:  atomPrice,
		Volume: atomVolume,
	}

	candles := make(map[string][]types.CandlePrice, 1)
	candles[pair.String()] = []types.CandlePrice{
		{
			Price:     atomPrice,
			Volume:    atomVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}

	success := SetProviderTickerPricesAndCandles(
		provider.ProviderGate,
		providerPrices,
		providerCandles,
		prices,
		candles,
		pair,
	)

	require.True(t, success, "It should successfully set the prices")
	require.Equal(t, atomPrice, providerPrices[provider.ProviderGate][pair.Base].Price)
	require.Equal(t, atomPrice, providerCandles[provider.ProviderGate][pair.Base][0].Price)
}

func TestFailedSetProviderTickerPricesAndCandles(t *testing.T) {
	success := SetProviderTickerPricesAndCandles(
		provider.ProviderCoinbase,
		make(provider.AggregatedProviderPrices, 1),
		make(provider.AggregatedProviderCandles, 1),
		make(map[string]types.TickerPrice, 1),
		make(map[string][]types.CandlePrice, 1),
		types.CurrencyPair{
			Base:  "ATOM",
			Quote: "USDT",
		},
	)

	require.False(t, success, "It should failed to set the prices, prices and candle are empty")
}

func (ots *OracleTestSuite) TestSuccessGetComputedPricesCandles() {
	providerCandles := make(provider.AggregatedProviderCandles, 1)
	pair := types.CurrencyPair{
		Base:  "ATOM",
		Quote: "USD",
	}

	atomPrice := sdk.MustNewDecFromStr("29.93")
	atomVolume := sdk.MustNewDecFromStr("894123.00")

	candles := make(map[string][]types.CandlePrice, 1)
	candles[pair.Base] = []types.CandlePrice{
		{
			Price:     atomPrice,
			Volume:    atomVolume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	providerCandles[provider.ProviderBinance] = candles

	providerPair := map[provider.Name][]types.CurrencyPair{
		provider.ProviderBinance: {pair},
	}

	prices, err := ots.oracle.GetComputedPrices(
		providerCandles,
		make(provider.AggregatedProviderPrices, 1),
		providerPair,
		make(map[string]sdk.Dec),
	)

	require.NoError(ots.T(), err, "It should successfully get computed candle prices")
	require.Equal(ots.T(), prices[pair.Base], atomPrice)
}

func (ots *OracleTestSuite) TestSuccessGetComputedPricesTickers() {
	providerPrices := make(provider.AggregatedProviderPrices, 1)
	pair := types.CurrencyPair{
		Base:  "ATOM",
		Quote: "USD",
	}

	atomPrice := sdk.MustNewDecFromStr("29.93")
	atomVolume := sdk.MustNewDecFromStr("894123.00")

	tickerPrices := make(map[string]types.TickerPrice, 1)
	tickerPrices[pair.Base] = types.TickerPrice{
		Price:  atomPrice,
		Volume: atomVolume,
	}
	providerPrices[provider.ProviderBinance] = tickerPrices

	providerPair := map[provider.Name][]types.CurrencyPair{
		provider.ProviderBinance: {pair},
	}

	prices, err := ots.oracle.GetComputedPrices(
		make(provider.AggregatedProviderCandles, 1),
		providerPrices,
		providerPair,
		make(map[string]sdk.Dec),
	)

	require.NoError(ots.T(), err, "It should successfully get computed ticker prices")
	require.Equal(ots.T(), prices[pair.Base], atomPrice)
}

func (ots *OracleTestSuite) TestGetComputedPricesCandlesConversion() {
	btcPair := types.CurrencyPair{
		Base:  "BTC",
		Quote: "ETH",
	}
	btcUSDPair := types.CurrencyPair{
		Base:  "BTC",
		Quote: "USD",
	}
	ethPair := types.CurrencyPair{
		Base:  "ETH",
		Quote: "USD",
	}
	btcEthPrice := sdk.MustNewDecFromStr("17.55")
	btcUSDPrice := sdk.MustNewDecFromStr("20962.601")
	ethUsdPrice := sdk.MustNewDecFromStr("1195.02")
	volume := sdk.MustNewDecFromStr("894123.00")
	providerCandles := make(provider.AggregatedProviderCandles, 4)

	// normal rates
	binanceCandles := make(map[string][]types.CandlePrice, 2)
	binanceCandles[btcPair.Base] = []types.CandlePrice{
		{
			Price:     btcEthPrice,
			Volume:    volume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	binanceCandles[ethPair.Base] = []types.CandlePrice{
		{
			Price:     ethUsdPrice,
			Volume:    volume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	providerCandles[provider.ProviderBinance] = binanceCandles

	// normal rates
	gateCandles := make(map[string][]types.CandlePrice, 1)
	gateCandles[ethPair.Base] = []types.CandlePrice{
		{
			Price:     ethUsdPrice,
			Volume:    volume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	gateCandles[btcPair.Base] = []types.CandlePrice{
		{
			Price:     btcEthPrice,
			Volume:    volume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	providerCandles[provider.ProviderGate] = gateCandles

	// abnormal eth rate
	okxCandles := make(map[string][]types.CandlePrice, 1)
	okxCandles[ethPair.Base] = []types.CandlePrice{
		{
			Price:     sdk.MustNewDecFromStr("1.0"),
			Volume:    volume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	providerCandles[provider.ProviderOkx] = okxCandles

	// btc / usd rate
	krakenCandles := make(map[string][]types.CandlePrice, 1)
	krakenCandles[btcUSDPair.Base] = []types.CandlePrice{
		{
			Price:     btcUSDPrice,
			Volume:    volume,
			TimeStamp: provider.PastUnixTime(1 * time.Minute),
		},
	}
	providerCandles[provider.ProviderKraken] = krakenCandles

	providerPair := map[provider.Name][]types.CurrencyPair{
		provider.ProviderBinance: {btcPair, ethPair},
		provider.ProviderGate:    {ethPair},
		provider.ProviderOkx:     {ethPair},
		provider.ProviderKraken:  {btcUSDPair},
	}

	prices, err := ots.oracle.GetComputedPrices(
		providerCandles,
		make(provider.AggregatedProviderPrices, 1),
		providerPair,
		make(map[string]sdk.Dec),
	)

	require.NoError(ots.T(), err,
		"It should successfully filter out bad candles and convert everything to USD",
	)
	require.Equal(ots.T(),
		ethUsdPrice.Mul(
			btcEthPrice).Add(btcUSDPrice).Quo(sdk.MustNewDecFromStr("2")),
		prices[btcPair.Base],
	)
}

func (ots *OracleTestSuite) TestGetComputedPricesTickersConversion() {
	btcPair := types.CurrencyPair{
		Base:  "BTC",
		Quote: "ETH",
	}
	btcUSDPair := types.CurrencyPair{
		Base:  "BTC",
		Quote: "USD",
	}
	ethPair := types.CurrencyPair{
		Base:  "ETH",
		Quote: "USD",
	}
	volume := sdk.MustNewDecFromStr("881272.00")
	btcEthPrice := sdk.MustNewDecFromStr("72.55")
	ethUsdPrice := sdk.MustNewDecFromStr("9989.02")
	btcUSDPrice := sdk.MustNewDecFromStr("724603.401")
	providerPrices := make(provider.AggregatedProviderPrices, 1)

	// normal rates
	binanceTickerPrices := make(map[string]types.TickerPrice, 2)
	binanceTickerPrices[btcPair.Base] = types.TickerPrice{
		Price:  btcEthPrice,
		Volume: volume,
	}
	binanceTickerPrices[ethPair.Base] = types.TickerPrice{
		Price:  ethUsdPrice,
		Volume: volume,
	}
	providerPrices[provider.ProviderBinance] = binanceTickerPrices

	// normal rates
	gateTickerPrices := make(map[string]types.TickerPrice, 4)
	gateTickerPrices[btcPair.Base] = types.TickerPrice{
		Price:  btcEthPrice,
		Volume: volume,
	}
	gateTickerPrices[ethPair.Base] = types.TickerPrice{
		Price:  ethUsdPrice,
		Volume: volume,
	}
	providerPrices[provider.ProviderGate] = gateTickerPrices

	// abnormal eth rate
	okxTickerPrices := make(map[string]types.TickerPrice, 1)
	okxTickerPrices[ethPair.Base] = types.TickerPrice{
		Price:  sdk.MustNewDecFromStr("1.0"),
		Volume: volume,
	}
	providerPrices[provider.ProviderOkx] = okxTickerPrices

	// btc / usd rate
	krakenTickerPrices := make(map[string]types.TickerPrice, 1)
	krakenTickerPrices[btcUSDPair.Base] = types.TickerPrice{
		Price:  btcUSDPrice,
		Volume: volume,
	}
	providerPrices[provider.ProviderKraken] = krakenTickerPrices

	providerPair := map[provider.Name][]types.CurrencyPair{
		provider.ProviderBinance: {ethPair, btcPair},
		provider.ProviderGate:    {ethPair},
		provider.ProviderOkx:     {ethPair},
		provider.ProviderKraken:  {btcUSDPair},
	}

	prices, err := ots.oracle.GetComputedPrices(
		make(provider.AggregatedProviderCandles, 1),
		providerPrices,
		providerPair,
		make(map[string]sdk.Dec),
	)

	require.NoError(ots.T(), err,
		"It should successfully filter out bad tickers and convert everything to USD",
	)
	require.Equal(ots.T(),
		ethUsdPrice.Mul(
			btcEthPrice).Add(btcUSDPrice).Quo(sdk.MustNewDecFromStr("2")),
		prices[btcPair.Base],
	)
}

func (ots *OracleTestSuite) TestGetComputedPricesEmptyTvwap() {
	symbolUSDT := "USDT"
	symbolUSD := "USD"
	symbolDAI := "DAI"
	symbolETH := "ETH"

	pairETHtoUSDT := types.CurrencyPair{
		Base:  symbolETH,
		Quote: symbolUSDT,
	}
	pairETHtoDAI := types.CurrencyPair{
		Base:  symbolETH,
		Quote: symbolDAI,
	}
	pairETHtoUSD := types.CurrencyPair{
		Base:  symbolETH,
		Quote: symbolUSD,
	}
	basePairsETH := []types.CurrencyPair{
		pairETHtoUSDT,
		pairETHtoDAI,
	}
	krakenPairsETH := append(basePairsETH, pairETHtoUSD)

	pairUSDTtoUSD := types.CurrencyPair{
		Base:  symbolUSDT,
		Quote: symbolUSD,
	}
	pairDAItoUSD := types.CurrencyPair{
		Base:  symbolDAI,
		Quote: symbolUSD,
	}
	stablecoinPairs := []types.CurrencyPair{
		pairUSDTtoUSD,
		pairDAItoUSD,
	}

	krakenPairs := append(krakenPairsETH, stablecoinPairs...)

	volume := sdk.MustNewDecFromStr("881272.00")
	ethUsdPrice := sdk.MustNewDecFromStr("9989.02")
	daiUsdPrice := sdk.MustNewDecFromStr("999890000000000000")
	ethTime := provider.PastUnixTime(1 * time.Minute)

	ethCandle := []types.CandlePrice{
		{
			Price:     ethUsdPrice,
			Volume:    volume,
			TimeStamp: ethTime,
		},
		{
			Price:     ethUsdPrice,
			Volume:    volume,
			TimeStamp: ethTime,
		},
	}
	daiCandle := []types.CandlePrice{
		{
			Price:     daiUsdPrice,
			Volume:    volume,
			TimeStamp: 1660829520000,
		},
	}

	prices := provider.AggregatedProviderPrices{}

	pairs := map[provider.Name][]types.CurrencyPair{
		provider.ProviderKraken: krakenPairs,
	}

	testCases := map[string]struct {
		expected string
		candles  provider.AggregatedProviderCandles
		prices   provider.AggregatedProviderPrices
		pairs    map[provider.Name][]types.CurrencyPair
	}{
		"Empty tvwap": {
			candles: provider.AggregatedProviderCandles{
				provider.ProviderKraken: {
					"USDT": ethCandle,
					"ETH":  ethCandle,
					"DAI":  daiCandle,
				},
			},
			prices:   prices,
			pairs:    pairs,
			expected: "error on computing tvwap for quote: DAI, base: ETH",
		},
		"No valid conversion rates DAI": {
			candles: provider.AggregatedProviderCandles{
				provider.ProviderKraken: {
					"USDT": ethCandle,
					"ETH":  ethCandle,
				},
			},
			prices:   prices,
			pairs:    pairs,
			expected: "there are no valid conversion rates for DAI",
		},
	}

	for name, tc := range testCases {
		tc := tc

		ots.Run(name, func() {
			_, err := ots.oracle.GetComputedPrices(
				tc.candles,
				tc.prices,
				tc.pairs,
				make(map[string]sdk.Dec),
			)

			require.ErrorContains(ots.T(), err, tc.expected)
		})
	}
}
