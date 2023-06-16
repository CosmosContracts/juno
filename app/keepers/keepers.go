package keepers

import (
	"path/filepath"

	"github.com/spf13/cast"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	globalfeetypes "github.com/CosmosContracts/juno/v16/x/globalfee/types"
	"github.com/CosmosContracts/juno/v16/x/ibchooks"
	ibchookskeeper "github.com/CosmosContracts/juno/v16/x/ibchooks/keeper"
	ibchookstypes "github.com/CosmosContracts/juno/v16/x/ibchooks/types"
	mintkeeper "github.com/CosmosContracts/juno/v16/x/mint/keeper"
	minttypes "github.com/CosmosContracts/juno/v16/x/mint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtype "github.com/cosmos/cosmos-sdk/x/auth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evedencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/nft"
	nftkeeper "github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcfee "github.com/cosmos/ibc-go/v7/modules/apps/29-fee"
	ibcfeekeeper "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	transfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	packetforward "github.com/strangelove-ventures/packet-forward-middleware/v7/router"
	packetforwardkeeper "github.com/strangelove-ventures/packet-forward-middleware/v7/router/keeper"
	packetforwardtypes "github.com/strangelove-ventures/packet-forward-middleware/v7/router/types"

	"github.com/CosmosContracts/juno/v16/x/tokenfactory/bindings"
	tokenfactorykeeper "github.com/CosmosContracts/juno/v16/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/CosmosContracts/juno/v16/x/tokenfactory/types"

	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"

	feesharekeeper "github.com/CosmosContracts/juno/v16/x/feeshare/keeper"
	feesharetypes "github.com/CosmosContracts/juno/v16/x/feeshare/types"

	icq "github.com/strangelove-ventures/async-icq/v7"
	icqkeeper "github.com/strangelove-ventures/async-icq/v7/keeper"
	icqtypes "github.com/strangelove-ventures/async-icq/v7/types"

	// ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/CosmosContracts/juno/v16/x/globalfee"
)

var (
	wasmCapabilities = "iterator,staking,stargate,token_factory,cosmwasm_1_1,cosmwasm_1_2"

	tokenFactoryCapabilities = []string{
		tokenfactorytypes.EnableBurnFrom,
		tokenfactorytypes.EnableForceTransfer,
		tokenfactorytypes.EnableSetMetadata,
	}
)

// module account permissions
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	distrtypes.ModuleName:          nil,
	minttypes.ModuleName:           {authtypes.Minter},
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:            {authtypes.Burner},
	nft.ModuleName:                 nil,
	icqtypes.ModuleName:            nil,
	ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
	icatypes.ModuleName:            nil,
	ibcfeetypes.ModuleName:         nil,
	wasm.ModuleName:                {authtypes.Burner},
	tokenfactorytypes.ModuleName:   {authtypes.Minter, authtypes.Burner},
}

type AppKeepers struct {
	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.BaseKeeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	IBCKeeper             *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	ICQKeeper             icqkeeper.Keeper
	IBCFeeKeeper          ibcfeekeeper.Keeper
	IBCHooksKeeper        *ibchookskeeper.Keeper
	PacketForwardKeeper   *packetforwardkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	TransferKeeper        ibctransferkeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	NFTKeeper             nftkeeper.Keeper
	FeeShareKeeper        feesharekeeper.Keeper
	ContractKeeper        *wasmkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper

	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedICQKeeper           capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedFeeMockKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper

	WasmKeeper         wasm.Keeper
	scopedWasmKeeper   capabilitykeeper.ScopedKeeper
	TokenFactoryKeeper tokenfactorykeeper.Keeper

	// Middleware wrapper
	Ics20WasmHooks   *ibchooks.WasmHooks
	HooksICS4Wrapper ibchooks.ICS4Middleware
}

