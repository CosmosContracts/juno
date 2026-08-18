package types_test

import (
	"math"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/CosmosContracts/juno/v31/x/feepay/types"
)

func TestContractBalanceBoundaries(t *testing.T) {
	max := sdkmath.NewIntFromUint64(math.MaxUint64)
	aboveMax := max.AddRaw(1)

	got, err := types.ContractBalanceAfterAddition(0, max)
	require.NoError(t, err)
	require.Equal(t, uint64(math.MaxUint64), got)

	_, err = types.ContractBalanceAfterAddition(0, aboveMax)
	require.ErrorIs(t, err, types.ErrFeePayAmountOutOfRange)

	_, err = types.ContractBalanceAfterAddition(math.MaxUint64, sdkmath.OneInt())
	require.ErrorIs(t, err, types.ErrFeePayBalanceOverflow)

	got, err = types.ContractBalanceAfterSubtraction(math.MaxUint64, max)
	require.NoError(t, err)
	require.Zero(t, got)

	_, err = types.ContractBalanceAfterSubtraction(math.MaxUint64, aboveMax)
	require.ErrorIs(t, err, types.ErrFeePayAmountOutOfRange)

	_, err = types.ContractBalanceAfterSubtraction(0, sdkmath.OneInt())
	require.ErrorIs(t, err, types.ErrContractNotEnoughFunds)
}
