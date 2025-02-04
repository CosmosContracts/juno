package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/CosmosContracts/juno/v27/testutil"
	"github.com/CosmosContracts/juno/v27/x/mint/keeper"
	"github.com/CosmosContracts/juno/v27/x/mint/types"
)

type KeeperTestSuite struct {
	testutil.KeeperTestHelper

	mintKeeper    keeper.Keeper
	accountKeeper authkeeper.AccountKeeper

	queryClient types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.mintKeeper = s.App.AppKeepers.MintKeeper
	s.accountKeeper = s.App.AppKeepers.AccountKeeper

	s.queryClient = types.NewQueryClient(s.QueryHelper)
}
