package v15

import (
	"fmt"

	"github.com/CosmosContracts/juno/v15/app/keepers"
	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/CosmosContracts/juno/v15/app/upgrades"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	tokenfactorytypes "github.com/CosmosContracts/juno/v15/x/tokenfactory/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icqtypes "github.com/strangelove-ventures/async-icq/v7/types"

	// SDK v47 modules
	// minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// External modules
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
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

		// TODO: Add postHandler (like antehandler but runs after runMsgs)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		// TODO: Our mint module needs to be migrated to v47 for minttypes.ModuleName
		// https://github.com/cosmos/cosmos-sdk/pull/12363/files
		// Set param key table for params module migration
		for _, subspace := range keepers.ParamsKeeper.GetSubspaces() {
			subspace := subspace

			var keyTable paramstypes.KeyTable
			switch subspace.Name() {
			case authtypes.ModuleName:
				keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
			case banktypes.ModuleName:
				keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
			case stakingtypes.ModuleName:
				keyTable = stakingtypes.ParamKeyTable() //nolint:staticcheck
			// TODO: mint module v47?
			// case minttypes.ModuleName:
			// 	keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
			case distrtypes.ModuleName:
				keyTable = distrtypes.ParamKeyTable() //nolint:staticcheck
			case slashingtypes.ModuleName:
				keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
			case govtypes.ModuleName:
				keyTable = govv1.ParamKeyTable() //nolint:staticcheck
			case crisistypes.ModuleName:
				keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck

			// ibc types
			case ibctransfertypes.ModuleName:
				keyTable = ibctransfertypes.ParamKeyTable()
			case icahosttypes.SubModuleName:
				keyTable = icahosttypes.ParamKeyTable()
			case icacontrollertypes.SubModuleName:
				keyTable = icacontrollertypes.ParamKeyTable()
			// wasm
			case wasmtypes.ModuleName:
				keyTable = wasmtypes.ParamKeyTable() //nolint:staticcheck
			}

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		// Migrate Tendermint consensus parameters from x/params module to a deprecated x/consensus module.
		// The old params module is required to still be imported in your app.go in order to handle this migration.
		baseAppLegacySS := keepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
		baseapp.MigrateParams(ctx, baseAppLegacySS, &keepers.ConsensusParamsKeeper)

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

		// update gov params to use a 20% initial deposit ratio, allowing us to remote the ante handler
		// TODO: Add test for this
		govParams := keepers.GovKeeper.GetParams(ctx)
		govParams.MinInitialDepositRatio = sdk.NewDec(20).Quo(sdk.NewDec(100)).String()
		if err := keepers.GovKeeper.SetParams(ctx, govParams); err != nil {
			return nil, err
		}

		// x/TokenFactory
		// Use denom creation gas consumtion instead of fee for contract developers
		updatedTf := tokenfactorytypes.Params{
			DenomCreationFee:        nil,
			DenomCreationGasConsume: NewDenomCreationGasConsume,
		}
		keepers.TokenFactoryKeeper.SetParams(ctx, updatedTf)
		logger.Info(fmt.Sprintf("updated tokenfactory params to %v", updatedTf))

		// x/Staking - set minimum commission to 0.050000000000000000
		stakingParams := keepers.StakingKeeper.GetParams(ctx)
		stakingParams.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
		err = keepers.StakingKeeper.SetParams(ctx, stakingParams)
		if err != nil {
			return nil, err
		}

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
