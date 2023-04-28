package feeshare_test

import (
	"encoding/json"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	junoapp "github.com/CosmosContracts/juno/v15/app"

	"github.com/CosmosContracts/juno/v15/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// returns context and an app with updated mint keeper
func CreateTestApp(isCheckTx bool) (*junoapp.App, sdk.Context) {
	app := Setup(isCheckTx)

	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.MintKeeper.SetParams(ctx, types.DefaultParams())
	app.MintKeeper.SetMinter(ctx, types.DefaultInitialMinter())

	return app, ctx
}

func Setup(isCheckTx bool) *junoapp.App {
	app, genesisState := GenApp(!isCheckTx, 5)
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

func GenApp(withGenesis bool, invCheckPeriod uint) (*junoapp.App, junoapp.GenesisState) {
	db := dbm.NewMemDB()
	encCdc := junoapp.MakeEncodingConfig()
	app := junoapp.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		junoapp.GetEnabledProposals(),
		simtestutil.EmptyAppOptions{},
		junoapp.GetWasmOpts(simtestutil.EmptyAppOptions{}),
	)

	if withGenesis {
		return app, junoapp.NewDefaultGenesisState(encCdc.Marshaler)
	}

	return app, junoapp.GenesisState{}
}
