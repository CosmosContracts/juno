package junoconformance

import (
	"context"
	"fmt"
	"testing"

	"github.com/CosmosContracts/juno/tests/interchaintest/helpers"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/stretchr/testify/require"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// ConformanceCosmWasm validates that store, instantiate, execute, and query work on a CosmWasm contract.
func ConformanceCosmWasm(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet) {
	// std(t, ctx, chain, user)
	subMsg(t, ctx, chain, user)
}

func std(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet) {
	_, contractAddr := helpers.SetupContract(t, ctx, chain, user.KeyName(), "contracts/cw_template.wasm", `{"count":0}`)
	helpers.ExecuteMsgWithFee(t, ctx, chain, user, contractAddr, "", "10000"+chain.Config().Denom, `{"increment":{}}`)

	var res helpers.GetCountResponse
	err := helpers.SmartQueryString(t, ctx, chain, contractAddr, `{"get_count":{}}`, &res)
	require.NoError(t, err)

	require.Equal(t, int64(1), res.Data.Count)
}

func subMsg(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet) {
	// ref: https://github.com/CosmWasm/wasmd/issues/1735

	// === execute a contract sub message ===
	_, senderContractAddr := helpers.SetupContract(t, ctx, chain, user.KeyName(), "contracts/cw721_base.wasm.gz", fmt.Sprintf(`{"name":"Reece #00001", "symbol":"juno-reece-test-#00001", "minter":"%s"}`, user.FormattedAddress()))
	_, receiverContractAddr := helpers.SetupContract(t, ctx, chain, user.KeyName(), "contracts/cw721_receiver.wasm.gz", `{}`)

	// mint a token
	res, err := helpers.ExecuteMsgWithFeeReturn(t, ctx, chain, user, senderContractAddr, "", "10000"+chain.Config().Denom, fmt.Sprintf(`{"mint":{"token_id":"00000", "owner":"%s"}}`, user.FormattedAddress()))
	fmt.Println("First", res)
	require.NoError(t, err)

	// this purposely will fail with the current, we are just validating the messsage is not unknown.
	// sub message of unknown means the `wasmkeeper.WithMessageHandlerDecorator` is not setup properly.
	fail := "ImZhaWwi"
	res2, err := helpers.ExecuteMsgWithFeeReturn(t, ctx, chain, user, senderContractAddr, "", "10000"+chain.Config().Denom, fmt.Sprintf(`{"send_nft": { "contract": "%s", "token_id": "00000", "msg": "%s" }}`, receiverContractAddr, fail))
	require.NoError(t, err)
	fmt.Println("Second", res2)
	require.NotEqualValues(t, wasmtypes.ErrUnknownMsg.ABCICode(), res2.Code)
	require.NotContains(t, res2.RawLog, "unknown message from the contract")

	success := "InN1Y2NlZWQi"
	res3, err := helpers.ExecuteMsgWithFeeReturn(t, ctx, chain, user, senderContractAddr, "", "10000"+chain.Config().Denom, fmt.Sprintf(`{"send_nft": { "contract": "%s", "token_id": "00000", "msg": "%s" }}`, receiverContractAddr, success))
	require.NoError(t, err)
	fmt.Println("Third", res3)
	require.EqualValues(t, 0, res3.Code)
	require.NotContains(t, res3.RawLog, "unknown message from the contract")
}
