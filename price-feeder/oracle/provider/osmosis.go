package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	"github.com/CosmosContracts/juno/price-feeder/oracle/util"
)

const (
	osmosisRestURL        = "https://api-osmosis.imperator.co"
	osmosisTokenEndpoint  = "/tokens/v2"
	osmosisCandleEndpoint = "/tokens/v2/historical"
	osmosisPairsEndpoint  = "/pairs/v1/summary"
)

var _ Provider = (*OsmosisProvider)(nil)

type (
	// OsmosisProvider defines an Oracle provider implemented by the Osmosis public
	// API.
	//
	// REF: https://api-osmosis.imperator.co/swagger/
	OsmosisProvider struct {
		baseURL string
		client  *http.Client
	}

	// OsmosisTokenResponse defines the response structure for an Osmosis token
	// request.
	OsmosisTokenResponse struct {
		Price  float64 `json:"price"`
		Symbol string  `json:"symbol"`
		Volume float64 `json:"volume_24h"`
	}

	// OsmosisCandleResponse defines the response structure for an Osmosis candle
	// request.
	OsmosisCandleResponse struct {
		Time   int64   `json:"time"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}

	// OsmosisPairsSummary defines the response structure for an Osmosis pairs
	// summary.
	OsmosisPairsSummary struct {
		Data []OsmosisPairData `json:"data"`
	}

	// OsmosisPairData defines the data response structure for an Osmosis pair.
	OsmosisPairData struct {
		Base  string `json:"base_symbol"`
		Quote string `json:"quote_symbol"`
	}
)

func NewOsmosisProvider(endpoint Endpoint) *OsmosisProvider {
	if endpoint.Name == ProviderOsmosis {
		return &OsmosisProvider{
			baseURL: endpoint.Rest,
			client:  newDefaultHTTPClient(),
		}
	}
	return &OsmosisProvider{
		baseURL: osmosisRestURL,
		client:  newDefaultHTTPClient(),
	}
}

func (p OsmosisProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	path := fmt.Sprintf("%s%s/all", p.baseURL, osmosisTokenEndpoint)

	resp, err := p.client.Get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to make Osmosis request: %w", err)
	}
	err = checkHTTPStatus(resp)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Osmosis response body: %w", err)
	}

	var tokensResp []OsmosisTokenResponse
	if err := json.Unmarshal(bz, &tokensResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Osmosis response body: %w", err)
	}

	baseDenomIdx := make(map[string]types.CurrencyPair)
	for _, cp := range pairs {
		baseDenomIdx[strings.ToUpper(cp.Base)] = cp
	}

	tickerPrices := make(map[string]types.TickerPrice, len(pairs))
	for _, tr := range tokensResp {
		symbol := strings.ToUpper(tr.Symbol) // symbol == base in a currency pair

		cp, ok := baseDenomIdx[symbol]
		if !ok {
			// skip tokens that are not requested
			continue
		}

		if _, ok := tickerPrices[symbol]; ok {
			return nil, fmt.Errorf("duplicate token found in Osmosis response: %s", symbol)
		}

		price, err := util.NewDecFromFloat(tr.Price)
		if err != nil {
			return nil, fmt.Errorf("failed to read Osmosis price (%f) for %s", tr.Price, symbol)
		}

		volume, err := util.NewDecFromFloat(tr.Volume)
		if err != nil {
			return nil, fmt.Errorf("failed to read Osmosis volume (%f) for %s", tr.Volume, symbol)
		}

		tickerPrices[cp.String()] = types.TickerPrice{Price: price, Volume: volume}
	}

	for _, cp := range pairs {
		if _, ok := tickerPrices[cp.String()]; !ok {
			return nil, fmt.Errorf(types.ErrMissingExchangeRate.Error(), cp.String())
		}
	}

	return tickerPrices, nil
}

func (p OsmosisProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candles := make(map[string][]types.CandlePrice)
	for _, pair := range pairs {
		if _, ok := candles[pair.Base]; !ok {
			candles[pair.String()] = []types.CandlePrice{}
		}

		path := fmt.Sprintf("%s%s/%s/chart?tf=5", p.baseURL, osmosisCandleEndpoint, pair.Base)

		resp, err := p.client.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to make Osmosis request: %w", err)
		}
		err = checkHTTPStatus(resp)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		bz, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read Osmosis response body: %w", err)
		}

		var candlesResp []OsmosisCandleResponse
		if err := json.Unmarshal(bz, &candlesResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Osmosis response body: %w", err)
		}

		staleTime := PastUnixTime(providerCandlePeriod)

		candlePrices := []types.CandlePrice{}
		for _, responseCandle := range candlesResp {
			if staleTime >= responseCandle.Time {
				continue
			}
			candlePrices = append(candlePrices, types.CandlePrice{
				Price:  util.MustNewDecFromFloat(responseCandle.Close),
				Volume: util.MustNewDecFromFloat(responseCandle.Volume),
				// convert osmosis timestamp seconds -> milliseconds
				TimeStamp: SecondsToMilli(responseCandle.Time),
			})
		}
		candles[pair.String()] = candlePrices
	}

	return candles, nil
}

// GetAvailablePairs return all available pairs symbol to susbscribe.
func (p OsmosisProvider) GetAvailablePairs() (map[string]struct{}, error) {
	path := fmt.Sprintf("%s%s", p.baseURL, osmosisPairsEndpoint)

	resp, err := p.client.Get(path)
	if err != nil {
		return nil, err
	}
	err = checkHTTPStatus(resp)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary OsmosisPairsSummary
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary.Data))
	for _, pair := range pairsSummary.Data {
		cp := types.CurrencyPair{
			Base:  strings.ToUpper(pair.Base),
			Quote: strings.ToUpper(pair.Quote),
		}
		availablePairs[cp.String()] = struct{}{}
	}

	return availablePairs, nil
}

// SubscribeCurrencyPairs performs a no-op since osmosis does not use websockets
func (p OsmosisProvider) SubscribeCurrencyPairs(pairs ...types.CurrencyPair) error {
	return nil
}
