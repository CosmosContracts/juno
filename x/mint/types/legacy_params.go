package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// TODO: Uhhh other params??? Does our fork not use them or what
// https://github.com/cosmos/cosmos-sdk/blob/828fcf2f05db0c4759ed370852b6dacc589ea472/x/mint/types/params_legacy.go

// Parameter store keys
var (
	KeyMintDenom     = []byte("MintDenom")
	KeyBlocksPerYear = []byte("BlocksPerYear")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyBlocksPerYear, &p.BlocksPerYear, validateBlocksPerYear),
	}
}
