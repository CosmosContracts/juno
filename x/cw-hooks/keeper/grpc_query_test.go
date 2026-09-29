package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v31/x/cw-hooks/types"
)

func (s *KeeperTestSuite) TestContracts() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	var registered []types.ContractInfo
	for range 5 {
		addr := s.InstantiateContract(sender.String(), "", wasmContract)
		info := types.ContractInfo{
			ContractAddress: addr,
			FailureCounter:  0,
		}
		registered = append(registered, info)

		s.Require().NoError(s.registerContract("staking", sender.String(), addr))
		s.Require().NoError(s.registerContract("gov", sender.String(), addr))
	}

	storedStaking, err := s.App.AppKeepers.CWHooksKeeper.GetAllContracts(s.Ctx, types.StakingPrefixKey)
	s.Require().NoError(err)
	s.T().Logf("stored staking contracts: %d", len(storedStaking))
	s.Require().Len(storedStaking, len(registered))

	iter, err := s.App.AppKeepers.CWHooksKeeper.Contracts.Iterate(s.Ctx, nil)
	s.Require().NoError(err)
	allContracts, err := iter.Values()
	s.Require().NoError(err)
	s.T().Logf("total contracts stored: %d", len(allContracts))
	s.Require().Len(allContracts, len(registered)*2)

	iterAll, err := s.App.AppKeepers.CWHooksKeeper.Contracts.Iterate(s.Ctx, nil)
	s.Require().NoError(err)
	keyValues, err := iterAll.KeyValues()
	s.Require().NoError(err)
	for _, kv := range keyValues {
		s.T().Logf("stored contract module=%q address=%s", string(kv.Key.K1()), kv.Value.ContractAddress)
	}

	storedGov, err := s.App.AppKeepers.CWHooksKeeper.GetAllContracts(s.Ctx, types.GovPrefixKey)
	s.Require().NoError(err)
	s.T().Logf("stored gov contracts: %d", len(storedGov))
	s.Require().Len(storedGov, len(registered))

	stakingResp, err := s.queryClient.Contracts(s.Ctx, &types.QueryContractsRequest{Module: "staking"})
	s.Require().NoError(err)
	s.Require().Len(stakingResp.Contracts, len(registered))
	s.Require().ElementsMatch(registered, stakingResp.Contracts)

	govResp, err := s.queryClient.Contracts(s.Ctx, &types.QueryContractsRequest{Module: "gov"})
	s.Require().NoError(err)
	s.Require().Len(govResp.Contracts, len(registered))
	s.Require().ElementsMatch(registered, govResp.Contracts)

	target := registered[0]
	infoResp, err := s.queryClient.ContractInfo(s.Ctx, &types.QueryContractInfoRequest{
		Module:          "staking",
		ContractAddress: target.ContractAddress,
	})
	s.Require().NoError(err)
	s.Require().Equal(target, infoResp.Contract)

	// invalid module should error
	_, err = s.queryClient.Contracts(s.Ctx, &types.QueryContractsRequest{Module: "invalid"})
	s.Require().Error(err)

	_, err = s.queryClient.ContractInfo(s.Ctx, &types.QueryContractInfoRequest{
		Module:          "invalid",
		ContractAddress: target.ContractAddress,
	})
	s.Require().Error(err)
}
