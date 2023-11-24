package keeper_test

import (
	_ "embed"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/CosmosContracts/juno/v18/x/clock/types"
)

func (s *IntegrationTestSuite) TestUpdateClockParams() {
	_, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()

	for _, tc := range []struct {
		desc              string
		isEnabled         bool
		ContractAddresses []string
		success           bool
	}{
		{
			desc:              "Success - Valid on",
			isEnabled:         true,
			ContractAddresses: []string{},
			success:           true,
		},
		{
			desc:              "Success - Valid off",
			isEnabled:         false,
			ContractAddresses: []string{},
			success:           true,
		},
		{
			desc:              "Success - On and 1 allowed address",
			isEnabled:         true,
			ContractAddresses: []string{addr.String()},
			success:           true,
		},
		{
			desc:              "Fail - On and 2 duplicate addresses",
			isEnabled:         true,
			ContractAddresses: []string{addr.String(), addr.String()},
			success:           false,
		},
		{
			desc:              "Success - On and 2 unique",
			isEnabled:         true,
			ContractAddresses: []string{addr.String(), addr2.String()},
			success:           true,
		},
		{
			desc:              "Success - On and 2 duplicate 1 unique",
			isEnabled:         true,
			ContractAddresses: []string{addr.String(), addr2.String(), addr.String()},
			success:           false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			params := types.DefaultParams()
			// params.ContractAddresses = tc.ContractAddresses

			err := s.app.AppKeepers.ClockKeeper.SetParams(s.ctx, params)

			if !tc.success {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
