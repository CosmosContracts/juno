package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"github.com/CosmosContracts/juno/price-feeder/oracle/util"
)

const (
	cryptoWSHost             = "stream.crypto.com"
	cryptoWSPath             = "/v2/market"
	cryptoReconnectTime      = time.Second * 30
	cryptoRestHost           = "https://api.crypto.com"
	cryptoRestPath           = "/v2/public/get-ticker"
	cryptoTickerChannel      = "ticker"
	cryptoCandleChannel      = "candlestick"
	cryptoHeartbeatMethod    = "public/heartbeat"
	cryptoHeartbeatReqMethod = "public/respond-heartbeat"
	cryptoTickerMsgPrefix    = "ticker."
	cryptoCandleMsgPrefix    = "candlestick.5m."
)

var _ Provider = (*CryptoProvider)(nil)

type (
	// CryptoProvider defines an Oracle provider implemented by the Crypto.com public
	// API.
	//
	// REF: https://exchange-docs.crypto.com/spot/index.html#introduction
	CryptoProvider struct {
		wsURL           url.URL
		wsClient        *websocket.Conn
		logger          zerolog.Logger
		mtx             sync.RWMutex
		endpoints       Endpoint
		tickers         map[string]types.TickerPrice   // Symbol => TickerPrice
		candles         map[string][]types.CandlePrice // Symbol => CandlePrice
		subscribedPairs map[string]types.CurrencyPair  // Symbol => types.CurrencyPair
	}

	CryptoTickerResponse struct {
		Result CryptoTickerResult `json:"result"`
	}
	CryptoTickerResult struct {
		InstrumentName string         `json:"instrument_name"` // ex.: ATOM_USDT
		Channel        string         `json:"channel"`         // ex.: ticker
		Data           []CryptoTicker `json:"data"`            // ticker data
	}
	CryptoTicker struct {
		InstrumentName string  `json:"i"` // Instrument Name, e.g. BTC_USDT, ETH_CRO, etc.
		Volume         float64 `json:"v"` // The total 24h traded volume
		LatestTrade    float64 `json:"a"` // The price of the latest trade, null if there weren't any trades
	}

	CryptoCandleResponse struct {
		Result CryptoCandleResult `json:"result"`
	}
	CryptoCandleResult struct {
		InstrumentName string         `json:"instrument_name"` // ex.: ATOM_USDT
		Channel        string         `json:"channel"`         // ex.: candlestick
		Data           []CryptoCandle `json:"data"`            // candlestick data
	}
	CryptoCandle struct {
		Close     float64 `json:"c"` // Price at close
		Volume    float64 `json:"v"` // Volume during interval
		Timestamp int64   `json:"t"` // End time of candlestick (Unix timestamp)
	}

	CryptoSubscriptionMsg struct {
		ID     int64                    `json:"id"`
		Method string                   `json:"method"` // subscribe, unsubscribe
		Params CryptoSubscriptionParams `json:"params"`
		Nonce  int64                    `json:"nonce"` // Current timestamp (milliseconds since the Unix epoch)
	}
	CryptoSubscriptionParams struct {
		Channels []string `json:"channels"` // Channels to be subscribed ex. ticker.ATOM_USDT
	}

	CryptoPairsSummary struct {
		Result CryptoInstruments `json:"result"`
	}
	CryptoInstruments struct {
		Data []CryptoTicker `json:"data"`
	}

	CryptoHeartbeatResponse struct {
		ID     int64  `json:"id"`
		Method string `json:"method"` // public/heartbeat
	}
	CryptoHeartbeatRequest struct {
		ID     int64  `json:"id"`
		Method string `json:"method"` // public/respond-heartbeat
	}
)

