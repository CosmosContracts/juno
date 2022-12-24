package v12

import (
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

	ica "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts"
	ibcfeetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"
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
		nativeDenom := GetChainsDenomToken(ctx.ChainID())
		// transfer module consensus version has been bumped to 2
		// the above is https://github.com/cosmos/ibc-go/blob/v5.1.0/docs/migrations/v3-to-v4.md

		// Oracle
		newOracleParams := oracletypes.DefaultParams()
		keepers.OracleKeeper.SetParams(ctx, newOracleParams)

		// TokenFactory
		newTokenFactoryParams := tokenfactorytypes.Params{
			DenomCreationFee: sdk.NewCoins(sdk.NewCoin(nativeDenom, sdk.NewInt(1000000))),
		}
		keepers.TokenFactoryKeeper.SetParams(ctx, newTokenFactoryParams)

		// FeeShare
		newFeeShareParams := feesharetypes.Params{
			EnableFeeShare:  true,
			DeveloperShares: sdk.NewDecWithPrec(50, 2), // = 50%
			AllowedDenoms:   []string{nativeDenom},
		}
		keepers.FeeShareKeeper.SetParams(ctx, newFeeShareParams)

		// ICA - https://github.com/CosmosContracts/juno/blob/integrate_ica_changes/app/app.go#L846-L885
		vm[icatypes.ModuleName] = mm.Modules[icatypes.ModuleName].ConsensusVersion()

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

		// IBCFee
		vm[ibcfeetypes.ModuleName] = mm.Modules[ibcfeetypes.ModuleName].ConsensusVersion()

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