func NewAppKeepers(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	cdc *codec.LegacyAmino,
	maccPerms map[string][]string,
	enabledProposals []wasm.ProposalType,
	appOpts servertypes.AppOptions,
	wasmOpts []wasm.Option,
) AppKeepers {
	appKeepers := AppKeepers{}

	// Set keys KVStoreKey, TransientStoreKey, MemoryStoreKey
	appKeepers.GenerateKeys()
	keys := appKeepers.GetKVStoreKey()
	tkeys := appKeepers.GetTransientStoreKey()

	appKeepers.ParamsKeeper = initParamsKeeper(
		appCodec,
		cdc,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	govModAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// set the BaseApp's parameter store
	appKeepers.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, keys[consensusparamtypes.StoreKey], govModAddress)
	bApp.SetParamStore(&appKeepers.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	appKeepers.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		appKeepers.keys[capabilitytypes.StoreKey],
		appKeepers.memKeys[capabilitytypes.MemStoreKey],
	)

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := appKeepers.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedICAControllerKeeper := appKeepers.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedICAHostKeeper := appKeepers.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedICQKeeper := appKeepers.CapabilityKeeper.ScopeToModule(icqtypes.ModuleName)
	scopedTransferKeeper := appKeepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedWasmKeeper := appKeepers.CapabilityKeeper.ScopeToModule(wasm.ModuleName)

	// add keepers
	Bech32Prefix := "juno"
	appKeepers.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		Bech32Prefix,
		govModAddress,
	)

	appKeepers.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		appKeepers.AccountKeeper,
		BlockedAddresses(),
		govModAddress,
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[stakingtypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		govModAddress,
	)
	appKeepers.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[minttypes.StoreKey],
		appKeepers.GetSubspace(minttypes.ModuleName),
		stakingKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		authtypes.FeeCollectorName,
	)
	appKeepers.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[distrtypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		govModAddress,
	)
	appKeepers.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		cdc,
		appKeepers.keys[slashingtypes.StoreKey],
		stakingKeeper,
		govModAddress,
	)

	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))
	appKeepers.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		keys[crisistypes.StoreKey],
		invCheckPeriod,
		appKeepers.BankKeeper,
		authtypes.FeeCollectorName,
		govModAddress,
	)

	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	// set the governance module account as the authority for conducting upgrades
	appKeepers.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		appKeepers.keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		bApp,
		govModAddress,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(appKeepers.DistrKeeper.Hooks(),
			appKeepers.SlashingKeeper.Hooks()),
	)
	appKeepers.StakingKeeper = stakingKeeper

	// ... other modules keepers

	// Create IBC Keeper
	appKeepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibcexported.StoreKey],
		appKeepers.GetSubspace(ibcexported.ModuleName),
		appKeepers.StakingKeeper,
		appKeepers.UpgradeKeeper,
		scopedIBCKeeper,
	)

	appKeepers.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[feegrant.StoreKey],
		appKeepers.AccountKeeper,
	)
	appKeepers.AuthzKeeper = authzkeeper.NewKeeper(
		appKeepers.keys[authzkeeper.StoreKey],
		appCodec,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
	)

	// register the proposal types
	govRouter := govv1beta.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(appKeepers.ParamsKeeper)). // This should be removed. It is still in place to avoid failures of modules that have not yet been upgraded
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(appKeepers.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(appKeepers.IBCKeeper.ClientKeeper))

	govKeeper := govkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[govtypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		bApp.MsgServiceRouter(),
		govtypes.DefaultConfig(),
		govModAddress,
	)
	appKeepers.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register governance hooks
		),
	)

	appKeepers.NFTKeeper = nftkeeper.NewKeeper(keys[nftkeeper.StoreKey], appCodec, appKeepers.AccountKeeper, appKeepers.BankKeeper)

	// Configure the hooks keeper
	hooksKeeper := ibchookskeeper.NewKeeper(
		appKeepers.keys[ibchookstypes.StoreKey],
	)
	appKeepers.IBCHooksKeeper = &hooksKeeper

	junoPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	wasmHooks := ibchooks.NewWasmHooks(appKeepers.IBCHooksKeeper, nil, junoPrefix) // The contract keeper needs to be set later
	appKeepers.Ics20WasmHooks = &wasmHooks
	appKeepers.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.Ics20WasmHooks,
	)

	// Do not use this middleware for anything except x/wasm requirement.
	// The spec currently requires new channels to be created, to use it.
	// We need to wait for Channel Upgradability before we can use this for any other middleware.
	appKeepers.IBCFeeKeeper = ibcfeekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibcfeetypes.StoreKey],
		appKeepers.HooksICS4Wrapper, // replaced with IBC middleware
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
	)

	// Initialize packet forward middleware router
	appKeepers.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec, appKeepers.keys[packetforwardtypes.StoreKey],
		appKeepers.GetSubspace(packetforwardtypes.ModuleName),
		appKeepers.TransferKeeper, // Will be zero-value here. Reference is set later on with SetTransferKeeper.
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.DistrKeeper,
		appKeepers.BankKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
	)

	// Create Transfer Keepers
	appKeepers.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibctransfertypes.StoreKey],
		appKeepers.GetSubspace(ibctransfertypes.ModuleName),
		// The ICS4Wrapper is replaced by the PacketForwardKeeper instead of the channel so that sending can be overridden by the middleware
		appKeepers.PacketForwardKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		scopedTransferKeeper,
	)

	appKeepers.PacketForwardKeeper.SetTransferKeeper(appKeepers.TransferKeeper)

	// ICQ Keeper
	appKeepers.ICQKeeper = icqkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[icqtypes.StoreKey],
		appKeepers.GetSubspace(icqtypes.ModuleName),
		appKeepers.IBCKeeper.ChannelKeeper, // may be replaced with middleware
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		scopedICQKeeper,
		NewQuerierWrapper(bApp),
	)

	appKeepers.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[icahosttypes.StoreKey],
		appKeepers.GetSubspace(icahosttypes.SubModuleName),
		appKeepers.HooksICS4Wrapper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		scopedICAHostKeeper,
		bApp.MsgServiceRouter(),
	)

	// ICA Controller keeper
	appKeepers.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec, appKeepers.keys[icacontrollertypes.StoreKey], appKeepers.GetSubspace(icacontrollertypes.SubModuleName),
		appKeepers.IBCFeeKeeper, // use ics29 fee as ics4Wrapper in middleware stack
		appKeepers.IBCKeeper.ChannelKeeper, &appKeepers.IBCKeeper.PortKeeper,
		scopedICAControllerKeeper, bApp.MsgServiceRouter(),
	)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, appKeepers.keys[evidencetypes.StoreKey], appKeepers.StakingKeeper, appKeepers.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	appKeepers.EvidenceKeeper = *evidenceKeeper

	// Create the TokenFactory Keeper
	appKeepers.TokenFactoryKeeper = tokenfactorykeeper.NewKeeper(
		appKeepers.keys[tokenfactorytypes.StoreKey],
		appKeepers.GetSubspace(tokenfactorytypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper,
		tokenFactoryCapabilities,
	)

	wasmDir := filepath.Join(homePath, "data")

	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic("error while reading wasm config: " + err.Error())
	}

	// Move custom query of token factory to stargate, still use custom msg which is tfOpts[1]
	tfOpts := bindings.RegisterCustomPlugins(&appKeepers.BankKeeper, &appKeepers.TokenFactoryKeeper)
	wasmOpts = append(wasmOpts, tfOpts...)

	// Stargate Queries
	// (Define custom modules here, then the next function takes care of the standard SDK types)
	accepted := wasmkeeper.AcceptedStargateQueries{
		// wasm
		"/cosmwasm.wasm.v1/Query/ContractInfo":       &wasmtypes.QueryContractInfoResponse{},
		"/cosmwasm.wasm.v1/Query/ContractHistory":    &wasmtypes.QueryContractHistoryResponse{},
		"/cosmwasm.wasm.v1/Query/ContractsByCode":    &wasmtypes.QueryContractsByCodeResponse{},
		"/cosmwasm.wasm.v1/Query/AllContractState":   &wasmtypes.QueryAllContractStateResponse{},
		"/cosmwasm.wasm.v1/Query/RawContractState":   &wasmtypes.QueryRawContractStateResponse{},
		"/cosmwasm.wasm.v1/Query/SmartContractState": &wasmtypes.QuerySmartContractStateResponse{},
		"/cosmwasm.wasm.v1/Query/Code":               &wasmtypes.QueryCodeResponse{},
		"/cosmwasm.wasm.v1/Query/Codes":              &wasmtypes.QueryCodesResponse{},
		"/cosmwasm.wasm.v1/Query/PinnedCodes":        &wasmtypes.QueryPinnedCodesResponse{},
		"/cosmwasm.wasm.v1/Query/Params":             &wasmtypes.QueryParamsResponse{},
		"/cosmwasm.wasm.v1/Query/ContractsByCreator": &wasmtypes.QueryContractsByCreatorResponse{},

		// ibc
		"/ibc.core.client.v1.Query/ClientState":    &ibcclienttypes.QueryClientStateResponse{},
		"/ibc.core.client.v1.Query/ConsensusState": &ibcclienttypes.QueryConsensusStateResponse{},
		"/ibc.core.connection.v1.Query/Connection": &ibcconnectiontypes.QueryConnectionResponse{},

		// token factory
		"/osmosis.tokenfactory.v1beta1.Query/Params":                 &tokenfactorytypes.QueryParamsResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata": &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator":      &tokenfactorytypes.QueryDenomsFromCreatorResponse{},

		// feeshare
		"/juno.feeshare.v1.Query/FeeShares":           &feesharetypes.QueryFeeSharesResponse{},
		"/juno.feeshare.v1.Query/FeeShare":            &feesharetypes.QueryFeeShareResponse{},
		"/juno.feeshare.v1.Query/Params":              &feesharetypes.QueryParamsResponse{},
		"/juno.feeshare.v1.Query/DeployerFeeShares":   &feesharetypes.QueryDeployerFeeSharesResponse{},
		"/juno.feeshare.v1.Query/WithdrawerFeeShares": &feesharetypes.QueryWithdrawerFeeSharesResponse{},

		// globalfee
		"/gaia.globalfee.v1beta1.Query/MinimumGasPrices": &globalfeetypes.QueryMinimumGasPricesResponse{},
	}
	accepted = addCosmosSDKStdStargateQueries(accepted)

	querierOpts := wasmkeeper.WithQueryPlugins(
		&wasmkeeper.QueryPlugins{
			Stargate: wasmkeeper.AcceptListStargateQuerier(accepted, bApp.GRPCQueryRouter(), appCodec),
		})
	wasmOpts = append(wasmOpts, querierOpts)

	appKeepers.WasmKeeper = wasm.NewKeeper(
		appCodec,
		appKeepers.keys[wasm.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		distrkeeper.NewQuerier(appKeepers.DistrKeeper),
		appKeepers.IBCFeeKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		scopedWasmKeeper,
		appKeepers.TransferKeeper,
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		wasmCapabilities,
		govModAddress,
		wasmOpts...,
	)

	appKeepers.FeeShareKeeper = feesharekeeper.NewKeeper(
		appKeepers.keys[feesharetypes.StoreKey],
		appCodec,
		appKeepers.BankKeeper,
		appKeepers.WasmKeeper,
		appKeepers.AccountKeeper,
		authtypes.FeeCollectorName,
		govModAddress,
	)

	// register wasm gov proposal types
	// The gov proposal types can be individually enabled
	if len(enabledProposals) != 0 {
		govRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(appKeepers.WasmKeeper, enabledProposals))
	}
	// Set legacy router for backwards compatibility with gov v1beta1
	appKeepers.GovKeeper.SetLegacyRouter(govRouter)

	// Create Transfer Stack
	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(appKeepers.TransferKeeper)
	transferStack = ibcfee.NewIBCMiddleware(transferStack, appKeepers.IBCFeeKeeper)
	transferStack = ibchooks.NewIBCMiddleware(transferStack, &appKeepers.HooksICS4Wrapper)
	transferStack = packetforward.NewIBCMiddleware(
		transferStack,
		appKeepers.PacketForwardKeeper,
		0,
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
		packetforwardkeeper.DefaultRefundTransferPacketTimeoutTimestamp,
	)

	// initialize ICA module with mock module as the authentication module on the controller side
	var icaControllerStack porttypes.IBCModule
	icaControllerStack = icacontroller.NewIBCMiddleware(icaControllerStack, appKeepers.ICAControllerKeeper)
	icaControllerStack = ibcfee.NewIBCMiddleware(icaControllerStack, appKeepers.IBCFeeKeeper)

	// RecvPacket, message that originates from core IBC and goes down to app, the flow is:
	// channel.RecvPacket -> fee.OnRecvPacket -> icaHost.OnRecvPacket
	var icaHostStack porttypes.IBCModule
	icaHostStack = icahost.NewIBCModule(appKeepers.ICAHostKeeper)
	icaHostStack = ibcfee.NewIBCMiddleware(icaHostStack, appKeepers.IBCFeeKeeper)

	// Create fee enabled wasm ibc Stack
	var wasmStack porttypes.IBCModule
	wasmStack = wasm.NewIBCHandler(appKeepers.WasmKeeper, appKeepers.IBCKeeper.ChannelKeeper, appKeepers.IBCFeeKeeper)
	wasmStack = ibcfee.NewIBCMiddleware(wasmStack, appKeepers.IBCFeeKeeper)

	// create ICQ module
	icqModule := icq.NewIBCModule(appKeepers.ICQKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter().
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		AddRoute(wasm.ModuleName, wasmStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack).
		AddRoute(icqtypes.ModuleName, icqModule)
	appKeepers.IBCKeeper.SetRouter(ibcRouter)

	appKeepers.ScopedIBCKeeper = scopedIBCKeeper
	appKeepers.ScopedTransferKeeper = scopedTransferKeeper
	appKeepers.ScopedICQKeeper = scopedICQKeeper
	appKeepers.scopedWasmKeeper = scopedWasmKeeper
	appKeepers.ScopedICAHostKeeper = scopedICAHostKeeper
	appKeepers.ScopedICAControllerKeeper = scopedICAControllerKeeper

	// set the contract keeper for the Ics20WasmHooks
	appKeepers.ContractKeeper = &appKeepers.WasmKeeper
	appKeepers.Ics20WasmHooks.ContractKeeper = appKeepers.ContractKeeper

	return appKeepers
}

