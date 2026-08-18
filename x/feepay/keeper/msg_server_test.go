package keeper_test

import (
	"math"

	_ "embed"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/CosmosContracts/juno/v31/x/feepay/types"
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

func (s *KeeperTestSuite) TestUnregisterFeePayContractPreservesStateWhenRefundFails() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewInt64Coin("stake", 1_000_000)))

	contractAddr := s.InstantiateContract(sender.String(), "", wasmContract)
	params, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	params.FeeDenom = "urefund"
	s.Require().NoError(s.App.AppKeepers.FeeMarketKeeper.SetParams(s.Ctx, params))

	s.registerFeePayContract(sender.String(), contractAddr, 0, 3)
	contract, err := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAddr)
	s.Require().NoError(err)
	s.App.AppKeepers.FeePayKeeper.SetContractBalance(s.Ctx, contract, 100)
	s.Require().Equal(uint64(100), contract.Balance)
	s.Require().True(s.bankKeeper.GetBalance(s.Ctx, s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName), "urefund").IsZero())
	s.Require().NoError(s.App.AppKeepers.FeePayKeeper.IncrementContractUses(s.Ctx, contract, sender.String(), 2))

	_, err = s.msgServer.UnregisterFeePayContract(s.Ctx, &types.MsgUnregisterFeePayContract{
		SenderAddress:   sender.String(),
		ContractAddress: contractAddr,
	})
	s.Require().Error(err)

	preserved, err := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAddr)
	s.Require().NoError(err)
	s.Require().Equal(uint64(100), preserved.Balance)
	uses, err := s.App.AppKeepers.FeePayKeeper.GetContractUses(s.Ctx, preserved, sender.String())
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), uses)
}

func (s *KeeperTestSuite) TestFundFeePayContract() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	// Contract instantiation consumes one stake, so fund beyond the exact
	// FeePay deposit to keep this fixture focused on denomination handling.
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(2_000_000)), sdk.NewCoin("ujuno", sdkmath.NewInt(100_000_000))))
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
			amount:          sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(100_000_000_000))),
			shouldErr:       true,
		},
		{
			desc:            "Success - Contract Funded",
			contractAddress: contract,
			senderAddress:   sender.String(),
			amount:          sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))),
			shouldErr:       false,
		},
	} {
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

func (s *KeeperTestSuite) TestConfiguredFeeDenomFundingAndUnregisterRefund() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	const feeDenom = "ufee"
	const amount = int64(1_000_000)

	params, err := s.App.AppKeepers.FeeMarketKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	params.FeeDenom = feeDenom
	s.Require().NoError(s.App.AppKeepers.FeeMarketKeeper.SetParams(s.Ctx, params))

	s.FundAcc(sender, sdk.NewCoins(
		sdk.NewInt64Coin("stake", 1_000_000),
		sdk.NewInt64Coin(feeDenom, amount),
	))
	contract := s.InstantiateContract(sender.String(), "", wasmContract)
	s.registerFeePayContract(sender.String(), contract, 0, 1)

	_, err = s.msgServer.FundFeePayContract(s.Ctx, &types.MsgFundFeePayContract{
		SenderAddress:   sender.String(),
		ContractAddress: contract,
		Amount:          sdk.NewCoins(sdk.NewInt64Coin(feeDenom, amount)),
	})
	s.Require().NoError(err)

	moduleAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName)
	s.Require().Equal(sdkmath.NewInt(amount), s.bankKeeper.GetBalance(s.Ctx, moduleAddr, feeDenom).Amount)
	funded, err := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contract)
	s.Require().NoError(err)
	s.Require().Equal(uint64(amount), funded.Balance)

	beforeRefund := s.bankKeeper.GetBalance(s.Ctx, sender, feeDenom).Amount
	_, err = s.msgServer.UnregisterFeePayContract(s.Ctx, &types.MsgUnregisterFeePayContract{
		SenderAddress:   sender.String(),
		ContractAddress: contract,
	})
	s.Require().NoError(err)
	s.Require().Equal(beforeRefund.AddRaw(amount), s.bankKeeper.GetBalance(s.Ctx, sender, feeDenom).Amount)
	s.Require().True(s.bankKeeper.GetBalance(s.Ctx, moduleAddr, feeDenom).IsZero())
}

