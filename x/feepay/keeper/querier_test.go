package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CosmosContracts/juno/v18/testutil/nullify"
	"github.com/CosmosContracts/juno/v18/x/feepay/types"
)

func (s *IntegrationTestSuite) TestQueryContract() {
	// Get & fund creator
	_, _, sender := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	// Instantiate the contractAddr
	contractAddr := s.InstantiateContract(sender.String(), "")

	s.registerFeePayContract(sender.String(), contractAddr, 0, 1)

	s.Run("QueryContract", func() {
		// Query for the contract
		res, err := s.queryClient.FeePayContract(s.ctx, &types.QueryFeePayContract{
			ContractAddress: contractAddr,
		})

		// Ensure no error and response exists
		s.Require().NoError(err)
		s.Require().Equal(res, &types.QueryFeePayContractResponse{
			FeePayContract: &types.FeePayContract{
				ContractAddress: contractAddr,
				WalletLimit:     1,
			},
		})
	})
}

// Test that when a Fee Pay contract is registered, the balance is always 0.
func (s *IntegrationTestSuite) TestQueryContractBalance() {
	// Get & fund creator
	_, _, sender := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	s.Run("QueryContract", func() {

		for _, bal := range []struct {
			balance uint64
		}{
			{balance: 0},
			{balance: 1_000_000},
		} {
			bal := bal

			// Instantiate the contractAddr
			contractAddr := s.InstantiateContract(sender.String(), "")
			s.registerFeePayContract(sender.String(), contractAddr, bal.balance, 1)

			// Query for the contract
			res, err := s.queryClient.FeePayContract(s.ctx, &types.QueryFeePayContract{
				ContractAddress: contractAddr,
			})

			// Ensure no error and response exists
			s.Require().NoError(err)
			s.Require().Equal(res, &types.QueryFeePayContractResponse{
				FeePayContract: &types.FeePayContract{
					ContractAddress: contractAddr,
					Balance:         0,
					WalletLimit:     1,
				},
			})
		}
	})
}

func (s *IntegrationTestSuite) TestQueryContracts() {
	// Get & fund creator
	_, _, sender := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	// Instantiate & register multiple fee pay contracts
	var contractAddressList []string
	var feePayContracts []types.FeePayContract
	for index := 0; index < 5; index++ {
		// Instantiate the contractAddr
		contractAddr := s.InstantiateContract(sender.String(), "")

		// Register the fee pay contract
		s.registerFeePayContract(sender.String(), contractAddr, 0, 1)

		// Query for the contract
		res, err := s.queryClient.FeePayContract(s.ctx, &types.QueryFeePayContract{
			ContractAddress: contractAddr,
		})

		// Ensure no error and response exists
		s.Require().NoError(err)
		s.Require().Equal(res, &types.QueryFeePayContractResponse{
			FeePayContract: &types.FeePayContract{
				ContractAddress: contractAddr,
				WalletLimit:     1,
			},
		})

		// Append to lists
		contractAddressList = append(contractAddressList, contractAddr)
		feePayContracts = append(feePayContracts, *res.FeePayContract)
	}

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryFeePayContracts {
		return &types.QueryFeePayContracts{
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
		goCtx := sdk.WrapSDKContext(s.ctx)
		for i := 0; i < len(contractAddressList); i += step {
			resp, err := s.queryClient.FeePayContracts(goCtx, request(nil, uint64(i), uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.FeePayContracts), step)
			s.Require().Subset(nullify.Fill(feePayContracts), nullify.Fill(resp.FeePayContracts))
		}
	})

	s.Run("ByKey", func() {
		step := 2
		var next []byte
		goCtx := sdk.WrapSDKContext(s.ctx)
		for i := 0; i < len(contractAddressList); i += step {
			resp, err := s.queryClient.FeePayContracts(goCtx, request(next, 0, uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.FeePayContracts), step)
			s.Require().Subset(nullify.Fill(feePayContracts), nullify.Fill(resp.FeePayContracts))
			next = resp.Pagination.NextKey
		}
	})

	s.Run("Total", func() {
		goCtx := sdk.WrapSDKContext(s.ctx)
		resp, err := s.queryClient.FeePayContracts(goCtx, request(nil, 0, 0, true))
		s.Require().NoError(err)
		s.Require().Equal(len(contractAddressList), int(resp.Pagination.Total))
		s.Require().ElementsMatch(nullify.Fill(feePayContracts), nullify.Fill(resp.FeePayContracts))
	})
}

func (s *IntegrationTestSuite) TestQueryEligibility() {
	// Get & fund creator
	_, _, sender := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000)), sdk.NewCoin("ujuno", sdk.NewInt(100_000_000))))

	// Instantiate the contractAddr
	contractAddr := s.InstantiateContract(sender.String(), "")

	// Register the fee pay contract
	s.registerFeePayContract(sender.String(), contractAddr, 0, 1)

	s.Run("QueryEligibilityNoFunds", func() {
		// Query for the contract
		res, err := s.queryClient.FeePayWalletIsEligible(s.ctx, &types.QueryFeePayWalletIsEligible{
			ContractAddress: contractAddr,
			WalletAddress:   sender.String(),
		})

		// Should not error, user has not exceeded limit
		s.Require().NoError(err)
		s.Require().True(res.Eligible)
	})

	// Add funds
	_, err := s.app.AppKeepers.FeePayKeeper.FundFeePayContract(s.ctx, &types.MsgFundFeePayContract{
		SenderAddress:   sender.String(),
		ContractAddress: contractAddr,
		Amount:          sdk.NewCoins(sdk.NewCoin("ujuno", sdk.NewInt(1_000_000))),
	})
	s.Require().NoError(err)

	s.Run("QueryEligibilityWithFunds", func() {
		// Query for the contract
		res, err := s.queryClient.FeePayWalletIsEligible(s.ctx, &types.QueryFeePayWalletIsEligible{
			ContractAddress: contractAddr,
			WalletAddress:   sender.String(),
		})

		// Should not error, user has not exceeded limit
		s.Require().NoError(err)
		s.Require().True(res.Eligible)
	})

	// Update usage limit to 0
	_, err = s.app.AppKeepers.FeePayKeeper.UpdateFeePayContractWalletLimit(s.ctx, &types.MsgUpdateFeePayContractWalletLimit{
		SenderAddress:   sender.String(),
		ContractAddress: contractAddr,
		WalletLimit:     0,
	})
	s.Require().NoError(err)

	s.Run("QueryEligibilityWithLimit", func() {
		// Query for the contract
		res, err := s.queryClient.FeePayWalletIsEligible(s.ctx, &types.QueryFeePayWalletIsEligible{
			ContractAddress: contractAddr,
			WalletAddress:   sender.String(),
		})

		// Should error, wallet exceeded limit of 0
		s.Require().Error(err)
		s.Require().Nil(res)
	})
}

func (s *IntegrationTestSuite) TestQueryUses() {
	// Get & fund creator
	_, _, sender := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	// Instantiate the contractAddr
	contractAddr := s.InstantiateContract(sender.String(), "")

	// Register the fee pay contract
	s.registerFeePayContract(sender.String(), contractAddr, 0, 1)

	s.Run("QueryUses", func() {
		// Query for the contract
		res, err := s.queryClient.FeePayContractUses(s.ctx, &types.QueryFeePayContractUses{
			ContractAddress: contractAddr,
			WalletAddress:   sender.String(),
		})

		// Ensure no error and response is false, contract should have no funds
		s.Require().NoError(err)
		s.Require().Equal(uint64(0), res.Uses)
	})
}
