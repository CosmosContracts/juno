package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CosmosContracts/juno/v28/testutil/common/nullify"
	"github.com/CosmosContracts/juno/v28/x/feepay/types"
)

func (s *KeeperTestSuite) TestQueryContract() {
	// Get & fund creator
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	// Instantiate the contractAddr
	contractAddr := s.InstantiateContract(sender.String(), "", wasmContract)

	s.registerFeePayContract(sender.String(), contractAddr, 0, 1)

	s.Run("QueryContract", func() {
		// Query for the contract
		res, err := s.queryClient.FeePayContract(s.Ctx, &types.QueryFeePayContractRequest{
			ContractAddress: contractAddr,
		})

		// Ensure no error and response exists
		s.Require().NoError(err)
		s.Require().Equal(res, &types.QueryFeePayContractResponse{
			FeePayContract: types.FeePayContract{
				ContractAddress: contractAddr,
				WalletLimit:     1,
			},
		})
	})
}

// Test that when a Fee Pay contract is registered, the balance is always 0.
func (s *KeeperTestSuite) TestQueryContractBalance() {
	// Get & fund creator
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	s.Run("QueryContract", func() {
		for _, bal := range []struct {
			balance uint64
		}{
			{balance: 0},
			{balance: 1_000_000},
		} {
			bal := bal

			// Instantiate the contractAddr
			contractAddr := s.InstantiateContract(sender.String(), "", wasmContract)
			s.registerFeePayContract(sender.String(), contractAddr, bal.balance, 1)

			// Query for the contract
			res, err := s.queryClient.FeePayContract(s.Ctx, &types.QueryFeePayContractRequest{
				ContractAddress: contractAddr,
			})

			// Ensure no error and response exists
			s.Require().NoError(err)
			s.Require().Equal(res, &types.QueryFeePayContractResponse{
				FeePayContract: types.FeePayContract{
					ContractAddress: contractAddr,
					Balance:         0,
					WalletLimit:     1,
				},
			})
		}
	})
}

func (s *KeeperTestSuite) TestQueryContracts() {
	// Get & fund creator
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	// Instantiate & register multiple fee pay contracts
	var contractAddressList []string
	var feePayContracts []types.FeePayContract
	for index := 0; index < 5; index++ {
		// Instantiate the contractAddr
		contractAddr := s.InstantiateContract(sender.String(), "", wasmContract)

		// Register the fee pay contract
		s.registerFeePayContract(sender.String(), contractAddr, 0, 1)

		// Query for the contract
		res, err := s.queryClient.FeePayContract(s.Ctx, &types.QueryFeePayContractRequest{
			ContractAddress: contractAddr,
		})

		// Ensure no error and response exists
		s.Require().NoError(err)
		s.Require().Equal(res, &types.QueryFeePayContractResponse{
			FeePayContract: types.FeePayContract{
				ContractAddress: contractAddr,
				WalletLimit:     1,
			},
		})

		// Append to lists
		contractAddressList = append(contractAddressList, contractAddr)
		feePayContracts = append(feePayContracts, res.FeePayContract)
	}

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryFeePayContractsRequest {
		return &types.QueryFeePayContractsRequest{
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
			resp, err := s.queryClient.FeePayContracts(s.Ctx, request(nil, uint64(i), uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.FeePayContracts), step)
			s.Require().Subset(nullify.Fill(feePayContracts), nullify.Fill(resp.FeePayContracts))
		}
	})

	s.Run("ByKey", func() {
		step := 2
		var next []byte
		for i := 0; i < len(contractAddressList); i += step {
			resp, err := s.queryClient.FeePayContracts(s.Ctx, request(next, 0, uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.FeePayContracts), step)
			s.Require().Subset(nullify.Fill(feePayContracts), nullify.Fill(resp.FeePayContracts))
			next = resp.Pagination.NextKey
		}
	})

	s.Run("Total", func() {
		resp, err := s.queryClient.FeePayContracts(s.Ctx, request(nil, 0, 0, true))
		s.Require().NoError(err)
		s.Require().Equal(len(contractAddressList), int(resp.Pagination.Total))
		s.Require().ElementsMatch(nullify.Fill(feePayContracts), nullify.Fill(resp.FeePayContracts))
	})
}

func (s *KeeperTestSuite) TestQueryEligibility() {
	// Get & fund creator
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000)), sdk.NewCoin("ujuno", sdkmath.NewInt(100_000_000))))

	// Instantiate the contractAddr
	contractAddr := s.InstantiateContract(sender.String(), "", wasmContract)

	// Register the fee pay contract
	s.registerFeePayContract(sender.String(), contractAddr, 0, 1)

	s.Run("QueryEligibilityNoFunds", func() {
		// Query for the contract
		res, err := s.queryClient.FeePayWalletIsEligible(s.Ctx, &types.QueryFeePayWalletIsEligibleRequest{
			ContractAddress: contractAddr,
			WalletAddress:   sender.String(),
		})

		// Should not error, user has not exceeded limit
		s.Require().NoError(err)
		s.Require().True(res.Eligible)
	})

	// Add funds
	_, err := s.msgServer.FundFeePayContract(s.Ctx, &types.MsgFundFeePayContract{
		SenderAddress:   sender.String(),
		ContractAddress: contractAddr,
		Amount:          sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(1_000_000))),
	})
	s.Require().NoError(err)

	s.Run("QueryEligibilityWithFunds", func() {
		// Query for the contract
		res, err := s.queryClient.FeePayWalletIsEligible(s.Ctx, &types.QueryFeePayWalletIsEligibleRequest{
			ContractAddress: contractAddr,
			WalletAddress:   sender.String(),
		})

		// Should not error, user has not exceeded limit
		s.Require().NoError(err)
		s.Require().True(res.Eligible)
	})

	// Update usage limit to 0
	_, err = s.msgServer.UpdateFeePayContractWalletLimit(s.Ctx, &types.MsgUpdateFeePayContractWalletLimit{
		SenderAddress:   sender.String(),
		ContractAddress: contractAddr,
		WalletLimit:     0,
	})
	s.Require().NoError(err)

	s.Run("QueryEligibilityWithLimit", func() {
		// Query for the contract
		res, err := s.queryClient.FeePayWalletIsEligible(s.Ctx, &types.QueryFeePayWalletIsEligibleRequest{
			ContractAddress: contractAddr,
			WalletAddress:   sender.String(),
		})

		// Should error, wallet exceeded limit of 0
		s.Require().Error(err)
		s.Require().Nil(res)
	})
}

func (s *KeeperTestSuite) TestQueryUses() {
	// Get & fund creator
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	// Instantiate the contractAddr
	contractAddr := s.InstantiateContract(sender.String(), "", wasmContract)

	// Register the fee pay contract
	s.registerFeePayContract(sender.String(), contractAddr, 0, 1)

	s.Run("QueryUses", func() {
		// Query for the contract
		res, err := s.queryClient.FeePayContractUses(s.Ctx, &types.QueryFeePayContractUsesRequest{
			ContractAddress: contractAddr,
			WalletAddress:   sender.String(),
		})

		// Ensure no error and response is false, contract should have no funds
		s.Require().NoError(err)
		s.Require().Equal(uint64(0), res.Uses)
	})
}
