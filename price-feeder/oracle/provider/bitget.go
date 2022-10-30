package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	bitgetWSHost        = "ws.bitget.com"
	bitgetWSPath        = "/spot/v1/stream"
	bitgetReconnectTime = time.Minute * 2
	bitgetRestHost      = "https://api.bitget.com"
	bitgetRestPath      = "/api/spot/v1/public/products"
	tickerChannel       = "ticker"
	candleChannel       = "candle5m"
	instType            = "SP"
)

var _ Provider = (*BitgetProvider)(nil)

type (
	// BitgetProvider defines an Oracle provider implemented by the Bitget public
	// API.
	//
	// REF: https://bitgetlimited.github.io/apidoc/en/spot/#tickers-channel
	// REF: https://bitgetlimited.github.io/apidoc/en/spot/#candlesticks-channel
	BitgetProvider struct {
		wsURL           url.URL
		wsClient        *websocket.Conn
		logger          zerolog.Logger
		mtx             sync.RWMutex
		endpoints       Endpoint
		tickers         map[string]BitgetTicker       // Symbol => BitgetTicker
		candles         map[string][]BitgetCandle     // Symbol => BitgetCandle
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
	}

	// BitgetSubscriptionMsg Msg to subscribe all at once.
	BitgetSubscriptionMsg struct {
		Operation string                  `json:"op"`   // Operation (e.x. "subscribe")
		Args      []BitgetSubscriptionArg `json:"args"` // Arguments to subscribe to
	}
	BitgetSubscriptionArg struct {
		InstType string `json:"instType"` // Instrument type (e.g. "sp")
		Channel  string `json:"channel"`  // Channel (e.x. "ticker" / "candle5m")
		InstID   string `json:"instId"`   // Instrument ID (e.x. BTCUSDT)
	}

	// BitgetErrResponse is the structure for bitget subscription errors.
	BitgetErrResponse struct {
		Event string `json:"event"` // e.x. "error"
		Code  uint64 `json:"code"`  // e.x. 30003 for invalid op
		Msg   string `json:"msg"`   // e.x. "INVALID op"
	}
	// BitgetSubscriptionResponse is the structure for bitget subscription confirmations.
	BitgetSubscriptionResponse struct {
		Event string                `json:"event"` // e.x. "subscribe"
		Arg   BitgetSubscriptionArg `json:"arg"`   // subscription event argument
	}

	// BitgetTickerResponse is the structure for bitget ticker messages.
	BitgetTicker struct {
		Action string                `json:"action"` // e.x. "snapshot"
		Arg    BitgetSubscriptionArg `json:"arg"`    // subscription event argument
		Data   []BitgetTickerData    `json:"data"`   // ticker data
	}
	BitgetTickerData struct {
		InstID string `json:"instId"`     // e.x. BTCUSD
		Price  string `json:"last"`       // last price e.x. "12.3907"
		Volume string `json:"baseVolume"` // volume in base asset (e.x. "112247.9173")
	}

	// BitgetCandleResponse is the response structure for the bitget ticker message.
	BitgetCandleResponse struct {
		Action string                `json:"action"` // e.x. "snapshot"
		Arg    BitgetSubscriptionArg `json:"arg"`    // subscription event argument
		Data   [][]string            `json:"data"`   // candle data in an array at data[0].
	}
	BitgetCandle struct {
		Arg       BitgetSubscriptionArg // subscription event argument
		TimeStamp int64                 // unix timestamp in milliseconds e.x. 1597026383085
		Close     string                // Most recent price e.x. "8533.02"
		Volume    string                // volume e.x. "45247"
	}

	// BitgetPairsSummary defines the response structure for a Bitget pairs
	// summary.
	BitgetPairsSummary struct {
		RespCode string           `json:"code"`
		Data     []BitgetPairData `json:"data"`
	}
	BitgetPairData struct {
		Base  string `json:"baseCoin"`
		Quote string `json:"quoteCoin"`
	}
)

