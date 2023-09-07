package v16

import (
	"fmt"

	log "github.com/cometbft/cometbft/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmosContracts/juno/v17/app/keepers"
	"github.com/CosmosContracts/juno/v17/app/upgrades"
	clocktypes "github.com/CosmosContracts/juno/v17/x/clock/types"
	driptypes "github.com/CosmosContracts/juno/v17/x/drip/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	// Verify the following with:
	// - https://daodao.zone/dao/<ADDRESS>
	subDaos = []string{
		"juno1j6glql3xmrcnga0gytecsucq3kd88jexxamxg3yn2xnqhunyvflqr7lxx3", // core-1
		"juno1q7ufzamrmwfw4w35azzkcxd5l44vy8zngm9ufcgryk2dt8clqznsp88lhd", // HackJuno
		"juno1xz54y0ktew0dcm00f9vjw0p7x29pa4j5p9rwq6zerkytugzg27qs4shxnt", // Growth Fund
		"juno1rw92sps9q4mm7ll3x9apnunlckchmn3v7cttchsf48dcdyajzj2sajfxcn", // Delegations
		"juno15zw5zt2pepx8n8675dz3k3yscdu94d24yhqqz00uzyx7ydf2vfmswz6nzw", // Communications
	}
)

func CreateV17UpgradeHandler(
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

		// x/Mint
		// Double blocks per year (from 6 seconds to 3 = 2x blocks per year)
		mintParams := keepers.MintKeeper.GetParams(ctx)
		mintParams.BlocksPerYear *= 2
		if err = keepers.MintKeeper.SetParams(ctx, mintParams); err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("updated minted blocks per year logic to %v", mintParams))

		// x/Slashing
		// Double slashing window due to double blocks per year
		slashingParams := keepers.SlashingKeeper.GetParams(ctx)
		slashingParams.SignedBlocksWindow *= 2
		if err := keepers.SlashingKeeper.SetParams(ctx, slashingParams); err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("updated slashing params to %v", slashingParams))

		// x/drip
		if err := keepers.DripKeeper.SetParams(ctx, driptypes.DefaultParams()); err != nil {
			return nil, err
		}

		// x/clock
		if err := keepers.ClockKeeper.SetParams(ctx, clocktypes.DefaultParams()); err != nil {
			return nil, err
		}

		// This function migrates all DAOs owned by the chain from the distribution module address -> the gov module.
		// While they chains till owns it, technically it makes more sense to store them in the gov account.
		if ctx.ChainID() == "juno-1" {
			migrateChainOwnedSubDaos(ctx, logger, keepers.AccountKeeper, &keepers.WasmKeeper, keepers.ContractKeeper)
		}

		return versionMap, err
	}
}

func migrateChainOwnedSubDaos(ctx sdk.Context, logger log.Logger, ak authkeeper.AccountKeeper, wk *wasmkeeper.Keeper, ck *wasmkeeper.PermissionedKeeper) error {
	logger.Info("migrating chain owned sub-daos")

	distrAddr := ak.GetModuleAddress(distrtypes.ModuleName)

	govAddr := ak.GetModuleAddress(govtypes.ModuleName)
	govBech32, err := sdk.AccAddressFromBech32(govAddr.String())
	if err != nil {
		return err
	}

	for _, dao := range subDaos {
		dao := dao

		cAddr, err := sdk.AccAddressFromBech32(dao)
		if err != nil {
			return err
		}

		// The dist module calls this to update its admin since its the admin currently.
		newAdmin := govBech32
		if err := ck.UpdateContractAdmin(ctx, cAddr, distrAddr, newAdmin); err != nil {
			return err
		}

	}

	return nil
}
