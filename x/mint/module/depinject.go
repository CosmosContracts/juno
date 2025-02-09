package module

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	modulev1 "github.com/CosmosContracts/juno/v28/api/juno/mint/module/v1"
	"github.com/CosmosContracts/juno/v28/x/mint/keeper"
	"github.com/CosmosContracts/juno/v28/x/mint/types"
)

func init() {
	appmodule.Register(&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

// nolint:revive
type ModuleInputs struct {
	depinject.In

	ModuleKey    depinject.OwnModuleKey
	Config       *modulev1.Module
	StoreService store.KVStoreService
	Cdc          codec.Codec

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
	StakingKeeper types.StakingKeeper
}

// nolint:revive
type ModuleOutputs struct {
	depinject.Out

	MintKeeper keeper.Keeper
	Module     appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	feeCollectorName := in.Config.FeeCollectorName
	if feeCollectorName == "" {
		feeCollectorName = authtypes.FeeCollectorName
	}

	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	k := keeper.NewKeeper(
		in.Cdc,
		in.StoreService,
		in.StakingKeeper,
		in.AccountKeeper,
		in.BankKeeper,
		feeCollectorName,
		authority.String(),
	)

	m := NewAppModule(in.Cdc, k, in.AccountKeeper)

	return ModuleOutputs{MintKeeper: k, Module: m}
}
