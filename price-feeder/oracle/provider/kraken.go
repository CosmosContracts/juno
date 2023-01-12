package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CosmosContracts/juno/price-feeder/oracle/types"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	krakenWSHost                  = "ws.kraken.com"
	KrakenRestHost                = "https://api.kraken.com"
	KrakenRestPath                = "/0/public/AssetPairs"
	krakenEventSystemStatus       = "systemStatus"
	krakenEventSubscriptionStatus = "subscriptionStatus"
)

var _ Provider = (*KrakenProvider)(nil)

type (
	// KrakenProvider defines an Oracle provider implemented by the Kraken public
	// API.
	//
	// REF: https://docs.kraken.com/websockets/#overview
	KrakenProvider struct {
		wsURL           url.URL
		wsClient        *websocket.Conn
		logger          zerolog.Logger
		mtx             sync.RWMutex
		endpoints       Endpoint
		tickers         map[string]types.TickerPrice  // Symbol => TickerPrice
		candles         map[string][]KrakenCandle     // Symbol => KrakenCandle
		subscribedPairs map[string]types.CurrencyPair // Symbol => types.CurrencyPair
	}

	// KrakenTicker ticker price response from Kraken ticker channel.
	// REF: https://docs.kraken.com/websockets/#message-ticker
	KrakenTicker struct {
		C []string `json:"c"` // Close with Price in the first position
		V []string `json:"v"` // Volume with the value over last 24 hours in the second position
	}

	// KrakenCandle candle response from Kraken candle channel.
	// REF: https://docs.kraken.com/websockets/#message-ohlc
	KrakenCandle struct {
		Close     string // Close price during this period
		TimeStamp int64  // Linux epoch timestamp
		Volume    string // Volume during this period
		Symbol    string // Symbol for this candle
	}

	// KrakenSubscriptionMsg Msg to subscribe to all the pairs at once.
	KrakenSubscriptionMsg struct {
		Event        string                    `json:"event"`        // subscribe/unsubscribe
		Pair         []string                  `json:"pair"`         // Array of currency pairs ex.: "BTC/USDT",
		Subscription KrakenSubscriptionChannel `json:"subscription"` // subscription object
	}

	// KrakenSubscriptionChannel Msg with the channel name to be subscribed.
	KrakenSubscriptionChannel struct {
		Name string `json:"name"` // channel to be subscribed ex.: ticker
	}

	// KrakenEvent wraps the possible events from the provider.
	KrakenEvent struct {
		Event string `json:"event"` // events from kraken ex.: systemStatus | subscriptionStatus
	}

	// KrakenEventSystemStatus parse the systemStatus event message.
	KrakenEventSystemStatus struct {
		Status string `json:"status"` // online|maintenance|cancel_only|limit_only|post_only
	}

	// KrakenEventSubscriptionStatus parse the subscriptionStatus event message.
	KrakenEventSubscriptionStatus struct {
		Status       string `json:"status"`       // subscribed|unsubscribed|error
		Pair         string `json:"pair"`         // Pair symbol base/quote ex.: "XBT/USD"
		ErrorMessage string `json:"errorMessage"` // error description
	}

	// KrakenPairsSummary defines the response structure for an Kraken pairs summary.
	KrakenPairsSummary struct {
		Result map[string]KrakenPairData `json:"result"`
	}

	// KrakenPairData defines the data response structure for an Kraken pair.
	KrakenPairData struct {
		WsName string `json:"wsname"`
	}
)

// NewKrakenProvider returns a new Kraken provider with the WS connection and msg handler.
func NewKrakenProvider(
	ctx context.Context,
	logger zerolog.Logger,
	endpoints Endpoint,
	pairs ...types.CurrencyPair,
) (*KrakenProvider, error) {
	if endpoints.Name != ProviderKraken {
		endpoints = Endpoint{
			Name:      ProviderKraken,
			Rest:      KrakenRestHost,
			Websocket: krakenWSHost,
		}
	}

	wsURL := url.URL{
		Scheme: "wss",
		Host:   endpoints.Websocket,
	}

	wsConn, resp, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf(
			types.ErrWebsocketDial.Error(),
			ProviderKraken,
			err,
		)
	}

	provider := &KrakenProvider{
		wsURL:           wsURL,
		wsClient:        wsConn,
		logger:          logger.With().Str("provider", string(ProviderKraken)).Logger(),
		endpoints:       endpoints,
		tickers:         map[string]types.TickerPrice{},
		candles:         map[string][]KrakenCandle{},
		subscribedPairs: map[string]types.CurrencyPair{},
	}

	if err := provider.SubscribeCurrencyPairs(pairs...); err != nil {
		return nil, err
	}

	go provider.handleWebSocketMsgs(ctx)

	return provider, nil
}

