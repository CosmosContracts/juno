package helpers

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	tokenfactorytypes "github.com/CosmosContracts/juno/v15/x/tokenfactory/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"
)

func debugOutput(t *testing.T, stdout string) {
	if true {
		t.Log(stdout)
	}
}

func CreateTokenFactoryDenom(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user *ibc.Wallet, subDenomName string) (fullDenom string) {
	// TF gas to create cost 2mil, so we set to 2.5 to be safe
	cmd := []string{"junod", "tx", "tokenfactory", "create-denom", subDenomName,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", user.KeyName,
		"--gas", "2500000",
		"--gas-adjustment", "2.0",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)

	return "factory/" + user.Bech32Address(chain.Config().Bech32Prefix) + "/" + subDenomName
}

func MintTokenFactoryDenom(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, admin *ibc.Wallet, amount uint64, fullDenom string) {
	denom := strconv.FormatUint(amount, 10) + fullDenom

	// mint new tokens to the account
	cmd := []string{"junod", "tx", "tokenfactory", "mint", denom,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", admin.KeyName,
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

func MintToTokenFactoryDenom(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, admin *ibc.Wallet, toWallet *ibc.Wallet, amount uint64, fullDenom string) {
	denom := strconv.FormatUint(amount, 10) + fullDenom

	receiver := toWallet.Bech32Address(chain.Config().Bech32Prefix)

	t.Log("minting", denom, "to", receiver)

	// mint new tokens to the account
	cmd := []string{"junod", "tx", "tokenfactory", "mint-to", receiver, denom,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", admin.KeyName,
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

func TransferTokenFactoryAdmin(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, currentAdmin *ibc.Wallet, newAdminBech32 string, fullDenom string) {
	cmd := []string{"junod", "tx", "tokenfactory", "change-admin", fullDenom, newAdminBech32,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", currentAdmin.KeyName,
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

// TODO:
// Getters
func GetTokenFactoryAdmin(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, fullDenom string) string {
	// $BINARY q tokenfactory denom-authority-metadata $FULL_DENOM

	// tokenfactorytypes.QueryDenomAuthorityMetadataRequest{
	// 	Denom: fullDenom,
	// }
	cmd := []string{"junod", "query", "tokenfactory", "denom-authority-metadata", fullDenom,
		"--node", chain.GetRPCAddress(),
		"--chain-id", chain.Config().ChainID,
		"--output", "json",
	}
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	debugOutput(t, string(stdout))

	// results := &tokenfactorytypes.DenomAuthorityMetadata{}
	results := &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{}
	err = json.Unmarshal(stdout, results)
	require.NoError(t, err)

	t.Log(results)

	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)

	// tokenfactorytypes.DenomAuthorityMetadata{
	// 	Admin: ...,
	// }
	return results.AuthorityMetadata.Admin
}
