package keeper_test

import (
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v31/x/cw-hooks/types"
)

func (s *KeeperTestSuite) firstValidatorAddr() sdk.ValAddress {
	vals, err := s.stakingKeeper.GetValidators(s.Ctx, 1)
	s.Require().NoError(err)
	s.Require().NotEmpty(vals)
	valAddr, err := sdk.ValAddressFromBech32(vals[0].GetOperator())
	s.Require().NoError(err)
	return valAddr
}

// Regression test for cw-hooks M-4: BeforeDelegationCreated must dispatch to
// registered contracts even for a FIRST delegation, where the delegation
// object does not exist yet. The previous implementation looked the delegation
// up and returned nil when it was absent, so the create event never fired.
func (s *KeeperTestSuite) TestBeforeDelegationCreatedDispatchesForFirstDelegation() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockHeight(11)

	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000_000))))

	contract := s.InstantiateContract(sender.String(), "", wasmContract)
	s.Require().NoError(s.registerContract("staking", sender.String(), contract))

	// Preseed a non-zero failure counter. A successful hook dispatch executes
	// the contract and resets the counter; if the hook early-returns without
	// dispatching (the M-4 bug), the counter stays untouched. This is a
	// contract-behaviour-independent probe of "was the hook dispatched".
	s.Require().NoError(s.App.AppKeepers.CWHooksKeeper.SetContract(s.Ctx, types.StakingPrefixKey, types.ContractInfo{
		ContractAddress: contract,
		FailureCounter:  7,
		LatestError:     "preseeded",
	}))

	valAddr := s.firstValidatorAddr()
	_, _, delAddr := testdata.KeyTestPubAddr()

	// First delegation: the delegation object does not exist yet. The create
	// event must still fire (built from the addresses directly, zero shares).
	err := s.App.AppKeepers.CWHooksKeeper.StakingHooks().BeforeDelegationCreated(s.Ctx, delAddr, valAddr)
	s.Require().NoError(err)

	// A successful dispatch reset the failure counter — proving the hook
	// dispatched to the contract for a first delegation.
	info := s.getContractInfo(contract)
	s.Require().EqualValues(0, info.FailureCounter, "BeforeDelegationCreated must dispatch for a first delegation")
	s.Require().Empty(info.LatestError)
}

// Containment regression test (cw-hooks suggested test #6): a registered hook
// contract that runs out of gas must NOT fail the staking state transition
// that triggered it. The hook returns nil, the failure is classified as
// ErrOutOfGas, and the contract's failure counter is incremented.
func (s *KeeperTestSuite) TestHookOutOfGasIsContained() {
	s.SetupTest()
	s.Ctx = s.Ctx.WithBlockHeight(11)

	_, _, sender := testdata.KeyTestPubAddr()
	s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000_000))))

	contract := s.InstantiateContract(sender.String(), "", wasmContract)
	s.Require().NoError(s.registerContract("staking", sender.String(), contract))

	// Starve the child gas meter so any execution runs out of gas.
	params := s.App.AppKeepers.CWHooksKeeper.GetParams(s.Ctx)
	params.ContractGasLimit = 1
	s.Require().NoError(s.App.AppKeepers.CWHooksKeeper.SetParams(s.Ctx, params))

	valAddr := s.firstValidatorAddr()
	_, _, delAddr := testdata.KeyTestPubAddr()

	// The hook dispatches a valid message but the contract OOGs. The staking
	// transition must not fail: the hook swallows the error and returns nil.
	err := s.App.AppKeepers.CWHooksKeeper.StakingHooks().BeforeDelegationCreated(s.Ctx, delAddr, valAddr)
	s.Require().NoError(err)

	// The failure is recorded and classified as out-of-gas.
	info := s.getContractInfo(contract)
	s.Require().EqualValues(1, info.FailureCounter)
	s.Require().True(
		strings.Contains(info.LatestError, "out of gas"),
		"expected out-of-gas classification, got: %s", info.LatestError,
	)
}

// Registration past the MaxContracts cap must be rejected (M1 DoS fix).
func (s *KeeperTestSuite) TestRegisterContractCap() {
	s.SetupTest()

	// Set a cap of 1 registered staking contract.
	params := s.App.AppKeepers.CWHooksKeeper.GetParams(s.Ctx)
	params.MaxContracts = 1
	s.Require().NoError(s.App.AppKeepers.CWHooksKeeper.SetParams(s.Ctx, params))

	ctxData := s.buildContractTestContext()

	// First registration succeeds.
	s.Require().NoError(s.registerContract("staking", ctxData.sender.String(), ctxData.contract))

	// Second (distinct) contract exceeds the cap and is rejected.
	err := s.registerContract("staking", ctxData.notAuthorized.String(), ctxData.contractWithAdmin)
	s.Require().Error(err)
	s.Require().True(strings.Contains(err.Error(), "maximum number of registered contracts"), "got: %v", err)

	// The gov module has its own independent cap, so registering there still
	// works even though staking is at capacity.
	s.Require().NoError(s.registerContract("gov", ctxData.sender.String(), ctxData.contract))
}
