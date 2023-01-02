package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestJunoProvider_GetTickerPrices(t *testing.T) {
	p := NewJunoProvider(Endpoint{})

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			var resp string
			if req.URL.String() == "/prices/tokens/current" { //nolint:goconst,gocritic
				resp = `{
					"JUNO": {
					"date": "2022-11-02 18:57:36",
					"price": 2.93,
					"denom": "ujuno"
					},
					"AKT": {
					"date": "2022-11-02 18:57:36",
					"price": 0.26169002,
					"denom": "ibc/DFC6F33796D5D0075C5FB54A4D7B8E76915ACF434CB1EE2A1BA0BB8334E17C3A"
					},
					"ARTO": {
					"date": "2022-11-02 18:57:36",
					"price": 0.08182318000000001,
					"denom": "arto"
					}
				}
			`
			} else if req.URL.String() == "/volumes/tokens/JUNO/current" { //nolint:goconst
				resp = `
				{
					"date": "2022-11-07",
					"volumes": 51
				}
				`
			} else if req.URL.String() == "/volumes/tokens/AKT/current" { //nolint:goconst
				resp = `
				{
					"date": "2022-11-07",
					"volumes": 500
				}
				`
			} else if req.URL.String() == "/volumes/tokens/ARTO/current" { //nolint:goconst
				resp = `
				{
					"date": "2022-11-07",
					"volumes": 500
				}
				`
			}
			rw.Write([]byte(resp)) //nolint:errcheck
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "JUNO", Quote: "USDT"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr("2.93"), prices["JUNOUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("51"), prices["JUNOUSDT"].Volume)
	})

	t.Run("valid_request_multi_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			println("check req: ", req.URL.String())
			var resp string
			if req.URL.String() == "/prices/tokens/current" { //nolint:gocritic
				resp = `{
					"JUNO": {
					"date": "2022-11-02 18:57:36",
					"price": 2.93,
					"denom": "ujuno"
					},
					"AKT": {
					"date": "2022-11-02 18:57:36",
					"price": 0.26169002,
					"denom": "ibc/DFC6F33796D5D0075C5FB54A4D7B8E76915ACF434CB1EE2A1BA0BB8334E17C3A"
					},
					"ARTO": {
					"date": "2022-11-02 18:57:36",
					"price": 0.08182318000000001,
					"denom": "arto"
					}
				}
			`
			} else if req.URL.String() == "/volumes/tokens/JUNO/current" {
				resp = `
				{
					"date": "2022-11-07",
					"volumes": 5
				}
				`
			} else if req.URL.String() == "/volumes/tokens/AKT/current" {
				resp = `
				{
					"date": "2022-11-07",
					"volumes": 5
				}
				`
			} else if req.URL.String() == "/volumes/tokens/ARTO/current" {
				resp = `
				{
					"date": "2022-11-07",
					"volumes": 0
				}
				`
			}
			rw.Write([]byte(resp)) //nolint:errcheck
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(
			types.CurrencyPair{Base: "JUNO", Quote: "USDT"},
			types.CurrencyPair{Base: "AKT", Quote: "USDT"},
		)
		require.NoError(t, err)
		require.Len(t, prices, 2)
		require.Equal(t, sdk.MustNewDecFromStr("2.93"), prices["JUNOUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("5"), prices["JUNOUSDT"].Volume)
		require.Equal(t, sdk.MustNewDecFromStr("0.26169002"), prices["AKTUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("5"), prices["AKTUSDT"].Volume)
	})

	t.Run("invalid_request_bad_response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte(`FOO`)) //nolint:errcheck
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "JUNO", Quote: "USDT"})
		require.Error(t, err)
		require.Nil(t, prices)
	})

	t.Run("invalid_request_invalid_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			println("check req: ", req.URL.String())
			var resp string
			if req.URL.String() == "/prices/tokens/current" { //nolint:gocritic
				resp = `{
					"JUNO": {
					"date": "2022-11-02 18:57:36",
					"price": 2.93,
					"denom": "ujuno"
					},
					"AKT": {
					"date": "2022-11-02 18:57:36",
					"price": 0.26169002,
					"denom": "ibc/DFC6F33796D5D0075C5FB54A4D7B8E76915ACF434CB1EE2A1BA0BB8334E17C3A"
					},
					"ARTO": {
					"date": "2022-11-02 18:57:36",
					"price": 0.08182318000000001,
					"denom": "arto"
					}
				}
			`
			} else if req.URL.String() == "/volumes/tokens/JUNO/current" {
				resp = `
				{
					"date": "2022-11-07",
					"volumes": 5
				}
				`
			} else if req.URL.String() == "/volumes/tokens/AKT/current" {
				resp = `
				{
					"date": "2022-11-07",
					"volumes": 0
				}
				`
			} else if req.URL.String() == "/volumes/tokens/ARTO/current" {
				resp = `
				{
					"date": "2022-11-07",
					"volumes": 0
				}
				`
			}
			rw.Write([]byte(resp))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "FOO", Quote: "BAR"})
		require.Error(t, err)
		require.Nil(t, prices)
	})
}

func TestJunoProvider_GetAvailablePairs(t *testing.T) {
	p := NewJunoProvider(Endpoint{})
	_, err := p.GetAvailablePairs()
	require.Error(t, err)

	t.Run("valid_available_pair", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/summary/pairs", req.URL.String())
			resp := `[
					{
						"base": "JUNO",
						"target": "RAW"
					},
					{
						"base": "JUNO",
						"target": "ATOM"
					}
				]`
			rw.Write([]byte(resp))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL

		availablePairs, err := p.GetAvailablePairs()
		require.Nil(t, err)

		_, exist := availablePairs["JUNORAW"]
		require.True(t, exist)

		_, exist = availablePairs["JUNOATOM"]
		require.True(t, exist)
	})
}
