package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
)

var _ types.QueryServer = querier{}

// Querier implements a QueryServer for the x/oracle module.
type querier struct {
	Keeper
}

// NewQuerier returns an implementation of the oracle QueryServer interface
// for the provided Keeper.
func NewQuerier(keeper Keeper) types.QueryServer {
	return &querier{Keeper: keeper}
}

// Params queries params of x/oracle module.
func (q querier) Params(
	goCtx context.Context,
	req *types.QueryParams,
) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// ExchangeRates queries exchange rates of all denoms, or, if specified, returns
// a single denom.
func (q querier) ExchangeRates(
	goCtx context.Context,
	req *types.QueryExchangeRates,
) (*types.QueryExchangeRatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var exchangeRates sdk.DecCoins

	if len(req.Denom) > 0 {
		exchangeRate, err := q.GetExchangeRate(ctx, req.Denom)
		if err != nil {
			return nil, err
		}

		exchangeRates = exchangeRates.Add(sdk.NewDecCoinFromDec(req.Denom, exchangeRate))
	} else {
		q.IterateExchangeRates(ctx, func(denom string, rate sdk.Dec) (stop bool) {
			exchangeRates = exchangeRates.Add(sdk.NewDecCoinFromDec(denom, rate))
			return false
		})
	}

	return &types.QueryExchangeRatesResponse{ExchangeRates: exchangeRates}, nil
}

// ActiveExchangeRates queries all denoms for which exchange rates exist.
func (q querier) ActiveExchangeRates(
	goCtx context.Context,
	req *types.QueryActiveExchangeRates,
) (*types.QueryActiveExchangeRatesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	denoms := []string{}
	q.IterateExchangeRates(ctx, func(denom string, _ sdk.Dec) (stop bool) {
		denoms = append(denoms, denom)
		return false
	})

	return &types.QueryActiveExchangeRatesResponse{ActiveRates: denoms}, nil
}

