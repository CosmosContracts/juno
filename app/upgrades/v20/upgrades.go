package v20

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmosContracts/juno/v19/app/keepers"
	"github.com/CosmosContracts/juno/v19/app/upgrades"
)

const (
	// Charter Council's SubDAO Address
	CharterCouncil = "juno1nmezpepv3lx45mndyctz2lzqxa6d9xzd2xumkxf7a6r4nxt0y95qypm6c0"
	JackKey        = "jack"
)

// JackValidatorAddress must be mutable for testing
var JackValidatorAddress = "junovaloper130mdu9a0etmeuw52qfxk73pn0ga6gawk2tz77l"

type IndividualAccount struct {
	Owner   string
	Address string
}

// Core1VestingAccounts https://daodao.zone/dao/juno1j6glql3xmrcnga0gytecsucq3kd88jexxamxg3yn2xnqhunyvflqr7lxx3/members
var Core1VestingAccounts = []IndividualAccount{
	{
		Owner:   "block",
		Address: "juno17py8gfneaam64vt9kaec0fseqwxvkq0flmsmhg",
	},
	{
		Owner:   "dimi",
		Address: "juno1s33zct2zhhaf60x4a90cpe9yquw99jj0zen8pt",
	},
	{
		Owner:   JackKey,
		Address: "juno130mdu9a0etmeuw52qfxk73pn0ga6gawk4k539x",
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

func CreateV20UpgradeHandler(
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

		// Migrate Core-1 vesting account remaining funds -> Council SubDAO
		if ctx.ChainID() == "juno-1" {
			if err := migrateCore1VestingAccounts(ctx, k, nativeDenom); err != nil {
				return nil, err
			}
		}

		return versionMap, err
	}
}

// Migrate balances from the Core-1 vesting accounts to the Council SubDAO.
func migrateCore1VestingAccounts(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string) error {
	for _, account := range Core1VestingAccounts {
		// A new vesting contract will not be created if the account name is 'wolf'.
		if err := MoveVestingCoinFromVestingAccount(ctx,
			keepers,
			bondDenom,
			account.Owner,
			sdk.MustAccAddressFromBech32(account.Address),
			sdk.MustAccAddressFromBech32(CharterCouncil),
		); err != nil {
			return err
		}
	}
	return nil
}
