package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CosmosContracts/juno/v27/testutil/common/nullify"
	"github.com/CosmosContracts/juno/v27/x/feeshare/types"
)

func (s *KeeperTestSuite) TestFeeShares() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	_, _, withdrawer := testdata.KeyTestPubAddr()

	var contractAddressList []string
	var index uint64
	for index < 5 {
		contractAddress := s.InstantiateContract(sender.String(), "", wasmContract)
		contractAddressList = append(contractAddressList, contractAddress)
		index++
	}

	// RegsisFeeShare
	var feeShares []types.FeeShare
	for _, contractAddress := range contractAddressList {
		msg := &types.MsgRegisterFeeShare{
			ContractAddress:   contractAddress,
			DeployerAddress:   sender.String(),
			WithdrawerAddress: withdrawer.String(),
		}

		feeShare := types.FeeShare{
			ContractAddress:   contractAddress,
			DeployerAddress:   sender.String(),
			WithdrawerAddress: withdrawer.String(),
		}

		feeShares = append(feeShares, feeShare)

		_, err := s.msgServer.RegisterFeeShare(s.Ctx, msg)
		s.Require().NoError(err)
	}

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryFeeSharesRequest {
		return &types.QueryFeeSharesRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	s.Run("ByOffset", func() {
		step := 2
		for i := 0; i < len(contractAddressList); i += step {
			resp, err := s.queryClient.FeeShares(s.Ctx, request(nil, uint64(i), uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.Feeshare), step)
			s.Require().Subset(nullify.Fill(feeShares), nullify.Fill(resp.Feeshare))
		}
	})
	s.Run("ByKey", func() {
		step := 2
		var next []byte
		for i := 0; i < len(contractAddressList); i += step {
			resp, err := s.queryClient.FeeShares(s.Ctx, request(next, 0, uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.Feeshare), step)
			s.Require().Subset(nullify.Fill(feeShares), nullify.Fill(resp.Feeshare))
			next = resp.Pagination.NextKey
		}
	})
	s.Run("Total", func() {
		resp, err := s.queryClient.FeeShares(s.Ctx, request(nil, 0, 0, true))
		s.Require().NoError(err)
		s.Require().Equal(len(feeShares), int(resp.Pagination.Total))
		s.Require().ElementsMatch(nullify.Fill(feeShares), nullify.Fill(resp.Feeshare))
	})
}

func (s *KeeperTestSuite) TestFeeShare() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	_, _, withdrawer := testdata.KeyTestPubAddr()

	contractAddress := s.InstantiateContract(sender.String(), "", wasmContract)

	msg := &types.MsgRegisterFeeShare{
		ContractAddress:   contractAddress,
		DeployerAddress:   sender.String(),
		WithdrawerAddress: withdrawer.String(),
	}

	feeShare := types.FeeShare{
		ContractAddress:   contractAddress,
		DeployerAddress:   sender.String(),
		WithdrawerAddress: withdrawer.String(),
	}
	_, err := s.msgServer.RegisterFeeShare(s.Ctx, msg)
	s.Require().NoError(err)

	req := &types.QueryFeeShareRequest{
		ContractAddress: contractAddress,
	}
	resp, err := s.queryClient.FeeShare(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(resp.Feeshare, feeShare)
}

func (s *KeeperTestSuite) TestDeployerFeeShares() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	_, _, withdrawer := testdata.KeyTestPubAddr()

	var contractAddressList []string
	var index uint64
	for index < 5 {
		contractAddress := s.InstantiateContract(sender.String(), "", wasmContract)
		contractAddressList = append(contractAddressList, contractAddress)
		index++
	}

	// RegsisFeeShare
	for _, contractAddress := range contractAddressList {
		msg := &types.MsgRegisterFeeShare{
			ContractAddress:   contractAddress,
			DeployerAddress:   sender.String(),
			WithdrawerAddress: withdrawer.String(),
		}

		_, err := s.msgServer.RegisterFeeShare(s.Ctx, msg)
		s.Require().NoError(err)
	}

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryDeployerFeeSharesRequest {
		return &types.QueryDeployerFeeSharesRequest{
			DeployerAddress: sender.String(),
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	s.Run("ByOffset", func() {
		step := 2
		for i := 0; i < len(contractAddressList); i += step {
			resp, err := s.queryClient.DeployerFeeShares(s.Ctx, request(nil, uint64(i), uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.ContractAddresses), step)
			s.Require().Subset(nullify.Fill(contractAddressList), nullify.Fill(resp.ContractAddresses))
		}
	})
	s.Run("ByKey", func() {
		step := 2
		var next []byte
		for i := 0; i < len(contractAddressList); i += step {
			resp, err := s.queryClient.DeployerFeeShares(s.Ctx, request(next, 0, uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.ContractAddresses), step)
			s.Require().Subset(nullify.Fill(contractAddressList), nullify.Fill(resp.ContractAddresses))
			next = resp.Pagination.NextKey
		}
	})
	s.Run("Total", func() {
		resp, err := s.queryClient.DeployerFeeShares(s.Ctx, request(nil, 0, 0, true))
		s.Require().NoError(err)
		s.Require().Equal(len(contractAddressList), int(resp.Pagination.Total))
		s.Require().ElementsMatch(nullify.Fill(contractAddressList), nullify.Fill(resp.ContractAddresses))
	})
}

func (s *KeeperTestSuite) TestWithdrawerFeeShares() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	_, _, withdrawer := testdata.KeyTestPubAddr()

	var contractAddressList []string
	var index uint64
	for index < 5 {
		contractAddress := s.InstantiateContract(sender.String(), "", wasmContract)
		contractAddressList = append(contractAddressList, contractAddress)
		index++
	}

	// RegsisFeeShare
	for _, contractAddress := range contractAddressList {
		msg := &types.MsgRegisterFeeShare{
			ContractAddress:   contractAddress,
			DeployerAddress:   sender.String(),
			WithdrawerAddress: withdrawer.String(),
		}

		_, err := s.msgServer.RegisterFeeShare(s.Ctx, msg)
		s.Require().NoError(err)
	}

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryWithdrawerFeeSharesRequest {
		return &types.QueryWithdrawerFeeSharesRequest{
			WithdrawerAddress: withdrawer.String(),
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	s.Run("ByOffset", func() {
		step := 2
		for i := 0; i < len(contractAddressList); i += step {
			resp, err := s.queryClient.WithdrawerFeeShares(s.Ctx, request(nil, uint64(i), uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.ContractAddresses), step)
			s.Require().Subset(nullify.Fill(contractAddressList), nullify.Fill(resp.ContractAddresses))
		}
	})
	s.Run("ByKey", func() {
		step := 2
		var next []byte
		for i := 0; i < len(contractAddressList); i += step {
			resp, err := s.queryClient.WithdrawerFeeShares(s.Ctx, request(next, 0, uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.ContractAddresses), step)
			s.Require().Subset(nullify.Fill(contractAddressList), nullify.Fill(resp.ContractAddresses))
			next = resp.Pagination.NextKey
		}
	})
	s.Run("Total", func() {
		resp, err := s.queryClient.WithdrawerFeeShares(s.Ctx, request(nil, 0, 0, true))
		s.Require().NoError(err)
		s.Require().Equal(len(contractAddressList), int(resp.Pagination.Total))
		s.Require().ElementsMatch(nullify.Fill(contractAddressList), nullify.Fill(resp.ContractAddresses))
	})
}
