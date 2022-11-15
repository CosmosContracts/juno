package simulation

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/CosmosContracts/juno/v12/x/oracle/keeper"
	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
//
//nolint:gosec
const (
	OpWeightMsgAggregateExchangeRatePrevote = "op_weight_msg_exchange_rate_aggregate_prevote"
	OpWeightMsgAggregateExchangeRateVote    = "op_weight_msg_exchange_rate_aggregate_vote"
	OpWeightMsgDelegateFeedConsent          = "op_weight_msg_exchange_feed_consent"

	salt = "89b8164ca0b4b8703ae9ab25962f3dd6d1de5d656f5442971a93b2ca7893f654"
)

var (
	acceptList = []string{types.JunoSymbol, types.USDDenom}
	umeePrice  = sdk.MustNewDecFromStr("25.71")
)

// GenerateExchangeRatesString generates a canonical string representation of
// the aggregated exchange rates.
func GenerateExchangeRatesString(prices map[string]sdk.Dec) string {
	exchangeRates := make([]string, len(prices))
	i := 0

	// aggregate exchange rates as "<base>:<price>"
	for base, avgPrice := range prices {
		exchangeRates[i] = fmt.Sprintf("%s:%s", base, avgPrice.String())
		i++
	}

	sort.Strings(exchangeRates)

	return strings.Join(exchangeRates, ",")
}

func WeightedOperations(
	simstate *module.SimulationState,
	ak types.AccountKeeper,
	bk bankkeeper.Keeper,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgAggregateExchangeRatePrevote int
		weightMsgAggregateExchangeRateVote    int
		weightMsgDelegateFeedConsent          int
		voteHashMap                           = make(map[string]string)
	)

	simstate.AppParams.GetOrGenerate(simstate.Cdc, OpWeightMsgAggregateExchangeRatePrevote, &weightMsgAggregateExchangeRatePrevote, nil,
		func(_ *rand.Rand) {
			weightMsgAggregateExchangeRatePrevote = simappparams.DefaultWeightMsgSend * 2
		},
	)
	simstate.AppParams.GetOrGenerate(simstate.Cdc, OpWeightMsgAggregateExchangeRateVote, &weightMsgAggregateExchangeRateVote, nil,
		func(_ *rand.Rand) {
			weightMsgAggregateExchangeRateVote = simappparams.DefaultWeightMsgSend * 2
		},
	)
	simstate.AppParams.GetOrGenerate(simstate.Cdc, OpWeightMsgDelegateFeedConsent, &weightMsgDelegateFeedConsent, nil,
		func(_ *rand.Rand) {
			weightMsgDelegateFeedConsent = simappparams.DefaultWeightMsgSetWithdrawAddress
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgAggregateExchangeRatePrevote,
			SimulateMsgAggregateExchangeRatePrevote(ak, bk, k, voteHashMap),
		),
		simulation.NewWeightedOperation(
			weightMsgAggregateExchangeRateVote,
			SimulateMsgAggregateExchangeRateVote(ak, bk, k, voteHashMap),
		),
		simulation.NewWeightedOperation(
			weightMsgDelegateFeedConsent,
			SimulateMsgDelegateFeedConsent(ak, bk, k),
		),
	}
}

// SimulateMsgAggregateExchangeRatePrevote generates a MsgAggregateExchangeRatePrevote with random values.
func SimulateMsgAggregateExchangeRatePrevote(
	ak types.AccountKeeper,
	bk bankkeeper.Keeper,
	k keeper.Keeper,
	voteHashMap map[string]string,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount := accs[1] // , _ := simtypes.RandomAcc(r, accs)
		address := sdk.ValAddress(simAccount.Address)
		noop := func(comment string) simtypes.OperationMsg {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(new(types.MsgAggregateExchangeRatePrevote)), comment)
		}

		// ensure the validator exists
		val := k.StakingKeeper.Validator(ctx, address)
		if val == nil || !val.IsBonded() {
			return noop("unable to find validator"), nil, nil
		}

		// check for an existing prevote
		_, err := k.GetAggregateExchangeRatePrevote(ctx, address)
		if err == nil {
			return noop("prevote already exists for this validator"), nil, nil
		}

		prices := make(map[string]sdk.Dec, len(acceptList))
		for _, denom := range acceptList {
			prices[denom] = umeePrice.Add(simtypes.RandomDecAmount(r, sdk.NewDec(1)))
		}

		exchangeRatesStr := GenerateExchangeRatesString(prices)
		voteHash := types.GetAggregateVoteHash(salt, exchangeRatesStr, address)
		feederAddr, _ := k.GetFeederDelegation(ctx, address)
		feederSimAccount, _ := simtypes.FindAccount(accs, feederAddr)
		msg := types.NewMsgAggregateExchangeRatePrevote(voteHash, feederAddr, address)
		voteHashMap[address.String()] = exchangeRatesStr

		return deliver(r, app, ctx, ak, bk, feederSimAccount, msg, nil)
	}
}

