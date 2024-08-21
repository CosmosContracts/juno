package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v24/x/cw-hooks/types"
)

func (s *IntegrationTestSuite) TestContracts() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	var contractAddressList []string
	var index uint64
	for index < 5 {
		contractAddress := s.InstantiateContract(sender.String(), "")
		contractAddressList = append(contractAddressList, contractAddress)
		index++
	}

	// Register Staking & Gov
	var staking []types.Contract
	var governance []types.Contract
	for _, contractAddress := range contractAddressList {
		goCtx := sdk.WrapSDKContext(s.ctx)

		c := types.Contract{
			ContractAddress: contractAddress,
			RegisterAddress: sender.String(),
		}

		_, err := s.msgServer.RegisterStaking(goCtx, &types.MsgRegisterStaking{
			ContractAddress: c.ContractAddress,
			RegisterAddress: c.RegisterAddress,
		})
		staking = append(staking, c)
		s.Require().NoError(err)

		_, err = s.msgServer.RegisterGovernance(goCtx, &types.MsgRegisterGovernance{
			ContractAddress: c.ContractAddress,
			RegisterAddress: c.RegisterAddress,
		})
		governance = append(governance, c)
		s.Require().NoError(err)
	}

	goCtx := sdk.WrapSDKContext(s.ctx)

	// verify outputs
	resp, err := s.queryClient.StakingContracts(goCtx, &types.QueryStakingContractsRequest{})
	s.Require().NoError(err)
	s.Require().LessOrEqual(len(resp.Contracts), len(staking))

	resp2, err := s.queryClient.GovernanceContracts(goCtx, &types.QueryGovernanceContractsRequest{})
	s.Require().NoError(err)
	s.Require().LessOrEqual(len(resp2.Contracts), len(governance))
}
