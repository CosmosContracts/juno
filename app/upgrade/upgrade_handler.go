package lupercalia

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	"github.com/cosmoscontracts/juno/app"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// Thank you, cosmos hub team.
// Juno team, we will need to add cw related messages here to ensure maximum interchain intercourse.
func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		app.UpgradeKeeper.SetUpgradeHandler(
			"lupercalia",
			func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {

				fromVM[icatypes.ModuleName] = app.icaModule.ConsensusVersion()
				// create ICS27 Controller submodule params
				controllerParams := icacontrollertypes.Params{}
				// create ICS27 Host submodule params
				hostParams := icahosttypes.Params{
					HostEnabled: true,
					AllowMessages: []string{
						"/cosmos.authz.v1beta1.MsgExec",
						"/cosmos.authz.v1beta1.MsgGrant",
						"/cosmos.authz.v1beta1.MsgRevoke",
						"/cosmos.bank.v1beta1.MsgSend",
						"/cosmos.bank.v1beta1.MsgMultiSend",
						"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress",
						"/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission",
						"/cosmos.distribution.v1beta1.MsgFundCommunityPool",
						"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
						"/cosmos.feegrant.v1beta1.MsgGrantAllowance",
						"/cosmos.feegrant.v1beta1.MsgRevokeAllowance",
						"/cosmos.gov.v1beta1.MsgVoteWeighted",
						"/cosmos.gov.v1beta1.MsgSubmitProposal",
						"/cosmos.gov.v1beta1.MsgDeposit",
						"/cosmos.gov.v1beta1.MsgVote",
						"/cosmos.staking.v1beta1.MsgEditValidator",
						"/cosmos.staking.v1beta1.MsgDelegate",
						"/cosmos.staking.v1beta1.MsgUndelegate",
						"/cosmos.staking.v1beta1.MsgBeginRedelegate",
						"/cosmos.staking.v1beta1.MsgCreateValidator",
						"/cosmos.vesting.v1beta1.MsgCreateVestingAccount",
						"/ibc.applications.transfer.v1.MsgTransfer",
					},
				}

				ctx.Logger().Info("start to init interchainaccount module...")
				// initialize ICS27 module
				app.icaModule.InitModule(ctx, controllerParams, hostParams)

				ctx.Logger().Info("start to run module migrations...")

				return app.mm.RunMigrations(ctx, app.configurator, fromVM)
			},
		)
		upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
		if err != nil {
			panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
		}

		if upgradeInfo.Name == "lupercalia" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
			storeUpgrades := store.StoreUpgrades{
				Added: []string{icahosttypes.StoreKey},
			}

			// configure store loader that checks if version == upgradeHeight and applies store upgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		}

		if loadLatest {
			if err := app.LoadLatestVersion(); err != nil {
				tmos.Exit(fmt.Sprintf("failed to load latest version: %s", err))
			}
		}

		app.ScopedIBCKeeper = scopedIBCKeeper
		app.ScopedTransferKeeper = scopedTransferKeeper

		return app
	}

}