// GetTickerPrices returns the tickerPrices based on the saved map.
func (p *KrakenProvider) GetTickerPrices(pairs ...types.CurrencyPair) (map[string]types.TickerPrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	tickerPrices := make(map[string]types.TickerPrice, len(pairs))

	for _, cp := range pairs {
		key := cp.String()
		tickerPrice, ok := p.tickers[key]
		if !ok {
			return nil, fmt.Errorf("kraken failed to get ticker price for %s", key)
		}
		tickerPrices[key] = tickerPrice
	}

	return tickerPrices, nil
}

// GetCandlePrices returns the candlePrices based on the saved map.
func (p *KrakenProvider) GetCandlePrices(pairs ...types.CurrencyPair) (map[string][]types.CandlePrice, error) {
	candlePrices := make(map[string][]types.CandlePrice, len(pairs))

	for _, cp := range pairs {
		key := cp.String()
		candlePrice, err := p.getCandlePrices(key)
		if err != nil {
			return nil, err
		}
		candlePrices[key] = candlePrice
	}

	return candlePrices, nil
}

// SubscribeCurrencyPairs subscribe all currency pairs into ticker and candle channels.
func (p *KrakenProvider) SubscribeCurrencyPairs(cps ...types.CurrencyPair) error {
	if len(cps) == 0 {
		return fmt.Errorf("currency pairs is empty")
	}

	if err := p.subscribeChannels(cps...); err != nil {
		return err
	}

	p.setSubscribedPairs(cps...)
	telemetryWebsocketSubscribeCurrencyPairs(ProviderKraken, len(cps))
	return nil
}

// subscribeChannels subscribe all currency pairs into ticker and candle channels.
func (p *KrakenProvider) subscribeChannels(cps ...types.CurrencyPair) error {
	pairs := make([]string, len(cps))

	for i, cp := range cps {
		pairs[i] = currencyPairToKrakenPair(cp)
	}

	if err := p.subscribeTickers(pairs...); err != nil {
		return err
	}

	return p.subscribeCandles(pairs...)
}

// subscribedPairsToSlice returns the map of subscribed pairs as slice
func (p *KrakenProvider) subscribedPairsToSlice() []types.CurrencyPair {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	return types.MapPairsToSlice(p.subscribedPairs)
}

func (candle KrakenCandle) toCandlePrice() (types.CandlePrice, error) {
	return types.NewCandlePrice(
		string(ProviderKraken),
		candle.Symbol,
		candle.Close,
		candle.Volume,
		candle.TimeStamp,
	)
}

func (p *KrakenProvider) getCandlePrices(key string) ([]types.CandlePrice, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	candles, ok := p.candles[key]
	if !ok {
		return []types.CandlePrice{}, fmt.Errorf("kraken failed to get candle prices for %s", key)
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

// handleWebSocketMsgs receive all the messages from the provider and controls the
// reconnect function to the web socket.
func (p *KrakenProvider) handleWebSocketMsgs(ctx context.Context) {
	reconnectTicker := time.NewTicker(defaultMaxConnectionTime)
	defer reconnectTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(defaultReadNewWSMessage):
			messageType, bz, err := p.wsClient.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					p.logger.Err(err).Msg("WebSocket closed unexpectedly")
					continue
				}

				// if some error occurs continue to try to read the next message.
				p.logger.Err(err).Msg("could not read message")
				if err := p.ping(); err != nil {
					p.logger.Err(err).Msg("failed to send ping")
				}
				continue
			}

			if len(bz) == 0 {
				continue
			}

			p.messageReceived(messageType, bz)

		case <-reconnectTicker.C:
			if err := p.reconnect(); err != nil {
				p.logger.Err(err).Msg("error reconnecting")
			}
		}
	}
}

