package lupercalia

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// set the ICS27 consensus version so InitGenesis is not run
		fromVM[icatypes.ModuleName] = icamodule.ConsensusVersion()

		// create ICS27 Controller submodule params
		controllerParams := icacontrollertypes.Params{
			ControllerEnabled: true,
		}

		// create ICS27 Host submodule params
		// NOTE TO AUDITORS: We might be able to add more message types, and I am wondering about wasm.
		hostParams := icahosttypes.Params{
			HostEnabled:   true,
			AllowMessages: []string{"/cosmos.bank.v1beta1.MsgSend"},
		}

		// initialize ICS27 module
		icamodule.InitModule(ctx, controllerParams, hostParams)

		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	}
}
