package provider

import (
	"encoding/json"
	"fmt"
	"io"
	_ "io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	"github.com/CosmosContracts/juno/price-feeder/oracle/util"
)

const (
	junoRestURL             = "https://api-junoswap.enigma-validator.com"
	junoPriceTokenEndpoint  = "/prices/tokens"
	junoVolumeTokenEndpoint = "/volumes/tokens"
	junoCandleEndpoint      = "/prices/tokens/historical"
	junoPairsEndpoint       = "/summary/pairs"
)

var _ Provider = (*JunoProvider)(nil)

type (
	// JunoProvider defines an Oracle provider implemented by the Juno public
	// API.
	//
	// REF: https://api-junoswap.enigma-validator.com/swagger/#/
	JunoProvider struct {
		baseURL string
		client  *http.Client
	}

	// JunoTokenPriceResponse defines the response structure of price for an Juno token
	// request.
	JunoTokenPriceResponse struct {
		Price float64 `json:"price"`
	}

	// JunoTokenVolumnResponse defines the response structure of volume for an Juno token
	// request.
	JunoTokenVolumnResponse struct {
		Volumne float64 `json:"volumes"`
	}

	// JunoTokenInfo defines the response structure of information of an Juno token
	// request.
	JunoTokenInfo struct {
		JunoTokenPriceResponse
		Symbol string
		Volume float64
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

	var junoTokenInfo []JunoTokenInfo
	// Get price and symbol of tokens
	pathPriceToken := fmt.Sprintf("%s%s/current", p.baseURL, junoPriceTokenEndpoint)
	resp, err := p.client.Get(pathPriceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to make Juno request: %w", err)
	}
	err = checkHTTPStatus(resp)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Juno response body: %w", err)
	}
	var tokensResps map[string]interface{}
	if err := json.Unmarshal(bz, &tokensResps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Juno response body: %w", err)
	}

	listSymbol := reflect.ValueOf(tokensResps).MapKeys()

	// Get symbol and price of Tokens
	for _, symbol := range listSymbol {
		tokenInfo, err := getSymbolAndPriceToken(symbol, tokensResps)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal Juno response body: %w", err)
		}
		junoTokenInfo = append(junoTokenInfo, tokenInfo)
	}

	// Get volume of tokens
	for id, symbol := range listSymbol {
		path := fmt.Sprintf("%s%s/%s/current", p.baseURL, junoVolumeTokenEndpoint, symbol)
		resp, err = p.client.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to make Juno request: %w", err)
		}
		err = checkHTTPStatus(resp)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		bz, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read Juno response body: %w", err)
		}
		var tokensResp JunoTokenVolumnResponse
		if err := json.Unmarshal(bz, &tokensResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Juno response body: %w", err)
		}
		junoTokenInfo[id].Volume = tokensResp.Volumne
	}

	baseDenomIdx := make(map[string]types.CurrencyPair)
	for _, cp := range pairs {
		baseDenomIdx[strings.ToUpper(cp.Base)] = cp
	}

	tickerPrices := make(map[string]types.TickerPrice, len(pairs))
	for _, tr := range junoTokenInfo {
		symbol := strings.ToUpper(tr.Symbol) // symbol == base in a currency pair

		cp, ok := baseDenomIdx[symbol]
		if !ok {
			// skip tokens that are not requested
			continue
		}

		if _, ok := tickerPrices[symbol]; ok {
			return nil, fmt.Errorf("duplicate token found in Juno response: %s", symbol)
		}

		price, err := util.NewDecFromFloat(tr.Price)
		if err != nil {
			return nil, fmt.Errorf("failed to read Juno price (%f) for %s", tr.Price, symbol)
		}

		volume, err := util.NewDecFromFloat(tr.Volume)
		if err != nil {
			return nil, fmt.Errorf("failed to read Juno volume (%f) for %s", tr.Volume, symbol)
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

// SubscribeCurrencyPairs performs a no-op since juno does not use websockets
func (p JunoProvider) SubscribeCurrencyPairs(pairs ...types.CurrencyPair) error {
	return nil
}

// Get symbol and price of token
func getSymbolAndPriceToken(symbol reflect.Value, tokensResps map[string]interface{}) (tokenInfo JunoTokenInfo, err error) {

	var tokenPrice JunoTokenPriceResponse
	tokenInfo.Symbol = symbol.String()
	dataOfToken := tokensResps[tokenInfo.Symbol]
	body, _ := json.Marshal(dataOfToken)
	if err := json.Unmarshal(body, &tokenPrice); err != nil {
		return tokenInfo, fmt.Errorf("failed to unmarshal Juno response body: %w", err)
	}
	tokenInfo.Price = tokenPrice.Price

	return tokenInfo, nil
}
