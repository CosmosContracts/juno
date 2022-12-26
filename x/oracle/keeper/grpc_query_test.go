package keeper_test

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	appparams "github.com/CosmosContracts/juno/v12/app/params"
	"github.com/CosmosContracts/juno/v12/testutil/nullify"
	"github.com/CosmosContracts/juno/v12/x/oracle/keeper"
	"github.com/CosmosContracts/juno/v12/x/oracle/types"
)

func (s *IntegrationTestSuite) TestQuerier_ActiveExchangeRates() {
	s.app.OracleKeeper.SetExchangeRate(s.ctx, displayDenom, sdk.OneDec())
	res, err := s.queryClient.ActiveExchangeRates(s.ctx.Context(), &types.QueryActiveExchangeRates{})
	s.Require().NoError(err)
	s.Require().Equal([]string{displayDenom}, res.ActiveRates)
}

func (s *IntegrationTestSuite) TestQuerier_ExchangeRates() {
	s.app.OracleKeeper.SetExchangeRate(s.ctx, displayDenom, sdk.OneDec())
	res, err := s.queryClient.ExchangeRates(s.ctx.Context(), &types.QueryExchangeRates{})
	s.Require().NoError(err)
	s.Require().Equal(sdk.DecCoins{
		sdk.NewDecCoinFromDec(displayDenom, sdk.OneDec()),
	}, res.ExchangeRates)

	res, err = s.queryClient.ExchangeRates(s.ctx.Context(), &types.QueryExchangeRates{
		Denom: displayDenom,
	})
	s.Require().NoError(err)
	s.Require().Equal(sdk.DecCoins{
		sdk.NewDecCoinFromDec(displayDenom, sdk.OneDec()),
	}, res.ExchangeRates)
}

