package v1

import (
	"time"

	"github.com/CosmosContracts/juno/price-feeder/oracle"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Oracle defines the Oracle interface contract that the v1 router depends on.
type Oracle interface {
	GetLastPriceSyncTimestamp() time.Time
	GetPrices() map[string]sdk.Dec
	GetTvwapPrices() oracle.PricesByProvider
	GetVwapPrices() oracle.PricesByProvider
}
