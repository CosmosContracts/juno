package interchaintest

import (

	// feesharetypes "github.com/CosmosContracts/juno/v15/x/feeshare/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

var (
	JunoE2ERepo  = "ghcr.io/cosmoscontracts/juno-e2e"
	JunoMainRepo = "ghcr.io/cosmoscontracts/juno"

	junoRepo, junoVersion = GetDockerImageInfo()

	JunoImage = ibc.DockerImage{
		Repository: junoRepo,
		Version:    junoVersion,
		UidGid:     "1025:1025",
	}

	junoConfig = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "juno",
		ChainID:             "juno-2",
		Images:              []ibc.DockerImage{JunoImage},
		Bin:                 "junod",
		Bech32Prefix:        "juno",
		Denom:               "ujuno",
		CoinType:            "118",
		GasPrices:           "0.0ujuno",
		GasAdjustment:       1.1,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		ModifyGenesis:       nil,
		ConfigFileOverrides: nil,
		EncodingConfig:      junoEncoding(),
	}

	pathJunoGaia        = "juno-gaia"
	genesisWalletAmount = int64(10_000_000)
)

// junoEncoding registers the Juno specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func junoEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	// feesharetypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}