func NewCryptoProvider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoints Endpoint,
	pairs ...types.CurrencyPair,
) (*CryptoProvider, error) {
	if endpoints.Name != ProviderCrypto {
		endpoints = Endpoint{
			Name:      ProviderCrypto,
			Rest:      cryptoRestHost,
			Websocket: cryptoWSHost,
		}
	}

	wsURL := url.URL{
		Scheme: "wss",
		Host:   endpoints.Websocket,
		Path:   cryptoWSPath,
	}

	wsConn, resp, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf(
			types.ErrWebsocketDial.Error(),
			ProviderCrypto,
			err,
		)
	}
	defer resp.Body.Close()

	provider := &CryptoProvider{
		wsURL:           wsURL,
		wsClient:        wsConn,
		logger:          logger.With().Str("provider", "crypto").Logger(),
		endpoints:       endpoints,
		tickers:         map[string]types.TickerPrice{},
		candles:         map[string][]types.CandlePrice{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	if err := provider.SubscribeCurrencyPairs(pairs...); err != nil {
		return nil, err
	}

	go provider.handleWebSocketMsgs(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map.
func (p *CryptoProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	tickerPrices := make(map[string]types.TickerPrice, len(pairs))

	for _, cp := range pairs {
		key := currencyPairToCryptoPair(cp)
		price, err := p.getTickerPrice(key)
		if err != nil {
			return nil, err
		}
		tickerPrices[cp.String()] = price
	}

	return tickerPrices, nil
}

// GetCandlePrices returns the candlePrices based on the saved map
func (p *CryptoProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candlePrices := make(map[string][]types.CandlePrice, len(pairs))

	for _, cp := range pairs {
		key := currencyPairToCryptoPair(cp)
		prices, err := p.getCandlePrices(key)
		if err != nil {
			return nil, err
		}
		candlePrices[cp.String()] = prices
	}

	return candlePrices, nil
}

// SubscribeCurrencyPairs subscribe all currency pairs into ticker and candle channels.
func (p *CryptoProvider) SubscribeCurrencyPairs(cps ...types.CurrencyPair) error {
	if len(cps) == 0 {
		return fmt.Errorf("currency pairs is empty")
	}

	if err := p.subscribeChannels(cps...); err != nil {
		return err
	}

	p.setSubscribedPairs(cps...)
	telemetryWebsocketSubscribeCurrencyPairs(ProviderCrypto, len(cps))
	return nil
}

// subscribeChannels subscribe all currency pairs into ticker and candle channels.
func (p *CryptoProvider) subscribeChannels(cps ...types.CurrencyPair) error {
	if err := p.subscribeTickers(cps...); err != nil {
		return err
	}

	return p.subscribeCandles(cps...)
}

// subscribeTickers subscribe all currency pairs into ticker channel.
func (p *CryptoProvider) subscribeTickers(cps ...types.CurrencyPair) error {
	pairs := make([]string, len(cps))

	for i, cp := range cps {
		pairs[i] = currencyPairToCryptoPair(cp)
	}

	channels := []string{}
	for _, pair := range pairs {
		channels = append(channels, cryptoTickerMsgPrefix+pair)
	}
	subsMsg := newCryptoSubscriptionMsg(channels)
	err := p.wsClient.WriteJSON(subsMsg)

	return err
}

// subscribeCandles subscribe all currency pairs into candle channel.
func (p *CryptoProvider) subscribeCandles(cps ...types.CurrencyPair) error {
	pairs := make([]string, len(cps))

	for i, cp := range cps {
		pairs[i] = currencyPairToCryptoPair(cp)
	}

	channels := []string{}
	for _, pair := range pairs {
		channels = append(channels, cryptoCandleMsgPrefix+pair)
	}
	subsMsg := newCryptoSubscriptionMsg(channels)
	err := p.wsClient.WriteJSON(subsMsg)

	return err
}

// subscribedPairsToSlice returns the map of subscribed pairs as a slice.
func (p *CryptoProvider) subscribedPairsToSlice() []types.CurrencyPair {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	return types.MapPairsToSlice(p.subscribedPairs)
}

func (p *CryptoProvider) getTickerPrice(key string) (types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	ticker, ok := p.tickers[key]
	if !ok {
		return types.TickerPrice{}, fmt.Errorf(
			types.ErrTickerNotFound.Error(),
			ProviderCrypto,
			key,
		)
	}

	return ticker, nil
}

func (p *CryptoProvider) getCandlePrices(key string) ([]types.CandlePrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	candles, ok := p.candles[key]
	if !ok {
		return []types.CandlePrice{}, fmt.Errorf(
			types.ErrCandleNotFound.Error(),
			ProviderCrypto,
			key,
		)
	}

	return candles, nil
}

func (p *CryptoProvider) messageReceived(messageType int, bz []byte, reconnectTicker *time.Ticker) {
	if messageType != websocket.TextMessage {
		return
	}

	var (
		heartbeatResp CryptoHeartbeatResponse
		heartbeatErr  error
		tickerResp    CryptoTickerResponse
		tickerErr     error
		candleResp    CryptoCandleResponse
		candleErr     error
	)

	// sometimes the message received is not a ticker or a candle response.
	heartbeatErr = json.Unmarshal(bz, &heartbeatResp)
	if heartbeatResp.Method == cryptoHeartbeatMethod {
		p.pong(heartbeatResp, reconnectTicker)
		return
	}

	tickerErr = json.Unmarshal(bz, &tickerResp)
	if tickerResp.Result.Channel == cryptoTickerChannel {
		for _, tickerPair := range tickerResp.Result.Data {
			p.setTickerPair(
				tickerResp.Result.InstrumentName,
				tickerPair,
			)
			telemetryWebsocketMessage(ProviderCrypto, MessageTypeTicker)
		}
		return
	}

	candleErr = json.Unmarshal(bz, &candleResp)
	if candleResp.Result.Channel == cryptoCandleChannel {
		for _, candlePair := range candleResp.Result.Data {
			p.setCandlePair(
				candleResp.Result.InstrumentName,
				candlePair,
			)
			telemetryWebsocketMessage(ProviderCrypto, MessageTypeCandle)
		}
		return
	}

	p.logger.Error().
		Int("length", len(bz)).
		AnErr("heartbeat", heartbeatErr).
		AnErr("ticker", tickerErr).
		AnErr("candle", candleErr).
		Msg("Error on receive message")
}

// pong return a heartbeat message when a "ping" is received and reset the
// recconnect ticker because the connection is alive. After connected to crypto.com's
// Websocket server, the server will send heartbeat periodically (30s interval).
// When client receives an heartbeat message, it must respond back with the
// public/respond-heartbeat method, using the same matching id,
// within 5 seconds, or the connection will break.
func (p *CryptoProvider) pong(heartbeatResp CryptoHeartbeatResponse, reconnectTicker *time.Ticker) {
	reconnectTicker.Reset(cryptoReconnectTime)

	heartbeatReq := CryptoHeartbeatRequest{
		ID:     heartbeatResp.ID,
		Method: cryptoHeartbeatReqMethod,
	}

	if err := p.wsClient.WriteJSON(heartbeatReq); err != nil {
		p.logger.Err(err).Msg("could not send pong message back")
	}
}

// ping to check websocket connection
func (p *CryptoProvider) ping() error {
	return p.wsClient.WriteMessage(websocket.PingMessage, ping)
}

func (p *CryptoProvider) setTickerPair(symbol string, tickerPair CryptoTicker) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	price, err := util.NewDecFromFloat(tickerPair.LatestTrade)
	if err != nil {
		p.logger.Warn().Err(err).Msg("crypto: failed to parse ticker price")
		return
	}
	volume, err := util.NewDecFromFloat(tickerPair.Volume)
	if err != nil {
		p.logger.Warn().Err(err).Msg("crypto: failed to parse ticker volume")
		return
	}

	p.tickers[symbol] = types.TickerPrice{
		Price:  price,
		Volume: volume,
	}
}

func (p *CryptoProvider) setCandlePair(symbol string, candlePair CryptoCandle) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	close, err := util.NewDecFromFloat(candlePair.Close)
	if err != nil {
		p.logger.Warn().Err(err).Msg("crypto: failed to parse candle close")
		return
	}
	volume, err := util.NewDecFromFloat(candlePair.Volume)
	if err != nil {
		p.logger.Warn().Err(err).Msg("crypto: failed to parse candle volume")
		return
	}
	candle := types.CandlePrice{
		Price:  close,
		Volume: volume,
		// convert seconds -> milli
		TimeStamp: SecondsToMilli(candlePair.Timestamp),
	}

	staleTime := PastUnixTime(providerCandlePeriod)
	candleList := []types.CandlePrice{}
	candleList = append(candleList, candle)

	for _, c := range p.candles[symbol] {
		if staleTime < c.TimeStamp {
			candleList = append(candleList, c)
		}
	}

	p.candles[symbol] = candleList
}

func (p *CryptoProvider) handleWebSocketMsgs(ctx context.Context) {
	reconnectTicker := time.NewTicker(cryptoReconnectTime)
	defer reconnectTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := p.wsClient.ReadMessage()
			if err != nil {
				// If some error occurs, check if connection is alive
				// and continue to try to read the next message.
				p.logger.Err(err).Msg("failed to read message")
				if err := p.ping(); err != nil {
					p.logger.Err(err).Msg("failed to send ping")
					if err := p.reconnect(); err != nil {
						p.logger.Err(err).Msg("error reconnecting websocket")
					}
				}
				continue
			}

			if len(bz) == 0 {
				continue
			}

			p.messageReceived(messageType, bz, reconnectTicker)

		case <-reconnectTicker.C:
			if err := p.reconnect(); err != nil {
				p.logger.Err(err).Msg("error reconnecting")
			}
		}
	}
}

