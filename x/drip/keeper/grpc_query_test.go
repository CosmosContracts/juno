package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/CosmosContracts/juno/v28/x/drip/types"
)

func (s *KeeperTestSuite) TestDripQueryParams() {
	_, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()

	for _, tc := range []struct {
		desc     string
		Expected types.Params
	}{
		{
			desc: "On empty",
			Expected: types.Params{
				EnableDrip:       true,
				AllowedAddresses: []string(nil),
			},
		},
		{
			desc: "off empty",
			Expected: types.Params{
				EnableDrip:       false,
				AllowedAddresses: []string(nil),
			},
		},
		{
			desc: "On 1 address",
			Expected: types.Params{
				EnableDrip:       true,
				AllowedAddresses: []string{addr.String()},
			},
		},
		{
			desc: "On 2 Unique",
			Expected: types.Params{
				EnableDrip:       true,
				AllowedAddresses: []string{addr.String(), addr2.String()},
			},
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			// Set the params to what is expected, then query and ensure the query is the same
			err := s.App.AppKeepers.DripKeeper.SetParams(s.Ctx, tc.Expected)
			s.Require().NoError(err)

			resp, err := s.queryClient.Params(s.Ctx, &types.QueryParamsRequest{})
			s.Require().NoError(err)
			s.Require().Equal(tc.Expected, resp.Params)
		})
	}
}
