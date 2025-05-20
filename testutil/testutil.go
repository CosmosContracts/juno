package testutil

import (
	"fmt"
	"os"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	s "github.com/stretchr/testify/suite"

	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/CosmosContracts/juno/v30/app"
	"github.com/CosmosContracts/juno/v30/cmd/junod/cmd"
	"github.com/CosmosContracts/juno/v30/testutil/common"
	"github.com/CosmosContracts/juno/v30/testutil/setup"
)

type KeeperTestHelper struct {
	s.Suite

	// defaults to false,
	// set to true if any method that potentially alters baseapp/abci is used.
	// this controls whether or not we can reuse the app instance, or have to set a new one.
	hasUsedAbci bool
	// defaults to false, set to true if we want to use a new app instance with caching enabled.
	// then on new setup test call, we just drop the current cache.
	// this is not always enabled, because some tests may take a painful performance hit due to CacheKv.
	withCaching bool

	App         *app.App
	Ctx         sdk.Context
	QueryHelper *baseapp.QueryServiceTestHelper
	TestAccs    []sdk.AccAddress
}

var (
	baseTestAccts        = []sdk.AccAddress{}
	defaultTestStartTime = time.Now().UTC()
)

func init() {
	baseTestAccts = common.CreateRandomAccounts(3)
}

// Setup sets up basic environment for suite (App, Ctx, and test accounts)
// preserves the caching enabled/disabled state.
func (suite *KeeperTestHelper) Setup() {
	sdk.DefaultBondDenom = "ujuno"
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(cmd.Bech32PrefixValAddr, cmd.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
	cfg.SetAddressVerifier(wasmtypes.VerifyAddressLen())

	suite.T().Log("Setting up KeeperTestHelper")
	dir, err := os.MkdirTemp("", "junod-test-home")
	if err != nil {
		panic(fmt.Sprintf("failed creating temporary directory: %v", err))
	}
	suite.T().Cleanup(func() {
		err = os.RemoveAll(dir)
		if err != nil {
			panic(fmt.Sprintf("failed removing temporary directory: %v", err))
		}
		suite.withCaching = false
	})
	if common.IsDebugLogEnabled() {
		suite.App = setup.Setup(false, dir, "juno-1", suite.T())
	} else {
		suite.App = setup.Setup(false, dir, "juno-1")
	}

	suite.Ctx = suite.App.NewContextLegacy(false, tmtypes.Header{Height: 1, ChainID: "juno-1", Time: defaultTestStartTime})
	if suite.withCaching {
		suite.Ctx, _ = suite.Ctx.CacheContext()
	}
	suite.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: suite.App.GRPCQueryRouter(),
		Ctx:             suite.Ctx,
	}

	suite.TestAccs = []sdk.AccAddress{}
	suite.TestAccs = append(suite.TestAccs, baseTestAccts...)

	suite.hasUsedAbci = false

	// Manually set validator signing info, otherwise we panic
	vals, err := suite.App.AppKeepers.StakingKeeper.GetAllValidators(suite.Ctx)
	if err != nil {
		panic(err)
	}
	for _, val := range vals {
		consAddr, _ := val.GetConsAddr()
		signingInfo := slashingtypes.NewValidatorSigningInfo(
			consAddr,
			suite.Ctx.BlockHeight(),
			0,
			time.Unix(0, 0),
			false,
			0,
		)
		err := suite.App.AppKeepers.SlashingKeeper.SetValidatorSigningInfo(suite.Ctx, consAddr, signingInfo)
		if err != nil {
			panic(err)
		}
	}
}

// resets the test environment
// requires that all commits go through helpers in s.
// On first reset, will instantiate a new app, with caching enabled.
// NOTE: If you are using ABCI methods, usage of Reset vs Setup has not been well tested.
// It is believed to work, but if you get an odd error, try changing the call to this for setup to sanity check.
// what's supposed to happen is a new setup call, and reset just does that in such a case.
func (suite *KeeperTestHelper) Reset() {
	if suite.hasUsedAbci || !suite.withCaching {
		suite.withCaching = true
		suite.Setup()
	} else {
		suite.Ctx = suite.App.NewContextLegacy(false, tmtypes.Header{Height: 1, ChainID: "juno-1", Time: defaultTestStartTime})
		if suite.withCaching {
			suite.Ctx, _ = suite.Ctx.CacheContext()
		}
		suite.QueryHelper = &baseapp.QueryServiceTestHelper{
			GRPCQueryRouter: suite.App.GRPCQueryRouter(),
			Ctx:             suite.Ctx,
		}
		suite.TestAccs = []sdk.AccAddress{}
		suite.TestAccs = append(suite.TestAccs, baseTestAccts...)
		suite.hasUsedAbci = false
	}
}

func (suite *KeeperTestHelper) SetupTestForInitGenesis() {
	dir, _ := os.MkdirTemp("", "junod-test-home")
	// Setting to True, leads to init genesis not running
	suite.App = setup.Setup(true, dir, "juno-1")
	suite.Ctx = suite.App.NewContextLegacy(true, tmtypes.Header{})
	suite.hasUsedAbci = true
}

func (suite *KeeperTestHelper) SetupWithLevelDB() func() {
	appSetup, cleanup := setup.SetupTestingAppWithLevelDB(false)
	suite.App = appSetup
	suite.Ctx = suite.App.NewContextLegacy(false, tmtypes.Header{Height: 1, ChainID: "juno-1", Time: defaultTestStartTime})
	if suite.withCaching {
		suite.Ctx, _ = suite.Ctx.CacheContext()
	}
	suite.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: suite.App.GRPCQueryRouter(),
		Ctx:             suite.Ctx,
	}
	suite.TestAccs = []sdk.AccAddress{}
	suite.TestAccs = append(suite.TestAccs, baseTestAccts...)
	suite.hasUsedAbci = false
	return cleanup
}
