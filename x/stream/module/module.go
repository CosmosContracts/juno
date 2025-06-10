package module

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
)

var (
	_ module.AppModuleBasic      = AppModule{}
	_ appmodule.AppModule        = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasServices         = AppModule{}
	_ appmodule.HasPreBlocker    = AppModule{}
)

// AppModuleBasic implements the AppModuleBasic interface for the stream module.
type AppModuleBasic struct {
	cdc codec.BinaryCodec
}

func NewAppModuleBasic(cdc codec.BinaryCodec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// Name returns the stream module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the mint module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// No need to register legacy codec for stream module
	// we do not have any messages as we just wrap queries
}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(r cdctypes.InterfaceRegistry) {
	// No need to register interfaces for stream module
	// we do not have any messages as we just wrap queries
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// No need to register gRPC Gateway routes for stream module
	// Users should connect directly to gRPC or WebSocket endpoints
}

// AppModule implements the AppModule interface for the stream module.
type AppModule struct {
	AppModuleBasic

	keeper *keeper.Keeper
}

func NewAppModule(
	cdc codec.Codec,
	keeper *keeper.Keeper,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
	}
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// ConsensusVersion implements ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// PreBlock is called before the beginning of each block
func (am AppModule) PreBlock(ctx context.Context) (appmodule.ResponsePreBlock, error) {
	err := am.keeper.PreBlocker(ctx)
	if err != nil {
		return nil, err
	}
	return &sdk.ResponsePreBlock{ConsensusParamsChanged: false}, nil
}
