package keeper_test

import (
	_ "embed"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CosmosContracts/juno/v29/x/feeshare/types"
)

func (s *KeeperTestSuite) TestGetContractAdminOrCreatorAddress() {
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	s.FundAcc(admin, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	noAdminContractAddress := s.InstantiateContract(sender.String(), "", wasmContract)
	withAdminContractAddress := s.InstantiateContract(sender.String(), admin.String(), wasmContract)

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
			desc:            "Error - Invalid deployer",
			contractAddress: noAdminContractAddress,
			deployerAddress: "Invalid",
			shouldErr:       true,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			if !tc.shouldErr {
				_, err := s.App.AppKeepers.FeeShareKeeper.GetContractAdminOrCreatorAddress(s.Ctx, sdk.MustAccAddressFromBech32(tc.contractAddress), tc.deployerAddress)
				s.Require().NoError(err)
			} else {
				_, err := s.App.AppKeepers.FeeShareKeeper.GetContractAdminOrCreatorAddress(s.Ctx, sdk.MustAccAddressFromBech32(tc.contractAddress), tc.deployerAddress)
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestRegisterFeeShare() {
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	gov := s.accountKeeper.GetModuleAddress(govtypes.ModuleName).String()
	govContract := s.InstantiateContract(sender.String(), gov, wasmContract)

	contractAddress := s.InstantiateContract(sender.String(), "", wasmContract)
	contractAddress2 := s.InstantiateContract(contractAddress, contractAddress, wasmContract)

	daodao := s.InstantiateContract(sender.String(), "", wasmContract)
	subContract := s.InstantiateContract(daodao, daodao, wasmContract)

	_, _, withdrawer := testdata.KeyTestPubAddr()

	for _, tc := range []struct {
		desc      string
		msg       *types.MsgRegisterFeeShare
		resp      *types.MsgRegisterFeeShareResponse
		shouldErr bool
	}{
		{
			desc: "Invalid contract address",
			msg: &types.MsgRegisterFeeShare{
				ContractAddress:   "Invalid",
				DeployerAddress:   sender.String(),
				WithdrawerAddress: withdrawer.String(),
			},
			resp:      &types.MsgRegisterFeeShareResponse{},
			shouldErr: true,
		},
		{
			desc: "Invalid deployer address",
			msg: &types.MsgRegisterFeeShare{
				ContractAddress:   contractAddress,
				DeployerAddress:   "Invalid",
				WithdrawerAddress: withdrawer.String(),
			},
			resp:      &types.MsgRegisterFeeShareResponse{},
			shouldErr: true,
		},
		{
			desc: "Invalid withdrawer address",
			msg: &types.MsgRegisterFeeShare{
				ContractAddress:   contractAddress,
				DeployerAddress:   sender.String(),
				WithdrawerAddress: "Invalid",
			},
			resp:      &types.MsgRegisterFeeShareResponse{},
			shouldErr: true,
		},
		{
			desc: "Success",
			msg: &types.MsgRegisterFeeShare{
				ContractAddress:   contractAddress,
				DeployerAddress:   sender.String(),
				WithdrawerAddress: withdrawer.String(),
			},
			resp:      &types.MsgRegisterFeeShareResponse{},
			shouldErr: false,
		},
		{
			desc: "Invalid withdraw address for factory contract",
			msg: &types.MsgRegisterFeeShare{
				ContractAddress:   contractAddress2,
				DeployerAddress:   sender.String(),
				WithdrawerAddress: sender.String(),
			},
			resp:      &types.MsgRegisterFeeShareResponse{},
			shouldErr: true,
		},
		{
			desc: "Success register factory contract to itself",
			msg: &types.MsgRegisterFeeShare{
				ContractAddress:   contractAddress2,
				DeployerAddress:   sender.String(),
				WithdrawerAddress: contractAddress2,
			},
			resp:      &types.MsgRegisterFeeShareResponse{},
			shouldErr: false,
		},
		{
			desc: "Invalid register gov contract withdraw address",
			msg: &types.MsgRegisterFeeShare{
				ContractAddress:   govContract,
				DeployerAddress:   sender.String(),
				WithdrawerAddress: sender.String(),
			},
			resp:      &types.MsgRegisterFeeShareResponse{},
			shouldErr: true,
		},
		{
			desc: "Success register gov contract withdraw address to self",
			msg: &types.MsgRegisterFeeShare{
				ContractAddress:   govContract,
				DeployerAddress:   sender.String(),
				WithdrawerAddress: govContract,
			},
			resp:      &types.MsgRegisterFeeShareResponse{},
			shouldErr: false,
		},
		{
			desc: "Success register contract from DAODAO contract as admin",
			msg: &types.MsgRegisterFeeShare{
				ContractAddress:   subContract,
				DeployerAddress:   daodao,
				WithdrawerAddress: daodao,
			},
			resp:      &types.MsgRegisterFeeShareResponse{},
			shouldErr: false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			if !tc.shouldErr {
				resp, err := s.msgServer.RegisterFeeShare(s.Ctx, tc.msg)
				s.Require().NoError(err)
				s.Require().Equal(resp, tc.resp)
			} else {
				resp, err := s.msgServer.RegisterFeeShare(s.Ctx, tc.msg)
				s.Require().Error(err)
				s.Require().Nil(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateFeeShare() {
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	contractAddress := s.InstantiateContract(sender.String(), "", wasmContract)
	_, _, withdrawer := testdata.KeyTestPubAddr()

	contractAddressNoRegisFeeShare := s.InstantiateContract(sender.String(), "", wasmContract)
	s.Require().NotEqual(contractAddress, contractAddressNoRegisFeeShare)

	// RegsisFeeShare
	msg := &types.MsgRegisterFeeShare{
		ContractAddress:   contractAddress,
		DeployerAddress:   sender.String(),
		WithdrawerAddress: withdrawer.String(),
	}
	_, err := s.msgServer.RegisterFeeShare(s.Ctx, msg)
	s.Require().NoError(err)
	_, _, newWithdrawer := testdata.KeyTestPubAddr()
	s.Require().NotEqual(withdrawer, newWithdrawer)

	for _, tc := range []struct {
		desc      string
		msg       *types.MsgUpdateFeeShare
		resp      *types.MsgCancelFeeShareResponse
		shouldErr bool
	}{
		{
			desc: "Invalid - contract not regis",
			msg: &types.MsgUpdateFeeShare{
				ContractAddress:   contractAddressNoRegisFeeShare,
				DeployerAddress:   sender.String(),
				WithdrawerAddress: newWithdrawer.String(),
			},
			resp:      nil,
			shouldErr: true,
		},
		{
			desc: "Invalid - Invalid DeployerAddress",
			msg: &types.MsgUpdateFeeShare{
				ContractAddress:   contractAddress,
				DeployerAddress:   "Invalid",
				WithdrawerAddress: newWithdrawer.String(),
			},
			resp:      nil,
			shouldErr: true,
		},
		{
			desc: "Invalid - Invalid WithdrawerAddress",
			msg: &types.MsgUpdateFeeShare{
				ContractAddress:   contractAddress,
				DeployerAddress:   sender.String(),
				WithdrawerAddress: "Invalid",
			},
			resp:      nil,
			shouldErr: true,
		},
		{
			desc: "Invalid - Invalid WithdrawerAddress not change",
			msg: &types.MsgUpdateFeeShare{
				ContractAddress:   contractAddress,
				DeployerAddress:   sender.String(),
				WithdrawerAddress: withdrawer.String(),
			},
			resp:      nil,
			shouldErr: true,
		},
		{
			desc: "Success",
			msg: &types.MsgUpdateFeeShare{
				ContractAddress:   contractAddress,
				DeployerAddress:   sender.String(),
				WithdrawerAddress: newWithdrawer.String(),
			},
			resp:      &types.MsgCancelFeeShareResponse{},
			shouldErr: false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			if !tc.shouldErr {
				_, err := s.msgServer.UpdateFeeShare(s.Ctx, tc.msg)
				s.Require().NoError(err)
			} else {
				resp, err := s.msgServer.UpdateFeeShare(s.Ctx, tc.msg)
				s.Require().Error(err)
				s.Require().Nil(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCancelFeeShare() {
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	contractAddress := s.InstantiateContract(sender.String(), "", wasmContract)
	_, _, withdrawer := testdata.KeyTestPubAddr()

	// RegsisFeeShare
	msg := &types.MsgRegisterFeeShare{
		ContractAddress:   contractAddress,
		DeployerAddress:   sender.String(),
		WithdrawerAddress: withdrawer.String(),
	}
	_, err := s.msgServer.RegisterFeeShare(s.Ctx, msg)
	s.Require().NoError(err)

	for _, tc := range []struct {
		desc      string
		msg       *types.MsgCancelFeeShare
		resp      *types.MsgCancelFeeShareResponse
		shouldErr bool
	}{
		{
			desc: "Invalid - contract Address",
			msg: &types.MsgCancelFeeShare{
				ContractAddress: "Invalid",
				DeployerAddress: sender.String(),
			},
			resp:      nil,
			shouldErr: true,
		},
		{
			desc: "Invalid - deployer Address",
			msg: &types.MsgCancelFeeShare{
				ContractAddress: contractAddress,
				DeployerAddress: "Invalid",
			},
			resp:      nil,
			shouldErr: true,
		},
		{
			desc: "Success",
			msg: &types.MsgCancelFeeShare{
				ContractAddress: contractAddress,
				DeployerAddress: sender.String(),
			},
			resp:      &types.MsgCancelFeeShareResponse{},
			shouldErr: false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			if !tc.shouldErr {
				resp, err := s.msgServer.CancelFeeShare(s.Ctx, tc.msg)
				s.Require().NoError(err)
				s.Require().Equal(resp, tc.resp)
			} else {
				resp, err := s.msgServer.CancelFeeShare(s.Ctx, tc.msg)
				s.Require().Error(err)
				s.Require().Equal(resp, tc.resp)
			}
		})
	}
}
