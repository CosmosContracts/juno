package keeper_test

import (
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v31/x/feepay/types"
)

func (s *KeeperTestSuite) TestFeeShareInitGenesis() {
	testCases := []struct {
		name    string
		genesis types.GenesisState
	}{
		{
			"Default Genesis - FeePay Enabled",
			s.genesis,
		},
		{
			"Custom Genesis - FeePay Disabled",
			types.GenesisState{
				Params: types.Params{
					EnableFeepay: false,
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			s.SetupTest() // reset

			s.Require().NotPanics(func() {
				s.App.AppKeepers.FeePayKeeper.InitGenesis(s.Ctx, tc.genesis)
			})

			params := s.App.AppKeepers.FeePayKeeper.GetParams(s.Ctx)
			s.Require().Equal(tc.genesis.Params, params)
		})
	}
}

// TestInitGenesisBalanceValidation asserts that imported contract balances
// must be backed by the feepay module account's bank balance.
func (s *KeeperTestSuite) TestInitGenesisBalanceValidation() {
	contractAddr := "juno1qsrercqegvs4ye0yqg93knv73ye5dc3prqwd6jcdcuj8ggp6w0us66deup"

	genesisWithBalance := types.GenesisState{
		Params: types.DefaultParams(),
		FeePayContracts: []types.FeePayContract{
			{
				ContractAddress: contractAddr,
				Balance:         1_000_000,
				WalletLimit:     10,
			},
		},
	}

	s.Run("unfunded module account panics", func() {
		s.SetupTest() // reset

		s.Require().Panics(func() {
			s.App.AppKeepers.FeePayKeeper.InitGenesis(s.Ctx, genesisWithBalance)
		})
	})

	s.Run("funded module account imports cleanly", func() {
		s.SetupTest() // reset

		s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("stake", 1_000_000)))

		s.Require().NotPanics(func() {
			s.App.AppKeepers.FeePayKeeper.InitGenesis(s.Ctx, genesisWithBalance)
		})

		contract, err := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAddr)
		s.Require().NoError(err)
		s.Require().Equal(uint64(1_000_000), contract.Balance)
	})
}

func (s *KeeperTestSuite) TestInitGenesisSupportsMaxUint64Balance() {
	s.SetupTest()
	contractAddr := sdk.AccAddress([]byte("12345678901234567890")).String()
	maxAmount := sdkmath.NewIntFromUint64(math.MaxUint64)
	s.FundModuleAcc(types.ModuleName, sdk.NewCoins(sdk.NewCoin("stake", maxAmount)))
	genesis := types.GenesisState{
		Params: types.DefaultParams(),
		FeePayContracts: []types.FeePayContract{{
			ContractAddress: contractAddr,
			Balance:         math.MaxUint64,
		}},
	}

	s.Require().NotPanics(func() {
		s.App.AppKeepers.FeePayKeeper.InitGenesis(s.Ctx, genesis)
	})
	contract, err := s.App.AppKeepers.FeePayKeeper.GetContract(s.Ctx, contractAddr)
	s.Require().NoError(err)
	s.Require().Equal(uint64(math.MaxUint64), contract.Balance)
	moduleAddr := s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName)
	s.Require().Equal(maxAmount, s.bankKeeper.GetBalance(s.Ctx, moduleAddr, "stake").Amount)
}
