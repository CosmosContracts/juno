package provider

import (
	"context"
	"strconv"
	"testing"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestFTXProvider_GetTickerPrices(t *testing.T) {
	p := NewFTXProvider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
	)

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		lastPrice := "34.689998626708984000"
		volume := "2396974.000000000000000000"

		lp, _ := strconv.ParseFloat(lastPrice, 32)
		v, _ := strconv.ParseFloat(volume, 32)
		marketCache := []FTXMarkets{
			{
				Base:   "ATOM",
				Quote:  "USDT",
				Price:  lp,
				Volume: v,
			},
		}

		p.setMarketsCache(marketCache)

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr(lastPrice), prices["ATOMUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr(volume), prices["ATOMUSDT"].Volume)
	})

	t.Run("invalid_request_invalid_ticker", func(t *testing.T) {
		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "FOO", Quote: "BAR"})
		require.EqualError(t, err, "missing exchange rate for FOOBAR")
		require.Nil(t, prices)
	})
}

func TestFTXProvider_GetCandlePrices(t *testing.T) {
	p := NewFTXProvider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
	)

	t.Run("valid_request_single_candle", func(t *testing.T) {
		price := sdk.MustNewDecFromStr("34.689998626708984000")
		volume := sdk.MustNewDecFromStr("2396974.000000000000000000")
		timeStamp := int64(1000000)

		candleCache := map[string][]types.CandlePrice{
			"ATOMUSDT": {
				types.CandlePrice{
					TimeStamp: timeStamp,
					Price:     price,
					Volume:    volume,
				},
			},
		}

		p.setCandleCache(candleCache)

		prices, err := p.GetCandlePrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, price, prices["ATOMUSDT"][0].Price)
		require.Equal(t, volume, prices["ATOMUSDT"][0].Volume)
		require.Equal(t, timeStamp, prices["ATOMUSDT"][0].TimeStamp)
	})

	t.Run("invalid_request_invalid_candle", func(t *testing.T) {
		prices, err := p.GetCandlePrices(types.CurrencyPair{Base: "FOO", Quote: "BAR"})
		require.EqualError(t, err, "missing candles for FOOBAR")
		require.Nil(t, prices)
	})
}

func TestFTXProvider_SubscribeCurrencyPairs(t *testing.T) {
	p := NewFTXProvider(
		context.TODO(),
		zerolog.Nop(),
		Endpoint{},
		types.CurrencyPair{Base: "ATOM", Quote: "USDT"},
	)

	t.Run("invalid_subscribe_channels_empty", func(t *testing.T) {
		err := p.SubscribeCurrencyPairs([]types.CurrencyPair{}...)
		require.NoError(t, err)
	})
}
