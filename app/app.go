package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	wasm "github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/spf13/cast"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/gogoproto/proto"

	wasmlckeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"

	"cosmossdk.io/x/tx/signing"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/CosmosContracts/juno/v27/app/keepers"
	upgrades "github.com/CosmosContracts/juno/v27/app/upgrades"
	testnetV18alpha2 "github.com/CosmosContracts/juno/v27/app/upgrades/testnet/v18.0.0-alpha.2"
	testnetV18alpha3 "github.com/CosmosContracts/juno/v27/app/upgrades/testnet/v18.0.0-alpha.3"
	testnetV18alpha4 "github.com/CosmosContracts/juno/v27/app/upgrades/testnet/v18.0.0-alpha.4"
	testnetV19alpha3 "github.com/CosmosContracts/juno/v27/app/upgrades/testnet/v19.0.0-alpha.3"
	testnetV21alpha1 "github.com/CosmosContracts/juno/v27/app/upgrades/testnet/v21.0.0-alpha.1"
	testnetV22alpha1 "github.com/CosmosContracts/juno/v27/app/upgrades/testnet/v22.0.0-alpha.1"
	testnetV23alpha1 "github.com/CosmosContracts/juno/v27/app/upgrades/testnet/v23.0.0-alpha.1"
	v10 "github.com/CosmosContracts/juno/v27/app/upgrades/v10"
	v11 "github.com/CosmosContracts/juno/v27/app/upgrades/v11"
	v12 "github.com/CosmosContracts/juno/v27/app/upgrades/v12"
	v13 "github.com/CosmosContracts/juno/v27/app/upgrades/v13"
	v14 "github.com/CosmosContracts/juno/v27/app/upgrades/v14"
	v15 "github.com/CosmosContracts/juno/v27/app/upgrades/v15"
	v16 "github.com/CosmosContracts/juno/v27/app/upgrades/v16"
	v17 "github.com/CosmosContracts/juno/v27/app/upgrades/v17"
	v18 "github.com/CosmosContracts/juno/v27/app/upgrades/v18"
	v19 "github.com/CosmosContracts/juno/v27/app/upgrades/v19"
	v21 "github.com/CosmosContracts/juno/v27/app/upgrades/v21"
	v22 "github.com/CosmosContracts/juno/v27/app/upgrades/v22"
	v23 "github.com/CosmosContracts/juno/v27/app/upgrades/v23"
	v24 "github.com/CosmosContracts/juno/v27/app/upgrades/v24"
	v25 "github.com/CosmosContracts/juno/v27/app/upgrades/v25"
	v26 "github.com/CosmosContracts/juno/v27/app/upgrades/v26"
)

const (
	Name = "juno"
)

// We pull these out so we can set them with LDFLAGS in the Makefile
var (
	NodeDir = ".juno"
	// DefaultNodeHome default home directories for Juno
	DefaultNodeHome = os.ExpandEnv("$HOME/") + NodeDir
	// If EnabledSpecificProposals is "", and this is "true", then enable all x/wasm proposals.
	// If EnabledSpecificProposals is "", and this is not "true", then disable all x/wasm proposals.
	ProposalsEnabled = "true"
	// If set to non-empty string it must be comma-separated list of values that are all a subset
	// of "EnableAllProposals" (takes precedence over ProposalsEnabled)
	// https://github.com/CosmWasm/wasmd/blob/02a54d33ff2c064f3539ae12d75d027d9c665f05/x/wasm/internal/types/proposal.go#L28-L34
	EnableSpecificProposals = ""

	Upgrades = []upgrades.Upgrade{
		// testnet
		testnetV18alpha2.Upgrade,
		testnetV18alpha3.Upgrade,
		testnetV18alpha4.Upgrade,
		testnetV19alpha3.Upgrade,
		testnetV21alpha1.Upgrade,
		testnetV22alpha1.Upgrade,
		testnetV23alpha1.Upgrade,

		v10.Upgrade,
		v11.Upgrade,
		v12.Upgrade,
		v13.Upgrade,
		v14.Upgrade,
		v15.Upgrade,
		v16.Upgrade,
		v17.Upgrade,
		v18.Upgrade,
		v19.Upgrade,
		v21.Upgrade,
		v22.Upgrade,
		v23.Upgrade,
		v24.Upgrade,
		v25.Upgrade,
		v26.Upgrade,
	}

	_ runtime.AppI            = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	// keepers
	AppKeepers keepers.AppKeepers

	// modules
	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

	// simulation
	configurator module.Configurator
	sm           *module.SimulationManager
}

