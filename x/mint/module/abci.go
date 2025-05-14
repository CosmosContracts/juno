package module

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v29/x/mint/keeper"
	"github.com/CosmosContracts/juno/v29/x/mint/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// fetch stored minter & params
	minter, err := k.GetMinter(ctx)
	if err != nil {
		return err
	}

	if minter.Inflation.Equal(sdkmath.LegacyZeroDec()) {
		return nil
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	currentBlock := uint64(sdkCtx.BlockHeight())
	totalSupply := k.TokenSupply(ctx, params.MintDenom)
	nextPhase := minter.NextPhase(params, totalSupply)

	if nextPhase != minter.Phase {
		newInflation := minter.PhaseInflationRate(nextPhase)
		minter.Inflation = newInflation
		minter.Phase = nextPhase
		minter.StartPhaseBlock = currentBlock
		minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)
		minter.TargetSupply = totalSupply.Add(minter.AnnualProvisions.TruncateInt())
		err = k.SetMinter(ctx, minter)
		if err != nil {
			return err
		}
	}

	// mint coins, update supply
	mintedCoin := minter.BlockProvision(params, totalSupply)
	mintedCoins := sdk.NewCoins(mintedCoin)

	err = k.MintCoins(ctx, mintedCoins)
	if err != nil {
		return err
	}

	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, mintedCoins)
	if err != nil {
		return err
	}

	bondedRatio := k.BondedRatio(ctx)

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
			sdk.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
			sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)

	return nil
}
