package v19

import (
	"fmt"

	wasmlctypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	decorators "github.com/CosmosContracts/juno/v19/app/decorators"
	"github.com/CosmosContracts/juno/v19/app/keepers"
	"github.com/CosmosContracts/juno/v19/app/upgrades"
)

const (
	// Core-1 Mainnet Address
	Core1SubDAOAddress = "juno1j6glql3xmrcnga0gytecsucq3kd88jexxamxg3yn2xnqhunyvflqr7lxx3"
)

// Core1VestingAccounts https://daodao.zone/dao/juno1j6glql3xmrcnga0gytecsucq3kd88jexxamxg3yn2xnqhunyvflqr7lxx3/members
var Core1VestingAccounts = map[string]string{
	"block": "juno17py8gfneaam64vt9kaec0fseqwxvkq0flmsmhg",
	"dimi":  "juno1s33zct2zhhaf60x4a90cpe9yquw99jj0zen8pt",
	"jack":  "juno130mdu9a0etmeuw52qfxk73pn0ga6gawk4k539x",
	"jake":  "juno18qw9ydpewh405w4lvmuhlg9gtaep79vy2gmtr2",
	// TODO: So, can the SubDAO be the owner of the init'ed contract to claim rewards?
	"multisig": "juno190g5j8aszqhvtg7cprmev8xcxs6csra7xnk3n3",
	"wolf":     "juno1a8u47ggy964tv9trjxfjcldutau5ls705djqyu",

	// "zlocalexample": "juno1xz599egrd3dhq5vx63mkwja38q5q3th8h3ukjj",
}

func CreateV19UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		// Run migrations
		logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		// Change Rate Decorator Migration
		// Ensure all Validators have a max change rate of 5%
		maxChangeRate := sdk.MustNewDecFromStr(decorators.MaxChangeRate)
		validators := k.StakingKeeper.GetAllValidators(ctx)

		for _, validator := range validators {
			if validator.Commission.MaxChangeRate.GT(maxChangeRate) {
				validator.Commission.MaxChangeRate.Set(maxChangeRate)
				k.StakingKeeper.SetValidator(ctx, validator)
			}
		}

		// https://github.com/cosmos/ibc-go/blob/main/docs/docs/03-light-clients/04-wasm/03-integration.md
		params := k.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = append(params.AllowedClients, wasmlctypes.Wasm)
		k.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		// Migrate Core-1 vesting account remaining funds -> Core-1, then create a new vesting contract for them (if not wolf).
		if ctx.ChainID() == "juno-1" {
			if err := migrateCore1VestingAccounts(ctx, k, nativeDenom); err != nil {
				return nil, err
			}
		}

		return versionMap, err
	}
}

func migrateCore1VestingAccounts(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string) error {
	for name, vestingAccount := range Core1VestingAccounts {
		// A new vesting contract will not be created if the account name is 'wolf'.
		if err := upgrades.MoveVestingCoinFromVestingAccount(ctx,
			keepers,
			bondDenom,
			name,
			sdk.MustAccAddressFromBech32(vestingAccount),
			sdk.MustAccAddressFromBech32(Core1SubDAOAddress),
		); err != nil {
			return err
		}
	}

	// return fmt.Errorf("DEBUGGING; not finished yet. (migrateCore1VestingAccounts)")
	return nil
}
