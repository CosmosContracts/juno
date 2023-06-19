package feeshare_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	junoapp "github.com/CosmosContracts/juno/v16/app"
	"github.com/CosmosContracts/juno/v16/x/mint/types"
)

// returns context and an app with updated mint keeper
func CreateTestApp(t *testing.T, isCheckTx bool) (*junoapp.App, sdk.Context) {
	app := Setup(t, isCheckTx)

	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{
		ChainID: "testing",
	})
	if err := app.AppKeepers.MintKeeper.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}
	app.AppKeepers.MintKeeper.SetMinter(ctx, types.DefaultInitialMinter())

	return app, ctx
}

func Setup(t *testing.T, isCheckTx bool) *junoapp.App {
	app, genesisState := GenApp(t, !isCheckTx)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				Validators: []abci.ValidatorUpdate{},
				// ConsensusParams: &tmproto.ConsensusParams{},
				ConsensusParams: junoapp.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
				ChainId:         "testing",
			},
		)
	}

	return app
}

func GenApp(t *testing.T, withGenesis bool, opts ...wasm.Option) (*junoapp.App, junoapp.GenesisState) {
	db := dbm.NewMemDB()
	nodeHome := t.TempDir()
	snapshotDir := filepath.Join(nodeHome, "data", "snapshots")

	snapshotDB, err := dbm.NewDB("metadata", dbm.GoLevelDBBackend, snapshotDir)
	require.NoError(t, err)
	t.Cleanup(func() { snapshotDB.Close() })
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	require.NoError(t, err)

	app := junoapp.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		wasm.EnableAllProposals,
		simtestutil.EmptyAppOptions{},
		opts,
		bam.SetChainID("testing"),
		bam.SetSnapshot(snapshotStore, snapshottypes.SnapshotOptions{KeepRecent: 2}),
	)

	if withGenesis {
		return app, junoapp.NewDefaultGenesisState(app.AppCodec())
	}

	return app, junoapp.GenesisState{}
}
