package v15

import (
	"fmt"

	"github.com/CosmosContracts/juno/v15/app/keepers"

	"github.com/CosmosContracts/juno/v15/app/upgrades"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	tokenfactorytypes "github.com/CosmosContracts/juno/v15/x/tokenfactory/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icqtypes "github.com/strangelove-ventures/async-icq/v7/types"
)

// We now charge 2 million gas * gas price to create a denom.
const NewDenomCreationGasConsume uint64 = 2_000_000

func CreateV15UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
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

		// Anything to do with ConsensusParamsKeeper?

		// Interchain Queries
		icqParams := icqtypes.NewParams(true, nil)
		keepers.ICQKeeper.SetParams(ctx, icqParams)

		// x/TokenFactory
		// Use denom creation gas consumtion instead of fee for contract developers
		updatedTf := tokenfactorytypes.Params{
			DenomCreationFee:        nil,
			DenomCreationGasConsume: NewDenomCreationGasConsume,
		}
		keepers.TokenFactoryKeeper.SetParams(ctx, updatedTf)
		logger.Info(fmt.Sprintf("updated tokenfactory params to %v", updatedTf))

		return versionMap, err
	}
}

// Previously planned Faster block time upgrade
//
// x/Mint
// Double blocks per year (from 6 seconds to 3 = 2x blocks per year)
// mintParams := keepers.MintKeeper.GetParams(ctx)
// mintParams.BlocksPerYear *= 2
// keepers.MintKeeper.SetParams(ctx, mintParams)
// logger.Info(fmt.Sprintf("updated minted blocks per year logic to %v", mintParams))
//
// x/Slashing
// Double slashing window due to double blocks per year
// slashingParams := keepers.SlashingKeeper.GetParams(ctx)
// slashingParams.SignedBlocksWindow *= 2
// err = keepers.SlashingKeeper.SetParams(ctx, slashingParams)
// if err != nil {
// 	return nil, err
// }
