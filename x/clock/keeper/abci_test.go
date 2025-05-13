package keeper_test

import (
	"encoding/json"

	_ "embed"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v29/x/clock/types"
)

// Register a contract. You must store the contract code before registering.
func (s *KeeperTestSuite) registerContract() string {
	// Create & fund accounts
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	s.FundAcc(admin, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	// Instantiate contract
	contractAddress := s.InstantiateContract(sender.String(), admin.String(), clockContract)

	// Register contract
	clockKeeper := s.App.AppKeepers.ClockKeeper
	err := clockKeeper.RegisterContract(s.Ctx, admin.String(), contractAddress)
	s.Require().NoError(err)

	// Assert contract is registered
	contract, err := clockKeeper.GetClockContract(s.Ctx, contractAddress)
	s.Require().NoError(err)
	s.Require().Equal(contractAddress, contract.ContractAddress)

	// Increment block height
	s.Ctx = s.Ctx.WithBlockHeight(11)

	return contract.ContractAddress
}

// Test the end blocker. This test registers a contract, executes it with enough gas,
// too little gas, and also ensures the unjailing process functions.
func (s *KeeperTestSuite) TestEndBlocker() {
	// Setup test
	clockKeeper := s.App.AppKeepers.ClockKeeper
	s.StoreCode(clockContract)
	contractAddress := s.registerContract()

	// Query contract
	val := s.queryContract(contractAddress)
	s.Require().Equal(int64(0), val)

	// Call end blocker
	s.EndBlock()

	// Query contract
	val = s.queryContract(contractAddress)
	s.Require().Equal(int64(1), val)

	// Update params with 10 gas limit
	s.updateGasLimit(65_000)

	// Call end blocker
	s.EndBlock()

	// Ensure contract is now jailed
	contract, err := clockKeeper.GetClockContract(s.Ctx, contractAddress)
	s.Require().NoError(err)
	s.Require().True(contract.IsJailed)

	// Update params to regular
	s.updateGasLimit(types.DefaultParams().ContractGasLimit)

	// Call end blocker
	s.EndBlock()

	// Unjail contract
	err = clockKeeper.SetJailStatus(s.Ctx, contractAddress, false)
	s.Require().NoError(err)

	// Ensure contract is no longer jailed
	contract, err = clockKeeper.GetClockContract(s.Ctx, contractAddress)
	s.Require().NoError(err)
	s.Require().False(contract.IsJailed)

	// Call end blocker
	s.EndBlock()

	// Query contract
	val = s.queryContract(contractAddress)
	s.Require().Equal(int64(2), val)
}

// Test a contract which does not handle the sudo EndBlock msg.
func (s *KeeperTestSuite) TestInvalidContract() {
	// Setup test
	clockKeeper := s.App.AppKeepers.ClockKeeper
	s.StoreCode(burnContract)
	contractAddress := s.registerContract()

	// Run the end blocker
	s.EndBlock()

	// Ensure contract is now jailed
	contract, err := clockKeeper.GetClockContract(s.Ctx, contractAddress)
	s.Require().NoError(err)
	s.Require().True(contract.IsJailed)
}

// Test the endblocker with numerous contracts that all panic
func (s *KeeperTestSuite) TestPerformance() {
	s.StoreCode(burnContract)

	numContracts := 1000

	// Register numerous contracts
	for x := 0; x < numContracts; x++ {
		// Register contract
		_ = s.registerContract()
	}

	// Ensure contracts exist
	clockKeeper := s.App.AppKeepers.ClockKeeper
	contracts, err := clockKeeper.GetAllContracts(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(contracts, numContracts)

	// Call end blocker
	s.EndBlock()

	// Ensure contracts are jailed
	contracts, err = clockKeeper.GetAllContracts(s.Ctx)
	s.Require().NoError(err)
	for _, contract := range contracts {
		s.Require().True(contract.IsJailed)
	}
}

// Update the gas limit
func (s *KeeperTestSuite) updateGasLimit(gasLimit uint64) {
	params := types.DefaultParams()
	params.ContractGasLimit = gasLimit
	k := s.App.AppKeepers.ClockKeeper

	store := runtime.KVStoreAdapter(k.GetStoreService().OpenKVStore(s.Ctx))
	bz := k.GetCdc().MustMarshal(&params)
	store.Set(types.ParamsKey, bz)

	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
}

// Query the clock contract
func (s *KeeperTestSuite) queryContract(contractAddress string) int64 {
	query := `{"get_config":{}}`
	output, err := s.App.AppKeepers.WasmKeeper.QuerySmart(s.Ctx, sdk.MustAccAddressFromBech32(contractAddress), []byte(query))
	s.Require().NoError(err)

	var val struct {
		Val int64 `json:"val"`
	}

	err = json.Unmarshal(output, &val)
	s.Require().NoError(err)

	return val.Val
}
