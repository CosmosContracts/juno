package provider

import (
	"encoding/json"
	"fmt"
	_ "io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	_ "github.com/CosmosContracts/juno/price-feeder/oracle/util"
)

const (
	junoRestURL        = "https://api-junoswap.enigma-validator.com"
	junoTokenEndpoint  = "/prices/tokens"
	junoCandleEndpoint = "/prices/tokens/historical"
	junoPairsEndpoint  = "/summary/pairs"
)

var _ Provider = (*JunoProvider)(nil)

type (
	// JunoProvider defines an Oracle provider implemented by the Juno public
	// API.
	//
	// REF: https://api-osmosis.imperator.co/swagger/#/
	JunoProvider struct {
		baseURL string
		client  *http.Client
	}

	// JunoPairsSummary defines the response structure for an Juno pairs
	// summary.
	JunoPairsSummary struct {
		Data []JunoPairData
	}

	// JunoPairData defines the data response structure for an Juno pair.
	JunoPairData struct {
		Base  string `json:"base"`
		Quote string `json:"target"`
	}
)

func NewJunoProvider(endpoint Endpoint) *JunoProvider {
	if endpoint.Name == ProviderJuno {
		return &JunoProvider{
			baseURL: endpoint.Rest,
			client:  newDefaultHTTPClient(),
		}
	}
	return &JunoProvider{
		baseURL: junoRestURL,
		client:  newDefaultHTTPClient(),
	}
}

func (p JunoProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	return nil, nil
}

func (p JunoProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	return nil, nil
}

func (p JunoProvider) GetAvailablePairs() (map[string]struct{}, error) {
	path := fmt.Sprintf("%s%s", p.baseURL, junoPairsEndpoint)
	resp, err := p.client.Get(path)
	if err != nil {
		return nil, err
	}

	err = checkHTTPStatus(resp)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var pairsSummary []JunoPairData
	if err := json.Unmarshal(body, &pairsSummary); err != nil {
		return nil, err
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary))
	for _, pair := range pairsSummary {
		cp := types.CurrencyPair{
			Base:  strings.ToUpper(pair.Base),
			Quote: strings.ToUpper(pair.Quote),
		}
		availablePairs[cp.String()] = struct{}{}
	}

	return availablePairs, nil
}

// SubscribeCurrencyPairs performs a no-op since osmosis does not use websockets
func (p JunoProvider) SubscribeCurrencyPairs(pairs ...types.CurrencyPair) error {
	return nil
}
