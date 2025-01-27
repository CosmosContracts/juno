package app

import (
	wasm "github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	icq "github.com/cosmos/ibc-apps/modules/async-icq/v8"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v8/types"
	ibc_hooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8/types"
	wasmlc "github.com/cosmos/ibc-go/modules/light-clients/08-wasm"
	wasmlctypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibcfee "github.com/cosmos/ibc-go/v8/modules/apps/29-fee"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	transfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"cosmossdk.io/x/evidence"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/nft"
	nftmodule "cosmossdk.io/x/nft/module"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	"github.com/CosmosContracts/juno/v27/x/clock"
	clocktypes "github.com/CosmosContracts/juno/v27/x/clock/types"
	cwhooks "github.com/CosmosContracts/juno/v27/x/cw-hooks"
	"github.com/CosmosContracts/juno/v27/x/drip"
	driptypes "github.com/CosmosContracts/juno/v27/x/drip/types"
	feepay "github.com/CosmosContracts/juno/v27/x/feepay"
	feepaytypes "github.com/CosmosContracts/juno/v27/x/feepay/types"
	feeshare "github.com/CosmosContracts/juno/v27/x/feeshare"
	feesharetypes "github.com/CosmosContracts/juno/v27/x/feeshare/types"
	"github.com/CosmosContracts/juno/v27/x/globalfee"
	"github.com/CosmosContracts/juno/v27/x/mint"
	minttypes "github.com/CosmosContracts/juno/v27/x/mint/types"
	"github.com/CosmosContracts/juno/v27/x/tokenfactory"
	tokenfactorytypes "github.com/CosmosContracts/juno/v27/x/tokenfactory/types"
)

func appModules(
	app *App,
	txConfig client.TxConfig,
	appCodec codec.Codec,
	skipGenesisInvariants bool,
) []module.AppModule {
	bondDenom := app.GetChainBondDenom()

	return []module.AppModule{
		// Cosmos SDK modules
		genutil.NewAppModule(
			app.AppKeepers.AccountKeeper,
			app.AppKeepers.StakingKeeper,
			app,
			txConfig,
		),
		auth.NewAppModule(appCodec, app.AppKeepers.AccountKeeper, nil, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper),
		bank.NewAppModule(appCodec, app.AppKeepers.BankKeeper, app.AppKeepers.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		feegrantmodule.NewAppModule(appCodec, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.AppKeepers.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.AppKeepers.GovKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		slashing.NewAppModule(appCodec, app.AppKeepers.SlashingKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.AppKeepers.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		distr.NewAppModule(appCodec, app.AppKeepers.DistrKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.AppKeepers.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, app.AppKeepers.StakingKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.AppKeepers.UpgradeKeeper, app.AppKeepers.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.AppKeepers.EvidenceKeeper),
		params.NewAppModule(app.AppKeepers.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AppKeepers.AuthzKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.interfaceRegistry),
		nftmodule.NewAppModule(appCodec, app.AppKeepers.NFTKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, app.AppKeepers.ConsensusParamsKeeper),
		crisis.NewAppModule(app.AppKeepers.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		// Juno modules
		mint.NewAppModule(appCodec, app.AppKeepers.MintKeeper, app.AppKeepers.AccountKeeper, bondDenom),
		tokenfactory.NewAppModule(app.AppKeepers.TokenFactoryKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.GetSubspace(tokenfactorytypes.ModuleName)),
		globalfee.NewAppModule(appCodec, app.AppKeepers.GlobalFeeKeeper, bondDenom),
		feepay.NewAppModule(app.AppKeepers.FeePayKeeper, app.AppKeepers.AccountKeeper),
		feeshare.NewAppModule(app.AppKeepers.FeeShareKeeper, app.AppKeepers.AccountKeeper, app.GetSubspace(feesharetypes.ModuleName)),
		drip.NewAppModule(app.AppKeepers.DripKeeper, app.AppKeepers.AccountKeeper),
		clock.NewAppModule(appCodec, app.AppKeepers.ClockKeeper),
		cwhooks.NewAppModule(appCodec, app.AppKeepers.CWHooksKeeper),
		// IBC modules
		capability.NewAppModule(appCodec, *app.AppKeepers.CapabilityKeeper, false),
		ibc.NewAppModule(app.AppKeepers.IBCKeeper),
		transfer.NewAppModule(app.AppKeepers.TransferKeeper),
		ica.NewAppModule(&app.AppKeepers.ICAControllerKeeper, &app.AppKeepers.ICAHostKeeper),
		ibcfee.NewAppModule(app.AppKeepers.IBCFeeKeeper),
		ibc_hooks.NewAppModule(app.AppKeepers.AccountKeeper),
		icq.NewAppModule(app.AppKeepers.ICQKeeper, app.GetSubspace(icqtypes.ModuleName)),
		packetforward.NewAppModule(app.AppKeepers.PacketForwardKeeper, app.GetSubspace(packetforwardtypes.ModuleName)),
		// Wasm modules
		wasm.NewAppModule(appCodec, &app.AppKeepers.WasmKeeper, app.AppKeepers.StakingKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		wasmlc.NewAppModule(app.AppKeepers.WasmClientKeeper),
	}
}

// orderBeginBlockers tell the app's module manager how to set the order of
// BeginBlockers, which are run at the beginning of every block.
func orderBeginBlockers() []string {
	return []string{
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		consensusparamtypes.ModuleName,
		// additional modules
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		packetforwardtypes.ModuleName,
		ibcfeetypes.ModuleName,
		icqtypes.ModuleName,
		tokenfactorytypes.ModuleName,
		driptypes.ModuleName,
		feepaytypes.ModuleName,
		feesharetypes.ModuleName,
		globalfee.ModuleName,
		wasmtypes.ModuleName,
		ibchookstypes.ModuleName,
		clocktypes.ModuleName,
		cwhooks.ModuleName,
		wasmlctypes.ModuleName,
	}
}

func orderEndBlockers() []string {
	return []string{
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		consensusparamtypes.ModuleName,
		// additional non simd modules
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		packetforwardtypes.ModuleName,
		ibcfeetypes.ModuleName,
		icqtypes.ModuleName,
		tokenfactorytypes.ModuleName,
		driptypes.ModuleName,
		feepaytypes.ModuleName,
		feesharetypes.ModuleName,
		globalfee.ModuleName,
		wasmtypes.ModuleName,
		ibchookstypes.ModuleName,
		clocktypes.ModuleName,
		cwhooks.ModuleName,
		wasmlctypes.ModuleName,
	}
}

// NOTE: The genutils module must occur after staking so that pools are
// properly initialized with tokens from genesis accounts.
//
// NOTE: Capability module must occur first so that it can initialize any capabilities
// so that other modules that want to create or claim capabilities afterwards in InitChain
// can do so safely.
func orderInitBlockers() []string {
	return []string{
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		feegrant.ModuleName,
		nft.ModuleName,
		consensusparamtypes.ModuleName,
		// additional non simd modules
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		packetforwardtypes.ModuleName,
		ibcfeetypes.ModuleName,
		icqtypes.ModuleName,
		tokenfactorytypes.ModuleName,
		driptypes.ModuleName,
		feepaytypes.ModuleName,
		feesharetypes.ModuleName,
		globalfee.ModuleName,
		wasmtypes.ModuleName,
		ibchookstypes.ModuleName,
		clocktypes.ModuleName,
		cwhooks.ModuleName,
		wasmlctypes.ModuleName,
	}
}