func addCosmosSDKStdStargateQueries(asq wasmkeeper.AcceptedStargateQueries) wasmkeeper.AcceptedStargateQueries {
	std := map[string]codec.ProtoMarshaler{
		// auth
		"/cosmos.auth.v1beta1.Query/Account":              &authtype.QueryAccountResponse{},
		"/cosmos.auth.v1beta1.Query/AccountAddressByID":   &authtype.QueryAccountAddressByIDResponse{},
		"/cosmos.auth.v1beta1.Query/Params":               &authtype.QueryParamsResponse{},
		"/cosmos.auth.v1beta1.Query/ModuleAccounts":       &authtype.QueryModuleAccountsResponse{},
		"/cosmos.auth.v1beta1.Query/ModuleAccountByName":  &authtype.QueryModuleAccountByNameResponse{},
		"/cosmos.auth.v1beta1.Query/Bech32Prefix":         &authtype.Bech32PrefixResponse{},
		"/cosmos.auth.v1beta1.Query/AddressBytesToString": &authtype.AddressBytesToStringResponse{},
		"/cosmos.auth.v1beta1.Query/AddressStringToBytes": &authtype.AddressStringToBytesResponse{},
		"/cosmos.auth.v1beta1.Query/AccountInfo":          &authtype.QueryAccountInfoResponse{},

		// authz
		"/cosmos.authz.v1beta1.Query/Grants":        &authz.QueryGrantsResponse{},
		"/cosmos.authz.v1beta1.Query/GranterGrants": &authz.QueryGranterGrantsResponse{},

		// bank
		"/cosmos.bank.v1beta1.Query/Balance":                 &banktypes.QueryBalanceResponse{},
		"/cosmos.bank.v1beta1.Query/AllBalances":             &banktypes.QueryAllBalancesResponse{},
		"/cosmos.bank.v1beta1.Query/SpendableBalances":       &banktypes.QuerySpendableBalancesResponse{},
		"/cosmos.bank.v1beta1.Query/SpendableBalanceByDenom": &banktypes.QuerySpendableBalanceByDenomResponse{},
		"/cosmos.bank.v1beta1.Query/TotalSupply":             &banktypes.QueryTotalSupplyResponse{},
		"/cosmos.bank.v1beta1.Query/SupplyOf":                &banktypes.QuerySupplyOfResponse{},
		"/cosmos.bank.v1beta1.Query/Params":                  &banktypes.QueryParamsResponse{},
		"/cosmos.bank.v1beta1.Query/DenomMetadata":           &banktypes.QueryDenomMetadataResponse{},
		"/cosmos.bank.v1beta1.Query/DenomsMetadata":          &banktypes.QueryDenomsMetadataResponse{},
		"/cosmos.bank.v1beta1.Query/DenomOwners":             &banktypes.QueryDenomOwnersResponse{},
		"/cosmos.bank.v1beta1.Query/SendEnabled":             &banktypes.QuerySendEnabledResponse{},

		// consensus
		"/cosmos.consensus.v1beta1.Query/Params": &consensustypes.QueryParamsResponse{},

		// distribution
		"/cosmos.distribution.v1beta1.Query/ParamsRequest":               &distrtypes.QueryParamsRequest{},
		"/cosmos.distribution.v1beta1.Query/ValidatorDistributionInfo":   &distrtypes.QueryValidatorDistributionInfoResponse{},
		"/cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards": &distrtypes.QueryValidatorOutstandingRewardsResponse{},
		"/cosmos.distribution.v1beta1.Query/ValidatorCommission":         &distrtypes.QueryValidatorCommissionResponse{},
		"/cosmos.distribution.v1beta1.Query/ValidatorSlashes":            &distrtypes.QueryValidatorSlashesResponse{},
		"/cosmos.distribution.v1beta1.Query/DelegationRewards":           &distrtypes.QueryDelegationRewardsResponse{},
		"/cosmos.distribution.v1beta1.Query/DelegationTotalRewards":      &distrtypes.QueryDelegationTotalRewardsResponse{},
		"/cosmos.distribution.v1beta1.Query/DelegatorValidators":         &distrtypes.QueryDelegatorValidatorsResponse{},
		"/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress":    &distrtypes.QueryDelegatorWithdrawAddressResponse{},
		"/cosmos.distribution.v1beta1.Query/CommunityPool":               &distrtypes.QueryCommunityPoolResponse{},

		// evidence
		"/cosmos.evidence.v1beta1.Query/Evidence":    &evedencetypes.QueryEvidenceResponse{},
		"/cosmos.evidence.v1beta1.Query/AllEvidence": &evedencetypes.QueryAllEvidenceResponse{},

		// feegrant
		"/cosmos.feegrant.v1beta1.Query/Allowance":           &feegrant.QueryAllowanceResponse{},
		"/cosmos.feegrant.v1beta1.Query/Allowances":          &feegrant.QueryAllowancesResponse{},
		"/cosmos.feegrant.v1beta1.Query/AllowancesByGranter": &feegrant.QueryAllowancesByGranterResponse{},

		// governance
		"/cosmos.gov.v1beta1.Query/Proposal":    &govv1.QueryProposalResponse{},
		"/cosmos.gov.v1beta1.Query/Proposals":   &govv1.QueryProposalsResponse{},
		"/cosmos.gov.v1beta1.Query/Vote":        &govv1.QueryVoteResponse{},
		"/cosmos.gov.v1beta1.Query/Votes":       &govv1.QueryVotesResponse{},
		"/cosmos.gov.v1beta1.Query/Params":      &govv1.QueryParamsResponse{},
		"/cosmos.gov.v1beta1.Query/Deposit":     &govv1.QueryDepositResponse{},
		"/cosmos.gov.v1beta1.Query/TallyResult": &govv1.QueryTallyResultResponse{},

		// mint
		"/juno.mint.Query/Params":           &minttypes.QueryParamsResponse{},
		"/juno.mint.Query/Inflation":        &minttypes.QueryInflationResponse{},
		"/juno.mint.Query/AnnualProvisions": &minttypes.QueryAnnualProvisionsResponse{},

		// nft
		"/cosmos.nft.v1beta1.Query/Balance": &nft.QueryBalanceResponse{},
		"/cosmos.nft.v1beta1.Query/Owner":   &nft.QueryOwnerResponse{},
		"/cosmos.nft.v1beta1.Query/Supply":  &nft.QuerySupplyResponse{},
		"/cosmos.nft.v1beta1.Query/NFTs":    &nft.QueryNFTsResponse{},
		"/cosmos.nft.v1beta1.Query/NFT":     &nft.QueryNFTResponse{},
		"/cosmos.nft.v1beta1.Query/Class":   &nft.QueryClassResponse{},
		"/cosmos.nft.v1beta1.Query/Classes": &nft.QueryClassesResponse{},

		// slashing
		"/cosmos.slashing.v1beta1.Query/Params":       &slashingtypes.QueryParamsResponse{},
		"/cosmos.slashing.v1beta1.Query/SigningInfo":  &slashingtypes.QuerySigningInfoResponse{},
		"/cosmos.slashing.v1beta1.Query/SigningInfos": &slashingtypes.QuerySigningInfosResponse{},

		// staking
		"/cosmos.staking.v1beta1.Query/Validators":                    &stakingtypes.QueryValidatorsResponse{},
		"/cosmos.staking.v1beta1.Query/Validator":                     &stakingtypes.QueryValidatorResponse{},
		"/cosmos.staking.v1beta1.Query/ValidatorDelegations":          &stakingtypes.QueryValidatorDelegationsResponse{},
		"/cosmos.staking.v1beta1.Query/ValidatorUnbondingDelegations": &stakingtypes.QueryValidatorUnbondingDelegationsResponse{},
		"/cosmos.staking.v1beta1.Query/Delegation":                    &stakingtypes.QueryDelegationResponse{},
		"/cosmos.staking.v1beta1.Query/UnbondingDelegation":           &stakingtypes.QueryUnbondingDelegationResponse{},
		"/cosmos.staking.v1beta1.Query/DelegatorDelegations":          &stakingtypes.QueryDelegatorDelegationsResponse{},
		"/cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations": &stakingtypes.QueryDelegatorUnbondingDelegationsResponse{},
		"/cosmos.staking.v1beta1.Query/Redelegations":                 &stakingtypes.QueryRedelegationsResponse{},
		"/cosmos.staking.v1beta1.Query/DelegatorValidators":           &stakingtypes.QueryDelegatorValidatorsResponse{},
		"/cosmos.staking.v1beta1.Query/DelegatorValidator":            &stakingtypes.QueryDelegatorValidatorResponse{},
		"/cosmos.staking.v1beta1.Query/HistoricalInfo":                &stakingtypes.QueryHistoricalInfoResponse{},
		"/cosmos.staking.v1beta1.Query/Params":                        &stakingtypes.QueryParamsResponse{},
		"/cosmos.staking.v1beta1.Query/Pool":                          &stakingtypes.QueryPoolResponse{},

		// upgrade
		"/cosmos.upgrade.v1beta1.Query/CurrentPlan":                 &upgradetypes.QueryCurrentPlanResponse{},
		"/cosmos.upgrade.v1beta1.Query/AppliedPlan":                 &upgradetypes.QueryAppliedPlanResponse{},
		"/cosmos.upgrade.v1beta1.Query/UpgradedConsensusState":      &upgradetypes.QueryUpgradedConsensusStateResponse{},
		"/cosmos.upgrade.v1beta1.Query/QueryModuleVersionsResponse": &upgradetypes.QueryModuleVersionsResponse{},
		"/cosmos.upgrade.v1beta1.Query/QueryAuthorityResponse":      &upgradetypes.QueryAuthorityResponse{},
	}

	updated := asq
	for k, v := range std {
		updated[k] = v
	}
	return updated
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	// https://github.com/cosmos/ibc-go/issues/2010
	// Will remove all of these in the future. For now we keep for legacy proposals to work properly.
	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)

	paramsKeeper.Subspace(stakingtypes.ModuleName).WithKeyTable(stakingtypes.ParamKeyTable()) // Used for GlobalFee
	paramsKeeper.Subspace(minttypes.ModuleName)

	// custom
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(icqtypes.ModuleName)
	paramsKeeper.Subspace(packetforwardtypes.ModuleName)
	paramsKeeper.Subspace(globalfee.ModuleName)
	paramsKeeper.Subspace(tokenfactorytypes.ModuleName)
	paramsKeeper.Subspace(feesharetypes.ModuleName)
	paramsKeeper.Subspace(wasm.ModuleName)

	return paramsKeeper
}

// GetSubspace returns a param subspace for a given module name.
func (appKeepers *AppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := appKeepers.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// GetStakingKeeper implements the TestingApp interface.
func (appKeepers *AppKeepers) GetStakingKeeper() *stakingkeeper.Keeper {
	return appKeepers.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (appKeepers *AppKeepers) GetIBCKeeper() *ibckeeper.Keeper {
	return appKeepers.IBCKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (appKeepers *AppKeepers) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return appKeepers.ScopedIBCKeeper
}

// GetWasmKeeper implements the TestingApp interface.
func (appKeepers *AppKeepers) GetWasmKeeper() wasm.Keeper {
	return appKeepers.WasmKeeper
}

// BlockedAddresses returns all the app's blocked account addresses.
func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range GetMaccPerms() {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}

	return dupMaccPerms
}
