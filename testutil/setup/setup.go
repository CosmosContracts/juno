package setup

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/log"
	cosmosdb "github.com/cosmos/cosmos-db"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	junoapp "github.com/CosmosContracts/juno/v27/app"
	"github.com/CosmosContracts/juno/v27/testutil/common"
)

var defaultGenesisStateBytes = []byte{}

// Setup initializes a new App.
func Setup(isCheckTx bool, homePath, chainId string, t ...*testing.T) *junoapp.App {
	db := cosmosdb.NewMemDB()
	var (
		l       log.Logger
		appOpts servertypes.AppOptions
	)
	if common.IsDebugLogEnabled() {
		appOpts = common.DebugAppOptions{}
	} else {
		appOpts = sims.EmptyAppOptions{}
	}

	if len(t) > 0 {
		testEnv := t[0]
		testEnv.Log("Using test environment logger")
		l = log.NewTestLogger(testEnv)
	} else {
		l = log.NewNopLogger()
	}

	app := junoapp.New(
		l,
		db,
		nil,
		true,
		homePath,
		appOpts,
		[]wasmkeeper.Option{},
		baseapp.SetChainID(chainId),
	)
	if !isCheckTx {
		if len(defaultGenesisStateBytes) == 0 {
			var err error
			valSet := common.GenerateValidatorSet(1)
			genesisState := junoapp.NewDefaultGenesisState(app.AppCodec())
			genAccs := common.GenerateGenesisAccounts(1)
			authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
			genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)
			balances := []banktypes.Balance{}
			genesisState = common.GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, genAccs, balances...)
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
