package keeper_test

import (
	"fmt"

	_ "embed"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v17/x/cw-hooks/types"
)

// TODO: TestContractDelete/Unregister (and that it only applies to 1, both both)

func (s *IntegrationTestSuite) TestRegisterContracts() {
	_, _, sender := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	contractAddress := s.InstantiateContract(sender.String(), "")

	DAODAO := s.InstantiateContract(sender.String(), "")
	daodaoSubContract := s.InstantiateContract(DAODAO, DAODAO)

	for _, tc := range []struct {
		desc string

		ContractAddress string
		RegisterAddress string

		shouldErr bool
	}{
		{
			desc:            "Invalid contract address",
			ContractAddress: "Invalid",
			shouldErr:       true,
		},
		{
			desc:            "Invalid sender address",
			ContractAddress: contractAddress,
			RegisterAddress: "Invalid",
			shouldErr:       true,
		},
		{
			desc:            "Success",
			ContractAddress: contractAddress,
			RegisterAddress: sender.String(),
			shouldErr:       false,
		},
		{
			desc:            "Failure register same contract",
			ContractAddress: contractAddress,
			RegisterAddress: sender.String(),
			shouldErr:       true,
		},
		{
			desc:            "Failure register a DAODAO contract",
			ContractAddress: daodaoSubContract,
			RegisterAddress: sender.String(),
			shouldErr:       true,
		},
		{
			desc:            "Success register DAODAO contract from factory",
			ContractAddress: daodaoSubContract,
			RegisterAddress: DAODAO,
			shouldErr:       false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			goCtx := sdk.WrapSDKContext(s.ctx)
			// staking
			sResp, err := s.msgServer.RegisterStaking(goCtx, &types.MsgRegisterStaking{
				ContractAddress: tc.ContractAddress,
				RegisterAddress: tc.RegisterAddress,
			})
			if !tc.shouldErr {
				s.Require().NoError(err)
				s.Require().Equal(sResp, &types.MsgRegisterStakingResponse{})
			} else {
				s.Require().Error(err)
				s.Require().Nil(sResp)
			}

			// governance
			gResp, err := s.msgServer.RegisterGovernance(goCtx, &types.MsgRegisterGovernance{
				ContractAddress: tc.ContractAddress,
				RegisterAddress: tc.RegisterAddress,
			})
			if !tc.shouldErr {
				s.Require().NoError(err)
				s.Require().Equal(gResp, &types.MsgRegisterGovernanceResponse{})
			} else {
				s.Require().Error(err)
				s.Require().Nil(gResp)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestContractExecution() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	coin := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10_000_000)), sdk.NewCoin("ujuno", sdk.NewInt(10_000_000)))
	_ = s.FundAccount(s.ctx, sender, coin)

	contractAddress := s.InstantiateContract(sender.String(), "")
	goCtx := sdk.WrapSDKContext(s.ctx)

	c := types.Contract{
		ContractAddress: contractAddress,
		RegisterAddress: sender.String(),
	}

	_, err := s.msgServer.RegisterStaking(goCtx, &types.MsgRegisterStaking{
		ContractAddress: contractAddress,
		RegisterAddress: sender.String(),
	})
	s.Require().NoError(err)

	// staking
	goCtx = sdk.WrapSDKContext(s.ctx)
	resp, err := s.queryClient.StakingContracts(goCtx, &types.QueryStakingContractsRequest{})
	s.Require().NoError(err)
	s.Require().Contains(resp.Contracts, c.ContractAddress)

	val := s.stakingKeeper.GetValidators(s.ctx, 1)[0]
	// print val
	fmt.Println(val)

	// == Delegate Tokens ==
	_, err = s.stakingKeeper.Delegate(s.ctx, sender, sdk.NewInt(1), stakingtypes.Bonded, val, false)
	s.Require().NoError(err)

	// query the contract to get the last modified shares (delegation)
	v, err := s.wasmKeeper.QuerySmart(s.ctx, sdk.MustAccAddressFromBech32(contractAddress), []byte(`{"last_delegation_change":{}}`))
	s.Require().NoError(err)

	shares := "0.000001000000000000"
	expected := fmt.Sprintf(`{"validator_address":"%s","delegator_address":"%s","shares":"%s"}`, val.GetOperator().String(), sender.String(), shares)
	s.Require().Equal(expected, string(v))

	// == Validator Slash ==
	cons, err := val.GetConsAddr()
	s.Require().NoError(err)

	s.stakingKeeper.Slash(s.ctx, cons, s.ctx.BlockHeight(), 1, sdk.NewDecWithPrec(5, 1))

	v, err = s.wasmKeeper.QuerySmart(s.ctx, sdk.MustAccAddressFromBech32(contractAddress), []byte(`{"last_validator_slash":{}}`))
	fmt.Println(string(v))
	s.Require().NoError(err)
}