// messageReceived handles any message sent by the provider.
func (p *KrakenProvider) messageReceived(messageType int, bz []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var (
		krakenEvent KrakenEvent
		krakenErr   error
		tickerErr   error
		candleErr   error
	)

	krakenErr = json.Unmarshal(bz, &krakenEvent)
	if krakenErr == nil {
		switch krakenEvent.Event {
		case krakenEventSystemStatus:
			p.messageReceivedSystemStatus(bz)
			return
		case krakenEventSubscriptionStatus:
			p.messageReceivedSubscriptionStatus(bz)
			return
		}
		return
	}

	tickerErr = p.messageReceivedTickerPrice(bz)
	if tickerErr == nil {
		return
	}

	candleErr = p.messageReceivedCandle(bz)
	if candleErr == nil {
		return
	}

	p.logger.Error().
		Int("length", len(bz)).
		AnErr("ticker", tickerErr).
		AnErr("candle", candleErr).
		AnErr("event", krakenErr).
		Msg("Error on receive message")
}

// messageReceivedTickerPrice handles the ticker price msg.
func (p *KrakenProvider) messageReceivedTickerPrice(bz []byte) error {
	// the provider response is an array with different types at each index
	// kraken documentation https://docs.kraken.com/websockets/#message-ticker
	var tickerMessage []interface{}
	if err := json.Unmarshal(bz, &tickerMessage); err != nil {
		return err
	}

	if len(tickerMessage) != 4 {
		return fmt.Errorf("received an unexpected structure")
	}

	channelName, ok := tickerMessage[2].(string)
	if !ok || channelName != "ticker" {
		return fmt.Errorf("received an unexpected channel name")
	}

	tickerBz, err := json.Marshal(tickerMessage[1])
	if err != nil {
		p.logger.Err(err).Msg("could not marshal ticker message")
		return err
	}

	var krakenTicker KrakenTicker
	if err := json.Unmarshal(tickerBz, &krakenTicker); err != nil {
		p.logger.Err(err).Msg("could not unmarshal ticker message")
		return err
	}

	krakenPair, ok := tickerMessage[3].(string)
	if !ok {
		p.logger.Debug().Msg("received an unexpected pair")
		return err
	}

	krakenPair = normalizeKrakenBTCPair(krakenPair)
	currencyPairSymbol := krakenPairToCurrencyPairSymbol(krakenPair)

	tickerPrice, err := krakenTicker.toTickerPrice(currencyPairSymbol)
	if err != nil {
		p.logger.Err(err).Msg("could not parse kraken ticker to ticker price")
		return err
	}

	p.setTickerPair(currencyPairSymbol, tickerPrice)
	telemetryWebsocketMessage(ProviderKraken, MessageTypeTicker)
	return nil
}

func (candle *KrakenCandle) UnmarshalJSON(buf []byte) error {
	var tmp []interface{}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if len(tmp) != 9 {
		return fmt.Errorf("wrong number of fields in candle")
	}

	// timestamps come as a float string
	time, ok := tmp[1].(string)
	if !ok {
		return fmt.Errorf("time field must be a string")
	}
	timeFloat, err := strconv.ParseFloat(time, 64)
	if err != nil {
		return fmt.Errorf("unable to convert time to float")
	}
	candle.TimeStamp = int64(timeFloat)

	close, ok := tmp[5].(string)
	if !ok {
		return fmt.Errorf("close field must be a string")
	}
	candle.Close = close

	volume, ok := tmp[7].(string)
	if !ok {
		return fmt.Errorf("volume field must be a string")
	}
	candle.Volume = volume

	return nil
}