// New returns a reference to an initialized Juno.
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmkeeper.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	if err != nil {
		panic(err)
	}

	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	bApp := baseapp.NewBaseApp(Name, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
	}

	app.AppKeepers = keepers.NewAppKeepers(
		appCodec,
		bApp,
		legacyAmino,
		keepers.GetMaccPerms(),
		appOpts,
		wasmOpts,
		app.GetChainBondDenom(),
	)

	// load state streaming if enabled
	if err := app.RegisterStreamingServices(appOpts, app.AppKeepers.GetKVStoreKeys()); err != nil {
		panic(err)
	}

	// optional: enable sign mode textual by overwriting the default tx config (after setting the bank keeper)
	enabledSignModes := append(authtx.DefaultSignModes, signingtypes.SignMode_SIGN_MODE_TEXTUAL)
	txConfigOpts := authtx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: authtxconfig.NewBankKeeperCoinMetadataQueryFn(app.AppKeepers.BankKeeper),
	}
	txConfig, err = authtx.NewTxConfigWithOptions(
		appCodec,
		txConfigOpts,
	)
	if err != nil {
		panic(err)
	}
	app.txConfig = txConfig

	if maxSize := os.Getenv("MAX_WASM_SIZE"); maxSize != "" {
		// https://github.com/CosmWasm/wasmd#compile-time-parameters
		val, _ := strconv.ParseInt(maxSize, 10, 32)
		wasmtypes.MaxWasmSize = int(val)
	}

	app.configurator = module.NewConfigurator(appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.ModuleManager = module.NewManager(appModules(app, txConfig, appCodec, skipGenesisInvariants)...)
	err = app.ModuleManager.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{
					paramsclient.ProposalHandler,
				},
			),
		},
	)
	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	app.ModuleManager.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)
	app.ModuleManager.SetOrderBeginBlockers(orderBeginBlockers()...)
	app.ModuleManager.SetOrderEndBlockers(orderEndBlockers()...)
	app.ModuleManager.SetOrderInitGenesis(orderInitBlockers()...)
	app.ModuleManager.SetOrderExportGenesis(orderInitBlockers()...)
	app.ModuleManager.RegisterInvariants(app.AppKeepers.CrisisKeeper)

	// initialize stores
	app.MountKVStores(app.AppKeepers.GetKVStoreKeys())
	app.MountTransientStores(app.AppKeepers.GetTransientStoreKeys())
	app.MountMemoryStores(app.AppKeepers.GetMemoryStoreKeys())

	autocliv1.RegisterQueryServer(
		app.GRPCQueryRouter(),
		runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules),
	)
	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	nodeConfig, err := wasm.ReadNodeConfig(appOpts)
	if err != nil {
		panic("error while reading wasm config: " + err.Error())
	}

	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AppKeepers.AccountKeeper,
				BankKeeper:      app.AppKeepers.BankKeeper,
				FeegrantKeeper:  app.AppKeepers.FeeGrantKeeper,
				SignModeHandler: app.txConfig.SignModeHandler(),
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
			GovKeeper:             app.AppKeepers.GovKeeper,
			IBCKeeper:             app.AppKeepers.IBCKeeper,
			FeePayKeeper:          app.AppKeepers.FeePayKeeper,
			FeeShareKeeper:        app.AppKeepers.FeeShareKeeper,
			BankKeeper:            app.AppKeepers.BankKeeper,
			TXCounterStoreService: runtime.NewKVStoreService(app.AppKeepers.GetKey(wasmtypes.StoreKey)),
			NodeConfig:            &nodeConfig,
			Cdc:                   appCodec,

			BypassMinFeeMsgTypes: GetDefaultBypassFeeMessages(),
			GlobalFeeKeeper:      app.AppKeepers.GlobalFeeKeeper,
			StakingKeeper:        *app.AppKeepers.StakingKeeper,

			TxEncoder: app.txConfig.TxEncoder(),
			BondDenom: app.GetChainBondDenom(),
		},
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	app.setPostHandler()

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	if manager := app.SnapshotManager(); manager != nil {
		err = manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.AppKeepers.WasmKeeper),
			// https://github.com/cosmos/ibc-go/pull/5439
			wasmlckeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.AppKeepers.WasmClientKeeper),
		)
		if err != nil {
			panic("failed to register snapshot extension: " + err.Error())
		}
	}

	app.setupUpgradeHandlers(app.configurator)
	app.setupUpgradeStoreLoaders()

	// At startup, after all modules have been registered, check that all proto
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		panic(err)
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Once we switch to using protoreflect-based antehandlers, we might
		// want to panic here instead of logging a warning.
		fmt.Fprintln(os.Stderr, err.Error())
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

		// https://github.com/cosmos/ibc-go/pull/5439
		if err := wasmlckeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("wasmlckeeper failed initialize pinned codes %s", err))
		}

		if err := app.AppKeepers.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("app.AppKeepers.WasmKeeper failed initialize pinned codes %s", err))
		}

		// Initialize and seal the capability keeper so all persistent capabilities
		// are loaded in-memory and prevent any further modules from creating scoped
		// sub-keepers.
		// This must be done during creation of baseapp rather than in InitChain so
		// that in-memory capabilities get regenerated on app restart.
		// Note that since this reads from the store, we can only perform it when
		// `loadLatest` is set to true.
		app.AppKeepers.CapabilityKeeper.Seal()
	}

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// no override for simulation for now, but we can add it in the future if needed
	app.sm = module.NewSimulationManagerFromAppModules(
		app.ModuleManager.Modules,
		make(map[string]module.AppModuleSimulation, 0),
	)
	app.sm.RegisterStoreDecoders()

	return app
}

