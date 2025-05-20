package helpers

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	tokenfactorytypes "github.com/CosmosContracts/juno/v30/x/tokenfactory/types"
)

func debugOutput(t *testing.T, stdout string) {
	if true {
		t.Log(stdout)
	}
}

func CreateTokenFactoryDenom(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, subDenomName, feeCoin string) (fullDenom string) {
	// TF gas to create cost 2mil, so we set to 2.5 to be safe
	cmd := []string{
		"junod", "tx", "tokenfactory", "create-denom", user.FormattedAddress(), subDenomName,
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--gas", "2500000",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}

	if feeCoin != "" {
		cmd = append(cmd, "--fees", feeCoin)
	}

	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)

	return "factory/" + user.FormattedAddress() + "/" + subDenomName
}

func MintTokenFactoryDenom(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, admin ibc.Wallet, amount uint64, fullDenom string) {
	denom := strconv.FormatUint(amount, 10) + fullDenom

	// mint new tokens to the account
	cmd := []string{
		"junod", "tx", "tokenfactory", "mint", admin.FormattedAddress(), denom, admin.FormattedAddress(),
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)
}

func MintToTokenFactoryDenom(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, admin ibc.Wallet, toWallet ibc.Wallet, amount uint64, fullDenom string) {
	denom := strconv.FormatUint(amount, 10) + fullDenom

	receiver := toWallet.FormattedAddress()

	t.Log("minting", denom, "to", receiver)

	// mint new tokens to the account
	cmd := []string{
		"junod", "tx", "tokenfactory", "mint", admin.FormattedAddress(), denom, receiver,
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)
}

func UpdateTokenFactoryMetadata(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, admin ibc.Wallet, fullDenom, ticker, desc, exponent string) {
	u, err := strconv.ParseUint(exponent, 10, 32)
	require.NoError(t, err)
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
	require.NoError(t, err)

	// junod tx tokenfactory modify-metadata [denom] [metadata as json string]
	cmd := []string{
		"junod", "tx", "tokenfactory", "modify-metadata", fullDenom, string(metadataJSON),
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)
}

func TransferTokenFactoryAdmin(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, currentAdmin ibc.Wallet, newAdminBech32 string, fullDenom string) {
	cmd := []string{
		"junod", "tx", "tokenfactory", "change-admin", currentAdmin.FormattedAddress(), fullDenom, newAdminBech32,
		"--home", chain.HomeDir(),
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)
}

// Getters
func GetTokenFactoryAdmin(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, fullDenom string) string {
	// $BINARY q tokenfactory denom-authority-metadata $FULL_DENOM
	cmd := []string{
		"junod", "query", "tokenfactory", "denom-authority-metadata", fullDenom,
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	results := &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{}
	err = json.Unmarshal(stdout, results)
	require.NoError(t, err)

	t.Log(results)

	return results.AuthorityMetadata.Admin
}

func GetTokenFactoryDenomMetadata(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, fullDenom string) banktypes.Metadata {
	cmd := []string{
		"junod", "query", "bank", "denom-metadata", "--denom", fullDenom,
		"--output", "json",
		"--node", chain.GetRPCAddress(),
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	results := &banktypes.QueryDenomMetadataResponse{}
	err = json.Unmarshal(stdout, results)
	require.NoError(t, err)

	t.Log(results)

	return results.Metadata
}
