package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestNewCandlePrice(t *testing.T) {
	price := "105473.43"
	volume := "48394"
	timeStamp := int64(1257894000)

	t.Run("when the inputs are valid", func(t *testing.T) {
		candlePrice, err := NewCandlePrice("binance", "BTC", price, volume, timeStamp)
		require.Nil(t, err, "expected the returned error to be nil")

		parsedPrice, _ := sdk.NewDecFromStr(price)
		require.Equal(t, candlePrice.Price, parsedPrice)

		parsedVolume, _ := sdk.NewDecFromStr(volume)
		require.Equal(t, candlePrice.Volume, parsedVolume)

		require.Equal(t, candlePrice.TimeStamp, timeStamp)
	})

	t.Run("when the lastPrice input is invalid", func(t *testing.T) {
		_, err := NewCandlePrice("binance", "BTC", "bad_price", volume, timeStamp)
		require.NotNil(t, err, "expected the returned error to not be nil")
	})

	t.Run("when the volume input is invalid", func(t *testing.T) {
		_, err := NewCandlePrice("binance", "BTC", price, "bad_volume", timeStamp)
		require.NotNil(t, err, "expected the returned error to not be nil")
	})
}