// NewBitgetProvider returns a new Bitget provider with the WS connection
// and msg handler.
func NewBitgetProvider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoints Endpoint,
	pairs ...types.CurrencyPair,
) (*BitgetProvider, error) {
	if endpoints.Name != ProviderBitget {
		endpoints = Endpoint{
			Name:      ProviderBitget,
			Rest:      bitgetRestHost,
			Websocket: bitgetWSHost,
		}
	}

	wsURL := url.URL{
		Scheme: "wss",
		Host:   endpoints.Websocket,
		Path:   bitgetWSPath,
	}

	wsConn, resp, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf(
			types.ErrWebsocketDial.Error(),
			ProviderBitget,
			err,
		)
	}
	defer resp.Body.Close()

	provider := &BitgetProvider{
		wsURL:           wsURL,
		wsClient:        wsConn,
		logger:          logger.With().Str("provider", string(ProviderBitget)).Logger(),
		endpoints:       endpoints,
		tickers:         map[string]BitgetTicker{},
		candles:         map[string][]BitgetCandle{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	if err := provider.SubscribeCurrencyPairs(pairs...); err != nil {
		return nil, err
	}

	go provider.handleWebSocketMsgs(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map.
func (p *BitgetProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	tickerPrices := make(map[string]types.TickerPrice, len(pairs))

	for _, cp := range pairs {
		price, err := p.getTickerPrice(cp)
		if err != nil {
			return nil, err
		}
		tickerPrices[cp.String()] = price
	}

	return tickerPrices, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map.
func (p *BitgetProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candlePrices := make(map[string][]types.CandlePrice, len(pairs))

	for _, cp := range pairs {
		price, err := p.getCandlePrices(cp)
		if err != nil {
			return nil, err
		}
		candlePrices[cp.String()] = price
	}

	return candlePrices, nil
}

// SubscribeCurrencyPairs subscribe all currency pairs into
// ticker and candle channels.
func (p *BitgetProvider) SubscribeCurrencyPairs(cps ...types.CurrencyPair) error {
	if len(cps) == 0 {
		return fmt.Errorf("currency pairs is empty")
	}

	if err := p.subscribeChannels(cps...); err != nil {
		return err
	}

	p.setSubscribedPairs(cps...)
	telemetryWebsocketSubscribeCurrencyPairs(ProviderBitget, len(cps))
	return nil
}

// subscribeChannels subscribe all currency pairs into ticker and candle channels.
func (p *BitgetProvider) subscribeChannels(cps ...types.CurrencyPair) error {
	bitgetSubscriptionMsg := newBitgetTickerSubscriptionMsg(cps)
	return p.wsClient.WriteJSON(bitgetSubscriptionMsg)
}

// subscribedPairsToSlice returns the map of subscribed pairs as slice
func (p *BitgetProvider) subscribedPairsToSlice() []types.CurrencyPair {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	return types.MapPairsToSlice(p.subscribedPairs)
}

func (p *BitgetProvider) handleWebSocketMsgs(ctx context.Context) {
	reconnectTicker := time.NewTicker(bitgetReconnectTime)
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
					if err := p.disconnect(); err != nil {
						p.logger.Err(err).Msg("error disconnecting websocket")
					}
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
			if err := p.disconnect(); err != nil {
				p.logger.Err(err).Msg("error disconnecting websocket")
			}
			if err := p.reconnect(); err != nil {
				p.logger.Err(err).Msg("error reconnecting websocket")
			}
		}
	}
}

// messageReceived handles the received data from the Bitget websocket.
func (p *BitgetProvider) messageReceived(messageType int, bz []byte, reconnectTicker *time.Ticker) {
	if messageType != websocket.TextMessage {
		return
	}

	if messageType == websocket.PingMessage {
		p.pong(bz, reconnectTicker)
		return
	}

	var (
		tickerResp           BitgetTicker
		tickerErr            error
		candleResp           BitgetCandleResponse
		candleErr            error
		errResponse          BitgetErrResponse
		subscriptionResponse BitgetSubscriptionResponse
	)

	err := json.Unmarshal(bz, &errResponse)
	if err == nil && errResponse.Code != 0 {
		p.logger.Error().
			Int("length", len(bz)).
			Str("msg", errResponse.Msg).
			Str("body", string(bz)).
			Msg("Error on receive bitget message")
		return
	}

	err = json.Unmarshal(bz, &subscriptionResponse)
	if err != nil && subscriptionResponse.Event == "subscribe" {
		p.logger.Debug().
			Str("InstID", subscriptionResponse.Arg.InstID).
			Str("Channel", subscriptionResponse.Arg.Channel).
			Str("InstType", subscriptionResponse.Arg.InstType).
			Msg("Bitget subscription confirmed")
		return
	}

	tickerErr = json.Unmarshal(bz, &tickerResp)
	if tickerResp.Arg.Channel == tickerChannel {
		p.setTickerPair(tickerResp)
		telemetryWebsocketMessage(ProviderBitget, MessageTypeTicker)
		return
	}

	candleErr = json.Unmarshal(bz, &candleResp)
	if candleResp.Arg.Channel == candleChannel {
		candle, err := candleResp.ToBitgetCandle()
		if err != nil {
			p.logger.Error().
				Int("length", len(bz)).
				AnErr("candle", err).
				Msg("Unable to parse bitget candle")
		}
		p.setCandlePair(candle)
		telemetryWebsocketMessage(ProviderBitget, MessageTypeCandle)
		return
	}

	p.logger.Error().
		Int("length", len(bz)).
		AnErr("ticker", tickerErr).
		AnErr("candle", candleErr).
		Msg("Error on receive message")
}

// ToBitgetCandle turns a BitgetCandleResponse into a more-readable
// BitgetCandle. The Close and Volume responses are at the [0][4] and
// [0][5] indexes respectively.
// Ref: https://bitgetlimited.github.io/apidoc/en/spot/#candlesticks-channel
func (bcr BitgetCandleResponse) ToBitgetCandle() (BitgetCandle, error) {
	if len(bcr.Data) < 1 || len(bcr.Data[0]) < 6 {
		return BitgetCandle{}, fmt.Errorf("invalid candle response")
	}

	ts, err := strconv.ParseInt(bcr.Data[0][0], 10, 64)
	if err != nil {
		return BitgetCandle{}, err
	}

	return BitgetCandle{
		Arg:       bcr.Arg,
		TimeStamp: ts,
		Close:     bcr.Data[0][4],
		Volume:    bcr.Data[0][5],
	}, nil
}

// pong return a heartbeat message when a "ping" is received and reset the
// recconnect ticker because the connection is alive. After connected to Bitget's
// Websocket server, the server will send heartbeat periodically (5s interval).
// When client receives an heartbeat message, it should respond with a matching
// "pong" message which has the same integer in it, e.g. {"ping": 1492420473027}
// and then the return pong message should be {"pong": 1492420473027}.
func (p *BitgetProvider) pong(bz []byte, reconnectTicker *time.Ticker) {
	reconnectTicker.Reset(bitgetReconnectTime)
	var heartbeat struct {
		Ping uint64 `json:"ping"`
	}

	if err := json.Unmarshal(bz, &heartbeat); err != nil {
		p.logger.Err(err).Msg("could not unmarshal heartbeat")
		return
	}

	if err := p.wsClient.WriteJSON(struct {
		Pong uint64 `json:"pong"`
	}{Pong: heartbeat.Ping}); err != nil {
		p.logger.Err(err).Msg("could not send pong message back")
	}
}

// ping to check websocket connection
func (p *BitgetProvider) ping() error {
	return p.wsClient.WriteMessage(websocket.PingMessage, ping)
}

func (p *BitgetProvider) setTickerPair(ticker BitgetTicker) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.tickers[ticker.Arg.InstID] = ticker
}

func (p *BitgetProvider) setCandlePair(candle BitgetCandle) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	staleTime := PastUnixTime(providerCandlePeriod)
	candleList := []BitgetCandle{}
	candleList = append(candleList, candle)

	for _, c := range p.candles[candle.Arg.InstID] {
		if staleTime < c.TimeStamp {
			candleList = append(candleList, c)
		}
	}
	p.candles[candle.Arg.InstID] = candleList
}

