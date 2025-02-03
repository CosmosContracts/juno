package testutil

import (
	"crypto/rand"
	"fmt"
	"time"

	coreheader "cosmossdk.io/core/header"
	log2 "cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	db2 "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v27/app"
)

// SetupAddr takes a balance, prefix, and address number. Then returns the respective account address byte array.
// If prefix is left blank, it will be replaced with a random prefix.
func SetupAddr(index int) sdk.AccAddress {
	prefixBz := make([]byte, 8)
	_, _ = rand.Read(prefixBz)
	prefix := string(prefixBz)
	addr := sdk.AccAddress(fmt.Sprintf("addr%s%8d", prefix, index))
	return addr
}

func SetupAddrs(numAddrs int) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, numAddrs)
	for i := 0; i < numAddrs; i++ {
		addrs[i] = SetupAddr(i)
	}
	return addrs
}

// These are for testing msg.ValidateBasic() functions
// which need to validate for valid/invalid addresses.
// Should not be used for anything else because these addresses
// are totally uninterpretable (100% random).
func GenerateTestAddrs() (string, string) {
	pk1 := ed25519.GenPrivKey().PubKey()
	validAddr := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("").String()
	return validAddr, invalidAddr
}

// CreateTestContext creates a test context.
func CreateTestContext() sdk.Context {
	ctx, _ := CreateTestContextWithMultiStore()
	return ctx
}

// CreateTestContextWithMultiStore creates a test context and returns it together with multi store.
func CreateTestContextWithMultiStore() (sdk.Context, store.CommitMultiStore) {
	db := db2.NewMemDB()
	logger := log2.NewNopLogger()

	ms := rootmulti.NewStore(db, logger, metrics.NewNoOpMetrics())

	return sdk.NewContext(ms, tmtypes.Header{}, false, logger), ms
}

// CreateTestContext creates a test context.
func Commit(app *app.App, ctx sdk.Context) sdk.Context {
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: ctx.BlockHeight(), Time: ctx.BlockTime()})
	if err != nil {
		panic(err)
	}
	_, err = app.Commit()
	if err != nil {
		panic(err)
	}

	newBlockTime := ctx.BlockTime().Add(time.Second)

	header := ctx.BlockHeader()
	header.Time = newBlockTime
	header.Height++

	return app.BaseApp.NewUncachedContext(false, header).WithHeaderInfo(coreheader.Info{
		Height: header.Height,
		Time:   header.Time,
	})
}
