package v12

import (
	"fmt"
	"strings"

	tokenfactorytypes "github.com/CosmWasm/token-factory/x/tokenfactory/types"
	"github.com/CosmosContracts/juno/v12/app/keepers"
	feesharetypes "github.com/CosmosContracts/juno/v12/x/feeshare/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/authz"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	ica "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"

	globalfeetypes "github.com/cosmos/gaia/v8/x/globalfee/types"

	// tendermint logger
	tmlog "github.com/tendermint/tendermint/libs/log"
)

// Returns "ujuno" if the chainID starts with "juno-" (ex: juno-1 or juno-t1 for local)
// else its the uni testnet
func GetChainsDenomToken(chainID string) string {
	if strings.HasPrefix(chainID, "juno-") {
		return "ujuno"
	}
	return "ujunox"
}

// CreateV12UpgradeHandler makes an upgrade handler for v12 of Juno
func CreateV12UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// transfer module consensus version has been bumped to 2
		// the above is https://github.com/cosmos/ibc-go/blob/v5.1.0/docs/migrations/v3-to-v4.md
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		upgradeTokenFactory(ctx, logger, nativeDenom, keepers)
		upgradeFeeShare(ctx, logger, nativeDenom, keepers)
		upgradeICAModule(ctx, logger, vm, mm)
		upgradeIBCFee(logger, vm, mm)

		// Run migrations
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)

		// GlobalFee must run after migrations to update the param space set by default
		upgradeGlobalFee(ctx, logger, nativeDenom, keepers)

		return versionMap, err
	}
}

func upgradeIBCFee(logger tmlog.Logger, vm module.VersionMap, mm *module.Manager) {
	// IBCFee
	vm[ibcfeetypes.ModuleName] = mm.Modules[ibcfeetypes.ModuleName].ConsensusVersion()
	logger.Info(fmt.Sprintf("ibcfee module version %s set", fmt.Sprint(vm[ibcfeetypes.ModuleName])))
}

func upgradeTokenFactory(ctx sdk.Context, logger tmlog.Logger, nativeDenom string, keepers *keepers.AppKeepers) {
	newTokenFactoryParams := tokenfactorytypes.Params{
		DenomCreationFee: sdk.NewCoins(sdk.NewCoin(nativeDenom, sdk.NewInt(1000000))),
	}
	keepers.TokenFactoryKeeper.SetParams(ctx, newTokenFactoryParams)
	logger.Info("upgraded token factory params")
}

func upgradeFeeShare(ctx sdk.Context, logger tmlog.Logger, nativeDenom string, keepers *keepers.AppKeepers) {
	newFeeShareParams := feesharetypes.Params{
		EnableFeeShare:  true,
		DeveloperShares: sdk.NewDecWithPrec(50, 2), // = 50%
		AllowedDenoms:   []string{nativeDenom},
	}
	keepers.FeeShareKeeper.SetParams(ctx, newFeeShareParams)
	logger.Info("upgraded fee share params")
}

func upgradeGlobalFee(ctx sdk.Context, logger tmlog.Logger, nativeDenom string, keepers *keepers.AppKeepers) {
	// This must run AFTER migrations to update the default param space.
	minGasPrices := sdk.DecCoins{
		// 0.0025ujuno
		sdk.NewDecCoinFromDec(nativeDenom, sdk.NewDecWithPrec(25, 4)),
		// 0.001 ATOM CHANNEL-1 -> `junod q ibc-transfer denom-trace ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9`
		sdk.NewDecCoinFromDec("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", sdk.NewDecWithPrec(1, 4)),
	}
	s, ok := keepers.ParamsKeeper.GetSubspace(globalfeetypes.ModuleName)
	if !ok {
		panic("global fee params subspace not found")
	}
	s.Set(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, minGasPrices)
	logger.Info(fmt.Sprintf("upgraded global fee params to %s", minGasPrices))
}

func upgradeICAModule(ctx sdk.Context, logger tmlog.Logger, vm module.VersionMap, mm *module.Manager) {
	// ICA - https://github.com/CosmosContracts/juno/blob/integrate_ica_changes/app/app.go#L846-L885
	vm[icatypes.ModuleName] = mm.Modules[icatypes.ModuleName].ConsensusVersion()
	logger.Info(fmt.Sprintf("upgraded ICA module version %s", fmt.Sprint(vm[icatypes.ModuleName])))

	// create ICS27 Controller submodule params with controller module enabled.
	controllerParams := icacontrollertypes.Params{
		ControllerEnabled: true,
	}

	// create ICS27 Host submodule params
	hostParams := icahosttypes.Params{
		HostEnabled: true,
		AllowMessages: []string{
			sdk.MsgTypeURL(&banktypes.MsgSend{}),
			sdk.MsgTypeURL(&banktypes.MsgMultiSend{}),
			sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
			sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
			sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
			sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
			sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
			sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
			sdk.MsgTypeURL(&govtypes.MsgVote{}),
			sdk.MsgTypeURL(&govtypes.MsgVoteWeighted{}),
			sdk.MsgTypeURL(&govtypes.MsgSubmitProposal{}),
			sdk.MsgTypeURL(&govtypes.MsgDeposit{}),
			sdk.MsgTypeURL(&authz.MsgExec{}),
			sdk.MsgTypeURL(&authz.MsgGrant{}),
			sdk.MsgTypeURL(&authz.MsgRevoke{}),
			// wasm
			sdk.MsgTypeURL(&wasmtypes.MsgStoreCode{}),
			sdk.MsgTypeURL(&wasmtypes.MsgInstantiateContract{}),
			sdk.MsgTypeURL(&wasmtypes.MsgInstantiateContract2{}),
			sdk.MsgTypeURL(&wasmtypes.MsgExecuteContract{}),
			sdk.MsgTypeURL(&wasmtypes.MsgMigrateContract{}),
			sdk.MsgTypeURL(&wasmtypes.MsgUpdateAdmin{}),
			sdk.MsgTypeURL(&wasmtypes.MsgClearAdmin{}),
			sdk.MsgTypeURL(&wasmtypes.MsgIBCSend{}),
			sdk.MsgTypeURL(&wasmtypes.MsgIBCCloseChannel{}),
			// tokenfactory
			sdk.MsgTypeURL(&tokenfactorytypes.MsgCreateDenom{}),
			sdk.MsgTypeURL(&tokenfactorytypes.MsgMint{}),
			sdk.MsgTypeURL(&tokenfactorytypes.MsgBurn{}),
			sdk.MsgTypeURL(&tokenfactorytypes.MsgChangeAdmin{}),
			sdk.MsgTypeURL(&tokenfactorytypes.MsgSetDenomMetadata{}),
			// feeshare
			sdk.MsgTypeURL(&feesharetypes.MsgRegisterFeeShare{}),
			sdk.MsgTypeURL(&feesharetypes.MsgUpdateFeeShare{}),
			sdk.MsgTypeURL(&feesharetypes.MsgUpdateFeeShare{}),
			sdk.MsgTypeURL(&feesharetypes.MsgCancelFeeShare{}),
		},
	}

	// initialize ICS27 module
	icamodule, correctTypecast := mm.Modules[icatypes.ModuleName].(ica.AppModule)
	if !correctTypecast {
		panic("mm.Modules[icatypes.ModuleName] is not of type ica.AppModule")
	}
	icamodule.InitModule(ctx, controllerParams, hostParams)
	logger.Info("icamodule initialized")
}
