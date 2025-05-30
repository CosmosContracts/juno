package cosmwasm_test

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
)

// UnityContractDeploy test to ensure the contract withdraw function works as expected on chain.
// - https://github.com/CosmosContracts/cw-unity-prop
func (s *CosmWasmTestSuite) TestJunoUnityContractDeploy() {
	t := s.T()
	require := s.Require()

	// Chains
	juno := s.Chain
	nativeDenom := juno.Config().Denom

	// Users
	user := s.GetAndFundTestUser("default", 10_000_000, juno)
	withdrawUser := s.GetAndFundTestUser("withdraw", 10_000_000, juno)
	withdrawAddr := withdrawUser.FormattedAddress()

	// TEST DEPLOY (./scripts/deploy_ci.sh)
	// Upload & init unity contract with no admin in test mode
	msg := fmt.Sprintf(`{"native_denom":"%s","withdraw_address":"%s","withdraw_delay_in_days":28}`, nativeDenom, withdrawAddr)
	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(100000)))
	_, contractAddr := s.SetupContract(juno, user.KeyName(), "../../contracts/cw_unity_prop.wasm", msg, false, fees)
	t.Log("testing Unity contractAddr", contractAddr)

	// Execute to start the withdrawal countdown
	_, err := juno.ExecuteContract(s.Ctx, withdrawUser.KeyName(), contractAddr, `{"start_withdraw":{}}`, "--fees", fees.String(), "--gas", "auto")
	require.NoError(err)

	// make a query with GetUnityContractWithdrawalReadyTime
	res := GetUnityContractWithdrawalReadyTime(t, s.Ctx, juno, contractAddr)
	t.Log("WithdrawalReadyTimestamp", res.Data.WithdrawalReadyTimestamp)
}

func GetUnityContractWithdrawalReadyTime(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contract string) e2esuite.WithdrawalTimestampResponse {
	// junod query wasm contract-state smart <contract> '{"get_withdrawal_ready_time":{}}' --output json
	var res e2esuite.WithdrawalTimestampResponse
	err := chain.QueryContract(ctx, contract, e2esuite.ContractQueryMsg{GetWithdrawalReadyTime: &struct{}{}}, &res)
	require.NoError(t, err)
	return res
}
