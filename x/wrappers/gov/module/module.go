package module

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	v1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/CosmosContracts/juno/v28/x/wrappers/gov/keeper"
)

type AppModuleBasic struct {
	gov.AppModuleBasic
}

type AppModule struct {
	gov.AppModule
	keeper         *keeper.KeeperWrapper
	accountKeeper  govtypes.AccountKeeper
	legacySubspace govtypes.ParamSubspace
}

func NewAppModule(
	cdc codec.Codec, keeper *keeper.KeeperWrapper, ak govtypes.AccountKeeper, bk govtypes.BankKeeper, ss govtypes.ParamSubspace,
) AppModule {
	govModule := gov.NewAppModule(cdc, keeper.Keeper, ak, bk, ss)
	return AppModule{
		AppModule:      govModule,
		keeper:         keeper,
		accountKeeper:  ak,
		legacySubspace: ss,
	}
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	msgServer := govkeeper.NewMsgServerImpl(am.keeper.Keeper)
	v1beta1.RegisterMsgServer(cfg.MsgServer(), govkeeper.NewLegacyMsgServerImpl(am.accountKeeper.GetModuleAddress(govtypes.ModuleName).String(), msgServer))
	v1.RegisterMsgServer(cfg.MsgServer(), msgServer)

	legacyQueryServer := keeper.NewLegacyQueryServer(am.keeper)
	v1beta1.RegisterQueryServer(cfg.QueryServer(), legacyQueryServer)
	v1.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))

	m := govkeeper.NewMigrator(am.keeper.Keeper, am.legacySubspace)
	if err := cfg.RegisterMigration(govtypes.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/gov from version 1 to 2: %v", err))
	}

	if err := cfg.RegisterMigration(govtypes.ModuleName, 2, m.Migrate2to3); err != nil {
		panic(fmt.Sprintf("failed to migrate x/gov from version 2 to 3: %v", err))
	}

	if err := cfg.RegisterMigration(govtypes.ModuleName, 3, m.Migrate3to4); err != nil {
		panic(fmt.Sprintf("failed to migrate x/gov from version 3 to 4: %v", err))
	}

	if err := cfg.RegisterMigration(govtypes.ModuleName, 4, m.Migrate4to5); err != nil {
		panic(fmt.Sprintf("failed to migrate x/gov from version 4 to 5: %v", err))
	}
}