// messageReceivedCandle handles the candle msg.
func (p *KrakenProvider) messageReceivedCandle(bz []byte) error {
	// the provider response is an array with different types at each index
	// kraken documentation https://docs.kraken.com/websockets/#message-ohlc
	var candleMessage []interface{}
	if err := json.Unmarshal(bz, &candleMessage); err != nil {
		return err
	}

	if len(candleMessage) != 4 {
		return fmt.Errorf("received something different than candle")
	}

	channelName, ok := candleMessage[2].(string)
	if !ok || channelName != "ohlc-1" {
		return fmt.Errorf("received an unexpected channel name")
	}

	tickerBz, err := json.Marshal(candleMessage[1])
	if err != nil {
		return fmt.Errorf("could not marshal candle message")
	}

	var krakenCandle KrakenCandle
	if err := krakenCandle.UnmarshalJSON(tickerBz); err != nil {
		return err
	}

	krakenPair, ok := candleMessage[3].(string)
	if !ok {
		return fmt.Errorf("received an unexpected pair")
	}

	krakenPair = normalizeKrakenBTCPair(krakenPair)
	currencyPairSymbol := krakenPairToCurrencyPairSymbol(krakenPair)
	krakenCandle.Symbol = currencyPairSymbol

	telemetryWebsocketMessage(ProviderKraken, MessageTypeCandle)
	p.setCandlePair(krakenCandle)
	return nil
}

// reconnect closes the last WS connection then create a new one.
func (p *KrakenProvider) reconnect() error {
	err := p.wsClient.Close()
	if err != nil {
		p.logger.Err(err).Msg("error closing Kraken websocket")
	}

	p.logger.Debug().Msg("trying to reconnect")
	wsConn, resp, err := websocket.DefaultDialer.Dial(p.wsURL.String(), nil)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf(
			types.ErrWebsocketDial.Error(),
			ProviderKraken,
			err,
		)
	}
	p.wsClient = wsConn
	telemetryWebsocketReconnect(ProviderKraken)

	return p.subscribeChannels(p.subscribedPairsToSlice()...)
}

// messageReceivedSubscriptionStatus handle the subscription status message
// sent by the provider.
func (p *KrakenProvider) messageReceivedSubscriptionStatus(bz []byte) {
	var subscriptionStatus KrakenEventSubscriptionStatus
	if err := json.Unmarshal(bz, &subscriptionStatus); err != nil {
		p.logger.Err(err).Msg("provider could not unmarshal KrakenEventSubscriptionStatus")
		return
	}

	switch subscriptionStatus.Status {
	case "error":
		p.logger.Error().Msg(subscriptionStatus.ErrorMessage)
		p.removeSubscribedTickers(krakenPairToCurrencyPairSymbol(subscriptionStatus.Pair))
		return
	case "unsubscribed":
		p.logger.Debug().Msgf("ticker %s was unsubscribed", subscriptionStatus.Pair)
		p.removeSubscribedTickers(krakenPairToCurrencyPairSymbol(subscriptionStatus.Pair))
		return
	}
}

// messageReceivedSystemStatus handle the system status and try to reconnect if it
// is not online.
func (p *KrakenProvider) messageReceivedSystemStatus(bz []byte) {
	var systemStatus KrakenEventSystemStatus
	if err := json.Unmarshal(bz, &systemStatus); err != nil {
		p.logger.Err(err).Msg("could not unmarshal event system status")
		return
	}

	if strings.EqualFold(systemStatus.Status, "online") {
		return
	}

	if err := p.reconnect(); err != nil {
		p.logger.Err(err).Msg("error reconnecting")
	}
}

// setTickerPair sets an ticker to the map thread safe by the mutex.
func (p *KrakenProvider) setTickerPair(symbol string, ticker types.TickerPrice) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.tickers[symbol] = ticker
}

func (p *KrakenProvider) setCandlePair(candle KrakenCandle) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	// convert kraken timestamp seconds -> milliseconds
	candle.TimeStamp = SecondsToMilli(candle.TimeStamp)
	staleTime := PastUnixTime(providerCandlePeriod)
	candleList := []KrakenCandle{}

	candleList = append(candleList, candle)
	for _, c := range p.candles[candle.Symbol] {
		if staleTime < c.TimeStamp {
			candleList = append(candleList, c)
		}
	}
	p.candles[candle.Symbol] = candleList
}

// ping to check websocket connection.
func (p *KrakenProvider) ping() error {
	return p.wsClient.WriteMessage(websocket.PingMessage, ping)
}

