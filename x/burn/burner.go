package burn

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	mintkeeper "github.com/CosmosContracts/juno/v26/x/mint/keeper"
)

// used to override Wasmd's NewBurnCoinMessageHandler

type BurnerWasmPlugin struct {
	bk bankkeeper.Keeper
	mk mintkeeper.Keeper
}

var _ wasmtypes.Burner = &BurnerWasmPlugin{}

func NewBurnerPlugin(bk bankkeeper.Keeper, mk mintkeeper.Keeper) *BurnerWasmPlugin {
	return &BurnerWasmPlugin{bk: bk, mk: mk}
}

func (k *BurnerWasmPlugin) BurnCoins(ctx sdk.Context, _ string, amt sdk.Coins) error {
	// first, try to burn the coins on bank module
	err := k.bk.BurnCoins(ctx, ModuleName, amt)
	if err != nil {
		return err
	}

	// get mint params
	params := k.mk.GetParams(ctx)

	// loop the burned coins
	for _, amount := range amt {
		// if we are burning mint denom, reduce the target staking supply
		if amount.Denom == params.MintDenom {
			if err := k.mk.ReduceTargetSupply(ctx, amount); err != nil {
				return err
			}
		}
	}

	return nil
}

func (k *BurnerWasmPlugin) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, _ string, amt sdk.Coins) error {
	return k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, ModuleName, amt)
}
