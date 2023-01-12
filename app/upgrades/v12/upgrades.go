package v12

import (
	"fmt"
	"strings"

	tokenfactorytypes "github.com/CosmWasm/token-factory/x/tokenfactory/types"
	"github.com/CosmosContracts/juno/v12/app/keepers"
	feesharetypes "github.com/CosmosContracts/juno/v12/x/feeshare/types"
	oracletypes "github.com/CosmosContracts/juno/v12/x/oracle/types"
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

	globalfeetypes "github.com/cosmos/gaia/v8/x/globalfee/types"

	ica "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"
)

// Returns "ujunox" if the chain is uni, else returns the standard ujuno token denom.
func GetChainsDenomToken(chainID string) string {
	if strings.HasPrefix(chainID, "uni-") {
		return "ujunox"
	}
	return "ujuno"
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

		// Oracle
		newOracleParams := oracletypes.DefaultParams()

		// add osmosis to the oracle params
		osmosisDenom := oracletypes.Denom{
			BaseDenom:   "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518",
			SymbolDenom: "OSMO",
			Exponent:    uint32(6),
		}

		allDenoms := oracletypes.DefaultAcceptList
		allDenoms = append(allDenoms, osmosisDenom)

		newOracleParams.AcceptList = allDenoms
		newOracleParams.PriceTrackingList = allDenoms
		logger.Info(fmt.Sprintf("Oracle params set: %s", newOracleParams.String()))

		keepers.OracleKeeper.SetParams(ctx, newOracleParams)

		// TokenFactory
		newTokenFactoryParams := tokenfactorytypes.Params{
			DenomCreationFee: sdk.NewCoins(sdk.NewCoin(nativeDenom, sdk.NewInt(1000000))),
		}
		keepers.TokenFactoryKeeper.SetParams(ctx, newTokenFactoryParams)
		logger.Info("set tokenfactory params")

		// FeeShare
		newFeeShareParams := feesharetypes.Params{
			EnableFeeShare:  true,
			DeveloperShares: sdk.NewDecWithPrec(50, 2), // = 50%
			AllowedDenoms:   []string{nativeDenom},
		}
		keepers.FeeShareKeeper.SetParams(ctx, newFeeShareParams)
		logger.Info("set feeshare params")

		// ICA - https://github.com/CosmosContracts/juno/blob/integrate_ica_changes/app/app.go#L846-L885
		vm[icatypes.ModuleName] = mm.Modules[icatypes.ModuleName].ConsensusVersion()
		logger.Info("upgraded icatypes version")

		// create ICS27 Controller submodule params, controller module not enabled.
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
		logger.Info("upgraded ica module")

		// IBCFee
		vm[ibcfeetypes.ModuleName] = mm.Modules[ibcfeetypes.ModuleName].ConsensusVersion()
		logger.Info(fmt.Sprintf("ibcfee module version %s set", fmt.Sprint(vm[ibcfeetypes.ModuleName])))

		// Run migrations
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)

		// GlobalFee - This must run AFTER migrations to update the default param space.
		minGasPrices := sdk.DecCoins{
			// 0.0025ujuno
			sdk.NewDecCoinFromDec(nativeDenom, sdk.NewDecWithPrec(25, 4)),
			// 0.001 ATOM CHANNEL-1 -> `junod q ibc-transfer denom-trace ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9`
			sdk.NewDecCoinFromDec("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9", sdk.NewDecWithPrec(1, 3)),
		}
		s, ok := keepers.ParamsKeeper.GetSubspace(globalfeetypes.ModuleName)
		if !ok {
			panic("global fee params subspace not found")
		}
		s.Set(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, minGasPrices)
		logger.Info(fmt.Sprintf("upgraded global fee params to %s", minGasPrices))

		return versionMap, err
	}
}
