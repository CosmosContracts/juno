package keepers

import (
	"fmt"
	"math"
	"path"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvm "github.com/CosmWasm/wasmvm/v3"
	"github.com/spf13/cast"

	packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/types"
	ibchooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v10"
	ibchookskeeper "github.com/cosmos/ibc-apps/modules/ibc-hooks/v10/keeper"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v10/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icacontroller "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	transfer "github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	storetypes "cosmossdk.io/store/types"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v30/wasmbindings"
	junoburn "github.com/CosmosContracts/juno/v30/x/burn"
	clockkeeper "github.com/CosmosContracts/juno/v30/x/clock/keeper"
	clocktypes "github.com/CosmosContracts/juno/v30/x/clock/types"
	cwhookskeeper "github.com/CosmosContracts/juno/v30/x/cw-hooks/keeper"
	cwhookstypes "github.com/CosmosContracts/juno/v30/x/cw-hooks/types"
	dripkeeper "github.com/CosmosContracts/juno/v30/x/drip/keeper"
	driptypes "github.com/CosmosContracts/juno/v30/x/drip/types"
	feemarketkeeper "github.com/CosmosContracts/juno/v30/x/feemarket/keeper"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	feepaykeeper "github.com/CosmosContracts/juno/v30/x/feepay/keeper"
	feepaytypes "github.com/CosmosContracts/juno/v30/x/feepay/types"
	feesharekeeper "github.com/CosmosContracts/juno/v30/x/feeshare/keeper"
	feesharetypes "github.com/CosmosContracts/juno/v30/x/feeshare/types"
	mintkeeper "github.com/CosmosContracts/juno/v30/x/mint/keeper"
	minttypes "github.com/CosmosContracts/juno/v30/x/mint/types"
	streamkeeper "github.com/CosmosContracts/juno/v30/x/stream/keeper"
	tokenfactorykeeper "github.com/CosmosContracts/juno/v30/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/CosmosContracts/juno/v30/x/tokenfactory/types"
	votingsnapshotkeeper "github.com/CosmosContracts/juno/v30/x/voting-snapshot/keeper"
	votingsnapshottypes "github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
	// wrappers
	wrappedgovkeeper "github.com/CosmosContracts/juno/v30/x/wrappers/gov/keeper"
)

var (
	wasmCapabilities = (append(wasmkeeper.BuiltInCapabilities(), "token_factory"))

	tokenFactoryCapabilities = []string{
		tokenfactorytypes.EnableBurnFrom,
		tokenfactorytypes.EnableForceTransfer,
		tokenfactorytypes.EnableSetMetadata,
	}
)

