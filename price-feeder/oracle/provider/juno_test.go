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
			//require.Equal(t, "/prices/tokens/current", req.URL.String())
			println("check req: ", req.URL.String())
			var resp string
			if req.URL.String() == "/prices/tokens/current" {
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

		prices, err := p.GetTickerPrices(types.CurrencyPair{Base: "JUNO", Quote: "USDT"})
		require.NoError(t, err)
		require.Len(t, prices, 1)
		require.Equal(t, sdk.MustNewDecFromStr("2.93"), prices["JUNOUSDT"].Price)
		require.Equal(t, sdk.MustNewDecFromStr("5"), prices["JUNOUSDT"].Volume)
	})
}

func TestJunoProvider_GetAvailablePairs(t *testing.T) {
	p := NewJunoProvider(Endpoint{})
	p.GetAvailablePairs()

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