func (s *KeeperTestSuite) TestFundFeePayContractRejectsUint64OverflowAtomically() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	contractAddr := sdk.AccAddress([]byte("12345678901234567890")).String()
	moduleAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName)

	s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, types.FeePayContract{
		ContractAddress: contractAddr,
		Balance:         math.MaxUint64 - 1,
	})
	s.FundAcc(sender, sdk.NewCoins(sdk.NewInt64Coin("stake", 2)))
	beforeSender := s.bankKeeper.GetBalance(s.Ctx, sender, "stake").Amount
	beforeModule := s.bankKeeper.GetBalance(s.Ctx, moduleAddr, "stake").Amount

	_, err := s.msgServer.FundFeePayContract(s.Ctx, &types.MsgFundFeePayContract{
		SenderAddress: sender.String(), ContractAddress: contractAddr,
		Amount: sdk.NewCoins(sdk.NewInt64Coin("stake", 2)),
	})
	s.Require().ErrorIs(err, types.ErrFeePayBalanceOverflow)

	contract, getErr := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAddr)
	s.Require().NoError(getErr)
	s.Require().Equal(uint64(math.MaxUint64-1), contract.Balance)
	s.Require().Equal(beforeSender, s.bankKeeper.GetBalance(s.Ctx, sender, "stake").Amount)
	s.Require().Equal(beforeModule, s.bankKeeper.GetBalance(s.Ctx, moduleAddr, "stake").Amount)
}

func (s *KeeperTestSuite) TestFundFeePayContractRejectsAmountAboveUint64Atomically() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	contractAddr := sdk.AccAddress([]byte("12345678901234567890")).String()
	s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, types.FeePayContract{ContractAddress: contractAddr})
	aboveMax := sdkmath.NewIntFromUint64(math.MaxUint64).AddRaw(1)
	moduleAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName)

	_, err := s.msgServer.FundFeePayContract(s.Ctx, &types.MsgFundFeePayContract{
		SenderAddress: sender.String(), ContractAddress: contractAddr,
		Amount: sdk.NewCoins(sdk.NewCoin("stake", aboveMax)),
	})
	s.Require().ErrorIs(err, types.ErrFeePayAmountOutOfRange)

	contract, getErr := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAddr)
	s.Require().NoError(getErr)
	s.Require().Zero(contract.Balance)
	s.Require().True(s.bankKeeper.GetBalance(s.Ctx, moduleAddr, "stake").IsZero())
}

func (s *KeeperTestSuite) TestFundFeePayContractSupportsMaxUint64AndPreservesBacking() {
	s.SetupTest()
	_, _, sender := testdata.KeyTestPubAddr()
	contractAddr := sdk.AccAddress([]byte("12345678901234567890")).String()
	s.App.AppKeepers.FeePayKeeper.SetFeePayContract(s.Ctx, types.FeePayContract{ContractAddress: contractAddr})
	maxAmount := sdkmath.NewIntFromUint64(math.MaxUint64)
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", maxAmount)))

	_, err := s.msgServer.FundFeePayContract(s.Ctx, &types.MsgFundFeePayContract{
		SenderAddress: sender.String(), ContractAddress: contractAddr,
		Amount: sdk.NewCoins(sdk.NewCoin("stake", maxAmount)),
	})
	s.Require().NoError(err)

	contract, getErr := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAddr)
	s.Require().NoError(getErr)
	s.Require().Equal(uint64(math.MaxUint64), contract.Balance)
	moduleAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName)
	s.Require().Equal(maxAmount, s.bankKeeper.GetBalance(s.Ctx, moduleAddr, "stake").Amount)
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
