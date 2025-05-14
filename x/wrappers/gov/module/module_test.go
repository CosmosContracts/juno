package module_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	googlegrpc "google.golang.org/grpc"

	"github.com/cosmos/gogoproto/grpc"

	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
)

var _ module.Configurator = &configuratorMock{}

type grpcServerMock struct{}

func (grpcServerMock) RegisterService(_ *googlegrpc.ServiceDesc, _ any) {}

type configuratorMock struct {
	msgServer                 grpcServerMock
	queryServer               grpcServerMock
	capturedMigrationVersions []uint64
}

func newConfiguratorMock() *configuratorMock {
	msgServer := grpcServerMock{}
	queryServer := grpcServerMock{}

	return &configuratorMock{
		msgServer:   msgServer,
		queryServer: queryServer,
	}
}

func (c *configuratorMock) MsgServer() grpc.Server {
	return c.msgServer
}

func (c *configuratorMock) QueryServer() grpc.Server {
	return c.queryServer
}

func (c *configuratorMock) RegisterMigration(
	_ string, forVersion uint64, _ module.MigrationHandler,
) error {
	c.capturedMigrationVersions = append(c.capturedMigrationVersions, forVersion)
	return nil
}

func (*configuratorMock) RegisterService(_ *googlegrpc.ServiceDesc, _ any) {
}

func (*configuratorMock) Error() error {
	return nil
}

// The test checks the migration registration of the original gov module.
//
// Since we override RegisterServices we want to be sure that
// the original gov module won't have unexpected migrations
// after a Cosmos SDK version upgrade.
func TestAppModuleOriginalGov_RegisterServices(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}).Codec
	govModule := gov.NewAppModule(cdc, &govkeeper.Keeper{}, keeper.AccountKeeper{}, bankkeeper.BaseKeeper{}, nil)
	configurator := newConfiguratorMock()
	govModule.RegisterServices(configurator)
	require.Equal(t, []uint64{1, 2, 3, 4}, configurator.capturedMigrationVersions)
	require.Equal(t, uint64(5), govModule.ConsensusVersion())
}
