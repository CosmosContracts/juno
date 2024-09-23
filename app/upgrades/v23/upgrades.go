package v23

import (
	"fmt"

	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmosContracts/juno/v25/app/keepers"
	"github.com/CosmosContracts/juno/v25/app/upgrades"
)

type IndividualAccount struct {
	Owner   string
	Address string
}

// Core1VestingAccounts https://daodao.zone/dao/juno1j6glql3xmrcnga0gytecsucq3kd88jexxamxg3yn2xnqhunyvflqr7lxx3/members
// we are including only lobo, dimi and jake because the other ones do not agree on giving up their vesting
var Core1VestingAccounts = []IndividualAccount{
	{
		Owner:   "dimi",
		Address: "juno1s33zct2zhhaf60x4a90cpe9yquw99jj0zen8pt",
	},
	{
		Owner:   "jake",
		Address: "juno18qw9ydpewh405w4lvmuhlg9gtaep79vy2gmtr2",
	},
	{
		Owner:   "wolf",
		Address: "juno1a8u47ggy964tv9trjxfjcldutau5ls705djqyu",
	},
}

func CreateV23UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		// migrate ICQ params
		for _, subspace := range keepers.ParamsKeeper.GetSubspaces() {
			subspace := subspace

			var keyTable paramstypes.KeyTable
			if subspace.Name() == icqtypes.ModuleName {
				keyTable = icqtypes.ParamKeyTable()
			} else {
				continue
			}

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		// Run migrations
		logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}

		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		// convert pob builder account to an actual module account
		// during upgrade from v15 to v16 it wasn't correctly created, and since it received tokens on mainnet is now a base account
		// it's like this on both mainnet and uni
		if ctx.ChainID() == "juno-1" || ctx.ChainID() == "uni-6" {
			logger.Info("converting x/pob builder module account")

			address := sdk.MustAccAddressFromBech32("juno1ma4sw9m2nvtucny6lsjhh4qywvh86zdh5dlkd4")

			acc := keepers.AccountKeeper.NewAccount(
				ctx,
				authtypes.NewModuleAccount(
					authtypes.NewBaseAccountWithAddress(address),
					"builder",
				),
			)
			keepers.AccountKeeper.SetAccount(ctx, acc)

			logger.Info("x/pob builder module address is now a module account")
		}

		// only on mainnet and uni, migrate core1 vesting accounts
		if ctx.ChainID() == "juno-1" || ctx.ChainID() == "uni-6" {
			if err := migrateCore1VestingAccounts(ctx, keepers, nativeDenom); err != nil {
				return nil, err
			}
		}

		return versionMap, err
	}
}

// Migrate balances from the Core-1 vesting accounts to the Council SubDAO.
func migrateCore1VestingAccounts(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string) error {
	for _, account := range Core1VestingAccounts {
		if err := MoveVestingCoinFromVestingAccount(ctx,
			keepers,
			bondDenom,
			account.Owner,
			sdk.MustAccAddressFromBech32(account.Address),
		); err != nil {
			return err
		}
	}
	return nil
}
