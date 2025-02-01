package testutil

import (
	"crypto/rand"
	"fmt"

	log2 "cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	"cosmossdk.io/store/types"

	"github.com/cometbft/cometbft/crypto/ed25519"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	db2 "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/v27/app"
	minttypes "github.com/CosmosContracts/juno/v27/x/mint/types"
)

type KeeperTestHelper struct {
	suite.Suite

	App *app.App
	Ctx sdk.Context
	// Used for testing queries end to end.
	// You can wrap this in a module-specific QueryClient()
	// and then make calls as you would a normal GRPC client.
	QueryHelper *baseapp.QueryServiceTestHelper
}

// Setup sets up basic environment for suite (App, Ctx, and test accounts)
func (s *KeeperTestHelper) Setup() {
	s.App = Setup(false, s.T())
	ctx := s.App.BaseApp.NewUncachedContext(false, tmtypes.Header{})
	s.Ctx = ctx.WithBlockGasMeter(types.NewInfiniteGasMeter())

	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx,
	}
}

// SetupAddr takes a balance, prefix, and address number. Then returns the respective account address byte array.
// If prefix is left blank, it will be replaced with a random prefix.
func SetupAddr(index int) sdk.AccAddress {
	prefixBz := make([]byte, 8)
	_, _ = rand.Read(prefixBz)
	prefix := string(prefixBz)
	addr := sdk.AccAddress(fmt.Sprintf("addr%s%8d", prefix, index))
	return addr
}

func (s *KeeperTestHelper) SetupAddr(index int) sdk.AccAddress {
	return SetupAddr(index)
}

func SetupAddrs(numAddrs int) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, numAddrs)
	for i := 0; i < numAddrs; i++ {
		addrs[i] = SetupAddr(i)
	}
	return addrs
}

func (s *KeeperTestHelper) SetupAddrs(numAddrs int) []sdk.AccAddress {
	return SetupAddrs(numAddrs)
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
func (s *KeeperTestHelper) CreateTestContext() sdk.Context {
	ctx, _ := s.CreateTestContextWithMultiStore()
	return ctx
}

// CreateTestContextWithMultiStore creates a test context and returns it together with multi store.
func (s *KeeperTestHelper) CreateTestContextWithMultiStore() (sdk.Context, store.CommitMultiStore) {
	db := db2.NewMemDB()
	logger := log2.NewNopLogger()

	ms := rootmulti.NewStore(db, logger, metrics.NewNoOpMetrics())

	return sdk.NewContext(ms, tmtypes.Header{}, false, logger), ms
}

// CreateTestContext creates a test context.
func (s *KeeperTestHelper) Commit() {
	// TODO: s.App.EndBlock(abci.RequestEndBlock{Height: s.Ctx.BlockHeight()})
	// oldHeight := s.Ctx.BlockHeight()
	// oldHeader := s.Ctx.BlockHeader()
	if _, err := s.App.Commit(); err != nil {
		panic(err)
	}
	// newHeader := tmtypes.Header{
	//	Height:  oldHeight + 1,
	//	ChainID: oldHeader.ChainID,
	//	Time:    oldHeader.Time.Add(time.Minute),
	//}
	//TODO: s.App.BeginBlock(abci.RequestBeginBlock{Header: newHeader})
	s.Ctx = s.App.BaseApp.NewContext(false)
}

// FundAcc funds target address with specified amount.
func (s *KeeperTestHelper) FundAcc(acc sdk.AccAddress, amounts sdk.Coins) {
	err := s.App.AppKeepers.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, amounts)
	s.Require().NoError(err)

	err = s.App.AppKeepers.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, acc, amounts)
	s.Require().NoError(err)
}
