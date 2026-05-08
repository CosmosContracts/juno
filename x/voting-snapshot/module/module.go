// Package module wires x/voting-snapshot into the app's module manager.
//
// MVP shape for v30: no proto-generated gRPC msg/query servers, no
// migrations. Delivery is via wasmbindings/query_plugin.go (which
// calls Keeper.VotingPowerAt / TotalVotingPowerAt directly). gRPC and
// REST surfaces are tracked for v30.x.
package module

import (
	"context"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmosContracts/juno/v30/x/voting-snapshot/keeper"
	"github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
)

const ConsensusVersion = 1

//nolint:staticcheck // module.AppModule deprecation deferred to v31 (see app/modules.go)
var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string                                                    { return types.ModuleName }
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino)                   {}
func (AppModuleBasic) RegisterInterfaces(_ cdctypes.InterfaceRegistry)                 {}
func (AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

func (AppModuleBasic) DefaultGenesis(_ codec.JSONCodec) json.RawMessage {
	b, err := json.Marshal(types.DefaultGenesis())
	if err != nil {
		panic(err)
	}
	return b
}

func (AppModuleBasic) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, raw json.RawMessage) error {
	var gs types.GenesisState
	return json.Unmarshal(raw, &gs)
}

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{keeper: k}
}

func (AppModule) IsOnePerModuleType()                    {}
func (AppModule) IsAppModule()                           {}
func (AppModule) ConsensusVersion() uint64               { return ConsensusVersion }
func (AppModule) RegisterServices(_ module.Configurator) {}

func (am AppModule) InitGenesis(ctx sdk.Context, _ codec.JSONCodec, raw json.RawMessage) {
	var gs types.GenesisState
	if err := json.Unmarshal(raw, &gs); err != nil {
		panic(err)
	}
	if err := am.keeper.InitGenesis(ctx, &gs); err != nil {
		panic(err)
	}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, _ codec.JSONCodec) json.RawMessage {
	gs, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	b, err := json.Marshal(gs)
	if err != nil {
		panic(err)
	}
	return b
}

// BeginBlock is a no-op — snapshot writes are driven by staking hooks,
// not block timing.
func (AppModule) BeginBlock(_ context.Context) error { return nil }

// EndBlock is a no-op — see BeginBlock.
func (AppModule) EndBlock(_ context.Context) error { return nil }
