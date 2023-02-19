package upgrades

import (
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icahostkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	mintkeeper "github.com/CosmosContracts/juno/v12/x/mint/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
)

func GetChainsDenomToken(chainID string) string {
	if strings.HasPrefix(chainID, "uni-") || strings.HasPrefix(chainID, "ares-") {
		return "ujunox"
	}
	return "ujuno"
}

// CreateV10UpgradeHandler makes an upgrade handler for v11 of Juno
func CreateV10UpgradeHandler(mm *module.Manager, cfg module.Configurator, icahostkeeper *icahostkeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

		// update ICA Host to catch missed msg
		// enumerate all because it's easier to reason about
		newIcaHostParams := icahosttypes.Params{
			HostEnabled: true,
			AllowMessages: []string{
				sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}), // missed but asked for

				sdk.MsgTypeURL(&banktypes.MsgSend{}),
				sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}), // this was missed last time
				sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
				sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
				sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
				sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
				sdk.MsgTypeURL(&govtypes.MsgVote{}),
				sdk.MsgTypeURL(&govtypes.MsgVoteWeighted{}), // required by quick
				sdk.MsgTypeURL(&authz.MsgExec{}),
				sdk.MsgTypeURL(&authz.MsgGrant{}),
				sdk.MsgTypeURL(&authz.MsgRevoke{}),
				// wasm msgs here
				// note we only support these three for now
				sdk.MsgTypeURL(&wasmtypes.MsgStoreCode{}),
				sdk.MsgTypeURL(&wasmtypes.MsgInstantiateContract{}),
				sdk.MsgTypeURL(&wasmtypes.MsgExecuteContract{}),
			},
		}
		icahostkeeper.SetParams(ctx, newIcaHostParams)

		// mint module consensus version bumped
		return mm.RunMigrations(ctx, cfg, vm)

	}
}

// CreateV11UpgradeHandler makes an upgrade handler for v11 of Juno
func CreateV11UpgradeHandler(mm *module.Manager, cfg module.Configurator, icahostkeeper *icahostkeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

		// update ICA Host to add new messages available
		// enumerate all because it's easier to reason about
		newIcaHostParams := icahosttypes.Params{
			HostEnabled: true,
			AllowMessages: []string{
				sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}), // added in v10

				sdk.MsgTypeURL(&banktypes.MsgSend{}),
				sdk.MsgTypeURL(&banktypes.MsgMultiSend{}), // this was missed last time
				sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}), // added in v10
				sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
				sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
				sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
				sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
				sdk.MsgTypeURL(&govtypes.MsgVote{}),
				sdk.MsgTypeURL(&govtypes.MsgVoteWeighted{}), // added in v10
				sdk.MsgTypeURL(&authz.MsgExec{}),
				sdk.MsgTypeURL(&authz.MsgGrant{}),
				sdk.MsgTypeURL(&authz.MsgRevoke{}),
				// wasm msgs here
				// note we only support three atm (well four inc instantiate2)
				sdk.MsgTypeURL(&wasmtypes.MsgStoreCode{}),
				sdk.MsgTypeURL(&wasmtypes.MsgInstantiateContract{}),
				sdk.MsgTypeURL(&wasmtypes.MsgInstantiateContract2{}), // added in wasmd 0.29.0
				sdk.MsgTypeURL(&wasmtypes.MsgExecuteContract{}),
			},
		}
		icahostkeeper.SetParams(ctx, newIcaHostParams)

		return mm.RunMigrations(ctx, cfg, vm)

	}
}

func CreateV12UpgradeHandler(mm *module.Manager, cfg module.Configurator, mk mintkeeper.Keeper, bk bankkeeper.Keeper, ck crisiskeeper.Keeper) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		nativeDenom := GetChainsDenomToken(ctx.ChainID())

		// Mint 100JUNO (100mil ujuno) to the distribution module
		// fixes invariance issue in wasm from governance
		amt := sdk.NewCoins(sdk.NewCoin(nativeDenom, sdk.NewInt(100_000_000)))
		if err := mk.MintCoins(ctx, amt); err != nil {
			panic(err)
		}

		if err := bk.SendCoinsFromModuleToModule(ctx, "mint", "distribution", amt); err != nil {
			panic(err)
		}

		// Increases crisis fee to 15000 JUNO (15000 000 000ujuno) to prevent DDoS attacks
		crisisAmt := sdk.NewCoin(nativeDenom, sdk.NewInt(15000_000_000))
		ck.SetConstantFee(ctx, crisisAmt)

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
