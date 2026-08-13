package keeper

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/CosmosContracts/juno/v30/x/feepay/types"
)

// InitGenesis import module genesis
func (k Keeper) InitGenesis(
	ctx sdk.Context,
	data types.GenesisState,
) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Sum the imported contract balances; every unit a contract claims must
	// be backed by funds in the feepay module account, otherwise feepay txs
	// would later fail mid-pipeline (or spend other contracts' escrow) when
	// the module account cannot cover a "funded" contract's fee.
	totalBalances := math.ZeroInt()
	for _, feepay := range data.FeePayContracts {
		// TODO: future, add all wallet interactions for exports?
		k.SetFeePayContract(ctx, feepay)
		totalBalances = totalBalances.Add(math.NewIntFromUint64(feepay.Balance))
	}

	feeDenom, err := k.feeDenom(ctx)
	if err != nil {
		panic(err)
	}
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := k.bankKeeper.GetBalance(ctx, moduleAddr, feeDenom).Amount
	if totalBalances.GT(moduleBalance) {
		panic(fmt.Sprintf(
			"feepay genesis: sum of imported contract balances (%s%s) exceeds feepay module account balance (%s%s)",
			totalBalances, feeDenom, moduleBalance, feeDenom,
		))
	}
}

// ExportGenesis export module state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	contracts := k.GetAllContracts(ctx)

	return &types.GenesisState{
		Params:          params,
		FeePayContracts: contracts,
	}
}
