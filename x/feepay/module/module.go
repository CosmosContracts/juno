package module

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/CosmosContracts/juno/v27/x/feepay/keeper"
	"github.com/CosmosContracts/juno/v27/x/feepay/types"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}

	_ appmodule.AppModule = AppModule{}
)

// ConsensusVersion defines the current x/feepay module consensus version.
const ConsensusVersion = 1

// AppModuleBasic type for the fees module
type AppModuleBasic struct{}

// Name returns the fees module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec performs a no-op as the fees do not support Amino
// encoding.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers interfaces and implementations of the feepay
// module.
func (AppModuleBasic) RegisterInterfaces(interfaceRegistry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(interfaceRegistry)
}

// DefaultGenesis returns default genesis state as raw bytes for the feepay
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the feepay module.
func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genesisState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesisState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genesisState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the feepay module
// module.
func (b AppModuleBasic) RegisterGRPCGatewayRoutes(c client.Context, serveMux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(c)); err != nil {
		panic(err)
	}
}

// AppModule implements the AppModule interface for the feepay module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
	ak     authkeeper.AccountKeeper
}

// NewAppModule creates a new AppModule Object
func NewAppModule(
	k keeper.Keeper,
	ak authkeeper.AccountKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		ak:             ak,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// InitGenesis performs the feepay module's genesis initialization. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	am.keeper.InitGenesis(ctx, genesisState)
}

// ExportGenesis returns the feepay module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion returns the consensus state-breaking version for the module.
func (AppModuleBasic) ConsensusVersion() uint64 {
	return ConsensusVersion
}