func GetDefaultBypassFeeMessages() []string {
	return []string{
		// IBC
		sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
		sdk.MsgTypeURL(&ibcclienttypes.MsgCreateClient{}),
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpgradeClient{}),
		sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeout{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeoutOnClose{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgChannelOpenTry{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgChannelOpenConfirm{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgChannelOpenAck{}),
	}
}

// AutoCLIOpts returns options based upon the modules used on Juno.
func (app *App) AutoCLIOpts(initClientCtx client.Context) autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule)
	for _, m := range app.ModuleManager.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
		ClientCtx:             initClientCtx,
	}
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (a *App) DefaultGenesis() map[string]json.RawMessage {
	return a.BasicModuleManager.DefaultGenesis(a.appCodec)
}

func (app *App) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

// Name returns the name of the App
func (app *App) Name() string {
	return app.BaseApp.Name()
}

// PreBlocker application updates every pre block
func (app *App) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	if err := app.AppKeepers.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap()); err != nil {
		panic(err)
	}

	response, err := app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
	if err != nil {
		panic(err)
	}
	return response, nil
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range keepers.GetMaccPerms() {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns Juno's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Juno's InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// InterfaceRegistry returns Juno's TxConfig
func (app *App) TxConfig() client.TxConfig {
	return app.txConfig
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: Still used in ibc-go, wait for them to remove params usage before removing this.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.AppKeepers.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new comet queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		cmtApp.Query,
	)
}

func (app *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// configure store loader that checks if version == upgradeHeight and applies store upgrades
func (app *App) setupUpgradeStoreLoaders() {
	upgradeInfo, err := app.AppKeepers.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic("failed to read upgrade info from disk" + err.Error())
	}

	if app.AppKeepers.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName {
			storeUpgrades := upgrade.StoreUpgrades
			app.SetStoreLoader(
				upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades),
			)
		}
	}
}

func (app *App) setupUpgradeHandlers(cfg module.Configurator) {
	for _, upgrade := range Upgrades {
		app.AppKeepers.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.ModuleManager,
				cfg,
				&app.AppKeepers,
			),
		)
	}
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

func (app *App) GetChainBondDenom() string {
	d := "ujuno"
	if strings.HasPrefix(app.ChainID(), "uni-") {
		d = "ujunox"
	}
	return d
}
