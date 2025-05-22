package suite

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	testutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	clocktypes "github.com/CosmosContracts/juno/v30/x/clock/types"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	feepaytypes "github.com/CosmosContracts/juno/v30/x/feepay/types"
	feesharetypes "github.com/CosmosContracts/juno/v30/x/feeshare/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v30/x/tokenfactory/types"
)

const (
	EnvKeepAlive = "JUNO_KEEP_ALIVE"
	InitBalance  = 30000000000000
)

var (
	VotingPeriod     = "10s"
	MaxDepositPeriod = "10s"
	Denom            = "ujuno"
	NumValidators    = 4
	NumFullNodes     = 1
	MinBaseGasPrice  = sdkmath.LegacyMustNewDecFromStr("0.00100000000000000")
	BaseGasPrice     = sdkmath.LegacyMustNewDecFromStr("0.002500000000000000")

	noHostMount = false

	JunoRepo              = "ghcr.io/cosmoscontracts/juno"
	junoRepo, junoVersion = GetDockerImageInfo()
	JunoImage             = ibc.DockerImage{
		Repository: junoRepo,
		Version:    junoVersion,
		UIDGID:     "1025:1025",
	}

	defaultGenesisKV = []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.params.voting_period",
			Value: VotingPeriod,
		},
		{
			Key:   "app_state.gov.params.max_deposit_period",
			Value: MaxDepositPeriod,
		},
		{
			Key:   "app_state.gov.params.min_deposit.0.denom",
			Value: Denom,
		},
		{
			Key:   "app_state.feepay.params.enable_feepay",
			Value: false,
		},
		{
			Key: "app_state.feemarket.params",
			Value: feemarkettypes.Params{
				Alpha:               feemarkettypes.DefaultAlpha,
				Beta:                feemarkettypes.DefaultBeta,
				Gamma:               feemarkettypes.DefaultAIMDGamma,
				Delta:               feemarkettypes.DefaultDelta,
				MinBaseGasPrice:     MinBaseGasPrice,
				MinLearningRate:     feemarkettypes.DefaultMinLearningRate,
				MaxLearningRate:     feemarkettypes.DefaultMaxLearningRate,
				MaxBlockUtilization: 15_000_000,
				Window:              feemarkettypes.DefaultWindow,
				FeeDenom:            Denom,
				Enabled:             true,
				DistributeFees:      false,
			},
		},
		{
			Key: "app_state.feemarket.state",
			Value: feemarkettypes.State{
				BaseGasPrice: BaseGasPrice,
				LearningRate: feemarkettypes.DefaultMaxLearningRate,
				Window:       make([]uint64, feemarkettypes.DefaultWindow),
				Index:        0,
			},
		},
	}

	Config = ibc.ChainConfig{
		Type:           "cosmos",
		Name:           "juno",
		ChainID:        "juno-2",
		Images:         []ibc.DockerImage{JunoImage},
		Bin:            "junod",
		Bech32Prefix:   "juno",
		Denom:          Denom,
		CoinType:       "118",
		GasPrices:      fmt.Sprintf("10%s", Denom),
		GasAdjustment:  2.0,
		TrustingPeriod: "112h",
		NoHostMount:    noHostMount,
		EncodingConfig: MakeJunoEncoding(),
		ModifyGenesis:  cosmos.ModifyGenesis(defaultGenesisKV),
	}

	// interchain specification
	Spec = &interchaintest.ChainSpec{
		ChainName:     "juno",
		Name:          "juno",
		NumValidators: &NumValidators,
		NumFullNodes:  &NumFullNodes,
		Version:       "latest",
		NoHostMount:   &noHostMount,
		ChainConfig:   Config,
	}

	TxCfg = TestTxConfig{
		SmallSendsNum:          1,
		LargeSendsNum:          400,
		TargetIncreaseGasPrice: sdkmath.LegacyMustNewDecFromStr("0.1"),
	}

	ibcConfig = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "ibc-chain",
		ChainID:             "ibc-1",
		Images:              []ibc.DockerImage{JunoImage},
		Bin:                 "junod",
		Bech32Prefix:        "juno",
		Denom:               "ujuno",
		CoinType:            "118",
		GasPrices:           fmt.Sprintf("0%s", Denom),
		GasAdjustment:       2.0,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		ConfigFileOverrides: nil,
		EncodingConfig:      MakeJunoEncoding(),
		ModifyGenesis:       cosmos.ModifyGenesis(defaultGenesisKV),
	}

	genesisWalletAmount = sdkmath.NewInt(10_000_000)
)

// func init() {
// 	sdk.GetConfig().SetBech32PrefixForAccount("juno", "juno")
// 	sdk.GetConfig().SetBech32PrefixForValidator("junovaloper", "juno")
// 	sdk.GetConfig().SetBech32PrefixForConsensusNode("junovalcons", "juno")
// 	sdk.GetConfig().SetCoinType(118)
// }

// junoEncoding registers the Juno specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func MakeJunoEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feesharetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feemarkettypes.RegisterInterfaces(cfg.InterfaceRegistry)
	tokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feepaytypes.RegisterInterfaces(cfg.InterfaceRegistry)
	clocktypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}
