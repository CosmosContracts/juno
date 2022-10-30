package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	"github.com/CosmosContracts/juno/price-feeder/oracle/util"
	"github.com/rs/zerolog"
)

const (
	ftxRestURL         = "https://ftx.com/api"
	ftxMarketsEndpoint = "/markets"
	ftxCandleEndpoint  = "/candles"
	ftxTimeFmt         = "2006-01-02T15:04:05+00:00"
	// candleWindowLength is the amount of seconds between
	// each candle
	candleWindowLength = 15
	cacheInterval      = 500 * time.Millisecond
)

var _ Provider = (*FTXProvider)(nil)

type (
	// FTXProvider defines an Oracle provider implemented by the FTX public
	// API.
	//
	// REF: https://docs.ftx.com/
	FTXProvider struct {
		baseURL string
		client  *http.Client

		logger zerolog.Logger
		mtx    sync.RWMutex

		// candleCache is the cache of candle prices for assets.
		candleCache map[string][]types.CandlePrice
		// marketsCache is the cache of token markets for assets.
		marketsCache []FTXMarkets
	}

	// FTXMarketsResponse is the response object used for
	// available exchange rates and tickers.
	FTXMarketsResponse struct {
		Success bool         `json:"success"`
		Markets []FTXMarkets `json:"result"`
	}
	FTXMarkets struct {
		Base   string  `json:"baseCurrency"`   // e.x. "BTC"
		Quote  string  `json:"quoteCurrency"`  // e.x. "USDT"
		Price  float64 `json:"price"`          // e.x. 10579.52
		Volume float64 `json:"quoteVolume24h"` // e.x. 28914.76
	}

	// FTXCandleResponse is the response object used for
	// candle information.
	FTXCandleResponse struct {
		Success bool        `json:"success"`
		Candle  []FTXCandle `json:"result"`
	}
	FTXCandle struct {
		Price     float64 `json:"close"`     // e.x. 11055.25
		Volume    float64 `json:"volume"`    // e.x. 464193.95725
		StartTime string  `json:"startTime"` // e.x. "2019-06-24T17:15:00+00:00"
	}
)

// parseTime parses a string such as "2022-08-29T20:23:00+00:00" into time.Time
func (c FTXCandle) parseTime() (time.Time, error) {
	t, err := time.Parse(ftxTimeFmt, c.StartTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse ftx timestamp %w", err)
	}
	return t, nil
}

func NewFTXProvider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoint Endpoint,
	pairs ...types.CurrencyPair,
) *FTXProvider {
	restURL := ftxRestURL

	if endpoint.Name == ProviderFTX {
		restURL = endpoint.Rest
	}

	ftx := FTXProvider{
		baseURL:      restURL,
		client:       newDefaultHTTPClient(),
		logger:       logger,
		candleCache:  nil,
		marketsCache: []FTXMarkets{},
	}

	go func() {
		logger.Debug().Msg("starting ftx polling...")
		err := ftx.pollCache(ctx, pairs...)
		if err != nil {
			logger.Err(err).Msg("ftx provider unable to poll new data")
		}
	}()

	return &ftx
}

