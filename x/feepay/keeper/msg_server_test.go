package keeper_test

import (
	_ "embed"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/CosmosContracts/juno/v17/x/feepay/types"
)

func (s *IntegrationTestSuite) TestRegisterFeePayContract() {
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))
	_ = s.FundAccount(s.ctx, admin, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	noAdminContractAddress := s.InstantiateContract(sender.String(), "")
	withAdminContractAddress := s.InstantiateContract(sender.String(), admin.String())
	tContract := s.InstantiateContract(sender.String(), admin.String())

	for _, tc := range []struct {
		desc            string
		contractAddress string
		deployerAddress string
		shouldErr       bool
	}{
		{
			desc:            "Success - Creator",
			contractAddress: noAdminContractAddress,
			deployerAddress: sender.String(),
			shouldErr:       false,
		},
		{
			desc:            "Success - Admin",
			contractAddress: withAdminContractAddress,
			deployerAddress: admin.String(),
			shouldErr:       false,
		},
		{
			desc:            "Error - contract already registered",
			contractAddress: withAdminContractAddress,
			deployerAddress: admin.String(),
			shouldErr:       true,
		},
		{
			desc:            "Error - Invalid deployer",
			contractAddress: tContract,
			deployerAddress: "Invalid",
			shouldErr:       true,
		},
		{
			desc:            "Error - Invalid contract",
			contractAddress: "Invalid",
			deployerAddress: admin.String(),
			shouldErr:       true,
		},
	} {
		tc := tc

		s.Run(tc.desc, func() {

			// TODO: test setting balances & wallet limit work in another test with the querier.
			err := s.app.AppKeepers.FeePayKeeper.RegisterContract(s.ctx, &types.MsgRegisterFeePayContract{
				SenderAddress: tc.deployerAddress,
				FeePayContract: &types.FeePayContract{
					ContractAddress: tc.contractAddress,
					WalletLimit:     1,
				},
			})

			if tc.shouldErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestUnRegisterFeePayContract() {
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))
	_ = s.FundAccount(s.ctx, admin, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	contract := s.InstantiateContract(sender.String(), "")

	err := s.app.AppKeepers.FeePayKeeper.RegisterContract(s.ctx, &types.MsgRegisterFeePayContract{
		SenderAddress: sender.String(),
		FeePayContract: &types.FeePayContract{
			ContractAddress: contract,
			WalletLimit:     1,
		},
	})
	s.Require().NoError(err)

	for _, tc := range []struct {
		desc            string
		contractAddress string
		deployerAddress string
		shouldErr       bool
	}{
		{
			desc:            "Fail - invalid address",
			contractAddress: contract,
			deployerAddress: "Invalid",
			shouldErr:       true,
		},
		// TODO: non creator, non admin, etc
		{
			desc:            "Success - unregister",
			contractAddress: contract,
			deployerAddress: sender.String(),
			shouldErr:       false,
		},
	} {
		tc := tc

		s.Run(tc.desc, func() {
			err := s.app.AppKeepers.FeePayKeeper.UnregisterContract(s.ctx, &types.MsgUnregisterFeePayContract{
				SenderAddress:   tc.deployerAddress,
				ContractAddress: tc.contractAddress,
			})

			if tc.shouldErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TODO: FundFeePayContract, UpdateFeePayContractWalletLimit
// UpdateParams? (Prob better to do in e2e)
// querier test
// genesis_test
// E2E test with interchaintest (both end of week) also handle the ante test. Fees etc.

// ---

// OLD Examples from feeshare
// func (s *IntegrationTestSuite) TestRegisterFeeShare() {
// 	_, _, sender := testdata.KeyTestPubAddr()
// 	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

// 	gov := s.accountKeeper.GetModuleAddress(govtypes.ModuleName).String()
// 	govContract := s.InstantiateContract(sender.String(), gov)

// 	contractAddress := s.InstantiateContract(sender.String(), "")
// 	contractAddress2 := s.InstantiateContract(contractAddress, contractAddress)

// 	DAODAO := s.InstantiateContract(sender.String(), "")
// 	subContract := s.InstantiateContract(DAODAO, DAODAO)

// 	_, _, withdrawer := testdata.KeyTestPubAddr()

// 	for _, tc := range []struct {
// 		desc      string
// 		msg       *types.MsgRegisterFeeShare
// 		resp      *types.MsgRegisterFeeShareResponse
// 		shouldErr bool
// 	}{
// 		{
// 			desc: "Invalid contract address",
// 			msg: &types.MsgRegisterFeeShare{
// 				ContractAddress:   "Invalid",
// 				DeployerAddress:   sender.String(),
// 				WithdrawerAddress: withdrawer.String(),
// 			},
// 			resp:      &types.MsgRegisterFeeShareResponse{},
// 			shouldErr: true,
// 		},
// 		{
// 			desc: "Invalid deployer address",
// 			msg: &types.MsgRegisterFeeShare{
// 				ContractAddress:   contractAddress,
// 				DeployerAddress:   "Invalid",
// 				WithdrawerAddress: withdrawer.String(),
// 			},
// 			resp:      &types.MsgRegisterFeeShareResponse{},
// 			shouldErr: true,
// 		},
// 		{
// 			desc: "Invalid withdrawer address",
// 			msg: &types.MsgRegisterFeeShare{
// 				ContractAddress:   contractAddress,
// 				DeployerAddress:   sender.String(),
// 				WithdrawerAddress: "Invalid",
// 			},
// 			resp:      &types.MsgRegisterFeeShareResponse{},
// 			shouldErr: true,
// 		},
// 		{
// 			desc: "Success",
// 			msg: &types.MsgRegisterFeeShare{
// 				ContractAddress:   contractAddress,
// 				DeployerAddress:   sender.String(),
// 				WithdrawerAddress: withdrawer.String(),
// 			},
// 			resp:      &types.MsgRegisterFeeShareResponse{},
// 			shouldErr: false,
// 		},
// 		{
// 			desc: "Invalid withdraw address for factory contract",
// 			msg: &types.MsgRegisterFeeShare{
// 				ContractAddress:   contractAddress2,
// 				DeployerAddress:   sender.String(),
// 				WithdrawerAddress: sender.String(),
// 			},
// 			resp:      &types.MsgRegisterFeeShareResponse{},
// 			shouldErr: true,
// 		},
// 		{
// 			desc: "Success register factory contract to itself",
// 			msg: &types.MsgRegisterFeeShare{
// 				ContractAddress:   contractAddress2,
// 				DeployerAddress:   sender.String(),
// 				WithdrawerAddress: contractAddress2,
// 			},
// 			resp:      &types.MsgRegisterFeeShareResponse{},
// 			shouldErr: false,
// 		},
// 		{
// 			desc: "Invalid register gov contract withdraw address",
// 			msg: &types.MsgRegisterFeeShare{
// 				ContractAddress:   govContract,
// 				DeployerAddress:   sender.String(),
// 				WithdrawerAddress: sender.String(),
// 			},
// 			resp:      &types.MsgRegisterFeeShareResponse{},
// 			shouldErr: true,
// 		},
// 		{
// 			desc: "Success register gov contract withdraw address to self",
// 			msg: &types.MsgRegisterFeeShare{
// 				ContractAddress:   govContract,
// 				DeployerAddress:   sender.String(),
// 				WithdrawerAddress: govContract,
// 			},
// 			resp:      &types.MsgRegisterFeeShareResponse{},
// 			shouldErr: false,
// 		},
// 		{
// 			desc: "Success register contract from DAODAO contract as admin",
// 			msg: &types.MsgRegisterFeeShare{
// 				ContractAddress:   subContract,
// 				DeployerAddress:   DAODAO,
// 				WithdrawerAddress: DAODAO,
// 			},
// 			resp:      &types.MsgRegisterFeeShareResponse{},
// 			shouldErr: false,
// 		},
// 	} {
// 		tc := tc
// 		s.Run(tc.desc, func() {
// 			goCtx := sdk.WrapSDKContext(s.ctx)
// 			if !tc.shouldErr {
// 				resp, err := s.feeShareMsgServer.RegisterFeeShare(goCtx, tc.msg)
// 				s.Require().NoError(err)
// 				s.Require().Equal(resp, tc.resp)
// 			} else {
// 				resp, err := s.feeShareMsgServer.RegisterFeeShare(goCtx, tc.msg)
// 				s.Require().Error(err)
// 				s.Require().Nil(resp)
// 			}
// 		})
// 	}
// }

// func (s *IntegrationTestSuite) TestUpdateFeeShare() {
// 	_, _, sender := testdata.KeyTestPubAddr()
// 	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

// 	contractAddress := s.InstantiateContract(sender.String(), "")
// 	_, _, withdrawer := testdata.KeyTestPubAddr()

// 	contractAddressNoRegisFeeShare := s.InstantiateContract(sender.String(), "")
// 	s.Require().NotEqual(contractAddress, contractAddressNoRegisFeeShare)

// 	// RegsisFeeShare
// 	goCtx := sdk.WrapSDKContext(s.ctx)
// 	msg := &types.MsgRegisterFeeShare{
// 		ContractAddress:   contractAddress,
// 		DeployerAddress:   sender.String(),
// 		WithdrawerAddress: withdrawer.String(),
// 	}
// 	_, err := s.feeShareMsgServer.RegisterFeeShare(goCtx, msg)
// 	s.Require().NoError(err)
// 	_, _, newWithdrawer := testdata.KeyTestPubAddr()
// 	s.Require().NotEqual(withdrawer, newWithdrawer)

// 	for _, tc := range []struct {
// 		desc      string
// 		msg       *types.MsgUpdateFeeShare
// 		resp      *types.MsgCancelFeeShareResponse
// 		shouldErr bool
// 	}{
// 		{
// 			desc: "Invalid - contract not regis",
// 			msg: &types.MsgUpdateFeeShare{
// 				ContractAddress:   contractAddressNoRegisFeeShare,
// 				DeployerAddress:   sender.String(),
// 				WithdrawerAddress: newWithdrawer.String(),
// 			},
// 			resp:      nil,
// 			shouldErr: true,
// 		},
// 		{
// 			desc: "Invalid - Invalid DeployerAddress",
// 			msg: &types.MsgUpdateFeeShare{
// 				ContractAddress:   contractAddress,
// 				DeployerAddress:   "Invalid",
// 				WithdrawerAddress: newWithdrawer.String(),
// 			},
// 			resp:      nil,
// 			shouldErr: true,
// 		},
// 		{
// 			desc: "Invalid - Invalid WithdrawerAddress",
// 			msg: &types.MsgUpdateFeeShare{
// 				ContractAddress:   contractAddress,
// 				DeployerAddress:   sender.String(),
// 				WithdrawerAddress: "Invalid",
// 			},
// 			resp:      nil,
// 			shouldErr: true,
// 		},
// 		{
// 			desc: "Invalid - Invalid WithdrawerAddress not change",
// 			msg: &types.MsgUpdateFeeShare{
// 				ContractAddress:   contractAddress,
// 				DeployerAddress:   sender.String(),
// 				WithdrawerAddress: withdrawer.String(),
// 			},
// 			resp:      nil,
// 			shouldErr: true,
// 		},
// 		{
// 			desc: "Success",
// 			msg: &types.MsgUpdateFeeShare{
// 				ContractAddress:   contractAddress,
// 				DeployerAddress:   sender.String(),
// 				WithdrawerAddress: newWithdrawer.String(),
// 			},
// 			resp:      &types.MsgCancelFeeShareResponse{},
// 			shouldErr: false,
// 		},
// 	} {
// 		tc := tc
// 		s.Run(tc.desc, func() {
// 			goCtx := sdk.WrapSDKContext(s.ctx)
// 			if !tc.shouldErr {
// 				_, err := s.feeShareMsgServer.UpdateFeeShare(goCtx, tc.msg)
// 				s.Require().NoError(err)
// 			} else {
// 				resp, err := s.feeShareMsgServer.UpdateFeeShare(goCtx, tc.msg)
// 				s.Require().Error(err)
// 				s.Require().Nil(resp)
// 			}
// 		})
// 	}
// }

// func (s *IntegrationTestSuite) TestCancelFeeShare() {
// 	_, _, sender := testdata.KeyTestPubAddr()
// 	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

// 	contractAddress := s.InstantiateContract(sender.String(), "")
// 	_, _, withdrawer := testdata.KeyTestPubAddr()

// 	// RegsisFeeShare
// 	goCtx := sdk.WrapSDKContext(s.ctx)
// 	msg := &types.MsgRegisterFeeShare{
// 		ContractAddress:   contractAddress,
// 		DeployerAddress:   sender.String(),
// 		WithdrawerAddress: withdrawer.String(),
// 	}
// 	_, err := s.feeShareMsgServer.RegisterFeeShare(goCtx, msg)
// 	s.Require().NoError(err)

// 	for _, tc := range []struct {
// 		desc      string
// 		msg       *types.MsgCancelFeeShare
// 		resp      *types.MsgCancelFeeShareResponse
// 		shouldErr bool
// 	}{
// 		{
// 			desc: "Invalid - contract Address",
// 			msg: &types.MsgCancelFeeShare{
// 				ContractAddress: "Invalid",
// 				DeployerAddress: sender.String(),
// 			},
// 			resp:      nil,
// 			shouldErr: true,
// 		},
// 		{
// 			desc: "Invalid - deployer Address",
// 			msg: &types.MsgCancelFeeShare{
// 				ContractAddress: contractAddress,
// 				DeployerAddress: "Invalid",
// 			},
// 			resp:      nil,
// 			shouldErr: true,
// 		},
// 		{
// 			desc: "Success",
// 			msg: &types.MsgCancelFeeShare{
// 				ContractAddress: contractAddress,
// 				DeployerAddress: sender.String(),
// 			},
// 			resp:      &types.MsgCancelFeeShareResponse{},
// 			shouldErr: false,
// 		},
// 	} {
// 		tc := tc
// 		s.Run(tc.desc, func() {
// 			goCtx := sdk.WrapSDKContext(s.ctx)
// 			if !tc.shouldErr {
// 				resp, err := s.feeShareMsgServer.CancelFeeShare(goCtx, tc.msg)
// 				s.Require().NoError(err)
// 				s.Require().Equal(resp, tc.resp)
// 			} else {
// 				resp, err := s.feeShareMsgServer.CancelFeeShare(goCtx, tc.msg)
// 				s.Require().Error(err)
// 				s.Require().Equal(resp, tc.resp)
// 			}
// 		})
// 	}
// }
