package testutil

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	junoapp "github.com/CosmosContracts/juno/v27/app"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmtypes "github.com/cometbft/cometbft/types"
	cosmosdb "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cosmoserver "github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func GenesisStateWithValSet(
	codec codec.Codec,
	genesisState map[string]json.RawMessage,
	valSet *tmtypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) junoapp.GenesisState {
	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction
	initValPowers := []abci.ValidatorUpdate{}

	for _, val := range valSet.Validators {
		pk, _ := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		pkAny, _ := codectypes.NewAnyWithValue(pk)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
			MinSelfDelegation: sdkmath.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress().String(), sdk.ValAddress(val.Address).String(), sdkmath.LegacyOneDec()))

		// add initial validator powers so consumer InitGenesis runs correctly
		pub, _ := val.ToProto()
		initValPowers = append(initValPowers, abci.ValidatorUpdate{
			Power:  val.VotingPower,
			PubKey: pub.PubKey,
		})
	}
	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = codec.MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = codec.MustMarshalJSON(bankGenesis)

	_, err := tmtypes.PB2TM.ValidatorUpdates(initValPowers)
	if err != nil {
		panic("failed to get vals")
	}

	return genesisState
}

var defaultGenesisStateBytes = []byte{}

// DebugAppOptions is a stub implementing AppOptions
type DebugAppOptions struct{}

// Get implements AppOptions
func (ao DebugAppOptions) Get(o string) interface{} {
	if o == cosmoserver.FlagTrace {
		return true
	}
	return nil
}

func IsDebugLogEnabled() bool {
	return os.Getenv("JUNO_KEEPER_DEBUG") != ""
}

// GenerateValidatorSet creates a ValidatorSet with n validators.
func GenerateValidatorSet(n int) *tmtypes.ValidatorSet {
	validators := make([]*tmtypes.Validator, n)
	for i := 0; i < n; i++ {
		pv := mock.NewPV()
		pubKey, _ := pv.GetPubKey()
		validator := tmtypes.NewValidator(pubKey, 1)
		validators[i] = validator
	}
	return tmtypes.NewValidatorSet(validators)
}

func GenerateGenesisAccounts(numAccounts int) []authtypes.GenesisAccount {
	genAccs := make([]authtypes.GenesisAccount, numAccounts)
	for i := 0; i < numAccounts; i++ {
		senderPrivKey := secp256k1.GenPrivKey()
		acc := authtypes.NewBaseAccountWithAddress(senderPrivKey.PubKey().Address().Bytes())
		genAccs[i] = acc
	}
	return genAccs
}

func SetupWithCustomHomeAndChainId(isCheckTx bool, dir, chainId string, t *testing.T, simultaneously bool) *junoapp.App {
	if simultaneously {
		t.Helper()
	}

	db := cosmosdb.NewMemDB()
	var (
		l       log.Logger
		appOpts servertypes.AppOptions
	)
	if IsDebugLogEnabled() {
		appOpts = DebugAppOptions{}
	} else {
		appOpts = sims.EmptyAppOptions{}
	}

	t.Log("Using test environment logger")
	l = log.NewTestLogger(t)

	app := junoapp.New(
		l,
		db,
		nil,
		true,
		dir,
		appOpts,
		[]wasmkeeper.Option{},
		baseapp.SetChainID(chainId),
	)
	if !isCheckTx {
		if len(defaultGenesisStateBytes) == 0 {
			var err error
			valSet := GenerateValidatorSet(1)
			genesisState := junoapp.NewDefaultGenesisState(app.AppCodec())
			genAccs := GenerateGenesisAccounts(1)
			authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
			genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)
			balances := []banktypes.Balance{}
			genesisState = GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, genAccs, balances...)
			defaultGenesisStateBytes, err = json.Marshal(genesisState)
			if err != nil {
				panic(err)
			}
		}

		_, err := app.InitChain(
			&abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: sims.DefaultConsensusParams,
				AppStateBytes:   defaultGenesisStateBytes,
				ChainId:         chainId,
			},
		)
		if err != nil {
			panic(err)
		}
	}

	return app
}

// Setup initializes a new App.
func Setup(isCheckTx bool, t *testing.T, simultaneously bool) *junoapp.App {
	return SetupWithCustomHomeAndChainId(isCheckTx, junoapp.DefaultNodeHome, "juno-1", t, simultaneously)
}

// SetupTestingAppWithLevelDb initializes a new App intended for testing,
// with LevelDB as a db.
func SetupTestingAppWithLevelDb(isCheckTx bool) (app *junoapp.App, cleanupFn func()) {
	dir, err := os.MkdirTemp(os.TempDir(), "juno_leveldb_testing")
	if err != nil {
		panic(err)
	}
	db, err := cosmosdb.NewGoLevelDB("juno_leveldb_testing", dir, nil)
	if err != nil {
		panic(err)
	}

	app = junoapp.New(log.NewNopLogger(), db, nil, true, dir, sims.EmptyAppOptions{}, []wasmkeeper.Option{}, baseapp.SetChainID("juno-1"))
	if !isCheckTx {
		valSet := GenerateValidatorSet(1)
		genesisState := junoapp.NewDefaultGenesisState(app.AppCodec())
		genAccs := GenerateGenesisAccounts(1)
		authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
		genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)
		balances := []banktypes.Balance{}
		genesisState = GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, genAccs, balances...)
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		_, err = app.InitChain(
			&abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: sims.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
				ChainId:         "juno-1",
			},
		)
		if err != nil {
			panic(err)
		}
	}

	cleanupFn = func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			panic(err)
		}
	}

	return app, cleanupFn
}
