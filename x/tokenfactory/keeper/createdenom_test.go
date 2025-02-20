package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CosmosContracts/juno/v28/x/tokenfactory/types"
)

func (s *KeeperTestSuite) TestMsgCreateDenom() {
	var (
		tokenFactoryKeeper = s.App.AppKeepers.TokenFactoryKeeper
		bankKeeper         = s.App.AppKeepers.BankKeeper
		denomCreationFee   = sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(1000000)))
	)

	// Set the denom creation fee. It is currently turned off in favor
	// of gas charge by default.
	params := s.App.AppKeepers.TokenFactoryKeeper.GetParams(s.Ctx)
	params.DenomCreationFee = denomCreationFee
	err := s.App.AppKeepers.TokenFactoryKeeper.SetParams(s.Ctx, params)
	s.Require().NoError(err)

	// Fund denom creation fee for every execution of MsgCreateDenom.
	s.FundAcc(s.TestAccs[0], denomCreationFee)
	s.FundAcc(s.TestAccs[0], denomCreationFee)
	s.FundAcc(s.TestAccs[1], denomCreationFee)

	// Get balance of acc 0 before creating a denom
	preCreateBalance := bankKeeper.GetBalance(s.Ctx, s.TestAccs[0], denomCreationFee[0].Denom)

	// Creating a denom should work
	res, err := s.msgServer.CreateDenom(s.Ctx, &types.MsgCreateDenom{
		Sender:   s.TestAccs[0].String(),
		Subdenom: "bitcoin",
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(res.GetNewTokenDenom())

	// Make sure that the admin is set correctly
	queryRes, err := s.queryClient.DenomAuthorityMetadata(s.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
		Denom: res.GetNewTokenDenom(),
	})
	s.Require().NoError(err)
	s.Require().Equal(s.TestAccs[0].String(), queryRes.AuthorityMetadata.Admin)

	// Make sure that creation fee was deducted
	postCreateBalance := bankKeeper.GetBalance(s.Ctx, s.TestAccs[0], tokenFactoryKeeper.GetParams(s.Ctx).DenomCreationFee[0].Denom)
	s.Require().True(preCreateBalance.Sub(postCreateBalance).IsEqual(denomCreationFee[0]))

	// Make sure that a second version of the same denom can't be recreated
	_, err = s.msgServer.CreateDenom(s.Ctx, &types.MsgCreateDenom{
		Sender:   s.TestAccs[0].String(),
		Subdenom: "bitcoin",
	})
	s.Require().Error(err)

	// Creating a second denom should work
	res, err = s.msgServer.CreateDenom(s.Ctx, &types.MsgCreateDenom{
		Sender:   s.TestAccs[0].String(),
		Subdenom: "litecoin",
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(res.GetNewTokenDenom())

	// Try querying all the denoms created by s.TestAccs[0]
	queryRes2, err := s.queryClient.DenomsFromCreator(s.Ctx.Context(), &types.QueryDenomsFromCreatorRequest{
		Creator: s.TestAccs[0].String(),
	})
	s.Require().NoError(err)
	s.Require().Len(queryRes2.Denoms, 2)

	// Make sure that a second account can create a denom with the same subdenom
	res, err = s.msgServer.CreateDenom(s.Ctx, &types.MsgCreateDenom{
		Sender:   s.TestAccs[1].String(),
		Subdenom: "bitcoin",
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(res.GetNewTokenDenom())

	// Make sure that an address with a "/" in it can't create denoms
	_, err = s.msgServer.CreateDenom(s.Ctx, &types.MsgCreateDenom{
		Sender:   "osmosis.eth/creator",
		Subdenom: "bitcoin",
	})
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestCreateDenom() {
	var (
		primaryDenom            = "ujuno"
		secondaryDenom          = "usecond"
		defaultDenomCreationFee = types.Params{DenomCreationFee: sdk.NewCoins(sdk.NewCoin(primaryDenom, sdkmath.NewInt(50000000)))}
		twoDenomCreationFee     = types.Params{DenomCreationFee: sdk.NewCoins(sdk.NewCoin(primaryDenom, sdkmath.NewInt(50000000)), sdk.NewCoin(secondaryDenom, sdkmath.NewInt(50000000)))}
		nilCreationFee          = types.Params{DenomCreationFee: nil}
		largeCreationFee        = types.Params{DenomCreationFee: sdk.NewCoins(sdk.NewCoin(primaryDenom, sdkmath.NewInt(5000000000)))}
	)

	for _, tc := range []struct {
		desc             string
		denomCreationFee types.Params
		setup            func()
		subdenom         string
		valid            bool
	}{
		{
			desc:             "subdenom too long",
			denomCreationFee: defaultDenomCreationFee,
			subdenom:         "assadsadsadasdasdsadsadsadsadsadsadsklkadaskkkdasdasedskhanhassyeunganassfnlksdflksafjlkasd",
			valid:            false,
		},
		{
			desc:             "subdenom and creator pair already exists",
			denomCreationFee: defaultDenomCreationFee,
			setup: func() {
				_, err := s.msgServer.CreateDenom(s.Ctx, &types.MsgCreateDenom{
					Sender:   s.TestAccs[0].String(),
					Subdenom: "bitcoin",
				})
				if err != nil {
					s.T().Error(err)
				}
				s.Require().NoError(err)
			},
			subdenom: "bitcoin",
			valid:    false,
		},
		{
			desc:             "success case: defaultDenomCreationFee",
			denomCreationFee: defaultDenomCreationFee,
			subdenom:         "evmos",
			valid:            true,
		},
		{
			desc:             "success case: twoDenomCreationFee",
			denomCreationFee: twoDenomCreationFee,
			subdenom:         "catcoin",
			valid:            true,
		},
		{
			desc:             "success case: nilCreationFee",
			denomCreationFee: nilCreationFee,
			subdenom:         "czcoin",
			valid:            true,
		},
		{
			desc:             "account doesn't have enough to pay for denom creation fee",
			denomCreationFee: largeCreationFee,
			subdenom:         "tooexpensive",
			valid:            false,
		},
	} {
		s.SetupTest()
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			if tc.setup != nil {
				tc.setup()
			}
			tokenFactoryKeeper := s.App.AppKeepers.TokenFactoryKeeper
			bankKeeper := s.App.AppKeepers.BankKeeper

			// Set denom creation fee in params
			s.FundAcc(s.TestAccs[0], defaultDenomCreationFee.DenomCreationFee)
			err := tokenFactoryKeeper.SetParams(s.Ctx, tc.denomCreationFee)
			s.Require().NoError(err)
			denomCreationFee := tokenFactoryKeeper.GetParams(s.Ctx).DenomCreationFee
			s.Require().Equal(tc.denomCreationFee.DenomCreationFee, denomCreationFee)

			// note balance, create a tokenfactory denom, then note balance again
			preCreateBalance := bankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			res, err := s.msgServer.CreateDenom(s.Ctx, &types.MsgCreateDenom{
				Sender:   s.TestAccs[0].String(),
				Subdenom: tc.subdenom,
			})
			postCreateBalance := bankKeeper.GetAllBalances(s.Ctx, s.TestAccs[0])
			if tc.valid {
				s.Require().NoError(err)
				s.Require().True(preCreateBalance.Sub(postCreateBalance...).Equal(denomCreationFee))

				// Make sure that the admin is set correctly
				queryRes, err := s.queryClient.DenomAuthorityMetadata(s.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
					Denom: res.GetNewTokenDenom(),
				})
				s.Require().NoError(err)
				s.Require().Equal(s.TestAccs[0].String(), queryRes.AuthorityMetadata.Admin)

				// Make sure that the denom metadata is initialized correctly
				metadata, found := bankKeeper.GetDenomMetaData(s.Ctx, res.GetNewTokenDenom())
				s.Require().True(found)
				s.Require().Equal(banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{{
						Denom:    res.GetNewTokenDenom(),
						Exponent: 0,
					}},
					Base:    res.GetNewTokenDenom(),
					Display: res.GetNewTokenDenom(),
					Name:    res.GetNewTokenDenom(),
					Symbol:  res.GetNewTokenDenom(),
				}, metadata)
			} else {
				s.Require().Error(err)
				// Ensure we don't charge if we expect an error
				s.Require().True(preCreateBalance.Equal(postCreateBalance))
			}
		})
	}
}
