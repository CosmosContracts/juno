package suite

import (
	"encoding/json"
	"strconv"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *E2ETestSuite) CreateTokenFactoryDenom(chain *cosmos.CosmosChain, user ibc.Wallet, subDenomName string, fees sdk.Coins) (fullDenom string) {
	t := s.T()
	require := s.Require()
	cmd := []string{
		"tokenfactory", "create-denom", user.FormattedAddress(), subDenomName,
		"--gas", "auto",
		"--fees", fees.String(),
	}
	txHash, _ := s.ExecTx(s.Chain, user.KeyName(), false, false, cmd...)
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	s.DebugOutput(string(txRes.RawLog))

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(err)

	return "factory/" + user.FormattedAddress() + "/" + subDenomName
}

func (s *E2ETestSuite) MintTokenFactoryDenom(chain *cosmos.CosmosChain, admin ibc.Wallet, amount uint64, fullDenom string, fees sdk.Coins) {
	t := s.T()
	require := s.Require()
	denom := strconv.FormatUint(amount, 10) + fullDenom

	// mint new tokens to the account
	cmd := []string{
		"tokenfactory", "mint", admin.FormattedAddress(), denom, admin.FormattedAddress(),
		"--gas", "auto",
		"--fees", fees.String(),
	}
	txHash, _ := s.ExecTx(s.Chain, admin.KeyName(), false, false, cmd...)
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	s.DebugOutput(string(txRes.RawLog))

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(err)
}

func (s *E2ETestSuite) MintToTokenFactoryDenom(chain *cosmos.CosmosChain, admin ibc.Wallet, toWallet ibc.Wallet, amount uint64, fullDenom string, fees sdk.Coins) {
	t := s.T()
	require := s.Require()
	denom := strconv.FormatUint(amount, 10) + fullDenom
	receiver := toWallet.FormattedAddress()

	t.Log("minting", denom, "to", receiver)

	// mint new tokens to the account
	cmd := []string{
		"tokenfactory", "mint", admin.FormattedAddress(), denom, receiver,
		"--gas", "auto",
		"--fees", fees.String(),
	}
	txHash, _ := s.ExecTx(s.Chain, admin.KeyName(), false, false, cmd...)
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	s.DebugOutput(string(txRes.RawLog))

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(err)
}

func (s *E2ETestSuite) UpdateTokenFactoryMetadata(chain *cosmos.CosmosChain, admin ibc.Wallet, fullDenom, ticker, desc, exponent string, fees sdk.Coins) {
	t := s.T()
	require := s.Require()
	u, err := strconv.ParseUint(exponent, 10, 32)
	require.NoError(err)
	exp := uint32(u)

	// Build the metadata JSON following the Metadata structure
	metadata := banktypes.Metadata{
		Description: desc,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    fullDenom,
				Exponent: 0,
				Aliases:  []string{},
			},
			{
				Denom:    ticker,
				Exponent: exp,
				Aliases:  []string{},
			},
		},
		Base:    fullDenom,
		Display: ticker,
		Name:    ticker,
		Symbol:  ticker,
	}

	metadataJSON, err := json.Marshal(metadata)
	require.NoError(err)

	// junod tx tokenfactory modify-metadata [denom] [metadata as json string]
	cmd := []string{
		"tokenfactory", "modify-metadata", fullDenom, string(metadataJSON),
		"--gas", "auto",
		"--fees", fees.String(),
	}
	txHash, _ := s.ExecTx(s.Chain, admin.KeyName(), false, false, cmd...)
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	s.DebugOutput(string(txRes.RawLog))

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(err)
}

func (s *E2ETestSuite) ChangeTokenFactoryAdmin(chain *cosmos.CosmosChain, currentAdmin ibc.Wallet, newAdminBech32 string, fullDenom string, fees sdk.Coins) {
	t := s.T()
	require := s.Require()
	cmd := []string{
		"tokenfactory", "change-admin", currentAdmin.FormattedAddress(), fullDenom, newAdminBech32,
		"--gas", "auto",
		"--fees", fees.String(),
	}
	txHash, _ := s.ExecTx(s.Chain, currentAdmin.KeyName(), false, false, cmd...)
	// convert stdout into a TxResponse
	txRes, err := chain.GetTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}

	s.DebugOutput(string(txRes.RawLog))

	err = testutil.WaitForBlocks(s.Ctx, 2, chain)
	require.NoError(err)
}

func (s *E2ETestSuite) GetUserTokenFactoryBalances(chain *cosmos.CosmosChain, contract string, uaddr string) GetAllBalancesResponse {
	t := s.T()
	var res GetAllBalancesResponse
	err := chain.QueryContract(s.Ctx, contract, ContractQueryMsg{GetAllBalances: &GetAllBalancesQuery{Address: uaddr}}, &res)
	require.NoError(t, err)
	return res
}