// GetTickerPrices returns the tickerPrices based on the provided pairs.
func (p *FTXProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	markets := p.getMarketsCache()

	baseDenomIdx := make(map[string]types.CurrencyPair)
	for _, cp := range pairs {
		baseDenomIdx[strings.ToUpper(cp.Base)] = cp
	}

	tickerPrices := make(map[string]types.TickerPrice, len(pairs))
	for _, tr := range markets {
		symbol := strings.ToUpper(tr.Base)

		cp, ok := baseDenomIdx[symbol]
		if !ok {
			// skip tokens that are not requested
			continue
		}

		if _, ok := tickerPrices[symbol]; ok {
			return nil, fmt.Errorf("duplicate token found in FTX response: %s", symbol)
		}

		price, err := util.NewDecFromFloat(tr.Price)
		if err != nil {
			return nil, fmt.Errorf("failed to read FTX price (%f) for %s", tr.Price, symbol)
		}

		volume, err := util.NewDecFromFloat(tr.Volume)
		if err != nil {
			return nil, fmt.Errorf("failed to read FTX volume (%f) for %s", tr.Volume, symbol)
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

// GetCandlePrices returns the cached candlePrices based on provided pairs.
func (p *FTXProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candleCache := p.getCandleCache()
	if len(candleCache) < 1 {
		return nil, fmt.Errorf("candles have not been cached")
	}

	candlePrices := make(map[string][]types.CandlePrice, len(pairs))
	for _, pair := range pairs {
		if _, ok := candleCache[pair.String()]; !ok {
			return nil, fmt.Errorf("missing candles for %s", pair.String())
		}
		candlePrices[pair.String()] = candleCache[pair.String()]
	}

	return candlePrices, nil
}

// GetAvailablePairs return all available pairs symbol to susbscribe.
func (p *FTXProvider) GetAvailablePairs() (map[string]struct{}, error) {
	markets := p.getMarketsCache()
	availablePairs := make(map[string]struct{}, len(markets))
	for _, pair := range markets {
		cp := types.CurrencyPair{
			Base:  strings.ToUpper(pair.Base),
			Quote: strings.ToUpper(pair.Quote),
		}
		availablePairs[cp.String()] = struct{}{}
	}

	return availablePairs, nil
}

// SubscribeCurrencyPairs performs a no-op since ftx does not use websockets
func (p *FTXProvider) SubscribeCurrencyPairs(pairs ...types.CurrencyPair) error {
	return nil
}

// pollCache polls the markets and candles endpoints,
// and updates the ftx cache.
func (p *FTXProvider) pollCache(ctx context.Context, pairs ...types.CurrencyPair) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		default:
			p.logger.Debug().Msg("querying ftx api")

			err := p.pollMarkets()
			if err != nil {
				return err
			}
			err = p.pollCandles(pairs...)
			if err != nil {
				return err
			}

			time.Sleep(cacheInterval)
		}
	}
}

// pollMarkets retrieves the markets response from the ftx api and
// places it in p.marketsCache.
func (p *FTXProvider) pollMarkets() error {
	path := fmt.Sprintf("%s%s", p.baseURL, ftxMarketsEndpoint)

	resp, err := p.client.Get(path)
	if err != nil {
		return err
	}
	err = checkHTTPStatus(resp)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var pairsSummary FTXMarketsResponse
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return err
	}

	if !pairsSummary.Success {
		return fmt.Errorf("ftx markets api returned with failure")
	}

	p.setMarketsCache(pairsSummary.Markets)
	return nil
}

// pollMarkets retrieves the candles response from the ftx api and
// places it in p.candleCache.
func (p *FTXProvider) pollCandles(pairs ...types.CurrencyPair) error {
	candles := make(map[string][]types.CandlePrice)
	now := time.Now()

	for _, pair := range pairs {
		if _, ok := candles[pair.Base]; !ok {
			candles[pair.String()] = []types.CandlePrice{}
		}

		path := fmt.Sprintf("%s%s/%s/%s%s?resolution=%d&start_time=%d&end_time=%d",
			p.baseURL,
			ftxMarketsEndpoint,
			pair.Base,
			pair.Quote,
			ftxCandleEndpoint,
			candleWindowLength,
			now.Add(providerCandlePeriod*-1).Unix(),
			now.Unix(),
		)

		resp, err := p.client.Get(path)
		if err != nil {
			return fmt.Errorf("failed to make FTX candle request: %w", err)
		}
		err = checkHTTPStatus(resp)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		bz, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read FTX candle response body: %w", err)
		}

		var candlesResp FTXCandleResponse
		if err := json.Unmarshal(bz, &candlesResp); err != nil {
			return fmt.Errorf("failed to unmarshal FTX response body: %w", err)
		}

		candlePrices := []types.CandlePrice{}
		for _, responseCandle := range candlesResp.Candle {
			// the ftx api does not provide the endtime for these candles,
			// so we have to calculate it
			candleStart, err := responseCandle.parseTime()
			if err != nil {
				return err
			}
			candleEnd := candleStart.Add(candleWindowLength).Unix() * int64(time.Second/time.Millisecond)

			candlePrices = append(candlePrices, types.CandlePrice{
				Price:     util.MustNewDecFromFloat(responseCandle.Price),
				Volume:    util.MustNewDecFromFloat(responseCandle.Volume),
				TimeStamp: candleEnd,
			})
		}
		candles[pair.String()] = candlePrices
	}

	p.setCandleCache(candles)
	return nil
}

func (p *FTXProvider) setCandleCache(c map[string][]types.CandlePrice) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.candleCache = c
}

func (p *FTXProvider) getCandleCache() map[string][]types.CandlePrice {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	return p.candleCache
}

func (p *FTXProvider) setMarketsCache(m []FTXMarkets) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.marketsCache = m
}

func (p *FTXProvider) getMarketsCache() []FTXMarkets {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	return p.marketsCache
}