func (s *IntegrationTestSuite) TestQuerier_FeeederDelegation() {
	feederAddr := sdk.AccAddress([]byte("addr________________"))
	feederAcc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, feederAddr)
	inactiveValidator := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	s.app.AccountKeeper.SetAccount(s.ctx, feederAcc)

	err := s.app.OracleKeeper.ValidateFeeder(s.ctx, feederAddr, valAddr)
	s.Require().Error(err)

	_, err = s.queryClient.FeederDelegation(s.ctx.Context(), &types.QueryFeederDelegation{
		ValidatorAddr: inactiveValidator,
	})
	s.Require().Error(err)

	s.app.OracleKeeper.SetFeederDelegation(s.ctx, valAddr, feederAddr)

	err = s.app.OracleKeeper.ValidateFeeder(s.ctx, feederAddr, valAddr)
	s.Require().NoError(err)

	res, err := s.queryClient.FeederDelegation(s.ctx.Context(), &types.QueryFeederDelegation{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(feederAddr.String(), res.FeederAddr)
}

func (s *IntegrationTestSuite) TestQuerier_MissCounter() {
	missCounter := uint64(rand.Intn(100))

	res, err := s.queryClient.MissCounter(s.ctx.Context(), &types.QueryMissCounter{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(res.MissCounter, uint64(0))

	s.app.OracleKeeper.SetMissCounter(s.ctx, valAddr, missCounter)

	res, err = s.queryClient.MissCounter(s.ctx.Context(), &types.QueryMissCounter{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(res.MissCounter, missCounter)
}

func (s *IntegrationTestSuite) TestQuerier_SlashWindow() {
	res, err := s.queryClient.SlashWindow(s.ctx.Context(), &types.QuerySlashWindow{})
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), res.WindowProgress)
}

func (s *IntegrationTestSuite) TestQuerier_AggregatePrevote() {
	prevote := types.AggregateExchangeRatePrevote{
		Hash:        "hash",
		Voter:       addr.String(),
		SubmitBlock: 0,
	}
	s.app.OracleKeeper.SetAggregateExchangeRatePrevote(s.ctx, valAddr, prevote)

	res, err := s.app.OracleKeeper.GetAggregateExchangeRatePrevote(s.ctx, valAddr)
	s.Require().NoError(err)
	s.Require().Equal(prevote, res)

	queryRes, err := s.queryClient.AggregatePrevote(s.ctx.Context(), &types.QueryAggregatePrevote{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(types.AggregateExchangeRatePrevote{
		Hash:        "hash",
		Voter:       addr.String(),
		SubmitBlock: 0,
	}, queryRes.AggregatePrevote)
}

func (s *IntegrationTestSuite) TestQuerier_AggregatePrevotes() {
	res, err := s.queryClient.AggregatePrevotes(s.ctx.Context(), &types.QueryAggregatePrevotes{})
	s.Require().Equal([]types.AggregateExchangeRatePrevote(nil), res.AggregatePrevotes)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestQuerier_AggregateVote() {
	var tuples types.ExchangeRateTuples
	tuples = append(tuples, types.ExchangeRateTuple{
		Denom:        appparams.DisplayDenom,
		ExchangeRate: sdk.ZeroDec(),
	})

	vote := types.AggregateExchangeRateVote{
		ExchangeRateTuples: tuples,
		Voter:              addr.String(),
	}
	s.app.OracleKeeper.SetAggregateExchangeRateVote(s.ctx, valAddr, vote)

	res, err := s.queryClient.AggregateVote(s.ctx.Context(), &types.QueryAggregateVote{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(types.AggregateExchangeRateVote{
		ExchangeRateTuples: tuples,
		Voter:              addr.String(),
	}, res.AggregateVote)
}

func (s *IntegrationTestSuite) TestQuerier_AggregateVotes() {
	res, err := s.queryClient.AggregateVotes(s.ctx.Context(), &types.QueryAggregateVotes{})
	s.Require().NoError(err)
	s.Require().Equal([]types.AggregateExchangeRateVote(nil), res.AggregateVotes)
}

func (s *IntegrationTestSuite) TestQuerier_AggregateVoteInvalidExchangeRate() {
	res, err := s.queryClient.AggregateVote(s.ctx.Context(), &types.QueryAggregateVote{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().Nil(res)
	s.Require().ErrorContains(err, "no aggregate vote")
}

func (s *IntegrationTestSuite) TestQuerier_AggregatePrevoteInvalidExchangeRate() {
	res, err := s.queryClient.AggregatePrevote(s.ctx.Context(), &types.QueryAggregatePrevote{
		ValidatorAddr: valAddr.String(),
	})
	s.Require().Nil(res)
	s.Require().ErrorContains(err, "no aggregate prevote")
}

func (s *IntegrationTestSuite) TestQuerier_Params() {
	res, err := s.queryClient.Params(s.ctx.Context(), &types.QueryParams{})
	s.Require().NoError(err)
	s.Require().Equal(types.DefaultGenesisState().Params, res.Params)
}

func (s *IntegrationTestSuite) TestQuerier_ExchangeRatesInvalidExchangeRate() {
	resExchangeRate, err := s.queryClient.ExchangeRates(s.ctx.Context(), &types.QueryExchangeRates{
		Denom: " ",
	})
	s.Require().Nil(resExchangeRate)
	s.Require().ErrorContains(err, "unknown denom")
}

func (s *IntegrationTestSuite) TestQuerier_AggregatePrevoteInvalidValAddr() {
	resExchangeRate, err := s.queryClient.AggregatePrevote(s.ctx.Context(), &types.QueryAggregatePrevote{
		ValidatorAddr: "valAddrInvalid",
	})
	s.Require().Nil(resExchangeRate)
	s.Require().ErrorContains(err, "decoding bech32 failed")
}

func (s *IntegrationTestSuite) TestQuerier_AggregatePrevotesAppendVotes() {
	s.app.OracleKeeper.SetAggregateExchangeRatePrevote(s.ctx, valAddr, types.NewAggregateExchangeRatePrevote(
		types.AggregateVoteHash{},
		valAddr,
		uint64(s.ctx.BlockHeight()),
	))

	_, err := s.queryClient.AggregatePrevotes(s.ctx.Context(), &types.QueryAggregatePrevotes{})
	s.Require().Nil(err)
}

func (s *IntegrationTestSuite) TestQuerier_AggregateVotesAppendVotes() {
	s.app.OracleKeeper.SetAggregateExchangeRateVote(s.ctx, valAddr, types.NewAggregateExchangeRateVote(
		types.DefaultGenesisState().ExchangeRates,
		valAddr,
	))

	_, err := s.queryClient.AggregateVotes(s.ctx.Context(), &types.QueryAggregateVotes{})
	s.Require().Nil(err)
}

func (s *IntegrationTestSuite) TestEmptyRequest() {
	q := keeper.NewQuerier(keeper.Keeper{})
	const emptyRequestErrorMsg = "empty request"

	resParams, err := q.Params(s.ctx.Context(), nil)
	s.Require().Nil(resParams)
	s.Require().ErrorContains(err, emptyRequestErrorMsg)

	resExchangeRate, err := q.ExchangeRates(s.ctx.Context(), nil)
	s.Require().Nil(resExchangeRate)
	s.Require().ErrorContains(err, emptyRequestErrorMsg)

	resActiveExchangeRates, err := q.ActiveExchangeRates(s.ctx.Context(), nil)
	s.Require().Nil(resActiveExchangeRates)
	s.Require().ErrorContains(err, emptyRequestErrorMsg)

	resFeederDelegation, err := q.FeederDelegation(s.ctx.Context(), nil)
	s.Require().Nil(resFeederDelegation)
	s.Require().ErrorContains(err, emptyRequestErrorMsg)

	resMissCounter, err := q.MissCounter(s.ctx.Context(), nil)
	s.Require().Nil(resMissCounter)
	s.Require().ErrorContains(err, emptyRequestErrorMsg)

	resAggregatePrevote, err := q.AggregatePrevote(s.ctx.Context(), nil)
	s.Require().Nil(resAggregatePrevote)
	s.Require().ErrorContains(err, emptyRequestErrorMsg)

	resAggregatePrevotes, err := q.AggregatePrevotes(s.ctx.Context(), nil)
	s.Require().Nil(resAggregatePrevotes)
	s.Require().ErrorContains(err, emptyRequestErrorMsg)

	resAggregateVote, err := q.AggregateVote(s.ctx.Context(), nil)
	s.Require().Nil(resAggregateVote)
	s.Require().ErrorContains(err, emptyRequestErrorMsg)

	resAggregateVotes, err := q.AggregateVotes(s.ctx.Context(), nil)
	s.Require().Nil(resAggregateVotes)
	s.Require().ErrorContains(err, emptyRequestErrorMsg)
}

func (s *IntegrationTestSuite) TestInvalidBechAddress() {
	q := keeper.NewQuerier(keeper.Keeper{})
	invalidAddressMsg := "empty address string is not allowed"

	resFeederDelegation, err := q.FeederDelegation(s.ctx.Context(), &types.QueryFeederDelegation{})
	s.Require().Nil(resFeederDelegation)
	s.Require().ErrorContains(err, invalidAddressMsg)

	resMissCounter, err := q.MissCounter(s.ctx.Context(), &types.QueryMissCounter{})
	s.Require().Nil(resMissCounter)
	s.Require().ErrorContains(err, invalidAddressMsg)

	resAggregatePrevote, err := q.AggregatePrevote(s.ctx.Context(), &types.QueryAggregatePrevote{})
	s.Require().Nil(resAggregatePrevote)
	s.Require().ErrorContains(err, invalidAddressMsg)

	resAggregateVote, err := q.AggregateVote(s.ctx.Context(), &types.QueryAggregateVote{})
	s.Require().Nil(resAggregateVote)
	s.Require().ErrorContains(err, invalidAddressMsg)
}

func (s *IntegrationTestSuite) TestQueryPriceTrackingLists() {
	s.SetupTest()

	res, err := s.queryClient.PriceTrackingLists(s.ctx.Context(), &types.QueryPriceTrackingLists{})
	s.Require().NoError(err)
	s.Require().NotNil(res)

	result := []string{"JUNO", "ATOM"} // default params

	s.Require().Equal(res.PriceTrakingLists, result)
}

func (s *IntegrationTestSuite) TestPriceHistoryAt() {
	s.SetupTest()
	timeNow := time.Now().UTC()

	phEntry := types.PriceHistoryEntry{
		Price:           sdk.OneDec(),
		VotePeriodCount: 10,
		PriceUpdateTime: timeNow,
	}

	s.app.OracleKeeper.SetPriceHistoryEntry(
		s.ctx,
		"JUNO",
		phEntry.PriceUpdateTime,
		phEntry.Price,
		phEntry.VotePeriodCount,
	)

	req := &types.QueryPriceHistoryAt{
		Denom: "JUNO",
		Time:  timeNow,
	}

	res, err := s.queryClient.PriceHistoryAt(s.ctx.Context(), req)
	s.Require().NoError(err)
	s.Require().Equal(phEntry, res.PriceHistoryEntry)

	req = &types.QueryPriceHistoryAt{
		Denom: "JUNO",
		Time:  timeNow.Add(time.Minute),
	}

	res, err = s.queryClient.PriceHistoryAt(s.ctx.Context(), req)
	s.Require().NoError(err)
	s.Require().Equal(phEntry, res.PriceHistoryEntry)
}

func (s *IntegrationTestSuite) TestAllPriceHistory() {
	s.SetupTest()
	timeNow := time.Now().UTC()

	phEntrys := []types.PriceHistoryEntry{
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 1,
			PriceUpdateTime: timeNow,
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 11,
			PriceUpdateTime: timeNow.Add(time.Minute),
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 12,
			PriceUpdateTime: timeNow.Add(2 * time.Minute),
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 13,
			PriceUpdateTime: timeNow.Add(3 * time.Minute),
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 14,
			PriceUpdateTime: timeNow.Add(4 * time.Minute),
		},
	}

	// Set price history
	for _, phEntry := range phEntrys {
		s.app.OracleKeeper.SetPriceHistoryEntry(
			s.ctx,
			"JUNO",
			phEntry.PriceUpdateTime,
			phEntry.Price,
			phEntry.VotePeriodCount,
		)
	}

	// Get price history
	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllPriceHistory {
		return &types.QueryAllPriceHistory{
			Denom: "JUNO",
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}

	s.Run("ByOffset", func() {
		step := 2
		goCtx := sdk.WrapSDKContext(s.ctx)
		for i := 0; i < len(phEntrys); i += step {
			resp, err := s.queryClient.AllPriceHistory(goCtx, request(nil, uint64(i), uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.PriceHistoryEntrys), step)
			s.Require().Subset(nullify.Fill(phEntrys), nullify.Fill(resp.PriceHistoryEntrys))
		}
	})
	s.Run("ByKey", func() {
		step := 2
		var next []byte
		goCtx := sdk.WrapSDKContext(s.ctx)
		for i := 0; i < len(phEntrys); i += step {
			resp, err := s.queryClient.AllPriceHistory(goCtx, request(next, 0, uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.PriceHistoryEntrys), step)
			s.Require().Subset(nullify.Fill(phEntrys), nullify.Fill(resp.PriceHistoryEntrys))
			next = resp.Pagination.NextKey
		}
	})
	s.Run("Total", func() {
		goCtx := sdk.WrapSDKContext(s.ctx)
		resp, err := s.queryClient.AllPriceHistory(goCtx, request(nil, 0, 0, true))
		s.Require().NoError(err)
		s.Require().Equal(len(phEntrys), int(resp.Pagination.Total))
		s.Require().ElementsMatch(nullify.Fill(phEntrys), nullify.Fill(resp.PriceHistoryEntrys))
	})
}

func (s *IntegrationTestSuite) TestTwapPrice() {
	s.SetupTest()
	timeNow := time.Now().UTC()

	phEntrys := []types.PriceHistoryEntry{
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 10,
			PriceUpdateTime: timeNow,
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 11,
			PriceUpdateTime: timeNow.Add(time.Minute),
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 12,
			PriceUpdateTime: timeNow.Add(2 * time.Minute),
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 13,
			PriceUpdateTime: timeNow.Add(3 * time.Minute),
		},
		{
			Price:           sdk.OneDec(),
			VotePeriodCount: 14,
			PriceUpdateTime: timeNow.Add(4 * time.Minute),
		},
	}

	// Set price history
	for _, phEntry := range phEntrys {
		s.app.OracleKeeper.SetPriceHistoryEntry(
			s.ctx,
			"JUNO",
			phEntry.PriceUpdateTime,
			phEntry.Price,
			phEntry.VotePeriodCount,
		)
	}

	for _, tc := range []struct {
		desc      string
		req       *types.QueryTwapBetween
		res       *types.QueryTwapBetweenRespone
		shouldErr bool
	}{
		{
			desc: "Success",
			req: &types.QueryTwapBetween{
				Denom:     "JUNO",
				StartTime: timeNow,
				EndTime:   timeNow.Add(4 * time.Minute),
			},
			res: &types.QueryTwapBetweenRespone{
				TwapPrice: sdk.NewDecCoinFromDec("JUNO", sdk.OneDec()),
			},
			shouldErr: false,
		},
		{
			desc: "Success",
			req: &types.QueryTwapBetween{
				Denom:     "JUNO",
				StartTime: timeNow.Add(30 * time.Second),
				EndTime:   timeNow.Add(4 * time.Minute),
			},
			res: &types.QueryTwapBetweenRespone{
				TwapPrice: sdk.NewDecCoinFromDec("JUNO", sdk.OneDec()),
			},
			shouldErr: false,
		},
		{
			desc: "Error - Start time before first entry",
			req: &types.QueryTwapBetween{
				Denom:     "JUNO",
				StartTime: timeNow.Add(-time.Minute),
				EndTime:   timeNow.Add(4 * time.Minute),
			},
			shouldErr: true,
		},
		{
			desc: "Error - End time before start time",
			req: &types.QueryTwapBetween{
				Denom:     "JUNO",
				StartTime: timeNow.Add(3 * time.Minute),
				EndTime:   timeNow.Add(2 * time.Minute),
			},
			shouldErr: true,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			if !tc.shouldErr {
				res, err := s.queryClient.TwapPrice(s.ctx.Context(), tc.req)
				s.Require().NoError(err)
				s.Require().Equal(tc.res, res)
			} else {
				_, err := s.queryClient.TwapPrice(s.ctx.Context(), tc.req)
				s.Require().Error(err)
			}
		})
	}
}
