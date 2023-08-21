package burn

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// used to override Wasmd's NewBurnCoinMessageHandler

type BurnerWasmPlugin struct {
	bk bankkeeper.Keeper
}

var _ wasmtypes.Burner = &BurnerWasmPlugin{}

func NewBurnerPlugin(bk bankkeeper.Keeper) *BurnerWasmPlugin {
	return &BurnerWasmPlugin{bk: bk}
}

func (k *BurnerWasmPlugin) BurnCoins(_ sdk.Context, _ string, _ sdk.Coins) error {
	// instead of burning, we just hold in balance and subtract from the x/mint module's total supply
	// return k.bk.BurnCoins(ctx, moduleName, amt)
	return nil
}

func (k *BurnerWasmPlugin) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, _ string, amt sdk.Coins) error {
	// we override the default send to instead sent to the "junoburn" module. Then we subtract that from the x/mint module in its query
	return k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, ModuleName, amt)
}
