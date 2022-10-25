package staking

import (
	"math/rand"

	"github.com/CosmosContracts/juno/v11/x/staking/simulation"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	staking "github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type SimAppModule struct {
	AppModule staking.AppModule

	keeper        stakingkeeper.Keeper
	accountKeeper stakingtypes.AccountKeeper
	bankKeeper    stakingtypes.BankKeeper
}

func NewSimAppModule(
	cdc codec.Codec,
	keeper stakingkeeper.Keeper,
	ak stakingtypes.AccountKeeper,
	bk stakingtypes.BankKeeper,
) SimAppModule {
	return SimAppModule{
		AppModule:     staking.NewAppModule(cdc, keeper, ak, bk),
		keeper:        keeper,
		accountKeeper: ak,
		bankKeeper:    bk,
	}
}

func (am SimAppModule) GenerateGenesisState(simState *module.SimulationState) {
	am.AppModule.GenerateGenesisState(simState)
}

func (SimAppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

func (am SimAppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return am.AppModule.RandomizedParams(r)
}

func (am SimAppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	am.AppModule.RegisterStoreDecoder(sdr)
}

func (am SimAppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc, am.accountKeeper, am.bankKeeper, am.keeper,
	)
}