// disconnect disconnects the existing websocket connection.
func (p *BitgetProvider) disconnect() error {
	err := p.wsClient.Close()
	if err != nil {
		return types.ErrProviderConnection.Wrapf("error closing Bitget websocket %v", err)
	}
	return nil
}

// reconnect creates a new websocket connection.
func (p *BitgetProvider) reconnect() error {
	p.logger.Debug().Msg("reconnecting websocket")
	wsConn, resp, err := websocket.DefaultDialer.Dial(p.wsURL.String(), nil)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf(
			types.ErrWebsocketDial.Error(),
			ProviderBitget,
			err,
		)
	}
	p.wsClient = wsConn

	currencyPairs := p.subscribedPairsToSlice()

	telemetryWebsocketReconnect(ProviderBitget)
	return p.subscribeChannels(currencyPairs...)
}

func (p *BitgetProvider) getTickerPrice(cp types.CurrencyPair) (types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	ticker, ok := p.tickers[cp.String()]
	if !ok {
		return types.TickerPrice{}, fmt.Errorf("bitget failed to get ticker price for %s", cp.String())
	}

	return ticker.toTickerPrice()
}

func (p *BitgetProvider) getCandlePrices(cp types.CurrencyPair) ([]types.CandlePrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	candles, ok := p.candles[cp.String()]
	if !ok {
		return []types.CandlePrice{}, fmt.Errorf("failed to get candles price for %s", cp.String())
	}

	candleList := []types.CandlePrice{}
	for _, candle := range candles {
		cp, err := candle.toCandlePrice()
		if err != nil {
			return []types.CandlePrice{}, err
		}
		candleList = append(candleList, cp)
	}
	return candleList, nil
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *BitgetProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// GetAvailablePairs returns all pairs to which the provider can subscribe.
func (p *BitgetProvider) GetAvailablePairs() (map[string]struct{}, error) {
	resp, err := http.Get(p.endpoints.Rest + bitgetRestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary BitgetPairsSummary
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}
	if pairsSummary.RespCode != "00000" {
		return nil, fmt.Errorf("unable to get bitget available pairs")
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary.Data))
	for _, pair := range pairsSummary.Data {
		cp := types.CurrencyPair{
			Base:  pair.Base,
			Quote: pair.Quote,
		}
		availablePairs[cp.String()] = struct{}{}
	}

	return availablePairs, nil
}

