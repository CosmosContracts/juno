package suite

import (
	"context"
	"encoding/json"

	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

// BANK

func (s *E2ETestSuite) QueryBankBalance(user ibc.Wallet) sdk.Coin {
	s.T().Helper()

	resp, err := s.QueryClients.BankClient.Balance(context.Background(), &banktypes.QueryBalanceRequest{
		Address: user.FormattedAddress(),
		Denom:   DefaultDenom,
	})
	s.Require().NoError(err)
	s.Require().NotNil(*resp.Balance)

	return *resp.Balance
}

// FEEMARKET

func (s *E2ETestSuite) QueryFeemarketParams() feemarkettypes.Params {
	s.T().Helper()

	resp, err := s.QueryClients.FeemarketClient.Params(context.Background(), &feemarkettypes.ParamsRequest{})
	s.Require().NoError(err)

	return resp.Params
}

func (s *E2ETestSuite) QueryFeemarketState() feemarkettypes.State {
	s.T().Helper()

	resp, err := s.QueryClients.FeemarketClient.State(context.Background(), &feemarkettypes.StateRequest{})
	s.Require().NoError(err)

	return resp.State
}

func (s *E2ETestSuite) QueryFeemarketGasPrice(denom string) sdk.DecCoin {
	s.T().Helper()

	resp, err := s.QueryClients.FeemarketClient.GasPrice(s.Ctx, &feemarkettypes.GasPriceRequest{
		Denom: denom,
	})
	s.Require().NoError(err)

	return resp.GetPrice()
}

// STAKING

// QueryValidators queries for all the network's validators
func (s *E2ETestSuite) QueryValidators(chain *cosmos.CosmosChain) []sdk.ValAddress {
	s.T().Helper()

	// query validators
	resp, err := s.QueryClients.StakingClient.Validators(s.Ctx, &stakingtypes.QueryValidatorsRequest{})
	s.Require().NoError(err)
	addrs := make([]sdk.ValAddress, len(resp.Validators))

	// unmarshal validators
	for i, val := range resp.Validators {
		addrBz, err := sdk.GetFromBech32(val.OperatorAddress, chain.Config().Bech32Prefix+sdk.PrefixValidator+sdk.PrefixOperator)
		s.Require().NoError(err)

		addrs[i] = sdk.ValAddress(addrBz)
	}
	return addrs
}

// NODE

func (s *E2ETestSuite) QueryAccountSequence(chain *cosmos.CosmosChain, address string) uint64 {
	s.T().Helper()

	// get nodes
	nodes := chain.Nodes()
	s.Require().True(len(nodes) > 0)

	resp, _, err := nodes[0].ExecQuery(context.Background(), "auth", "account", address)
	s.Require().NoError(err)
	// unmarshal json response
	var accResp codectypes.Any
	s.Require().NoError(json.Unmarshal(resp, &accResp))

	// unmarshal into baseAccount
	var acc authtypes.BaseAccount
	s.Require().NoError(acc.Unmarshal(accResp.Value))

	return acc.GetSequence()
}

// Block returns the block at the given height
func (s *E2ETestSuite) QueryBlock(chain *cosmos.CosmosChain, height int64) *coretypes.ResultBlock {
	s.T().Helper()

	// get nodes
	nodes := chain.Nodes()
	s.Require().True(len(nodes) > 0)

	client := nodes[0].Client

	resp, err := client.Block(context.Background(), &height)
	s.Require().NoError(err)

	return resp
}
