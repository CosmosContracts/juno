package module

import (
	"context"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	appmodule "cosmossdk.io/core/appmodule"
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v28/x/clock/keeper"
	"github.com/CosmosContracts/juno/v28/x/clock/types"
)

const (
	// ConsensusVersion defines the current x/clock module consensus version.
	ConsensusVersion = 1
)

var (
	_ module.AppModuleBasic = AppModule{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}

	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasEndBlocker = AppModule{}
)

// AppModuleBasic defines the basic application module used by the wasm module.
type AppModuleBasic struct {
	cdc codec.Codec
}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(r codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(r)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (AppModuleBasic) ValidateGenesis(marshaler codec.JSONCodec, _ client.TxEncodingConfig, message json.RawMessage) error {
	var data types.GenesisState
	err := marshaler.UnmarshalJSON(message, &data)
	if err != nil {
		return err
	}
	if err := data.Params.Validate(); err != nil {
		return errorsmod.Wrap(err, "params")
	}
	return nil
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

// NewAppModule constructor
func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
) *AppModule {
	return &AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// InitGenesis performs genesis initialization for the clock module.
func (am AppModule) InitGenesis(ctx sdk.Context, marshaler codec.JSONCodec, message json.RawMessage) {
	var genesisState types.GenesisState
	marshaler.MustUnmarshalJSON(message, &genesisState)

	am.keeper.InitGenesis(ctx, genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the mint
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, marshaler codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return marshaler.MustMarshalJSON(genState)
}

func (am AppModule) EndBlock(ctx context.Context) error {
	return keeper.EndBlocker(ctx, am.keeper)
}

// ConsensusVersion is a sequence number for state-breaking change of the
// module. It should be incremented on each consensus-breaking change
// introduced by the module. To avoid wrong/empty versions, the initial version
// should be set to 1.
func (am AppModule) ConsensusVersion() uint64 {
	return ConsensusVersion
}
