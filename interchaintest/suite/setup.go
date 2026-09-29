package suite

import (
	"math/rand/v2"
	"sync"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	interchaintest "github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"

	testutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	clocktypes "github.com/CosmosContracts/juno/v31/x/clock/types"
	driptypes "github.com/CosmosContracts/juno/v31/x/drip/types"
	feemarkettypes "github.com/CosmosContracts/juno/v31/x/feemarket/types"
	feepaytypes "github.com/CosmosContracts/juno/v31/x/feepay/types"
	feesharetypes "github.com/CosmosContracts/juno/v31/x/feeshare/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v31/x/tokenfactory/types"
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
	DefaultMinBaseGasPrice  = sdkmath.LegacyMustNewDecFromStr("0.002")
	DefaultBaseGasPrice     = sdkmath.LegacyMustNewDecFromStr("1")
	DefaultNoHostMount      = false

	JunoRepo, JunoVersion = GetDockerImageInfo()
	JunoImage             = ibc.DockerImage{
		Repository: JunoRepo,
		Version:    JunoVersion,
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
			// Mirrors mainnet juno-1 (25M). The previous 5M cap was below the
			// gas-limit produced by `--gas auto --gas-adjustment 3` for larger
			// wasm-store payloads (cw721_base.wasm.gz lands ~6M after the ×3
			// multiplier and was failing with code 41 "invalid gas limit").
			// feemarket.MaxBlockUtilization stays at 5M — that's the dynamic-fee
			// target, not a hard cap, and tests assert on --fees paid rather
			// than gas consumed.
			Key:   "consensus.params.block.max_gas",
			Value: "25000000",
		},
		{
			Key:   "consensus.params.abci.vote_extensions_enable_height",
			Value: "2",
		},
		{
			// Enable feepay so the fees suite's TestFeePay can register a
			// contract and exercise the zero-fee execute path. With feepay
			// disabled, IsValidFeePayTransaction short-circuits to false and
			// the v30 feemarket ante rejects --fees 0 with "no fee coin
			// provided". Other suites don't register feepay contracts so
			// this default doesn't change their behavior.
			Key:   "app_state.feepay.params.enable_feepay",
			Value: true,
		},
		{
			// this resembles the params from the v30 upgrade handler for prod mirroring e2e tests
			// max block utilization is set to 1M gas to reach target block utilization in tests easier
			Key: "app_state.feemarket.params",
			Value: feemarkettypes.Params{
				Alpha:               sdkmath.LegacyMustNewDecFromStr("0.004"),
				Beta:                sdkmath.LegacyMustNewDecFromStr("0.983"),
				Gamma:               sdkmath.LegacyMustNewDecFromStr("0.2"),
				Delta:               sdkmath.LegacyMustNewDecFromStr("0.00000000000125"),
				MinBaseGasPrice:     sdkmath.LegacyMustNewDecFromStr("0.075"),
				MinLearningRate:     sdkmath.LegacyMustNewDecFromStr("0.0015"),
				MaxLearningRate:     sdkmath.LegacyMustNewDecFromStr("0.05"),
				MaxBlockUtilization: 5000000,
				Window:              60,
				FeeDenom:            DefaultDenom,
				Enabled:             true,
				DistributeFees:      true,
			},
		},
		{
			Key: "app_state.feemarket.state",
			Value: feemarkettypes.State{
				BaseGasPrice: sdkmath.LegacyMustNewDecFromStr("0.075"),
				LearningRate: sdkmath.LegacyMustNewDecFromStr("0.0015"),
				Window:       make([]uint64, 60),
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
		GasPrices:      "",
		GasAdjustment:  3,
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
		Version:       JunoVersion,
		NoHostMount:   &DefaultNoHostMount,
		ChainConfig:   DefaultConfig,
	}

	DefaultTxCfg = TestTxConfig{
		SmallSendsNum: 1,
		LargeSendsNum: 400,
	}
)

func init() {
	sdk.GetConfig().SetBech32PrefixForAccount("juno", "juno")
	sdk.GetConfig().SetBech32PrefixForValidator("junovaloper", "juno")
	sdk.GetConfig().SetBech32PrefixForConsensusNode("junovalcons", "juno")
	sdk.GetConfig().SetCoinType(118)
}

// MakeJunoEncoding registers the Juno specific module codecs so that the associated types and msgs
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
