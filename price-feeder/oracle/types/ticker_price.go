package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TickerPrice defines price and volume information for a symbol or ticker exchange rate.
type TickerPrice struct {
	Price  sdk.Dec // last trade price
	Volume sdk.Dec // 24h volume
}

// NewTickerPrice parses the lastPrice and volume to a decimal and returns a TickerPrice
func NewTickerPrice(provider, symbol, lastPrice, volume string) (TickerPrice, error) {
	price, err := sdk.NewDecFromStr(lastPrice)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse %s price (%s) for %s: %w", provider, lastPrice, symbol, err)
	}

	volumeDec, err := sdk.NewDecFromStr(volume)
	if err != nil {
		return TickerPrice{}, fmt.Errorf("failed to parse %s volume (%s) for %s: %w", provider, volume, symbol, err)
	}

	return TickerPrice{Price: price, Volume: volumeDec}, nil
}
