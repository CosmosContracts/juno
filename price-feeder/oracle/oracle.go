package oracle

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/CosmosContracts/juno/price-feeder/config"
	"github.com/CosmosContracts/juno/price-feeder/oracle/client"
	"github.com/CosmosContracts/juno/price-feeder/oracle/provider"
	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	pfsync "github.com/CosmosContracts/juno/price-feeder/pkg/sync"
	oracletypes "github.com/CosmosContracts/juno/v12/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

// We define tickerSleep as the minimum timeout between each oracle loop. We
// define this value empirically based on enough time to collect exchange rates,
// and broadcast pre-vote and vote transactions such that they're committed in
// at least one block during each voting period.
const (
	tickerSleep = 1000 * time.Millisecond
)

// PreviousPrevote defines a structure for defining the previous prevote
// submitted on-chain.
type PreviousPrevote struct {
	ExchangeRates     string
	Salt              string
	SubmitBlockHeight int64
}

func NewPreviousPrevote() *PreviousPrevote {
	return &PreviousPrevote{
		Salt:              "",
		ExchangeRates:     "",
		SubmitBlockHeight: 0,
	}
}

// Oracle implements the core component responsible for fetching exchange rates
// for a given set of currency pairs and determining the correct exchange rates
// to submit to the on-chain price oracle adhering the oracle specification.
type Oracle struct {
	logger zerolog.Logger
	closer *pfsync.Closer

	providerTimeout    time.Duration
	providerPairs      map[provider.Name][]types.CurrencyPair
	previousPrevote    *PreviousPrevote
	previousVotePeriod float64
	priceProviders     map[provider.Name]provider.Provider
	oracleClient       client.OracleClient
	deviations         map[string]sdk.Dec
	endpoints          map[provider.Name]provider.Endpoint
	paramCache         ParamCache

	pricesMutex     sync.RWMutex
	lastPriceSyncTS time.Time
	prices          map[string]sdk.Dec

	tvwapsByProvider PricesWithMutex
	vwapsByProvider  PricesWithMutex
}

func New(
	logger zerolog.Logger,
	oc client.OracleClient,
	currencyPairs []config.CurrencyPair,
	providerTimeout time.Duration,
	deviations map[string]sdk.Dec,
	endpoints map[provider.Name]provider.Endpoint,
) *Oracle {
	providerPairs := make(map[provider.Name][]types.CurrencyPair)

	for _, pair := range currencyPairs {
		for _, provider := range pair.Providers {
			providerPairs[provider] = append(providerPairs[provider], types.CurrencyPair{
				Base:  pair.Base,
				Quote: pair.Quote,
			})
		}
	}

	return &Oracle{
		logger:          logger.With().Str("module", "oracle").Logger(),
		closer:          pfsync.NewCloser(),
		oracleClient:    oc,
		providerPairs:   providerPairs,
		priceProviders:  make(map[provider.Name]provider.Provider),
		previousPrevote: nil,
		providerTimeout: providerTimeout,
		deviations:      deviations,
		paramCache:      ParamCache{},
		endpoints:       endpoints,
	}
}

// Start starts the oracle process in a blocking fashion.
func (o *Oracle) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			o.closer.Close()

		default:
			o.logger.Debug().Msg("starting oracle tick")

			startTime := time.Now()

			if err := o.tick(ctx); err != nil {
				telemetry.IncrCounter(1, "failure", "tick")
				o.logger.Err(err).Msg("oracle tick failed")
			}

			o.lastPriceSyncTS = time.Now()

			telemetry.MeasureSince(startTime, "runtime", "tick")
			telemetry.IncrCounter(1, "new", "tick")

			time.Sleep(tickerSleep)
		}
	}
}

// Stop stops the oracle process and waits for it to gracefully exit.
func (o *Oracle) Stop() {
	o.closer.Close()
	<-o.closer.Done()
}

// GetLastPriceSyncTimestamp returns the latest timestamp at which prices where
// fetched from the oracle's set of exchange rate providers.
func (o *Oracle) GetLastPriceSyncTimestamp() time.Time {
	o.pricesMutex.RLock()
	defer o.pricesMutex.RUnlock()

	return o.lastPriceSyncTS
}

