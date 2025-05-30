package suite

import (
	"fmt"
	"math/rand/v2"
	"sync"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	testutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	clocktypes "github.com/CosmosContracts/juno/v30/x/clock/types"
	driptypes "github.com/CosmosContracts/juno/v30/x/drip/types"
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
	random *rand.Rand
	mu     sync.Mutex

	DefaultVotingPeriod     = "10s"
	DefaultMaxDepositPeriod = "10s"
	DefaultDenom            = "ujuno"
	DefaultAuthority        = "juno10d07y265gmmuvt4z0w9aw880jnsr700jvss730"
	DefaultNumValidators    = 1
	DefaultNumFullNodes     = 0
	DefaultHaltHeightDelta  = int64(9) // will propose upgrade this many blocks in the future
	DefaultMinBaseGasPrice  = sdkmath.LegacyMustNewDecFromStr("0.001")
	DefaultBaseGasPrice     = sdkmath.LegacyMustNewDecFromStr("0.01")
	DefaultNoHostMount      = false

	JunoRepo              = "ghcr.io/cosmoscontracts/juno"
	HubRepo               = "ghcr.io/cosmos/gaia"
	junoRepo, junoVersion = GetDockerImageInfo()
	JunoImage             = ibc.DockerImage{
		Repository: junoRepo,
		Version:    junoVersion,
		UIDGID:     "1025:1025",
	}
	HubImage = ibc.DockerImage{
		Repository: HubRepo,
		Version:    "v23.3.0",
		UIDGID:     "1025:1025",
	}

	DefaultGenesisKV = []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.params.voting_period",
			Value: DefaultVotingPeriod,
		},
		{
			Key:   "app_state.gov.params.max_deposit_period",
			Value: DefaultMaxDepositPeriod,
		},
		{
			Key:   "app_state.gov.params.min_deposit.0.denom",
			Value: DefaultDenom,
		},
		{
			Key:   "app_state.cw-hooks.params.contract_gas_limit",
			Value: 500000,
		},
		{
			Key:   "consensus.params.block.max_gas",
			Value: "100000000",
		},
		{
			Key:   "consensus.params.abci.vote_extensions_enable_height",
			Value: "2",
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
				MinBaseGasPrice:     DefaultMinBaseGasPrice,
				MinLearningRate:     feemarkettypes.DefaultMinLearningRate,
				MaxLearningRate:     feemarkettypes.DefaultMaxLearningRate,
				MaxBlockUtilization: 100_000_000,
				Window:              feemarkettypes.DefaultWindow,
				FeeDenom:            DefaultDenom,
				Enabled:             false,
				DistributeFees:      false,
			},
		},
		{
			Key: "app_state.feemarket.state",
			Value: feemarkettypes.State{
				BaseGasPrice: DefaultBaseGasPrice,
				LearningRate: feemarkettypes.DefaultMaxLearningRate,
				Window:       make([]uint64, feemarkettypes.DefaultWindow),
				Index:        0,
			},
		},
	}

	DefaultConfig = ibc.ChainConfig{
		Type:           "cosmos",
		Name:           "juno",
		ChainID:        "juno-2",
		Images:         []ibc.DockerImage{JunoImage},
		Bin:            "junod",
		Bech32Prefix:   "juno",
		Denom:          DefaultDenom,
		Gas:            "auto",
		CoinType:       "118",
		GasPrices:      fmt.Sprintf("%v%s", DefaultBaseGasPrice, DefaultDenom),
		GasAdjustment:  10.0,
		TrustingPeriod: "112h",
		NoHostMount:    DefaultNoHostMount,
		EncodingConfig: MakeJunoEncoding(),
		ModifyGenesis:  cosmos.ModifyGenesis(DefaultGenesisKV),
	}
	// interchain specification
	DefaultSpec = &interchaintest.ChainSpec{
		ChainName:     "juno",
		Name:          "juno",
		NumValidators: &DefaultNumValidators,
		NumFullNodes:  &DefaultNumFullNodes,
		Version:       junoVersion,
		NoHostMount:   &DefaultNoHostMount,
		ChainConfig:   DefaultConfig,
	}

	DefaultTxCfg = TestTxConfig{
		SmallSendsNum:          1,
		LargeSendsNum:          325,
		TargetIncreaseGasPrice: sdkmath.LegacyMustNewDecFromStr("0.0011"),
	}
)

func init() {
	sdk.GetConfig().SetBech32PrefixForAccount("juno", "juno")
	sdk.GetConfig().SetBech32PrefixForValidator("junovaloper", "juno")
	sdk.GetConfig().SetBech32PrefixForConsensusNode("junovalcons", "juno")
	sdk.GetConfig().SetCoinType(118)
}

// junoEncoding registers the Juno specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func MakeJunoEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feesharetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feemarkettypes.RegisterInterfaces(cfg.InterfaceRegistry)
	driptypes.RegisterInterfaces(cfg.InterfaceRegistry)
	tokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feepaytypes.RegisterInterfaces(cfg.InterfaceRegistry)
	clocktypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}
