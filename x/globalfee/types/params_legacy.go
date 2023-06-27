package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ParamStoreKeyMinGasPrices store key
var ParamStoreKeyMinGasPrices = []byte("MinimumGasPricesParam")

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			ParamStoreKeyMinGasPrices, &p.MinimumGasPrices, validateMinimumGasPrices,
		),
	}
}
