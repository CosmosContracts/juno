package mint

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v16/x/mint/keeper"
	"github.com/CosmosContracts/juno/v16/x/mint/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored minter
	minter := k.GetMinter(ctx)

	// inflation phase end
	if minter.Inflation.Equal(sdk.ZeroDec()) {
		return
	}

	// fetch stored params
	params := k.GetParams(ctx)
	currentBlock := uint64(ctx.BlockHeight())

	// fetch current total supply
	totalSupply := k.TokenSupply(ctx, params.MintDenom)

	// check if we need to change phase
	nextPhase := minter.NextPhase(params, totalSupply)

	if nextPhase != minter.Phase {
		// store new inflation rate by phase
		newInflation := minter.PhaseInflationRate(nextPhase)
		minter.Inflation = newInflation
		minter.Phase = nextPhase
		minter.StartPhaseBlock = currentBlock
		minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)
		minter.TargetSupply = totalSupply.Add(minter.AnnualProvisions.TruncateInt())
		k.SetMinter(ctx, minter)

		// inflation phase end
		if minter.Inflation.Equal(sdk.ZeroDec()) {
			return
		}
	}

	// mint coins, update supply
	mintedCoin := minter.BlockProvision(params, totalSupply)
	mintedCoins := sdk.NewCoins(mintedCoin)

	err := k.MintCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
			sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)
}