// reconnect closes the last WS connection then create a new one and subscribes to
// all subscribed pairs in the ticker and candle pairs. If no ping is received
// within 1 minute, the connection will be disconnected. It is recommended to
// send a ping for 10-20 seconds
func (p *CryptoProvider) reconnect() error {
	err := p.wsClient.Close()
	if err != nil {
		p.logger.Err(err).Msg("error closing crypto websocket")
	}

	p.logger.Debug().Msg("crypto: reconnecting websocket")

	wsConn, resp, err := websocket.DefaultDialer.Dial(p.wsURL.String(), nil)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf(
			types.ErrWebsocketDial.Error(),
			ProviderCrypto,
			err,
		)
	}
	p.wsClient = wsConn
	telemetryWebsocketReconnect(ProviderCrypto)

	return p.subscribeChannels(p.subscribedPairsToSlice()...)
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *CryptoProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// GetAvailablePairs returns all pairs to which the provider can subscribe.
// ex.: map["ATOMUSDT" => {}, "UMEEUSDC" => {}].
func (p *CryptoProvider) GetAvailablePairs() (map[string]struct{}, error) {
	resp, err := http.Get(p.endpoints.Rest + cryptoRestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary CryptoPairsSummary
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary.Result.Data))
	for _, pair := range pairsSummary.Result.Data {
		splitInstName := strings.Split(pair.InstrumentName, "_")
		if len(splitInstName) != 2 {
			continue
		}

		cp := types.CurrencyPair{
			Base:  strings.ToUpper(splitInstName[0]),
			Quote: strings.ToUpper(splitInstName[1]),
		}

		availablePairs[cp.String()] = struct{}{}
	}

	return availablePairs, nil
}

// currencyPairToCryptoPair receives a currency pair and return crypto
// ticker symbol atomusdt@ticker.
func currencyPairToCryptoPair(cp types.CurrencyPair) string {
	return strings.ToUpper(cp.Base + "_" + cp.Quote)
}

// newCryptoSubscriptionMsg returns a new subscription Msg.
func newCryptoSubscriptionMsg(channels []string) CryptoSubscriptionMsg {
	return CryptoSubscriptionMsg{
		ID:     1,
		Method: "subscribe",
		Params: CryptoSubscriptionParams{
			Channels: channels,
		},
		Nonce: time.Now().UnixMilli(),
	}
}
