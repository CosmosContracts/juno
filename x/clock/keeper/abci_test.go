package keeper_test

import (
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	_ "embed"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v31/x/clock/types"
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

// Regression test for H1: a failing (low-address) contract must NOT cascade
// into jailing higher-address healthy contracts. Before the fix, the stale
// shared `err` out-param caused every contract sorting after the first failure
// to be jailed and skipped without executing.
func (s *KeeperTestSuite) TestEndBlockerNoCascadeJail() {
	clockKeeper := s.App.AppKeepers.ClockKeeper

	// Code id 1 = failing contract (burn does not handle the EndBlock sudo msg).
	s.StoreCode(burnContract)
	// Code id 2 = healthy clock contract.
	clockCodeID := s.storeCodeReturnID(clockContract)
	s.Require().Equal(uint64(2), clockCodeID)

	// Raise the cap so we can register enough contracts.
	s.setMaxContracts(1000)

	// Register one failing contract.
	failAddr := s.instantiateAndRegisterCode(1)

	// Register healthy contracts until at least one sorts AFTER the failing
	// contract (store iteration order is lexicographic over the bech32 key),
	// which is precisely the case the cascade bug would have mis-jailed.
	healthyAddrs := []string{}
	for {
		healthyAddrs = append(healthyAddrs, s.instantiateAndRegisterCode(clockCodeID))

		hasHigher := false
		for _, h := range healthyAddrs {
			if h > failAddr {
				hasHigher = true
				break
			}
		}
		if hasHigher && len(healthyAddrs) >= 3 {
			break
		}
	}

	// Advance past gentx skip height and run the end blocker.
	s.Ctx = s.Ctx.WithBlockHeight(11)
	s.EndBlock()

	// The failing contract is jailed.
	fc, err := clockKeeper.GetClockContract(s.Ctx, failAddr)
	s.Require().NoError(err)
	s.Require().True(fc.IsJailed, "failing contract should be jailed")

	// Every healthy contract executed exactly once and is NOT jailed —
	// including those sorting after the failing contract.
	for _, a := range healthyAddrs {
		c, err := clockKeeper.GetClockContract(s.Ctx, a)
		s.Require().NoError(err)
		s.Require().False(c.IsJailed, "healthy contract must not be cascade-jailed: %s", a)
		s.Require().Equal(int64(1), s.queryContract(a), "healthy contract must have executed: %s", a)
	}
}

// Registration past the MaxContracts cap must be rejected (M1 DoS fix).
func (s *KeeperTestSuite) TestRegisterContractCap() {
	s.StoreCode(clockContract)

	// Set a small cap.
	s.setMaxContracts(2)

	// First two registrations succeed.
	first := s.instantiateAndRegisterCode(1)
	s.Require().NotEmpty(first)
	second := s.instantiateAndRegisterCode(1)
	s.Require().NotEmpty(second)

	// The third registration is rejected by the cap.
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	s.FundAcc(admin, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	addr := s.instantiateCode(1, sender.String(), admin.String())
	err := s.App.AppKeepers.ClockKeeper.RegisterContract(s.Ctx, admin.String(), addr)
	s.Require().ErrorIs(err, types.ErrMaxContractsRegistered)
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

	// Raise the registered-contract cap above the number under test.
	s.setMaxContracts(uint64(numContracts) + 1)

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

// storeCodeReturnID stores wasm code and returns its assigned code id.
func (s *KeeperTestSuite) storeCodeReturnID(code []byte) uint64 {
	_, _, sender := testdata.KeyTestPubAddr()
	msg := wasmtypes.MsgStoreCodeFixture(func(m *wasmtypes.MsgStoreCode) {
		m.WASMByteCode = code
		m.Sender = sender.String()
	})
	rsp, err := s.App.MsgServiceRouter().Handler(msg)(s.Ctx, msg)
	s.Require().NoError(err)
	var result wasmtypes.MsgStoreCodeResponse
	s.Require().NoError(s.App.AppCodec().Unmarshal(rsp.Data, &result))
	return result.CodeID
}

// instantiateCode instantiates the given code id with the given sender/admin
// and returns the contract address.
func (s *KeeperTestSuite) instantiateCode(codeID uint64, sender, admin string) string {
	msgInstantiate := wasmtypes.MsgInstantiateContractFixture(func(m *wasmtypes.MsgInstantiateContract) {
		m.Sender = sender
		m.Admin = admin
		m.CodeID = codeID
		m.Msg = []byte(`{}`)
	})
	resp, err := s.App.MsgServiceRouter().Handler(msgInstantiate)(s.Ctx, msgInstantiate)
	s.Require().NoError(err)
	var result wasmtypes.MsgInstantiateContractResponse
	s.Require().NoError(s.App.AppCodec().Unmarshal(resp.Data, &result))
	return result.Address
}

// instantiateAndRegisterCode instantiates the given code id under a fresh
// admin and registers it as a clock contract, returning the address.
func (s *KeeperTestSuite) instantiateAndRegisterCode(codeID uint64) string {
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	s.FundAcc(admin, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	addr := s.instantiateCode(codeID, sender.String(), admin.String())
	err := s.App.AppKeepers.ClockKeeper.RegisterContract(s.Ctx, admin.String(), addr)
	s.Require().NoError(err)
	return addr
}

// Raise/lower the MaxContracts cap
func (s *KeeperTestSuite) setMaxContracts(maxContracts uint64) {
	k := s.App.AppKeepers.ClockKeeper
	params := k.GetParams(s.Ctx)
	params.MaxContracts = maxContracts
	err := k.SetParams(s.Ctx, params)
	s.Require().NoError(err)
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
