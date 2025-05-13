package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v29/x/clock/types"
)

// Query Clock Params
func (s *KeeperTestSuite) TestQueryClockParams() {
	for _, tc := range []struct {
		desc   string
		params types.Params
	}{
		{
			desc:   "On default",
			params: types.DefaultParams(),
		},
		{
			desc: "On 500_000",
			params: types.Params{
				ContractGasLimit: 500_000,
			},
		},
		{
			desc: "On 1_000_000",
			params: types.Params{
				ContractGasLimit: 1_000_000,
			},
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			// Set params
			err := s.App.AppKeepers.ClockKeeper.SetParams(s.Ctx, tc.params)
			s.Require().NoError(err)

			// Query params
			resp, err := s.queryClient.Params(s.Ctx, &types.QueryParamsRequest{})

			// Response check
			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().Equal(tc.params, resp.Params)
		})
	}
}

// Query Clock Contracts
func (s *KeeperTestSuite) TestQueryClockContracts() {
	_, _, addr := testdata.KeyTestPubAddr()
	s.FundAcc(addr, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	s.StoreCode(clockContract)

	for _, tc := range []struct {
		desc      string
		contracts []string
	}{
		{
			desc:      "On empty",
			contracts: []string(nil),
		},
		{
			desc: "On Single",
			contracts: []string{
				s.InstantiateContract(addr.String(), "", clockContract),
			},
		},
		{
			desc: "On Multiple",
			contracts: []string{
				s.InstantiateContract(addr.String(), "", clockContract),
				s.InstantiateContract(addr.String(), "", clockContract),
				s.InstantiateContract(addr.String(), "", clockContract),
			},
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			// Loop through contracts & register
			for _, contract := range tc.contracts {
				s.RegisterClockContract(addr.String(), contract)
			}

			// Contracts check
			resp, err := s.queryClient.ClockContracts(s.Ctx, &types.QueryClockContractsRequest{})

			// Response check
			s.Require().NoError(err)
			s.Require().NotNil(resp)
			for _, contract := range resp.ClockContracts {
				s.Require().Contains(tc.contracts, contract.ContractAddress)
				s.Require().False(contract.IsJailed)
			}

			// Remove all contracts
			for _, contract := range tc.contracts {
				s.App.AppKeepers.ClockKeeper.RemoveContract(s.Ctx, contract)
			}
		})
	}
}

// Query Jailed Clock Contracts
func (s *KeeperTestSuite) TestQueryJailedClockContracts() {
	_, _, addr := testdata.KeyTestPubAddr()
	s.FundAcc(addr, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	s.StoreCode(clockContract)

	for _, tc := range []struct {
		desc      string
		contracts []string
	}{
		{
			desc:      "On empty",
			contracts: []string(nil),
		},
		{
			desc: "On Single",
			contracts: []string{
				s.InstantiateContract(addr.String(), "", clockContract),
			},
		},
		{
			desc: "On Multiple",
			contracts: []string{
				s.InstantiateContract(addr.String(), "", clockContract),
				s.InstantiateContract(addr.String(), "", clockContract),
				s.InstantiateContract(addr.String(), "", clockContract),
			},
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			// Loop through contracts & register
			for _, contract := range tc.contracts {
				s.RegisterClockContract(addr.String(), contract)
				s.JailClockContract(contract)
			}

			// Contracts check
			resp, err := s.queryClient.ClockContracts(s.Ctx, &types.QueryClockContractsRequest{})

			// Response check
			s.Require().NoError(err)
			s.Require().NotNil(resp)
			for _, contract := range resp.ClockContracts {
				s.Require().Contains(tc.contracts, contract.ContractAddress)
				s.Require().True(contract.IsJailed)
			}

			// Remove all contracts
			for _, contract := range tc.contracts {
				s.App.AppKeepers.ClockKeeper.RemoveContract(s.Ctx, contract)
			}
		})
	}
}

// Query Clock Contract
func (s *KeeperTestSuite) TestQueryClockContract() {
	_, _, addr := testdata.KeyTestPubAddr()
	s.FundAcc(addr, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	_, _, invalidAddr := testdata.KeyTestPubAddr()

	s.StoreCode(clockContract)

	unjailedContract := s.InstantiateContract(addr.String(), "", clockContract)
	_ = s.App.AppKeepers.ClockKeeper.SetClockContract(s.Ctx, types.ClockContract{
		ContractAddress: unjailedContract,
		IsJailed:        false,
	})

	jailedContract := s.InstantiateContract(addr.String(), "", clockContract)
	_ = s.App.AppKeepers.ClockKeeper.SetClockContract(s.Ctx, types.ClockContract{
		ContractAddress: jailedContract,
		IsJailed:        true,
	})

	for _, tc := range []struct {
		desc     string
		contract string
		isJailed bool
		success  bool
	}{
		{
			desc:     "On Unjailed",
			contract: unjailedContract,
			isJailed: false,
			success:  true,
		},
		{
			desc:     "On Jailed",
			contract: jailedContract,
			isJailed: true,
			success:  true,
		},
		{
			desc:     "Invalid Contract - Unjailed",
			contract: invalidAddr.String(),
			isJailed: false,
			success:  false,
		},
		{
			desc:     "Invalid Contract - Jailed",
			contract: invalidAddr.String(),
			isJailed: true,
			success:  false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			// Query contract
			resp, err := s.queryClient.ClockContract(s.Ctx, &types.QueryClockContractRequest{
				ContractAddress: tc.contract,
			})

			// Validate responses
			if tc.success {
				s.Require().NoError(err)
				s.Require().Equal(resp.ClockContract.ContractAddress, tc.contract)
				s.Require().Equal(resp.ClockContract.IsJailed, tc.isJailed)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