// GetPrices returns a copy of the current prices fetched from the oracle's
// set of exchange rate providers.
func (o *Oracle) GetPrices() map[string]sdk.Dec {
	o.pricesMutex.RLock()
	defer o.pricesMutex.RUnlock()

	// Creates a new array for the prices in the oracle
	prices := make(map[string]sdk.Dec, len(o.prices))
	for k, v := range o.prices {
		// Fills in the prices with each value in the oracle
		prices[k] = v
	}

	return prices
}

// GetTvwapPrices returns a copy of the tvwapsByProvider map
func (o *Oracle) GetTvwapPrices() PricesByProvider {
	return o.tvwapsByProvider.GetPricesClone()
}

// GetVwapPrices returns the vwapsByProvider map using a read lock
func (o *Oracle) GetVwapPrices() PricesByProvider {
	return o.vwapsByProvider.GetPricesClone()
}

// SetPrices retrieves all the prices and candles from our set of providers as
// determined in the config. If candles are available, uses TVWAP in order
// to determine prices. If candles are not available, uses the most recent prices
// with VWAP. Warns the the user of any missing prices, and filters out any faulty
// providers which do not report prices or candles within 2𝜎 of the others.
func (o *Oracle) SetPrices(ctx context.Context) error {
	g := new(errgroup.Group)
	mtx := new(sync.Mutex)
	providerPrices := make(provider.AggregatedProviderPrices)
	providerCandles := make(provider.AggregatedProviderCandles)
	requiredRates := make(map[string]struct{})

	for providerName, currencyPairs := range o.providerPairs {
		providerName := providerName
		currencyPairs := currencyPairs

		priceProvider, err := o.getOrSetProvider(ctx, providerName)
		if err != nil {
			return err
		}

		for _, pair := range currencyPairs {
			if _, ok := requiredRates[pair.Base]; !ok {
				requiredRates[pair.Base] = struct{}{}
			}
		}

		g.Go(func() error {
			prices := make(map[string]types.TickerPrice, 0)
			candles := make(map[string][]types.CandlePrice, 0)
			ch := make(chan struct{})
			errCh := make(chan error, 1)

			go func() {
				defer close(ch)
				prices, err = priceProvider.GetTickerPrices(currencyPairs...)
				if err != nil {
					provider.TelemetryFailure(providerName, provider.MessageTypeTicker)
					errCh <- err
				}

				candles, err = priceProvider.GetCandlePrices(currencyPairs...)
				if err != nil {
					provider.TelemetryFailure(providerName, provider.MessageTypeCandle)
					errCh <- err
				}
			}()

			select {
			case <-ch:
				break
			case err := <-errCh:
				return err
			case <-time.After(o.providerTimeout):
				telemetry.IncrCounter(1, "failure", "provider", "type", "timeout")
				return fmt.Errorf("provider timed out")
			}

			// flatten and collect prices based on the base currency per provider
			//
			// e.g.: {ProviderKraken: {"ATOM": <price, volume>, ...}}
			mtx.Lock()
			for _, pair := range currencyPairs {
				success := SetProviderTickerPricesAndCandles(providerName, providerPrices, providerCandles, prices, candles, pair)
				if !success {
					mtx.Unlock()
					return fmt.Errorf("failed to find any exchange rates in provider responses")
				}
			}

			mtx.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		o.logger.Err(err).Msg("failed to get ticker prices from provider")
	}

	computedPrices, err := o.GetComputedPrices(
		providerCandles,
		providerPrices,
		o.providerPairs,
		o.deviations,
	)
	if err != nil {
		return err
	}

	if len(computedPrices) != len(requiredRates) {
		return fmt.Errorf("unable to get prices for all exchange candles")
	}
	for base := range requiredRates {
		if _, ok := computedPrices[base]; !ok {
			return fmt.Errorf("reported prices were not equal to required rates, missed: %s", base)
		}
	}

	o.pricesMutex.Lock()
	o.prices = computedPrices
	o.pricesMutex.Unlock()
	return nil
}

// GetComputedPrices gets the candle and ticker prices and computes it.
// It returns candles' TVWAP if possible, if not possible (not available
// or due to some staleness) it will use the most recent ticker prices
// and the VWAP formula instead.
func (o *Oracle) GetComputedPrices(
	providerCandles provider.AggregatedProviderCandles,
	providerPrices provider.AggregatedProviderPrices,
	providerPairs map[provider.Name][]types.CurrencyPair,
	deviations map[string]sdk.Dec,
) (prices map[string]sdk.Dec, err error) {

	// convert any non-USD denominated candles into USD
	convertedCandles, err := convertCandlesToUSD(
		o.logger,
		providerCandles,
		providerPairs,
		deviations,
	)
	if err != nil {
		return nil, err
	}

	// filter out any erroneous candles
	filteredCandles, err := FilterCandleDeviations(
		o.logger,
		convertedCandles,
		deviations,
	)
	if err != nil {
		return nil, err
	}

	computedPrices, _ := ComputeTvwapsByProvider(filteredCandles)
	o.tvwapsByProvider.SetPrices(computedPrices)

	// attempt to use candles for TVWAP calculations
	tvwapPrices, err := ComputeTVWAP(filteredCandles)
	if err != nil {
		return nil, err
	}

	// If TVWAP candles are not available or were filtered out due to staleness,
	// use most recent prices & VWAP instead.
	if len(tvwapPrices) == 0 {
		convertedTickers, err := convertTickersToUSD(
			o.logger,
			providerPrices,
			providerPairs,
			deviations,
		)
		if err != nil {
			return nil, err
		}

		filteredProviderPrices, err := FilterTickerDeviations(
			o.logger,
			convertedTickers,
			deviations,
		)
		if err != nil {
			return nil, err
		}

		o.vwapsByProvider.SetPrices(ComputeVwapsByProvider(filteredProviderPrices))

		vwapPrices := ComputeVWAP(filteredProviderPrices)

		return vwapPrices, nil
	}

	return tvwapPrices, nil
}

// SetProviderTickerPricesAndCandles flattens and collects prices for
// candles and tickers based on the base currency per provider.
// Returns true if at least one of price or candle exists.
func SetProviderTickerPricesAndCandles(
	providerName provider.Name,
	providerPrices provider.AggregatedProviderPrices,
	providerCandles provider.AggregatedProviderCandles,
	prices map[string]types.TickerPrice,
	candles map[string][]types.CandlePrice,
	pair types.CurrencyPair,
) (success bool) {
	if _, ok := providerPrices[providerName]; !ok {
		providerPrices[providerName] = make(map[string]types.TickerPrice)
	}
	if _, ok := providerCandles[providerName]; !ok {
		providerCandles[providerName] = make(map[string][]types.CandlePrice)
	}

	tp, pricesOk := prices[pair.String()]
	cp, candlesOk := candles[pair.String()]

	if pricesOk {
		providerPrices[providerName][pair.Base] = tp
	}
	if candlesOk {
		providerCandles[providerName][pair.Base] = cp
	}

	return pricesOk || candlesOk
}

// GetParamCache returns the last updated parameters of the x/oracle module
// if the current ParamCache is outdated, we will query it again.
func (o *Oracle) GetParamCache(ctx context.Context, currentBlockHeigh int64) (oracletypes.Params, error) {
	if !o.paramCache.IsOutdated(currentBlockHeigh) {
		return *o.paramCache.params, nil
	}

	params, err := o.GetParams(ctx)
	if err != nil {
		return oracletypes.Params{}, err
	}

	o.checkAcceptList(params)
	o.paramCache.Update(currentBlockHeigh, params)
	return params, nil
}

// GetParams returns the current on-chain parameters of the x/oracle module.
func (o *Oracle) GetParams(ctx context.Context) (oracletypes.Params, error) {
	grpcConn, err := grpc.Dial(
		o.oracleClient.GRPCEndpoint,
		// the Cosmos SDK doesn't support any transport security mechanism
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerFunc),
	)
	if err != nil {
		return oracletypes.Params{}, fmt.Errorf("failed to dial Cosmos gRPC service: %w", err)
	}

	defer grpcConn.Close()
	queryClient := oracletypes.NewQueryClient(grpcConn)

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	queryResponse, err := queryClient.Params(ctx, &oracletypes.QueryParams{})
	if err != nil {
		return oracletypes.Params{}, fmt.Errorf("failed to get x/oracle params: %w", err)
	}

	return queryResponse.Params, nil
}

func (o *Oracle) getOrSetProvider(ctx context.Context, providerName provider.Name) (provider.Provider, error) {
	var (
		priceProvider provider.Provider
		ok            bool
	)

	priceProvider, ok = o.priceProviders[providerName]
	if !ok {
		newProvider, err := NewProvider(
			ctx,
			providerName,
			o.logger,
			o.endpoints[providerName],
			o.providerPairs[providerName]...,
		)
		if err != nil {
			return nil, err
		}
		priceProvider = newProvider

		o.priceProviders[providerName] = priceProvider
	}

	return priceProvider, nil
}

func NewProvider(
	ctx context.Context,
	providerName provider.Name,
	logger zerolog.Logger,
	endpoint provider.Endpoint,
	providerPairs ...types.CurrencyPair,
) (provider.Provider, error) {
	switch providerName {
	case provider.ProviderBinance:
		return provider.NewBinanceProvider(ctx, logger, endpoint, providerPairs...)

	case provider.ProviderKraken:
		return provider.NewKrakenProvider(ctx, logger, endpoint, providerPairs...)

	case provider.ProviderOsmosis:
		return provider.NewOsmosisProvider(endpoint), nil

	case provider.ProviderJuno:
		return provider.NewJunoProvider(endpoint), nil

	case provider.ProviderHuobi:
		return provider.NewHuobiProvider(ctx, logger, endpoint, providerPairs...)

	case provider.ProviderCoinbase:
		return provider.NewCoinbaseProvider(ctx, logger, endpoint, providerPairs...)

	case provider.ProviderOkx:
		return provider.NewOkxProvider(ctx, logger, endpoint, providerPairs...)

	case provider.ProviderGate:
		return provider.NewGateProvider(ctx, logger, endpoint, providerPairs...)

	case provider.ProviderBitget:
		return provider.NewBitgetProvider(ctx, logger, endpoint, providerPairs...)

	case provider.ProviderMexc:
		return provider.NewMexcProvider(ctx, logger, endpoint, providerPairs...)

	case provider.ProviderCrypto:
		return provider.NewCryptoProvider(ctx, logger, endpoint, providerPairs...)

	case provider.ProviderMock:
		return provider.NewMockProvider(), nil
	}

	return nil, fmt.Errorf("provider %s not found", providerName)
}

func (o *Oracle) checkAcceptList(params oracletypes.Params) {
	for _, denom := range params.AcceptList {
		symbol := strings.ToUpper(denom.SymbolDenom)
		if _, ok := o.prices[symbol]; !ok {
			o.logger.Warn().Str("denom", symbol).Msg("price missing for required denom")
		}
	}
}

func (o *Oracle) tick(ctx context.Context) error {
	o.logger.Debug().Msg("executing oracle tick")

	blockHeight, err := o.oracleClient.ChainHeight.GetChainHeight()
	if err != nil {
		return err
	}
	if blockHeight < 1 {
		return fmt.Errorf("expected positive block height")
	}

	oracleParams, err := o.GetParamCache(ctx, blockHeight)
	if err != nil {
		return err
	}

	if err := o.SetPrices(ctx); err != nil {
		return err
	}

	// Get oracle vote period, next block height, current vote period, and index
	// in the vote period.
	oracleVotePeriod := int64(oracleParams.VotePeriod)
	nextBlockHeight := blockHeight + 1
	currentVotePeriod := math.Floor(float64(nextBlockHeight) / float64(oracleVotePeriod))
	indexInVotePeriod := nextBlockHeight % oracleVotePeriod

	// Skip until new voting period. Specifically, skip when:
	// index [0, oracleVotePeriod - 1] > oracleVotePeriod - 2 OR index is 0
	if (o.previousVotePeriod != 0 && currentVotePeriod == o.previousVotePeriod) ||
		oracleVotePeriod-indexInVotePeriod < 2 {
		o.logger.Info().
			Int64("vote_period", oracleVotePeriod).
			Float64("previous_vote_period", o.previousVotePeriod).
			Float64("current_vote_period", currentVotePeriod).
			Msg("skipping until next voting period")

		return nil
	}

	// If we're past the voting period we needed to hit, reset and submit another
	// prevote.
	if o.previousVotePeriod != 0 && currentVotePeriod-o.previousVotePeriod != 1 {
		o.logger.Info().
			Int64("vote_period", oracleVotePeriod).
			Float64("previous_vote_period", o.previousVotePeriod).
			Float64("current_vote_period", currentVotePeriod).
			Msg("missing vote during voting period")
		telemetry.IncrCounter(1, "vote", "failure", "missed")

		o.previousVotePeriod = 0
		o.previousPrevote = nil
		return nil
	}

	salt, err := GenerateSalt(32)
	if err != nil {
		return err
	}

	valAddr, err := sdk.ValAddressFromBech32(o.oracleClient.ValidatorAddrString)
	if err != nil {
		return err
	}

	exchangeRatesStr := GenerateExchangeRatesString(o.prices)
	hash := oracletypes.GetAggregateVoteHash(salt, exchangeRatesStr, valAddr)
	preVoteMsg := &oracletypes.MsgAggregateExchangeRatePrevote{
		Hash:      hash.String(), // hash of prices from the oracle
		Feeder:    o.oracleClient.OracleAddrString,
		Validator: valAddr.String(),
	}

	isPrevoteOnlyTx := o.previousPrevote == nil
	if isPrevoteOnlyTx {
		// This timeout could be as small as oracleVotePeriod-indexInVotePeriod,
		// but we give it some extra time just in case.
		//
		// Ref : https://github.com/terra-money/oracle-feeder/blob/baef2a4a02f57a2ffeaa207932b2e03d7fb0fb25/feeder/src/vote.ts#L222
		o.logger.Info().
			Str("hash", hash.String()).
			Str("validator", preVoteMsg.Validator).
			Str("feeder", preVoteMsg.Feeder).
			Msg("broadcasting pre-vote")
		if err := o.oracleClient.BroadcastTx(nextBlockHeight, oracleVotePeriod*2, preVoteMsg); err != nil {
			return err
		}

		currentHeight, err := o.oracleClient.ChainHeight.GetChainHeight()
		if err != nil {
			return err
		}

		o.previousVotePeriod = math.Floor(float64(currentHeight) / float64(oracleVotePeriod))
		o.previousPrevote = &PreviousPrevote{
			Salt:              salt,
			ExchangeRates:     exchangeRatesStr,
			SubmitBlockHeight: currentHeight,
		}
	} else {
		// otherwise, we're in the next voting period and thus we vote
		voteMsg := &oracletypes.MsgAggregateExchangeRateVote{
			Salt:          o.previousPrevote.Salt,
			ExchangeRates: o.previousPrevote.ExchangeRates,
			Feeder:        o.oracleClient.OracleAddrString,
			Validator:     valAddr.String(),
		}

		o.logger.Info().
			Str("exchange_rates", voteMsg.ExchangeRates).
			Str("validator", voteMsg.Validator).
			Str("feeder", voteMsg.Feeder).
			Msg("broadcasting vote")
		if err := o.oracleClient.BroadcastTx(
			nextBlockHeight,
			oracleVotePeriod-indexInVotePeriod,
			voteMsg,
		); err != nil {
			return err
		}

		o.previousPrevote = nil
		o.previousVotePeriod = 0
	}

	return nil
}

// GenerateSalt generates a random salt, size length/2,  as a HEX encoded string.
func GenerateSalt(length int) (string, error) {
	if length == 0 {
		return "", fmt.Errorf("failed to generate salt: zero length")
	}

	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// GenerateExchangeRatesString generates a canonical string representation of
// the aggregated exchange rates.
func GenerateExchangeRatesString(prices map[string]sdk.Dec) string {
	exchangeRates := make([]string, len(prices))
	i := 0

	// aggregate exchange rates as "<base>:<price>"
	for base, avgPrice := range prices {
		exchangeRates[i] = fmt.Sprintf("%s:%s", base, avgPrice.String())
		i++
	}

	sort.Strings(exchangeRates)

	return strings.Join(exchangeRates, ",")
}