// module account permissions
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:      nil,
	distrtypes.ModuleName:           nil,
	minttypes.ModuleName:            {authtypes.Minter},
	stakingtypes.BondedPoolName:     {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:  {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:             {authtypes.Burner},
	ibctransfertypes.ModuleName:     {authtypes.Minter, authtypes.Burner},
	icatypes.ModuleName:             nil,
	wasmtypes.ModuleName:            {},
	tokenfactorytypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
	feepaytypes.ModuleName:          nil,
	feemarkettypes.FeeCollectorName: nil,
	junoburn.ModuleName:             {authtypes.Burner},
}

type AppKeepers struct {
	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey

	// keepers
	AccountKeeper        authkeeper.AccountKeeper
	BankKeeper           bankkeeper.Keeper
	CapabilityKeeper     *capabilitykeeper.Keeper
	StakingKeeper        *stakingkeeper.Keeper
	SlashingKeeper       slashingkeeper.Keeper
	MintKeeper           mintkeeper.Keeper
	DistrKeeper          distrkeeper.Keeper
	GovKeeper            *wrappedgovkeeper.KeeperWrapper // x/wrappers/gov wrapper to modify the gov module without forking it
	UpgradeKeeper        *upgradekeeper.Keeper
	IBCKeeper            *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	TmLightClientModule  ibctm.LightClientModule
	IBCHooksKeeper       *ibchookskeeper.Keeper
	PacketForwardKeeper  *packetforwardkeeper.Keeper
	EvidenceKeeper       evidencekeeper.Keeper
	TransferKeeper       ibctransferkeeper.Keeper
	AuthzKeeper          authzkeeper.Keeper
	FeeGrantKeeper       feegrantkeeper.Keeper
	FeePayKeeper         feepaykeeper.Keeper
	FeeShareKeeper       feesharekeeper.Keeper
	ContractKeeper       wasmtypes.ContractOpsKeeper
	ClockKeeper          clockkeeper.Keeper
	CWHooksKeeper        cwhookskeeper.Keeper
	VotingSnapshotKeeper votingsnapshotkeeper.Keeper

	ConsensusParamsKeeper consensusparamkeeper.Keeper

	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedFeeMockKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper

	WasmKeeper         wasmkeeper.Keeper
	scopedWasmKeeper   capabilitykeeper.ScopedKeeper
	TokenFactoryKeeper tokenfactorykeeper.Keeper

	DripKeeper   dripkeeper.Keeper
	StreamKeeper *streamkeeper.Keeper

	FeeMarketKeeper *feemarketkeeper.Keeper

	// Middleware wrapper
	Ics20WasmHooks   *ibchooks.WasmHooks
	HooksICS4Wrapper ibchooks.ICS4Middleware
}

func NewAppKeepers(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	cdc *codec.LegacyAmino,
	maccPerms map[string][]string,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmkeeper.Option,
	bondDenom string,
	homePath string,
) AppKeepers {
	appKeepers := AppKeepers{}

	// Set keys KVStoreKey, TransientStoreKey, MemoryStoreKey
	appKeepers.GenerateKeys()
	keys := appKeepers.GetKVStoreKeys()

	govModAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	bech32Prefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	ac := authcodec.NewBech32Codec(bech32Prefix)
	dataDir := path.Join(homePath, "data")
	wasmDir := path.Join(dataDir, "wasm")

	// set the BaseApp's parameter store
	appKeepers.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		govModAddress,
		runtime.EventService{},
	)
	bApp.SetParamStore(&appKeepers.ConsensusParamsKeeper.ParamsStore)

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
	scopedTransferKeeper := appKeepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedWasmKeeper := appKeepers.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)

	// add keepers
	appKeepers.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		ac,
		bech32Prefix,
		govModAddress,
	)

	appKeepers.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		appKeepers.AccountKeeper,
		BlockedAddresses(),
		govModAddress,
		bApp.Logger(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[stakingtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		govModAddress,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)

	appKeepers.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[minttypes.StoreKey]),
		stakingKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		authtypes.FeeCollectorName,
		govModAddress,
	)

	appKeepers.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[distrtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		govModAddress,
	)

	appKeepers.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		cdc,
		runtime.NewKVStoreService(appKeepers.keys[slashingtypes.StoreKey]),
		stakingKeeper,
		govModAddress,
	)

	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	appKeepers.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(appKeepers.keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		bApp,
		govModAddress,
	)

	// Create IBC Keeper (ibc-go v10: drops StakingKeeper + scoped capability keeper)
	appKeepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[ibcexported.StoreKey]),
		nil,
		appKeepers.UpgradeKeeper,
		govModAddress,
	)

	// Tendermint light client module — ibc-go v10 requires explicit registration
	clientKeeper := appKeepers.IBCKeeper.ClientKeeper
	storeProvider := clientKeeper.GetStoreProvider()
	appKeepers.TmLightClientModule = ibctm.NewLightClientModule(appCodec, storeProvider)
	clientKeeper.AddRoute(ibctm.ModuleName, &appKeepers.TmLightClientModule)

	appKeepers.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[feegrant.StoreKey]),
		appKeepers.AccountKeeper,
	)

	appKeepers.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(appKeepers.keys[authzkeeper.StoreKey]),
		appCodec,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
	)

	// register the proposal types
	govRouter := govv1beta.NewRouter()
	// This should be removed. It is still in place to avoid failures of modules that have not yet been upgraded
	govRouter.AddRoute(govtypes.RouterKey, govv1beta.ProposalHandler)
	// Update the max metadata length to be >255
	govConfig := govtypes.DefaultConfig()
	govConfig.MaxMetadataLen = math.MaxUint64

	govKeeper := wrappedgovkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[govtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		stakingKeeper,
		appKeepers.DistrKeeper,
		bApp.MsgServiceRouter(),
		govConfig,
		govModAddress,
	)

	// Configure the hooks keeper
	hooksKeeper := ibchookskeeper.NewKeeper(
		appKeepers.keys[ibchookstypes.StoreKey],
	)
	appKeepers.IBCHooksKeeper = &hooksKeeper

	wasmHooks := ibchooks.NewWasmHooks(
		appKeepers.IBCHooksKeeper,
		&appKeepers.WasmKeeper,
		bech32Prefix,
	)
	appKeepers.Ics20WasmHooks = &wasmHooks
	// The contract keeper needs to be set later
	appKeepers.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.Ics20WasmHooks,
	)

	// Initialize packet forward middleware router (PFM v10: KVStoreService)
	appKeepers.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[packetforwardtypes.StoreKey]),
		appKeepers.TransferKeeper, // Will be zero-value here. Reference is set later on with SetTransferKeeper.
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.BankKeeper,
		appKeepers.HooksICS4Wrapper,
		govModAddress,
	)

	// Create Transfer Keeper (ibc-go v10: drops PortKeeper + scoped capability keeper, adds MessageRouter)
	appKeepers.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[ibctransfertypes.StoreKey]),
		nil,
		// The ICS4Wrapper is replaced by the PacketForwardKeeper instead of the channel so that sending can be overridden by the middleware
		appKeepers.PacketForwardKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		govModAddress,
	)

	appKeepers.PacketForwardKeeper.SetTransferKeeper(appKeepers.TransferKeeper)

	// async-icq dropped from v30: ibc-apps maintainers haven't published a /v9/v10/v11
	// line, and the latest /v8 commit (2026-04-27) still pins ibc-go/v8 — incompatible
	// with our ibc-go/v10. Reintroduce in v30.x once ibc-apps publishes /v10.

	// ICA Host Keeper (ibc-go v10: QueryRouter now passed to NewKeeper, WithQueryRouter removed)
	appKeepers.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[icahosttypes.StoreKey]),
		nil,
		appKeepers.HooksICS4Wrapper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.AccountKeeper,
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		govModAddress,
	)

	// ICA Controller keeper (ibc-go v10: ICS-29 dropped, capability keepers no longer needed)
	appKeepers.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[icacontrollertypes.StoreKey]),
		nil,
		appKeepers.IBCKeeper.ChannelKeeper, // ics4Wrapper (was IBCFeeKeeper before 29-fee removal)
		appKeepers.IBCKeeper.ChannelKeeper,
		bApp.MsgServiceRouter(),
		govModAddress,
	)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[evidencetypes.StoreKey]),
		stakingKeeper,
		appKeepers.SlashingKeeper,
		ac,
		runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	appKeepers.EvidenceKeeper = *evidenceKeeper

	// Create the TokenFactory Keeper
	appKeepers.TokenFactoryKeeper = tokenfactorykeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[tokenfactorytypes.StoreKey]),
		maccPerms,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper,
		tokenFactoryCapabilities,
		govModAddress,
	)

	appKeepers.FeeMarketKeeper = feemarketkeeper.NewKeeper(
		appCodec,
		keys[feemarkettypes.StoreKey],
		appKeepers.AccountKeeper,
		// Fees are payable ONLY in the fee (bond) denom. ErrorDenomResolver
		// rejects everything else — a permissive resolver would let anyone
		// mint a tokenfactory denom and pay fees with it 1:1.
		&feemarkettypes.ErrorDenomResolver{},
		govModAddress,
	)

	// VotingSnapshotKeeper MUST be constructed before it is handed to the
	// wasm query plugin below: RegisterCustomPlugins captures the keeper BY
	// VALUE, so a zero-value keeper here would nil-deref (chain halt) the
	// first time a contract queries voting power. Its staking hooks are
	// registered further down alongside the other staking hooks.
	appKeepers.VotingSnapshotKeeper = votingsnapshotkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[votingsnapshottypes.StoreKey]),
		stakingKeeper,
		govModAddress,
		votingsnapshottypes.NewTransientKVStoreService(appKeepers.tkeys[votingsnapshottypes.TransientStoreKey]),
	)

	wasmConfig, err := wasm.ReadNodeConfig(appOpts)
	if err != nil {
		panic("error while reading wasm config: " + err.Error())
	}

	// C1 regression guard: never hand a zero-value keeper to the wasm query
	// plugin (it is captured by value — see comment on the construction above).
	if appKeepers.VotingSnapshotKeeper.Authority() == "" {
		panic("voting-snapshot keeper must be constructed before wasm plugin registration")
	}

	// Move custom query of token factory to stargate, still use custom msg which is tfOpts[1]
	tfOpts := wasmbindings.RegisterCustomPlugins(appKeepers.BankKeeper, &appKeepers.TokenFactoryKeeper, appKeepers.VotingSnapshotKeeper)
	wasmOpts = append(wasmOpts, tfOpts...)

	// Stargate Queries
	acceptedQueries := AcceptedQueries()
	querierOpts := wasmkeeper.WithQueryPlugins(
		&wasmkeeper.QueryPlugins{
			Stargate: wasmkeeper.AcceptListStargateQuerier(acceptedQueries, bApp.GRPCQueryRouter(), appCodec),
			Grpc:     wasmkeeper.AcceptListGrpcQuerier(acceptedQueries, bApp.GRPCQueryRouter(), appCodec),
		})
	wasmOpts = append(wasmOpts, querierOpts)

	junoBurnerPlugin := junoburn.NewBurnerPlugin(appKeepers.BankKeeper, appKeepers.MintKeeper)
	// ref: https://github.com/CosmWasm/wasmd/issues/1735
	burnMessageHandler := wasmkeeper.WithMessageHandlerDecorator(func(nested wasmkeeper.Messenger) wasmkeeper.Messenger {
		return wasmkeeper.NewMessageHandlerChain(
			wasmkeeper.NewBurnCoinMessageHandler(junoBurnerPlugin),
			nested,
		)
	})
	wasmOpts = append(wasmOpts, burnMessageHandler)

	wasmOpts = append(wasmOpts, wasmkeeper.WithGasRegister(NewJunoWasmGasRegister()))

	wasmer, err := wasmvm.NewVM(wasmDir, wasmCapabilities, 32, wasmConfig.ContractDebugMode, wasmConfig.MemoryCacheSize)
	if err != nil {
		panic(fmt.Sprintf("failed to create juno wasmvm: %s", err))
	}
	wasmOpts = append(wasmOpts, wasmkeeper.WithWasmEngine(wasmer))

	// wasmd v0.61: IBCFeeKeeper + PortKeeper + scoped capability keeper dropped;
	// adds ChannelKeeperV2 between ChannelKeeper and TransferKeeper
	appKeepers.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[wasmtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		stakingKeeper,
		distrkeeper.NewQuerier(appKeepers.DistrKeeper),
		appKeepers.IBCKeeper.ChannelKeeper, // ICS4Wrapper (was IBCFeeKeeper)
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.ChannelKeeperV2,
		appKeepers.TransferKeeper,
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		dataDir,
		wasmConfig,
		wasmtypes.VMConfig{},
		wasmCapabilities,
		govModAddress,
		wasmOpts...,
	)

	appKeepers.FeePayKeeper = feepaykeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[feepaytypes.StoreKey]),
		appKeepers.BankKeeper,
		appKeepers.WasmKeeper,
		appKeepers.AccountKeeper,
		appKeepers.FeeMarketKeeper,
		govModAddress,
	)
	appKeepers.FeeMarketKeeper.SetFeePayLiabilityChecker(appKeepers.FeePayKeeper)

	// set the contract keeper for the Ics20WasmHooks
	appKeepers.ContractKeeper = wasmkeeper.NewDefaultPermissionKeeper(appKeepers.WasmKeeper)
	appKeepers.Ics20WasmHooks.ContractKeeper = &appKeepers.WasmKeeper

	appKeepers.FeeShareKeeper = feesharekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[feesharetypes.StoreKey]),
		appKeepers.BankKeeper,
		appKeepers.WasmKeeper,
		appKeepers.AccountKeeper,
		govModAddress,
	)

	appKeepers.DripKeeper = dripkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[driptypes.StoreKey]),
		appKeepers.BankKeeper,
		authtypes.FeeCollectorName,
		govModAddress,
	)

	appKeepers.StreamKeeper = streamkeeper.NewKeeper(
		appCodec,
		homePath,
		bApp,
	)

	// if err := appKeepers.registerStreamModuleCodecs(appCodec); err != nil {
	// 	panic(fmt.Sprintf("failed to register stream module codecs: %v", err))
	// }

	appKeepers.ClockKeeper = clockkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[clocktypes.StoreKey]),
		appKeepers.WasmKeeper,
		appKeepers.ContractKeeper,
		govModAddress,
	)

	appKeepers.CWHooksKeeper = cwhookskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[cwhookstypes.StoreKey]),
		*stakingKeeper,
		*govKeeper.Keeper,
		appKeepers.WasmKeeper,
		appKeepers.ContractKeeper,
		govModAddress,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	// this must be at the end so CWHooksKeeper can use the contractKeeper
	stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			appKeepers.DistrKeeper.Hooks(),
			appKeepers.SlashingKeeper.Hooks(),
			appKeepers.CWHooksKeeper.StakingHooks(),
			appKeepers.VotingSnapshotKeeper.Hooks(),
		),
	)
	appKeepers.StakingKeeper = stakingKeeper

	appKeepers.GovKeeper = govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
			appKeepers.CWHooksKeeper.GovHooks(),
		),
	)

	// Create Transfer Stack
	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(appKeepers.TransferKeeper)
	transferStack = ibchooks.NewIBCMiddleware(transferStack, &appKeepers.HooksICS4Wrapper)
	transferStack = packetforward.NewIBCMiddleware(
		transferStack,
		appKeepers.PacketForwardKeeper,
		0,
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
	)

	// ICA controller stack (ibc-go v10: NewIBCMiddleware takes only the keeper now;
	// the auth-module stack arg was dropped along with the capability scoping)
	var icaControllerStack porttypes.IBCModule = icacontroller.NewIBCMiddleware(appKeepers.ICAControllerKeeper)

	// ICA host stack
	var icaHostStack porttypes.IBCModule = icahost.NewIBCModule(appKeepers.ICAHostKeeper)

	// Wasm IBC stack — Phase 2: NewIBCHandler signature changed in wasmd v0.61
	// (3rd arg is now ICS20TransferPortSource, 4th is appVersionGetter)
	var wasmStack porttypes.IBCModule = wasm.NewIBCHandler(appKeepers.WasmKeeper, appKeepers.IBCKeeper.ChannelKeeper, appKeepers.TransferKeeper, appKeepers.IBCKeeper.ChannelKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter().
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		AddRoute(wasmtypes.ModuleName, wasmStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack)
	appKeepers.IBCKeeper.SetRouter(ibcRouter)

	appKeepers.ScopedIBCKeeper = scopedIBCKeeper
	appKeepers.ScopedTransferKeeper = scopedTransferKeeper
	appKeepers.scopedWasmKeeper = scopedWasmKeeper
	appKeepers.ScopedICAHostKeeper = scopedICAHostKeeper
	appKeepers.ScopedICAControllerKeeper = scopedICAControllerKeeper

	return appKeepers
}

// TODO: we need to wait until ALL cosmos sdk and juno modules FULLY implement SDK collections
// until we can simplify the state decode by A LOT. Keeping this here for future reference.
// type moduleEntry struct {
// 	storeKey string
// 	schema   collections.Schema
// }

// func (appKeepers *AppKeepers) registerStreamModuleCodecs(appCodec codec.Codec) error {
// 	modules := []moduleEntry{
// 		{storeKey: distrtypes.StoreKey, schema: appKeepers.DistrKeeper.Schema},
// 	}

// 	for _, module := range modules {
// 		if err := appKeepers.StreamKeeper.RegisterModuleSchema(module.storeKey, module.schema, collections.IndexingOptions{}); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// BlockedAddresses returns all the app's blocked account addresses.
func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range GetMaccPerms() {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	delete(modAccAddrs, authtypes.NewModuleAddress(feepaytypes.ModuleName).String())

	return modAccAddrs
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	return maccPerms
}
