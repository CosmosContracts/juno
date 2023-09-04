package unia

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmosContracts/juno/v17/app/keepers"
	"github.com/CosmosContracts/juno/v17/app/upgrades"
	minttypes "github.com/CosmosContracts/juno/v17/x/mint/types"
)

const (
	// Reece's validator account.
	ReeceBech32 = "juno15twk6xu5rnrrlnf7c5zy92gvykcs4h5ucxzzqq"
)

// TokensAmount is for testnet only.
var TokensAmount = sdk.NewCoins(sdk.NewCoin("ujunox", sdk.NewInt(100_000_000_000000)))

func CreateUniAUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		// Run migrations
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}

		// Turn off inflation
		minter := keepers.MintKeeper.GetMinter(ctx)
		minter.Inflation = sdk.NewDecWithPrec(0, 2)
		keepers.MintKeeper.SetMinter(ctx, minter)

		// set blocks per year to be very high
		mp := keepers.MintKeeper.GetParams(ctx)
		mp.BlocksPerYear = 1_000_000_000_000
		if err := keepers.MintKeeper.SetParams(ctx, mp); err != nil {
			return nil, err
		}

		// Mint Tokens for Account
		acc := sdk.MustAccAddressFromBech32(ReeceBech32)
		if err := keepers.BankKeeper.MintCoins(ctx, minttypes.ModuleName, TokensAmount); err != nil {
			return nil, err
		}

		if err := keepers.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, acc, TokensAmount); err != nil {
			return nil, err
		}

		return versionMap, err
	}
}
