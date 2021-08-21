package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinter(inflation, annualProvisions sdk.Dec) Minter {
	return Minter{
		Inflation:        inflation,
		AnnualProvisions: annualProvisions,
	}
}

// InitialMinter returns an initial Minter object with a given inflation value.
func InitialMinter(inflation sdk.Dec) Minter {
	return NewMinter(
		inflation,
		sdk.NewDec(0),
	)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain
// which uses an inflation rate of 13%.
func DefaultInitialMinter() Minter {
	return InitialMinter(
		sdk.NewDecWithPrec(13, 2),
	)
}

// validate minter
func ValidateMinter(minter Minter) error {
	if minter.Inflation.IsNegative() {
		return fmt.Errorf("mint parameter Inflation should be positive, is %s",
			minter.Inflation.String())
	}
	return nil
}

// NextInflationRate returns the new inflation rate for the next hour.
func (m Minter) NextInflationRate(params Params, currentBlock sdk.Dec) sdk.Dec {
	phase := currentBlock.Quo(sdk.NewDec(int64(params.BlocksPerYear))).Ceil()

	switch {
	case phase.GT(sdk.NewDec(12)):
		return sdk.ZeroDec()

	case phase.Equal(sdk.NewDec(1)):
		return sdk.NewDecWithPrec(40, 2)

	case phase.Equal(sdk.NewDec(2)):
		return sdk.NewDecWithPrec(20, 2)

	case phase.Equal(sdk.NewDec(3)):
		return sdk.NewDecWithPrec(10, 2)

	default:
		// Phase4:  9%
		// Phase5:  8%
		// Phase6:  7%
		// ...
		// Phase12: 1%
		return sdk.NewDecWithPrec(13-phase.RoundInt64(), 2)
	}
}

// NextAnnualProvisions returns the annual provisions based on current total
// supply and inflation rate.
func (m Minter) NextAnnualProvisions(_ Params, totalSupply sdk.Int) sdk.Dec {
	return m.Inflation.MulInt(totalSupply)
}

// BlockProvision returns the provisions for a block based on the annual
// provisions rate.
func (m Minter) BlockProvision(params Params) sdk.Coin {
	provisionAmt := m.AnnualProvisions.QuoInt(sdk.NewInt(int64(params.BlocksPerYear)))
	return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}
