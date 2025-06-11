package suite

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
)

// ConformanceCosmWasm validates that store, instantiate, execute, and query work on a CosmWasm contract.
func (s *E2ETestSuite) ConformanceCosmWasm(chain *cosmos.CosmosChain, user ibc.Wallet) {
	s.StdExecute(chain, user)
	s.subMsg(chain, user)
}

func (s *E2ETestSuite) StdExecute(chain *cosmos.CosmosChain, user ibc.Wallet) (contractAddr string) {
	t := s.T()
	fees := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(100000)))
	_, contractAddr = s.SetupContract(chain, user.KeyName(), "../../contracts/cw_template.wasm", `{"count":0}`, false, fees)
	tx, err := s.ExecuteMsgWithFeeReturn(chain, user, contractAddr, "", `{"increment":{}}`, false, fees)
	require.NoError(t, err)
	t.Log(tx)

	var res GetCountResponse
	err = s.SmartQueryString(chain, contractAddr, `{"get_count":{}}`, &res)
	require.NoError(t, err)
	require.Equal(t, int64(1), res.Data.Count)

	return contractAddr
}

func (s *E2ETestSuite) subMsg(chain *cosmos.CosmosChain, user ibc.Wallet) {
	// ref: https://github.com/CosmWasm/wasmd/issues/1735
	require := s.Require()
	fees := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(100000)))

	// === execute a contract sub message ===
	_, senderContractAddr := s.SetupContract(chain, user.KeyName(), "../../contracts/cw721_base.wasm.gz", fmt.Sprintf(`{"name":"Reece #00001", "symbol":"juno-reece-test-#00001", "minter":"%s"}`, user.FormattedAddress()), false, fees)
	_, receiverContractAddr := s.SetupContract(chain, user.KeyName(), "../../contracts/cw721_receiver.wasm.gz", `{}`, false, fees)

	// mint a token
	res, err := s.ExecuteMsgWithFeeReturn(
		chain,
		user,
		senderContractAddr,
		"10000"+chain.Config().Denom,
		fmt.Sprintf(`{"mint":{"token_id":"00000", "owner":"%s"}}`, user.FormattedAddress()),
		true,
		fees,
	)
	fmt.Println("First", res)
	require.NoError(err)

	success := "InN1Y2NlZWQi"
	res3, err := s.ExecuteMsgWithFeeReturn(chain, user, senderContractAddr, "", fmt.Sprintf(`{"send_nft": { "contract": "%s", "token_id": "00000", "msg": "%s" }}`, receiverContractAddr, success), false, fees)
	require.NoError(err)
	fmt.Println("Third", res3)
	require.EqualValues(0, res3.Code)
	require.NotContains(res3.RawLog, "unknown message from the contract")
}
