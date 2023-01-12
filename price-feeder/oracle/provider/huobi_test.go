package provider

import (
	"context"
	"testing"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	"github.com/CosmosContracts/juno/price-feeder/oracle/util"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestHuobiProvider_GetTickerPrices(t *testing.T) {
	p, err := NewHuobiProvider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
	)
	require.NoError(t, err)

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		lastPrice := 34.69000000
		volume := 2396974.02000000

		tickerMap := map[string]HuobiTicker{}
		tickerMap["market.atomusdt.ticker"] = HuobiTicker{
			CH: "market.atomusdt.ticker",
			Tick: HuobiTick{
				LastPrice: lastPrice,
				Vol:       volume,
			},
		}

		p.tickers = tickerMap

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, util.MustNewDecFromFloat(lastPrice), prices["ATOMUSDT"].Price)
		require.Equal(t, util.MustNewDecFromFloat(volume), prices["ATOMUSDT"].Volume)
	})

	t.Run("valid_request_multi_ticker", func(t *testing.T) {
		lastPriceAtom := 34.69000000
		lastPriceLuna := 41.35000000
		volume := 2396974.02000000

		tickerMap := map[string]HuobiTicker{}
		tickerMap["market.atomusdt.ticker"] = HuobiTicker{
			CH: "market.atomusdt.ticker",
			Tick: HuobiTick{
				LastPrice: lastPriceAtom,
				Vol:       volume,
			},
		}

		tickerMap["market.lunausdt.ticker"] = HuobiTicker{
			CH: "market.lunausdt.ticker",
			Tick: HuobiTick{
				LastPrice: lastPriceLuna,
				Vol:       volume,
			},
		}

		p.tickers = tickerMap
		prices, err := p.GetTickerPrices(
			types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
			types.CurrencyPair{Base: "LUNA", Quote: "USDT"},
		)
		require.NoError(t, err)
		require.Len(t, prices, 2)
		require.Equal(t, util.MustNewDecFromFloat(lastPriceAtom), prices["ATOMUSDT"].Price)
		require.Equal(t, util.MustNewDecFromFloat(volume), prices["ATOMUSDT"].Volume)
		require.Equal(t, util.MustNewDecFromFloat(lastPriceLuna), prices["LUNAUSDT"].Price)
		require.Equal(t, util.MustNewDecFromFloat(volume), prices["LUNAUSDT"].Volume)
	})

	t.Run("invalid_request_invalid_ticker", func(t *testing.T) {
		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "FOO", Quote: "BAR"})
		require.EqualError(t, err, "huobi failed to get ticker price for FOOBAR")
		require.Nil(t, prices)
	})
}

func TestHuobiProvider_SubscribeCurrencyPairs(t *testing.T) {
	p, err := NewHuobiProvider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
	)
	require.NoError(t, err)

	t.Run("invalid_subscribe_channels_empty", func(t *testing.T) {
		err = p.SubscribeCurrencyPairs([]types.CurrencyPair{}...)
		require.ErrorContains(t, err, "currency pairs is empty")
	})
}

func TestHuobiCurrencyPairToHuobiPair(t *testing.T) {
	cp := types.CurrencyPair{Base: "ATOM", Quote: "USDT"}
	binanceSymbol := currencyPairToHuobiTickerPair(cp)
	require.Equal(t, binanceSymbol, "market.atomusdt.ticker")
}
