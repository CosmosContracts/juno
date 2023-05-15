package app

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	encparams "github.com/CosmosContracts/juno/v15/app/params"
	feeshare "github.com/CosmosContracts/juno/v15/x/feeshare"
	feesharetypes "github.com/CosmosContracts/juno/v15/x/feeshare/types"
	"github.com/CosmosContracts/juno/v15/x/globalfee"
	"github.com/CosmosContracts/juno/v15/x/mint"
	minttypes "github.com/CosmosContracts/juno/v15/x/mint/types"
	"github.com/CosmosContracts/juno/v15/x/tokenfactory"
	tokenfactorytypes "github.com/CosmosContracts/juno/v15/x/tokenfactory/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
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
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibcfee "github.com/cosmos/ibc-go/v7/modules/apps/29-fee"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	transfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	icq "github.com/strangelove-ventures/async-icq/v7"
	icqtypes "github.com/strangelove-ventures/async-icq/v7/types"
)

// ModuleBasics defines the module BasicManager is in charge of setting up basic,
// non-dependant module elements, such as codec registration
// and genesis verification.
var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	distr.AppModuleBasic{},
	gov.NewAppModuleBasic(getGovProposalHandlers()),
	params.AppModuleBasic{},
	crisis.AppModuleBasic{},
	slashing.AppModuleBasic{},
	upgrade.AppModuleBasic{},
	evidence.AppModuleBasic{},
	authzmodule.AppModuleBasic{},
	vesting.AppModuleBasic{},
	consensus.AppModuleBasic{},
	// non sdk modules
	wasm.AppModuleBasic{},
	ibc.AppModuleBasic{},
	ibctm.AppModuleBasic{},
	transfer.AppModuleBasic{},
	ica.AppModuleBasic{},
	ibcfee.AppModuleBasic{},
	icq.AppModuleBasic{},
	feegrantmodule.AppModuleBasic{},
	tokenfactory.AppModuleBasic{},
	feeshare.AppModuleBasic{},
	globalfee.AppModuleBasic{},
	// ibchooks.AppModuleBasic{},
	// packetforward.AppModuleBasic{},
)

func appModules(
	app *App,
	encodingConfig encparams.EncodingConfig,
	skipGenesisInvariants bool,
) []module.AppModule {
	appCodec := encodingConfig.Marshaler

	return []module.AppModule{
		genutil.NewAppModule(
			app.AppKeepers.AccountKeeper,
			app.AppKeepers.StakingKeeper,
			app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.AppKeepers.AccountKeeper, nil, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper),
		bank.NewAppModule(appCodec, app.AppKeepers.BankKeeper, app.AppKeepers.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.AppKeepers.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.AppKeepers.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.AppKeepers.GovKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.AppKeepers.MintKeeper, app.AppKeepers.AccountKeeper),
		slashing.NewAppModule(appCodec, app.AppKeepers.SlashingKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.AppKeepers.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		distr.NewAppModule(appCodec, app.AppKeepers.DistrKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.AppKeepers.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, app.AppKeepers.StakingKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.AppKeepers.UpgradeKeeper),
		evidence.NewAppModule(app.AppKeepers.EvidenceKeeper),
		ibc.NewAppModule(app.AppKeepers.IBCKeeper),
		params.NewAppModule(app.AppKeepers.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AppKeepers.AuthzKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, app.AppKeepers.ConsensusParamsKeeper),
		transfer.NewAppModule(app.AppKeepers.TransferKeeper),
		ibcfee.NewAppModule(app.AppKeepers.IBCFeeKeeper),
		tokenfactory.NewAppModule(app.AppKeepers.TokenFactoryKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper),
		globalfee.NewAppModule(app.AppKeepers.GetSubspace(globalfee.ModuleName)),
		feeshare.NewAppModule(app.AppKeepers.FeeShareKeeper, app.AppKeepers.AccountKeeper),
		wasm.NewAppModule(appCodec, &app.AppKeepers.WasmKeeper, app.AppKeepers.StakingKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		ica.NewAppModule(&app.AppKeepers.ICAControllerKeeper, &app.AppKeepers.ICAHostKeeper),
		// ibchooks.NewAppModule(app.AppKeepers.AccountKeeper),
		crisis.NewAppModule(app.AppKeepers.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		// IBC modules
		icq.NewAppModule(app.AppKeepers.ICQKeeper),
		// packetforward.NewAppModule(app.AppKeepers.PacketForwardKeeper),
	}
}

// simulationModules returns modules for simulation manager
// define the order of the modules for deterministic simulationss
func simulationModules(
	app *App,
	encodingConfig encparams.EncodingConfig,
	_ bool,
) []module.AppModuleSimulation {
	appCodec := encodingConfig.Marshaler

	return []module.AppModuleSimulation{
		auth.NewAppModule(appCodec, app.AppKeepers.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, app.AppKeepers.BankKeeper, app.AppKeepers.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.AppKeepers.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.AppKeepers.FeeGrantKeeper, app.interfaceRegistry),
		authzmodule.NewAppModule(appCodec, app.AppKeepers.AuthzKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.AppKeepers.GovKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.AppKeepers.MintKeeper, app.AppKeepers.AccountKeeper),
		staking.NewAppModule(appCodec, app.AppKeepers.StakingKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		distr.NewAppModule(appCodec, app.AppKeepers.DistrKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.AppKeepers.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		slashing.NewAppModule(appCodec, app.AppKeepers.SlashingKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.AppKeepers.StakingKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		params.NewAppModule(app.AppKeepers.ParamsKeeper),
		evidence.NewAppModule(app.AppKeepers.EvidenceKeeper),
		wasm.NewAppModule(appCodec, &app.AppKeepers.WasmKeeper, app.AppKeepers.StakingKeeper, app.AppKeepers.AccountKeeper, app.AppKeepers.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		ibc.NewAppModule(app.AppKeepers.IBCKeeper),
		transfer.NewAppModule(app.AppKeepers.TransferKeeper),
		feeshare.NewAppModule(app.AppKeepers.FeeShareKeeper, app.AppKeepers.AccountKeeper),
		ibcfee.NewAppModule(app.AppKeepers.IBCFeeKeeper),
	}
}

// orderBeginBlockers tell the app's module manager how to set the order of
// BeginBlockers, which are run at the beginning of every block.
func orderBeginBlockers() []string {
	return []string{
		upgradetypes.ModuleName,
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
		consensusparamtypes.ModuleName,
		// additional modules
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		// packetforwardtypes.ModuleName,
		ibcfeetypes.ModuleName,
		icqtypes.ModuleName,
		tokenfactorytypes.ModuleName,
		feesharetypes.ModuleName,
		globalfee.ModuleName,
		wasm.ModuleName,
		// ibchookstypes.ModuleName,
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
		consensusparamtypes.ModuleName,
		// additional non simd modules
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		// packetforwardtypes.ModuleName,
		ibcfeetypes.ModuleName,
		icqtypes.ModuleName,
		tokenfactorytypes.ModuleName,
		feesharetypes.ModuleName,
		globalfee.ModuleName,
		wasm.ModuleName,
		// ibchookstypes.ModuleName,
	}
}

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
		consensusparamtypes.ModuleName,
		// additional non simd modules
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		// packetforwardtypes.ModuleName,
		ibcfeetypes.ModuleName,
		icqtypes.ModuleName,
		tokenfactorytypes.ModuleName,
		feesharetypes.ModuleName,
		globalfee.ModuleName,
		wasm.ModuleName,
		// ibchookstypes.ModuleName,
	}
}