// toTickerPrice converts current BitgetTicker to TickerPrice.
func (ticker BitgetTicker) toTickerPrice() (types.TickerPrice, error) {
	if len(ticker.Data) < 1 {
		return types.TickerPrice{}, fmt.Errorf("ticker has no data")
	}
	return types.NewTickerPrice(
		string(ProviderBitget),
		ticker.Arg.InstID,
		ticker.Data[0].Price,
		ticker.Data[0].Volume,
	)
}

func (candle BitgetCandle) toCandlePrice() (types.CandlePrice, error) {
	return types.NewCandlePrice(
		string(ProviderBitget),
		candle.Arg.InstID,
		candle.Close,
		candle.Volume,
		candle.TimeStamp,
	)
}

// newBitgetTickerSubscriptionMsg returns a new ticker subscription Msg.
func newBitgetTickerSubscriptionMsg(cps []types.CurrencyPair) BitgetSubscriptionMsg {
	args := []BitgetSubscriptionArg{}
	for _, cp := range cps {
		args = append(args, BitgetSubscriptionArg{
			InstType: instType,
			Channel:  tickerChannel,
			InstID:   cp.String(),
		})
		args = append(args, BitgetSubscriptionArg{
			InstType: instType,
			Channel:  candleChannel,
			InstID:   cp.String(),
		})
	}

	return BitgetSubscriptionMsg{
		Operation: "subscribe",
		Args:      args,
	}
}
