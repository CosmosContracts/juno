package keeper_test

import (
	_ "embed"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v29/x/cw-hooks/types"
)

func (s *KeeperTestSuite) TestRegisterContracts() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, notAuthorizedAcc := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	s.FundAcc(notAuthorizedAcc, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	contractAddress := s.InstantiateContract(sender.String(), "", wasmContract)
	contractAddressWithAdmin := s.InstantiateContract(notAuthorizedAcc.String(), sender.String(), wasmContract)

	DAODAO := s.InstantiateContract(sender.String(), "", wasmContract)
	daodaoSubContract := s.InstantiateContract(DAODAO, DAODAO, wasmContract)

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
			desc:            "Invalid not authorized creator",
			ContractAddress: contractAddress,
			RegisterAddress: notAuthorizedAcc.String(),
			shouldErr:       true,
		},
		{
			desc:            "Invalid not authorized admin",
			ContractAddress: contractAddressWithAdmin,
			RegisterAddress: notAuthorizedAcc.String(),
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
			desc:            "Success register DAODAO contract from factory",
			ContractAddress: daodaoSubContract,
			RegisterAddress: DAODAO,
			shouldErr:       false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			// staking
			sResp, err := s.msgServer.RegisterStaking(s.Ctx, &types.MsgRegisterStaking{
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
			gResp, err := s.msgServer.RegisterGovernance(s.Ctx, &types.MsgRegisterGovernance{
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

func (s *KeeperTestSuite) TestUnRegisterContracts() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, notAuthorizedAcc := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	s.FundAcc(notAuthorizedAcc, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	contractAddress := s.InstantiateContract(sender.String(), "", wasmContract)

	_, err := s.msgServer.RegisterStaking(s.Ctx, &types.MsgRegisterStaking{
		ContractAddress: contractAddress,
		RegisterAddress: sender.String(),
	})
	s.Require().NoError(err)

	_, err = s.msgServer.RegisterGovernance(s.Ctx, &types.MsgRegisterGovernance{
		ContractAddress: contractAddress,
		RegisterAddress: sender.String(),
	})
	s.Require().NoError(err)

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
			desc:            "Invalid not authorized creator",
			ContractAddress: contractAddress,
			RegisterAddress: notAuthorizedAcc.String(),
			shouldErr:       true,
		},
		{
			desc:            "Success",
			ContractAddress: contractAddress,
			RegisterAddress: sender.String(),
			shouldErr:       false,
		},
		{
			desc:            "Failure contract already deleted",
			ContractAddress: contractAddress,
			RegisterAddress: sender.String(),
			shouldErr:       true,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			// staking
			sResp, err := s.msgServer.UnregisterStaking(s.Ctx, &types.MsgUnregisterStaking{
				ContractAddress: tc.ContractAddress,
				RegisterAddress: tc.RegisterAddress,
			})
			if !tc.shouldErr {
				s.Require().NoError(err)
				s.Require().Equal(sResp, &types.MsgUnregisterStakingResponse{})
			} else {
				s.Require().Error(err)
				s.Require().Nil(sResp)
			}

			// governance
			gResp, err := s.msgServer.UnregisterGovernance(s.Ctx, &types.MsgUnregisterGovernance{
				ContractAddress: tc.ContractAddress,
				RegisterAddress: tc.RegisterAddress,
			})
			if !tc.shouldErr {
				s.Require().NoError(err)
				s.Require().Equal(gResp, &types.MsgUnregisterGovernanceResponse{})
			} else {
				s.Require().Error(err)
				s.Require().Nil(gResp)
			}
		})
	}

	sc, err := s.queryClient.StakingContracts(s.Ctx, &types.QueryStakingContractsRequest{})
	s.Require().NoError(err)
	s.Require().Nil(sc.Contracts)

	gc, err := s.queryClient.GovernanceContracts(s.Ctx, &types.QueryGovernanceContractsRequest{})
	s.Require().NoError(err)
	s.Require().Nil(gc.Contracts)
}

// TODO: Reimplement, e2e is passing so this is not a priority
// func (s *KeeperTestSuite) TestContractExecution() {
// 	s.SetupTest()
// 	_, _, sender := testdata.KeyTestPubAddr()
// 	coin := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10_000_000)), sdk.NewCoin("ujuno", sdkmath.NewInt(10_000_000)))
// 	s.FundAcc(sender, coin)

// 	contractAddress := s.InstantiateContract(sender.String(), "", wasmContract)

// 	c := types.Contract{
// 		ContractAddress: contractAddress,
// 	}

// 	_, err := s.msgServer.RegisterStaking(s.Ctx, &types.MsgRegisterStaking{
// 		ContractAddress: contractAddress,
// 		RegisterAddress: sender.String(),
// 	})
// 	s.Require().NoError(err)

// 	// staking
// 	resp, err := s.queryClient.StakingContracts(s.Ctx, &types.QueryStakingContractsRequest{})
// 	s.Require().NoError(err)
// 	s.Require().Contains(resp.Contracts, c.ContractAddress)

// 	vals, err := s.stakingKeeper.GetValidators(s.Ctx, 1)
// 	s.Require().NoError(err)
// 	val := vals[0]

// 	// == Delegate Tokens ==
// 	_, err = s.stakingKeeper.Delegate(s.Ctx, sender, sdkmath.NewInt(1), stakingtypes.Bonded, val, false)
// 	s.Require().NoError(err)

// 	// query the contract to get the last modified shares (delegation)
// 	v, err := s.wasmKeeper.QuerySmart(s.Ctx, sdk.MustAccAddressFromBech32(contractAddress), []byte(`{"last_delegation_change":{}}`))
// 	s.Require().NoError(err)

// 	shares := "0.000001000000000000"
// 	expected := fmt.Sprintf(`{"validator_address":"%s","delegator_address":"%s","shares":"%s"}`, val.GetOperator(), sender.String(), shares)
// 	s.Require().Equal(expected, string(v))

// 	// == Validator Slash ==
// 	cons, err := val.GetConsAddr()
// 	s.Require().NoError(err)

// 	_, err = s.stakingKeeper.Slash(s.Ctx, cons, s.Ctx.BlockHeight(), 1, sdkmath.LegacyNewDecWithPrec(5, 1))
// 	s.Require().NoError(err)

// 	_, err = s.wasmKeeper.QuerySmart(s.Ctx, sdk.MustAccAddressFromBech32(contractAddress), []byte(`{"last_validator_slash":{}}`))
// 	s.Require().NoError(err)
// }
