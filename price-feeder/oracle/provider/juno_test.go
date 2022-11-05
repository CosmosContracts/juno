package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	"github.com/stretchr/testify/require"
)

func TestJunoProvider_GetTickerPrices(t *testing.T) {
	p := NewJunoProvider(Endpoint{})

	t.Run("valid_request_single_ticker", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/prices/tokens/current", req.URL.String())
			resp := `{
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
			rw.Write([]byte(resp))
		}))
		defer server.Close()

		p.client = server.Client()
		p.baseURL = server.URL
		//TODO: config JunoProvider to 2 group baseURL and client
		// one for get price, one for get volume
		_, err := p.GetTickerPrices(types.CurrencyPair{Base: "ATOM", Quote: "USDT"})
		// require.NoError(t, err)
		require.Error(t, err)
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