// FeederDelegation queries the account address to which the validator operator
// delegated oracle vote rights.
func (q querier) FeederDelegation(
	goCtx context.Context,
	req *types.QueryFeederDelegation,
) (*types.QueryFeederDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	feederAddr, err := q.GetFeederDelegation(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	return &types.QueryFeederDelegationResponse{
		FeederAddr: feederAddr.String(),
	}, nil
}

// MissCounter queries oracle miss counter of a validator.
func (q querier) MissCounter(
	goCtx context.Context,
	req *types.QueryMissCounter,
) (*types.QueryMissCounterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryMissCounterResponse{
		MissCounter: q.GetMissCounter(ctx, valAddr),
	}, nil
}

// SlashWindow queries the current slash window progress of the oracle.
func (q querier) SlashWindow(
	goCtx context.Context,
	req *types.QuerySlashWindow,
) (*types.QuerySlashWindowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.GetParams(ctx)

	return &types.QuerySlashWindowResponse{
		WindowProgress: (uint64(ctx.BlockHeight()) % params.SlashWindow) /
			params.VotePeriod,
	}, nil
}

// AggregatePrevote queries an aggregate prevote of a validator.
func (q querier) AggregatePrevote(
	goCtx context.Context,
	req *types.QueryAggregatePrevote,
) (*types.QueryAggregatePrevoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	prevote, err := q.GetAggregateExchangeRatePrevote(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	return &types.QueryAggregatePrevoteResponse{
		AggregatePrevote: prevote,
	}, nil
}

// AggregatePrevotes queries aggregate prevotes of all validators
func (q querier) AggregatePrevotes(
	goCtx context.Context,
	req *types.QueryAggregatePrevotes,
) (*types.QueryAggregatePrevotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var prevotes []types.AggregateExchangeRatePrevote
	q.IterateAggregateExchangeRatePrevotes(ctx, func(_ sdk.ValAddress, prevote types.AggregateExchangeRatePrevote) bool {
		prevotes = append(prevotes, prevote)
		return false
	})

	return &types.QueryAggregatePrevotesResponse{
		AggregatePrevotes: prevotes,
	}, nil
}

// AggregateVote queries an aggregate vote of a validator
func (q querier) AggregateVote(
	goCtx context.Context,
	req *types.QueryAggregateVote,
) (*types.QueryAggregateVoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	vote, err := q.GetAggregateExchangeRateVote(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	return &types.QueryAggregateVoteResponse{
		AggregateVote: vote,
	}, nil
}

// AggregateVotes queries aggregate votes of all validators
func (q querier) AggregateVotes(
	goCtx context.Context,
	req *types.QueryAggregateVotes,
) (*types.QueryAggregateVotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var votes []types.AggregateExchangeRateVote
	q.IterateAggregateExchangeRateVotes(ctx, func(_ sdk.ValAddress, vote types.AggregateExchangeRateVote) bool {
		votes = append(votes, vote)
		return false
	})

	return &types.QueryAggregateVotesResponse{
		AggregateVotes: votes,
	}, nil
}

// PriceTrackingLists queries all price tracking lists
func (q querier) PriceTrackingLists(
	goCtx context.Context,
	req *types.QueryPriceTrackingLists,
) (*types.QueryPriceTrackingListsRespone, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.GetParams(ctx)

	var result []string
	for _, trackingDenom := range params.PriceTrackingList {
		result = append(result, trackingDenom.SymbolDenom)
	}

	return &types.QueryPriceTrackingListsRespone{
		PriceTrakingLists: result,
	}, nil
}

func (q querier) PriceHistoryAt(
	goCtx context.Context,
	req *types.QueryPriceHistoryAt,
) (*types.QueryPriceHistoryAtRespone, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	found, denom := q.isInTrackingList(ctx, req.Denom)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "Denom %s not in tracking list", req.Denom)
	}

	priceHistoryEntry, err := q.GetDenomPriceHistoryWithBlockHeight(ctx, req.Denom, req.BlockHeight)
	if err != nil {
		return nil, err
	}

	return &types.QueryPriceHistoryAtRespone{
		Denom:             denom,
		PriceHistoryEntry: priceHistoryEntry,
	}, nil
}

func (q querier) AllPriceHistory(
	goCtx context.Context,
	req *types.QueryAllPriceHistory,
) (*types.QueryAllPriceHistoryRespone, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	found, denom := q.isInTrackingList(ctx, req.Denom)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "Denom %s not in tracking list", req.Denom)
	}

	store := ctx.KVStore(q.storeKey)
	priceHistoryStore := prefix.NewStore(store, types.FormatHistoricalDenomIndexPrefix(req.Denom))

	var priceHistoryEntrys []types.PriceHistoryEntry

	pageRes, err := query.Paginate(priceHistoryStore, req.Pagination, func(key []byte, value []byte) error {
		var priceHistoryEntry types.PriceHistoryEntry
		if err := q.cdc.Unmarshal(value, &priceHistoryEntry); err != nil {
			return err
		}
		priceHistoryEntrys = append(priceHistoryEntrys, priceHistoryEntry)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllPriceHistoryRespone{
		Denom:              denom,
		PriceHistoryEntrys: priceHistoryEntrys,
		Pagination:         pageRes,
	}, nil
}

func (q querier) CurrentVotePeriodCount(
	goCtx context.Context,
	req *types.QueryCurrentVotePeriodCount,
) (*types.QueryCurrentVotePeriodCountRespone, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	found, _ := q.isInTrackingList(ctx, req.Denom)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "Denom %s not in tracking list", req.Denom)
	}

	store := ctx.KVStore(q.storeKey)
	priceHistoryStore := prefix.NewStore(store, types.GetPriceHistoryKey(req.Denom))
	iter := sdk.KVStoreReversePrefixIterator(priceHistoryStore, []byte{})

	defer iter.Close()

	var currentVotePeriodCount uint64
	if iter.Valid() {
		currentVotePeriodCount = sdk.BigEndianToUint64(iter.Key())
	}

	if currentVotePeriodCount == 0 {
		return nil, status.Errorf(codes.Internal, "Denom %s not have price tracking data", req.Denom)
	}

	return &types.QueryCurrentVotePeriodCountRespone{
		VotePeriodCount: currentVotePeriodCount,
	}, nil
}
