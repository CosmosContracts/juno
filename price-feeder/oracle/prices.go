package oracle

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/price-feeder/oracle/provider"
)

type (
	PricesByProvider map[provider.Name]map[string]sdk.Dec

	PricesWithMutex struct {
		prices PricesByProvider
		mx     sync.RWMutex
	}
)

// SetPrices sets the PricesWithMutex.prices value surrounded by a write lock
func (pwm *PricesWithMutex) SetPrices(prices PricesByProvider) {
	pwm.mx.Lock()
	defer pwm.mx.Unlock()

	pwm.prices = prices
}

// GetPricesClone retrieves a clone of PricesWithMutex.prices
// surrounded by a read lock
func (pwm *PricesWithMutex) GetPricesClone() PricesByProvider {
	pwm.mx.RLock()
	defer pwm.mx.RUnlock()
	return pwm.clonePrices()
}

// clonePrices returns a deep copy of PricesWithMutex.prices
func (pwm *PricesWithMutex) clonePrices() PricesByProvider {
	clone := make(PricesByProvider, len(pwm.prices))
	for provider, prices := range pwm.prices {
		pricesClone := make(map[string]sdk.Dec, len(prices))
		for denom, price := range prices {
			pricesClone[denom] = price
		}
		clone[provider] = pricesClone
	}
	return clone
}
