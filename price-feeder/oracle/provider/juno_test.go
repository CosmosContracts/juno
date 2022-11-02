package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

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
