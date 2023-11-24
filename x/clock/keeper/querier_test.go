package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v18/x/clock/types"
)

func (s *IntegrationTestSuite) TestClockQueryParams() {
	// _, _, addr := testdata.KeyTestPubAddr()
	// _, _, addr2 := testdata.KeyTestPubAddr()

	defaultParams := types.DefaultParams()

	for _, tc := range []struct {
		desc     string
		Expected types.Params
	}{
		{
			desc: "On empty",
			Expected: types.Params{
				ContractGasLimit: defaultParams.ContractGasLimit,
			},
		},
		{
			desc: "On 1 address",
			Expected: types.Params{
				ContractGasLimit: defaultParams.ContractGasLimit,
			},
		},
		{
			desc: "On 2 Unique",
			Expected: types.Params{
				ContractGasLimit: defaultParams.ContractGasLimit,
			},
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			// Set the params to what is expected, then query and ensure the query is the same
			err := s.app.AppKeepers.ClockKeeper.SetParams(s.ctx, tc.Expected)
			s.Require().NoError(err)

			// Contracts check
			goCtx := sdk.WrapSDKContext(s.ctx)
			resp, err := s.queryClient.ClockContracts(goCtx, &types.QueryClockContracts{})
			s.Require().NoError(err)
			s.Require().NotNil(resp)

			// All Params Check
			resp2, err := s.queryClient.Params(goCtx, &types.QueryParamsRequest{})
			s.Require().NoError(err)
			s.Require().Equal(tc.Expected, *resp2.Params)
		})
	}
}
