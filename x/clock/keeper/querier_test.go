package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v17/x/clock/types"
)

func (s *IntegrationTestSuite) TestClockQueryParams() {
	_, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()

	for _, tc := range []struct {
		desc     string
		Expected types.Params
	}{
		{
			desc: "On empty",
			Expected: types.Params{
				ContractAddresses: []string(nil),
			},
		},
		{
			desc: "On 1 address",
			Expected: types.Params{
				ContractAddresses: []string{addr.String()},
			},
		},
		{
			desc: "On 2 Unique",
			Expected: types.Params{
				ContractAddresses: []string{addr.String(), addr2.String()},
			},
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			// Set the params to what is expected, then query and ensure the query is the same
			err := s.app.AppKeepers.ClockKeeper.SetParams(s.ctx, tc.Expected)
			s.Require().NoError(err)

			goCtx := sdk.WrapSDKContext(s.ctx)
			resp, err := s.queryClient.ClockContracts(goCtx, &types.QueryClockContracts{})
			s.Require().NoError(err)
			s.Require().Equal(tc.Expected.ContractAddresses, resp.ContractAddresses)
		})
	}
}
