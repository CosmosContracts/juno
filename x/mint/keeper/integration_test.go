package keeper_test

import (
	"encoding/json"

	"github.com/CosmWasm/wasmd/x/wasm"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	junoapp "github.com/CosmosContracts/juno/v15/app"
)

func setup(isCheckTx bool) *junoapp.App {
	app, genesisState := genApp(!isCheckTx)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simtestutil.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

func genApp(withGenesis bool) (*junoapp.App, junoapp.GenesisState) {
	db := dbm.NewMemDB()
	encCdc := junoapp.MakeEncodingConfig()

	var emptyWasmOpts []wasm.Option

	app := junoapp.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		wasm.EnableAllProposals,
		simtestutil.EmptyAppOptions{},
		emptyWasmOpts,
	)

	if withGenesis {
		return app, junoapp.NewDefaultGenesisState(encCdc.Marshaler)
	}

	return app, junoapp.GenesisState{}
}
