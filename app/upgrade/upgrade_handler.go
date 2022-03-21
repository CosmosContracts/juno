package upgrade

import (
	"fmt"

	"github.com/CosmosContracts/juno/app"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icamodule "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	"github.com/tendermint/tendermint/libs/os"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// Thank you, cosmos hub team.
// Juno team, we will need to add cw related messages here to ensure maximum interchain intercourse.
func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator, staking *stakingkeeper.Keeper, bank *bankkeeper.BaseKeeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		upgradekeeper.Keeper.SetUpgradeHandler(
			"lupercalia",
			func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {

				fromVM[icatypes.ModuleName] = icamodule.ConsensusVersion()
				// create ICS27 Controller submodule params
				controllerParams := icacontrollertypes.Params{}
				// create ICS27 Host submodule params
				hostParams := icahosttypes.Params{
					HostEnabled: true,
					AllowMessages: []string{
						"cosmwasm.wasm.v1.MsgInstantiate",
						"cosmwasm.wasm.v1.MsgExecute",
						"cosmwasm.wasm.v1.MsgStoreCode",
						"cosmwasm.wasm.v1.MsgMigrateContract",
						"cosmwasm.wasm.v1.UpdateAdmin",
						"cosmwasm.wasm.v1.MsgClearAdmin",
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
				icamodule.AppModule.InitModule(ctx, controllerParams, hostParams)

				ctx.Logger().Info("start to run module migrations...")

				return mm.RunMigrations(ctx, mm.configurator, fromVM)
			},
		)

		upgradeInfo, err := juno.UpgradeKeeper.ReadUpgradeInfoFromDisk(ctx, UpgradeName)
		if err != nil {
			panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
		}

		if upgradeInfo.Name == "lupercalia" && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
			storeUpgrades := storetypes.StoreUpgrades{
				Added: []string{icahosttypes.StoreKey},
			}

			// configure store loader that checks if version == upgradeHeight and applies store upgrades
			mm.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		}

		if app.LoadLatest() {
			if err := app.LoadLatestVersion(); err != nil {
				os.Exit(fmt.Sprintf("failed to load latest version: %s", err))
			}
		}

		return mm.RunMigrations(ctx, app.configurator, fromVM)

	}
}