// subscribeTickers write the subscription msg to the provider.
func (p *KrakenProvider) subscribeTickers(pairs ...string) error {
	subsMsg := newKrakenTickerSubscriptionMsg(pairs...)
	return p.wsClient.WriteJSON(subsMsg)
}

// subscribeCandles write the subscription msg to the provider.
func (p *KrakenProvider) subscribeCandles(pairs ...string) error {
	subsMsg := newKrakenCandleSubscriptionMsg(pairs...)
	return p.wsClient.WriteJSON(subsMsg)
}

// setSubscribedPairs sets N currency pairs to the map of subscribed pairs.
func (p *KrakenProvider) setSubscribedPairs(cps ...types.CurrencyPair) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, cp := range cps {
		p.subscribedPairs[cp.String()] = cp
	}
}

// removeSubscribedTickers delete N pairs from the subscribed map.
func (p *KrakenProvider) removeSubscribedTickers(tickerSymbols ...string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, tickerSymbol := range tickerSymbols {
		delete(p.subscribedPairs, tickerSymbol)
	}
}

// GetAvailablePairs returns all pairs to which the provider can subscribe.
func (p *KrakenProvider) GetAvailablePairs() (map[string]struct{}, error) {
	resp, err := http.Get(p.endpoints.Rest + KrakenRestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pairsSummary KrakenPairsSummary
	if err := json.NewDecoder(resp.Body).Decode(&pairsSummary); err != nil {
		return nil, err
	}

	availablePairs := make(map[string]struct{}, len(pairsSummary.Result))
	for _, pair := range pairsSummary.Result {
		splitPair := strings.Split(pair.WsName, "/")
		if len(splitPair) != 2 {
			continue
		}

		cp := types.CurrencyPair{
			Base:  strings.ToUpper(splitPair[0]),
			Quote: strings.ToUpper(splitPair[1]),
		}
		availablePairs[cp.String()] = struct{}{}
	}

	return availablePairs, nil
}

// toTickerPrice return a TickerPrice based on the KrakenTicker.
func (ticker KrakenTicker) toTickerPrice(symbol string) (types.TickerPrice, error) {
	if len(ticker.C) != 2 || len(ticker.V) != 2 {
		return types.TickerPrice{}, fmt.Errorf("error converting KrakenTicker to TickerPrice")
	}
	// ticker.C has the Price in the first position.
	// ticker.V has the totla	Value over last 24 hours in the second position.
	return types.NewTickerPrice(string(ProviderKraken), symbol, ticker.C[0], ticker.V[1])
}

// newKrakenTickerSubscriptionMsg returns a new subscription Msg.
func newKrakenTickerSubscriptionMsg(pairs ...string) KrakenSubscriptionMsg {
	return KrakenSubscriptionMsg{
		Event: "subscribe",
		Pair:  pairs,
		Subscription: KrakenSubscriptionChannel{
			Name: "ticker",
		},
	}
}

// newKrakenSubscriptionMsg returns a new subscription Msg.
func newKrakenCandleSubscriptionMsg(pairs ...string) KrakenSubscriptionMsg {
	return KrakenSubscriptionMsg{
		Event: "subscribe",
		Pair:  pairs,
		Subscription: KrakenSubscriptionChannel{
			Name: "ohlc",
		},
	}
}

// krakenPairToCurrencyPairSymbol receives a kraken pair formated
// ex.: ATOM/USDT and return currencyPair Symbol ATOMUSDT.
func krakenPairToCurrencyPairSymbol(krakenPair string) string {
	return strings.ReplaceAll(krakenPair, "/", "")
}

// currencyPairToKrakenPair receives a currency pair
// and return kraken ticker symbol ATOM/USDT.
func currencyPairToKrakenPair(cp types.CurrencyPair) string {
	return strings.ToUpper(cp.Base + "/" + cp.Quote)
}

// normalizeKrakenBTCPair changes XBT pairs to BTC,
// since other providers list bitcoin as BTC.
func normalizeKrakenBTCPair(ticker string) string {
	return strings.Replace(ticker, "XBT", "BTC", 1)
}