// SimulateMsgAggregateExchangeRateVote generates a MsgAggregateExchangeRateVote with random values.
func SimulateMsgAggregateExchangeRateVote(
	ak types.AccountKeeper,
	bk bankkeeper.Keeper,
	k keeper.Keeper,
	voteHashMap map[string]string,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		address := sdk.ValAddress(simAccount.Address)
		noop := func(comment string) simtypes.OperationMsg {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(new(types.MsgAggregateExchangeRateVote)), comment)
		}

		// ensure the validator exists
		val := k.StakingKeeper.Validator(ctx, address)
		if val == nil || !val.IsBonded() {
			return noop("unable to find validator"), nil, nil
		}

		// ensure vote hash exists
		exchangeRatesStr, ok := voteHashMap[address.String()]
		if !ok {
			return noop("vote hash does not exist"), nil, nil
		}

		// get prevote
		prevote, err := k.GetAggregateExchangeRatePrevote(ctx, address)
		if err != nil {
			return noop("prevote not found"), nil, nil
		}

		params := k.GetParams(ctx)
		if (uint64(ctx.BlockHeight())/params.VotePeriod)-(prevote.SubmitBlock/params.VotePeriod) != 1 {
			return noop("reveal period of submitted vote does not match with registered prevote"), nil, nil
		}

		feederAddr, _ := k.GetFeederDelegation(ctx, address)
		feederSimAccount, _ := simtypes.FindAccount(accs, feederAddr)
		msg := types.NewMsgAggregateExchangeRateVote(salt, exchangeRatesStr, feederAddr, address)

		return deliver(r, app, ctx, ak, bk, feederSimAccount, msg, nil)
	}
}

// SimulateMsgDelegateFeedConsent generates a MsgDelegateFeedConsent with random values.
func SimulateMsgDelegateFeedConsent(ak types.AccountKeeper, bk bankkeeper.Keeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		delegateAccount, _ := simtypes.RandomAcc(r, accs)
		valAddress := sdk.ValAddress(simAccount.Address)
		delegateValAddress := sdk.ValAddress(delegateAccount.Address)
		noop := func(comment string) simtypes.OperationMsg {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(new(types.MsgDelegateFeedConsent)), comment)
		}

		// ensure the validator exists
		val := k.StakingKeeper.Validator(ctx, valAddress)
		if val == nil {
			return noop("unable to find validator"), nil, nil
		}

		// ensure the target address is not a validator
		val2 := k.StakingKeeper.Validator(ctx, delegateValAddress)
		if val2 != nil {
			return noop("unable to delegate to validator"), nil, nil
		}

		msg := types.NewMsgDelegateFeedConsent(valAddress, delegateAccount.Address)
		return deliver(r, app, ctx, ak, bk, simAccount, msg, nil)
	}
}

func deliver(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, ak simulation.AccountKeeper,
	bk bankkeeper.Keeper, from simtypes.Account, msg sdk.Msg, coins sdk.Coins,
) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	cfg := simappparams.MakeTestEncodingConfig()
	txCtx := simulation.OperationInput{
		R:               r,
		App:             app,
		TxGen:           cfg.TxConfig,
		Cdc:             cfg.Marshaler.(*codec.ProtoCodec),
		Msg:             msg,
		MsgType:         sdk.MsgTypeURL(msg),
		Context:         ctx,
		SimAccount:      from,
		AccountKeeper:   ak,
		Bankkeeper:      bk,
		ModuleName:      types.ModuleName,
		CoinsSpentInMsg: coins,
	}

	return simulation.GenAndDeliverTxWithRandFees(txCtx)
}
