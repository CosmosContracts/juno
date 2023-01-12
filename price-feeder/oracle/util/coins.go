package util

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewDecFromFloat(f float64) (sdk.Dec, error) {
	return sdk.NewDecFromStr(strconv.FormatFloat(f, 'f', -1, 64))
}

func MustNewDecFromFloat(f float64) sdk.Dec {
	return sdk.MustNewDecFromStr(strconv.FormatFloat(f, 'f', -1, 64))
}
