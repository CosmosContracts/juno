package keepers

import (
	"path/filepath"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	builderkeeper "github.com/skip-mev/pob/x/builder/keeper"
	buildertypes "github.com/skip-mev/pob/x/builder/types"
	"github.com/spf13/cast"

	packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/types"
	icq "github.com/cosmos/ibc-apps/modules/async-icq/v7"
	icqkeeper "github.com/cosmos/ibc-apps/modules/async-icq/v7/keeper"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"
	ibc_hooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7"
	ibchookskeeper "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/keeper"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/types"
	icacontroller "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
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

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
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

	feesharekeeper "github.com/CosmosContracts/juno/v16/x/feeshare/keeper"
	feesharetypes "github.com/CosmosContracts/juno/v16/x/feeshare/types"
	"github.com/CosmosContracts/juno/v16/x/globalfee"
	globalfeekeeper "github.com/CosmosContracts/juno/v16/x/globalfee/keeper"
	globalfeetypes "github.com/CosmosContracts/juno/v16/x/globalfee/types"
	mintkeeper "github.com/CosmosContracts/juno/v16/x/mint/keeper"
	minttypes "github.com/CosmosContracts/juno/v16/x/mint/types"
	"github.com/CosmosContracts/juno/v16/x/tokenfactory/bindings"
	tokenfactorykeeper "github.com/CosmosContracts/juno/v16/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/CosmosContracts/juno/v16/x/tokenfactory/types"
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
	globalfee.ModuleName:           nil,
	buildertypes.ModuleName:        nil,
}

type AppKeepers struct {
	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper       authkeeper.AccountKeeper
	BankKeeper          bankkeeper.BaseKeeper
	BuildKeeper         builderkeeper.Keeper
	CapabilityKeeper    *capabilitykeeper.Keeper
	StakingKeeper       *stakingkeeper.Keeper
	SlashingKeeper      slashingkeeper.Keeper
	MintKeeper          mintkeeper.Keeper
	DistrKeeper         distrkeeper.Keeper
	GovKeeper           govkeeper.Keeper
	CrisisKeeper        *crisiskeeper.Keeper
	UpgradeKeeper       *upgradekeeper.Keeper
	ParamsKeeper        paramskeeper.Keeper
	IBCKeeper           *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	ICQKeeper           icqkeeper.Keeper
	IBCFeeKeeper        ibcfeekeeper.Keeper
	IBCHooksKeeper      *ibchookskeeper.Keeper
	PacketForwardKeeper *packetforwardkeeper.Keeper
	EvidenceKeeper      evidencekeeper.Keeper
	TransferKeeper      ibctransferkeeper.Keeper
	AuthzKeeper         authzkeeper.Keeper
	FeeGrantKeeper      feegrantkeeper.Keeper
	NFTKeeper           nftkeeper.Keeper
	FeeShareKeeper      feesharekeeper.Keeper
	GlobalFeeKeeper     globalfeekeeper.Keeper
	ContractKeeper      *wasmkeeper.PermissionedKeeper

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
	Ics20WasmHooks   *ibc_hooks.WasmHooks
	HooksICS4Wrapper ibc_hooks.ICS4Middleware
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
		stakingKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		authtypes.FeeCollectorName,
		govModAddress,
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
	wasmHooks := ibc_hooks.NewWasmHooks(appKeepers.IBCHooksKeeper, &appKeepers.WasmKeeper, junoPrefix) // The contract keeper needs to be set later
	appKeepers.Ics20WasmHooks = &wasmHooks
	appKeepers.HooksICS4Wrapper = ibc_hooks.NewICS4Middleware(
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
		appCodec,
		appKeepers.keys[tokenfactorytypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper,
		tokenFactoryCapabilities,
		govModAddress,
	)

	// Create the Skip Builder Keeper
	appKeepers.BuildKeeper = builderkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[buildertypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper,
		appKeepers.StakingKeeper,
		govModAddress,
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
	accepted := wasmkeeper.AcceptedStargateQueries{
		// ibc
		"/ibc.core.client.v1.Query/ClientState":    &ibcclienttypes.QueryClientStateResponse{},
		"/ibc.core.client.v1.Query/ConsensusState": &ibcclienttypes.QueryConsensusStateResponse{},
		"/ibc.core.connection.v1.Query/Connection": &ibcconnectiontypes.QueryConnectionResponse{},

		// governance
		"/cosmos.gov.v1beta1.Query/Vote": &govv1.QueryVoteResponse{},

		// distribution
		"/cosmos.distribution.v1beta1.Query/DelegationRewards": &distrtypes.QueryDelegationRewardsResponse{},

		// staking
		"/cosmos.staking.v1beta1.Query/Delegation":          &stakingtypes.QueryDelegationResponse{},
		"/cosmos.staking.v1beta1.Query/Redelegations":       &stakingtypes.QueryRedelegationsResponse{},
		"/cosmos.staking.v1beta1.Query/UnbondingDelegation": &stakingtypes.QueryUnbondingDelegationResponse{},
		"/cosmos.staking.v1beta1.Query/Validator":           &stakingtypes.QueryValidatorResponse{},
		"/cosmos.staking.v1beta1.Query/Params":              &stakingtypes.QueryParamsResponse{},
		"/cosmos.staking.v1beta1.Query/Pool":                &stakingtypes.QueryPoolResponse{},

		// token factory
		"/osmosis.tokenfactory.v1beta1.Query/Params":                 &tokenfactorytypes.QueryParamsResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata": &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator":      &tokenfactorytypes.QueryDenomsFromCreatorResponse{},
	}

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

	appKeepers.GlobalFeeKeeper = globalfeekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[globalfeetypes.StoreKey],
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
	transferStack = ibc_hooks.NewIBCMiddleware(transferStack, &appKeepers.HooksICS4Wrapper)
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
	appKeepers.ContractKeeper = wasmkeeper.NewDefaultPermissionKeeper(appKeepers.WasmKeeper)
	appKeepers.Ics20WasmHooks.ContractKeeper = &appKeepers.WasmKeeper

	return appKeepers
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
	paramsKeeper.Subspace(buildertypes.ModuleName)

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
