package keeper_test

import (
	_ "embed"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v18/x/clock/types"
)

// Test register clock contract.
func (s *IntegrationTestSuite) TestRegisterClockContract() {
	_, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, addr, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	// Store code
	s.StoreCode()
	contractAddress := s.InstantiateContract(addr.String(), "")

	for _, tc := range []struct {
		desc     string
		sender   string
		contract string
		isJailed bool
		success  bool
	}{
		{
			desc:     "Success - Register Contract",
			sender:   addr.String(),
			contract: contractAddress,
			success:  true,
		},
		{
			desc:     "Error - Invalid Sender",
			sender:   addr2.String(),
			contract: contractAddress,
			success:  false,
		},
		{
			desc:     "Fail - Invalid Contract Address",
			sender:   addr.String(),
			contract: "Invalid",
			success:  false,
		},
		{
			desc:     "Fail - Invalid Sender Address",
			sender:   "Invalid",
			contract: contractAddress,
			success:  false,
		},
		{
			desc:     "Fail - Contract Already Jailed",
			sender:   addr.String(),
			contract: contractAddress,
			isJailed: true,
			success:  false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			// Set params
			params := types.DefaultParams()
			err := s.app.AppKeepers.ClockKeeper.SetParams(s.ctx, params)
			s.Require().NoError(err)

			// Jail contract if needed
			if tc.isJailed {
				err := s.app.AppKeepers.ClockKeeper.JailContract(s.ctx, contractAddress)
				s.Require().NoError(err)
			}

			// Try to register contract
			res, err := s.clockMsgServer.RegisterClockContract(s.ctx, &types.MsgRegisterClockContract{
				SenderAddress:   tc.sender,
				ContractAddress: tc.contract,
			})

			if !tc.success {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(res, &types.MsgRegisterClockContractResponse{})
			}
		})
	}
}

// Test standard unregistration of clock contracts.
func (s *IntegrationTestSuite) TestUnregisterClockContract() {
	_, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, addr, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	s.StoreCode()
	contractAddress := s.InstantiateContract(addr.String(), "")

	for _, tc := range []struct {
		desc     string
		sender   string
		contract string
		success  bool
	}{
		{
			desc:     "Success - Unregister Contract",
			sender:   addr.String(),
			contract: contractAddress,
			success:  true,
		},
		{
			desc:     "Error - Invalid Sender",
			sender:   addr2.String(),
			contract: contractAddress,
			success:  false,
		},
		{
			desc:     "Fail - Invalid Contract Address",
			sender:   addr.String(),
			contract: "Invalid",
			success:  false,
		},
		{
			desc:     "Fail - Invalid Sender Address",
			sender:   "Invalid",
			contract: contractAddress,
			success:  false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			// Ensure contract is unregistered, ignore unregister error
			_ = s.app.AppKeepers.ClockKeeper.UnregisterContract(s.ctx, addr.String(), contractAddress)
			s.RegisterClockContract(addr.String(), contractAddress)

			// Set params
			params := types.DefaultParams()
			err := s.app.AppKeepers.ClockKeeper.SetParams(s.ctx, params)
			s.Require().NoError(err)

			// Try to register all contracts
			res, err := s.clockMsgServer.UnregisterClockContract(s.ctx, &types.MsgUnregisterClockContract{
				SenderAddress:   tc.sender,
				ContractAddress: tc.contract,
			})

			if !tc.success {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(res, &types.MsgUnregisterClockContractResponse{})
			}
		})
	}
}

// Test duplicate register/unregister clock contracts.
func (s *IntegrationTestSuite) TestDuplicateRegistrationChecks() {
	_, _, addr := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, addr, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	s.StoreCode()
	contractAddress := s.InstantiateContract(addr.String(), "")

	// Test double register, first succeed, second fail
	_, err := s.clockMsgServer.RegisterClockContract(s.ctx, &types.MsgRegisterClockContract{
		SenderAddress:   addr.String(),
		ContractAddress: contractAddress,
	})
	s.Require().NoError(err)

	_, err = s.clockMsgServer.RegisterClockContract(s.ctx, &types.MsgRegisterClockContract{
		SenderAddress:   addr.String(),
		ContractAddress: contractAddress,
	})
	s.Require().Error(err)

	// Test double unregister, first succeed, second fail
	_, err = s.clockMsgServer.UnregisterClockContract(s.ctx, &types.MsgUnregisterClockContract{
		SenderAddress:   addr.String(),
		ContractAddress: contractAddress,
	})
	s.Require().NoError(err)

	_, err = s.clockMsgServer.UnregisterClockContract(s.ctx, &types.MsgUnregisterClockContract{
		SenderAddress:   addr.String(),
		ContractAddress: contractAddress,
	})
	s.Require().Error(err)
}

// Test unjailing clock contracts.
func (s *IntegrationTestSuite) TestUnjailClockContract() {
	_, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, addr, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	s.StoreCode()
	contractAddress := s.InstantiateContract(addr.String(), "")
	s.RegisterClockContract(addr.String(), contractAddress)

	for _, tc := range []struct {
		desc     string
		sender   string
		contract string
		unjail   bool
		success  bool
	}{
		{
			desc:     "Success - Unjail Contract",
			sender:   addr.String(),
			contract: contractAddress,
			success:  true,
		},
		{
			desc:     "Error - Invalid Sender",
			sender:   addr2.String(),
			contract: contractAddress,
			success:  false,
		},
		{
			desc:     "Fail - Invalid Contract Address",
			sender:   addr.String(),
			contract: "Invalid",
			success:  false,
		},
		{
			desc:     "Fail - Invalid Sender Address",
			sender:   "Invalid",
			contract: contractAddress,
			success:  false,
		},
		{
			desc:     "Fail - Contract Not Jailed",
			sender:   addr.String(),
			contract: contractAddress,
			unjail:   true,
			success:  false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			_ = s.app.AppKeepers.ClockKeeper.UnjailContract(s.ctx, addr.String(), contractAddress)
			s.JailClockContract(contractAddress)

			// Unjail contract if needed
			if tc.unjail {
				s.UnjailClockContract(addr.String(), contractAddress)
			}

			// Set params
			params := types.DefaultParams()
			err := s.app.AppKeepers.ClockKeeper.SetParams(s.ctx, params)
			s.Require().NoError(err)

			// Try to register all contracts
			res, err := s.clockMsgServer.UnjailClockContract(s.ctx, &types.MsgUnjailClockContract{
				SenderAddress:   tc.sender,
				ContractAddress: tc.contract,
			})

			if !tc.success {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(res, &types.MsgUnjailClockContractResponse{})
			}
		})
	}
}
