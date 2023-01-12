package provider

import (
	"fmt"
	"net/http"
	"time"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
)

const (
	defaultTimeout           = 10 * time.Second
	defaultReadNewWSMessage  = 50 * time.Millisecond
	defaultMaxConnectionTime = time.Hour * 23 // should be < 24h
	defaultReconnectTime     = 2 * time.Minute
	defaultPingDuration      = 15 * time.Second
	providerCandlePeriod     = 10 * time.Minute

	ProviderKraken   Name = "kraken"
	ProviderBinance  Name = "binance"
	ProviderOsmosis  Name = "osmosis"
	ProviderJuno     Name = "juno"
	ProviderHuobi    Name = "huobi"
	ProviderOkx      Name = "okx"
	ProviderGate     Name = "gate"
	ProviderCoinbase Name = "coinbase"
	ProviderBitget   Name = "bitget"
	ProviderMexc     Name = "mexc"
	ProviderCrypto   Name = "crypto"
	ProviderMock     Name = "mock"
)

var ping = []byte("ping")

type (
	// Provider defines an interface an exchange price provider must implement.
	Provider interface {
		// GetTickerPrices returns the tickerPrices based on the provided pairs.
		GetTickerPrices(...types.CurrencyPair) (map[string]types.TickerPrice, error)

		// GetCandlePrices returns the candlePrices based on the provided pairs.
		GetCandlePrices(...types.CurrencyPair) (map[string][]types.CandlePrice, error)

		// GetAvailablePairs return all available pairs symbol to susbscribe.
		GetAvailablePairs() (map[string]struct{}, error)

		// SubscribeCurrencyPairs subscribe to ticker and candle channels for all pairs.
		SubscribeCurrencyPairs(...types.CurrencyPair) error
	}

	// Name name of an oracle provider. Usually it is an exchange
	// but this can be any provider name that can give token prices
	// examples.: "binance", "osmosis", "kraken".
	Name string

	// AggregatedProviderPrices defines a type alias for a map
	// of provider -> asset -> TickerPrice
	AggregatedProviderPrices map[Name]map[string]types.TickerPrice

	// AggregatedProviderCandles defines a type alias for a map
	// of provider -> asset -> []types.CandlePrice
	AggregatedProviderCandles map[Name]map[string][]types.CandlePrice

	// Endpoint defines an override setting in our config for the
	// hardcoded rest and websocket api endpoints.
	Endpoint struct {
		// Name of the provider, ex. "binance"
		Name Name `toml:"name"`

		// Rest endpoint for the provider, ex. "https://api1.binance.com"
		Rest string `toml:"rest"`

		// Websocket endpoint for the provider, ex. "stream.binance.com:9443"
		Websocket string `toml:"websocket"`
	}
)

// String cast provider name to string.
func (n Name) String() string {
	return string(n)
}

// preventRedirect avoid any redirect in the http.Client the request call
// will not return an error, but a valid response with redirect response code.
func preventRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func newDefaultHTTPClient() *http.Client {
	return newHTTPClientWithTimeout(defaultTimeout)
}

func newHTTPClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: preventRedirect,
	}
}

// PastUnixTime returns a millisecond timestamp that represents the unix time
// minus t.
func PastUnixTime(t time.Duration) int64 {
	return time.Now().Add(t*-1).Unix() * int64(time.Second/time.Millisecond)
}

// SecondsToMilli converts seconds to milliseconds for our unix timestamps.
func SecondsToMilli(t int64) int64 {
	return t * int64(time.Second/time.Millisecond)
}

func checkHTTPStatus(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}
