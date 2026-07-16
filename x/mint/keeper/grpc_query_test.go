package keeper_test

import (
	"github.com/CosmosContracts/juno/v30/x/mint/types"
)

func (s *KeeperTestSuite) TestGRPCParams() {
	s.SetupTest()
	params, err := s.queryClient.Params(s.Ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	kparams, err := s.mintKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(params.Params, kparams)

	inflation, err := s.queryClient.Inflation(s.Ctx, &types.QueryInflationRequest{})
	s.Require().NoError(err)
	minter, err := s.mintKeeper.GetMinter(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(inflation.Inflation, minter.Inflation)

	annualProvisions, err := s.queryClient.AnnualProvisions(s.Ctx, &types.QueryAnnualProvisionsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(annualProvisions.AnnualProvisions, minter.AnnualProvisions)
}
