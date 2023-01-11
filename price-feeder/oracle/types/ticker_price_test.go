package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestNewTicketPrice(t *testing.T) {
	price := "105473.43"
	volume := "48394"

	t.Run("when the inputs are valid", func(t *testing.T) {
		tickerPrice, err := NewTickerPrice("binance", "BTC", price, volume)
		require.Nil(t, err, "expected the returned error to be nil")

		parsedPrice, _ := sdk.NewDecFromStr(price)
		require.Equal(t, tickerPrice.Price, parsedPrice)

		parsedVolume, _ := sdk.NewDecFromStr(volume)
		require.Equal(t, tickerPrice.Volume, parsedVolume)
	})

	t.Run("when the lastPrice input is invalid", func(t *testing.T) {
		_, err := NewTickerPrice("binance", "BTC", "bad_price", volume)
		require.NotNil(t, err, "expected the returned error to not be nil")
	})

	t.Run("when the volume input is invalid", func(t *testing.T) {
		_, err := NewTickerPrice("binance", "BTC", price, "bad_volume")
		require.NotNil(t, err, "expected the returned error to not be nil")
	})
}
