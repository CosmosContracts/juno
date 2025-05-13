package keeper_test

import (
	_ "embed"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/CosmosContracts/juno/v29/x/feepay/types"
)

func (s *KeeperTestSuite) TestRegisterFeePayContract() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	s.FundAcc(admin, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	noAdminContractAddress := s.InstantiateContract(sender.String(), "", wasmContract)
	withAdminContractAddress := s.InstantiateContract(sender.String(), admin.String(), wasmContract)
	tContract := s.InstantiateContract(sender.String(), admin.String(), wasmContract)

	for _, tc := range []struct {
		desc            string
		contractAddress string
		senderAddress   string
		shouldErr       bool
	}{
		{
			desc:            "Success - Creator",
			contractAddress: noAdminContractAddress,
			senderAddress:   sender.String(),
			shouldErr:       false,
		},
		{
			desc:            "Fail - Already Registered Contract",
			contractAddress: noAdminContractAddress,
			senderAddress:   sender.String(),
			shouldErr:       true,
		},
		{
			desc:            "Success - Admin",
			contractAddress: withAdminContractAddress,
			senderAddress:   admin.String(),
			shouldErr:       false,
		},
		{
			desc:            "Error - Contract Already Registered",
			contractAddress: withAdminContractAddress,
			senderAddress:   admin.String(),
			shouldErr:       true,
		},
		{
			desc:            "Error - Invalid Sender",
			contractAddress: tContract,
			senderAddress:   "Invalid",
			shouldErr:       true,
		},
		{
			desc:            "Error - Invalid Contract",
			contractAddress: "Invalid",
			senderAddress:   admin.String(),
			shouldErr:       true,
		},
	} {
		tc := tc

		s.Run(tc.desc, func() {
			_, err := s.msgServer.RegisterFeePayContract(s.Ctx, &types.MsgRegisterFeePayContract{
				SenderAddress: tc.senderAddress,
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

func (s *KeeperTestSuite) TestUnregisterFeePayContract() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	s.FundAcc(admin, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	creatorContract := s.InstantiateContract(sender.String(), "", wasmContract)
	adminContract := s.InstantiateContract(sender.String(), admin.String(), wasmContract)

	s.registerFeePayContract(sender.String(), creatorContract, 0, 1)
	s.registerFeePayContract(admin.String(), adminContract, 0, 0)

	for _, tc := range []struct {
		desc            string
		contractAddress string
		senderAddress   string
		shouldErr       bool
	}{
		{
			desc:            "Fail - Invalid Contract Address",
			contractAddress: "Invalid",
			senderAddress:   sender.String(),
			shouldErr:       true,
		},
		{
			desc:            "Fail - Invalid Sender Address",
			contractAddress: creatorContract,
			senderAddress:   "Invalid",
			shouldErr:       true,
		},
		{
			desc:            "Success - Unregister Creator Contract as Creator",
			contractAddress: creatorContract,
			senderAddress:   sender.String(),
			shouldErr:       false,
		},
		{
			desc:            "Fail - Unregister Admin Contract As Creator",
			contractAddress: adminContract,
			senderAddress:   sender.String(),
			shouldErr:       true,
		},
		{
			desc:            "Success - Unregister Admin Contract As Admin",
			contractAddress: adminContract,
			senderAddress:   admin.String(),
			shouldErr:       false,
		},
		{
			desc:            "Fail - Already Unregistered",
			contractAddress: creatorContract,
			senderAddress:   sender.String(),
			shouldErr:       true,
		},
	} {
		tc := tc

		s.Run(tc.desc, func() {
			_, err := s.msgServer.UnregisterFeePayContract(s.Ctx, &types.MsgUnregisterFeePayContract{
				SenderAddress:   tc.senderAddress,
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

func (s *KeeperTestSuite) TestFundFeePayContract() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000)), sdk.NewCoin("ujuno", sdkmath.NewInt(100_000_000))))
	s.FundAcc(admin, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	contract := s.InstantiateContract(sender.String(), "", wasmContract)

	s.registerFeePayContract(sender.String(), contract, 0, 1)

	for _, tc := range []struct {
		desc            string
		contractAddress string
		senderAddress   string
		amount          sdk.Coins
		shouldErr       bool
	}{
		{
			desc:            "Fail - Invalid Contract Address",
			contractAddress: "Invalid",
			senderAddress:   sender.String(),
			amount:          sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(1_000_000))),
			shouldErr:       true,
		},
		{
			desc:            "Fail - Invalid Sender Address",
			contractAddress: contract,
			senderAddress:   "Invalid",
			amount:          sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(1_000_000))),
			shouldErr:       true,
		},
		{
			desc:            "Fail - Invalid Funds",
			contractAddress: contract,
			senderAddress:   sender.String(),
			amount:          sdk.NewCoins(sdk.NewCoin("invalid-denom", sdkmath.NewInt(1_000_000))),
			shouldErr:       true,
		},
		{
			desc:            "Fail - Wallet Not Enough Funds",
			contractAddress: contract,
			senderAddress:   sender.String(),
			amount:          sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(100_000_000_000))),
			shouldErr:       true,
		},
		{
			desc:            "Success - Contract Funded",
			contractAddress: contract,
			senderAddress:   sender.String(),
			amount:          sdk.NewCoins(sdk.NewCoin("ujuno", sdkmath.NewInt(1_000_000))),
			shouldErr:       false,
		},
	} {
		tc := tc

		s.Run(tc.desc, func() {
			_, err := s.msgServer.FundFeePayContract(s.Ctx, &types.MsgFundFeePayContract{
				SenderAddress:   tc.senderAddress,
				ContractAddress: tc.contractAddress,
				Amount:          tc.amount,
			})

			if tc.shouldErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateFeePayContractWalletLimit() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	s.FundAcc(admin, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	creatorContract := s.InstantiateContract(sender.String(), "", wasmContract)
	adminContract := s.InstantiateContract(sender.String(), admin.String(), wasmContract)

	s.registerFeePayContract(sender.String(), creatorContract, 0, 1)
	s.registerFeePayContract(admin.String(), adminContract, 0, 0)

	for _, tc := range []struct {
		desc            string
		contractAddress string
		senderAddress   string
		walletLimit     uint64
		shouldErr       bool
	}{
		{
			desc:            "Success - Update Admin Contract As Admin",
			contractAddress: adminContract,
			senderAddress:   admin.String(),
			walletLimit:     10,
			shouldErr:       false,
		},
		{
			desc:            "Fail - Update Admin Contract As Creator",
			contractAddress: adminContract,
			senderAddress:   sender.String(),
			walletLimit:     150,
			shouldErr:       true,
		},
		{
			desc:            "Success - Update Admin Contract As Admin (lower bounds)",
			contractAddress: adminContract,
			senderAddress:   admin.String(),
			walletLimit:     0,
			shouldErr:       false,
		},
		{
			desc:            "Success - Update Admin Contract As Admin (upper bounds)",
			contractAddress: adminContract,
			senderAddress:   admin.String(),
			walletLimit:     1_000_000,
			shouldErr:       false,
		},
		{
			desc:            "Fail - Update Admin Contract As Admin (out of bounds)",
			contractAddress: adminContract,
			senderAddress:   admin.String(),
			walletLimit:     1_000_001,
			shouldErr:       true,
		},
		{
			desc:            "Fail - Update Creator Contract As Non Creator",
			contractAddress: creatorContract,
			senderAddress:   admin.String(),
			walletLimit:     1,
			shouldErr:       true,
		},
		{
			desc:            "Success - Update Creator Contract As Creator",
			contractAddress: creatorContract,
			senderAddress:   sender.String(),
			walletLimit:     21,
			shouldErr:       false,
		},
	} {
		tc := tc

		s.Run(tc.desc, func() {
			_, err := s.msgServer.UpdateFeePayContractWalletLimit(s.Ctx, &types.MsgUpdateFeePayContractWalletLimit{
				SenderAddress:   tc.senderAddress,
				ContractAddress: tc.contractAddress,
				WalletLimit:     tc.walletLimit,
			})

			if tc.shouldErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
