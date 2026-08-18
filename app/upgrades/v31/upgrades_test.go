package v31

import (
	"context"
	"errors"
	"testing"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/stretchr/testify/require"

	clocktypes "github.com/CosmosContracts/juno/v31/x/clock/types"
	cwhookstypes "github.com/CosmosContracts/juno/v31/x/cw-hooks/types"
)

type clockParamsStoreStub struct {
	params clocktypes.Params
}

func (s *clockParamsStoreStub) GetParams(context.Context) clocktypes.Params { return s.params }
func (s *clockParamsStoreStub) SetParams(_ context.Context, params clocktypes.Params) error {
	s.params = params
	return nil
}

type cwHooksParamsStoreStub struct {
	params cwhookstypes.Params
}

func (s *cwHooksParamsStoreStub) Get(context.Context) (cwhookstypes.Params, error) {
	return s.params, nil
}
func (s *cwHooksParamsStoreStub) Set(_ context.Context, params cwhookstypes.Params) error {
	s.params = params
	return nil
}

func TestMigrateLegacyContractCaps(t *testing.T) {
	clockStore := &clockParamsStoreStub{params: clocktypes.Params{ContractGasLimit: 250_000}}
	cwHooksStore := &cwHooksParamsStoreStub{params: cwhookstypes.Params{
		ContractGasLimit:                250_000,
		ContractFailureRemovalThreshold: 3,
	}}

	err := migrateLegacyContractCaps(context.Background(), clockStore, cwHooksStore)

	require.NoError(t, err)
	require.Equal(t, clocktypes.DefaultMaxContracts, clockStore.params.MaxContracts)
	require.Equal(t, uint64(250_000), clockStore.params.ContractGasLimit)
	require.Equal(t, cwhookstypes.DefaultMaxContracts, cwHooksStore.params.MaxContracts)
	require.Equal(t, uint64(250_000), cwHooksStore.params.ContractGasLimit)
	require.Equal(t, uint64(3), cwHooksStore.params.ContractFailureRemovalThreshold)
}

type migrationRunnerStub struct {
	gotVersionMap module.VersionMap
	versionMap    module.VersionMap
	err           error
}

func (s *migrationRunnerStub) RunMigrations(_ context.Context, _ module.Configurator, vm module.VersionMap) (module.VersionMap, error) {
	s.gotVersionMap = vm
	return s.versionMap, s.err
}

func TestUpgradeIdentityAndStoreUpgrades(t *testing.T) {
	require.Equal(t, "v31", UpgradeName)
	require.Equal(t, UpgradeName, Upgrade.UpgradeName)
	require.Equal(t, storetypes.StoreUpgrades{}, Upgrade.StoreUpgrades)
	require.Empty(t, Upgrade.StoreUpgrades.Added)
	require.Empty(t, Upgrade.StoreUpgrades.Deleted)
	require.Empty(t, Upgrade.StoreUpgrades.Renamed)
}

func TestUpgradeHandlerReturnsMigrationVersionMap(t *testing.T) {
	input := module.VersionMap{"bank": 3}
	expected := module.VersionMap{"bank": 4}
	runner := &migrationRunnerStub{versionMap: expected}

	got, err := createV31UpgradeHandler(runner, nil)(context.Background(), upgradetypes.Plan{Name: UpgradeName}, input)

	require.NoError(t, err)
	require.Equal(t, input, runner.gotVersionMap)
	require.Equal(t, expected, got)
}

func TestUpgradeHandlerPropagatesMigrationError(t *testing.T) {
	expectedErr := errors.New("migration failed")
	runner := &migrationRunnerStub{err: expectedErr}

	got, err := createV31UpgradeHandler(runner, nil)(context.Background(), upgradetypes.Plan{Name: UpgradeName}, module.VersionMap{"bank": 3})

	require.Nil(t, got)
	require.ErrorIs(t, err, expectedErr)
}
