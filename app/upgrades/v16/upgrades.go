package v16

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	buildertypes "github.com/skip-mev/pob/x/builder/types"

	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	exported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	// External modules
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	// SDK v47 modules
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CosmosContracts/juno/v16/app/keepers"
	"github.com/CosmosContracts/juno/v16/app/upgrades"
	// Juno modules
	feesharetypes "github.com/CosmosContracts/juno/v16/x/feeshare/types"
	globalfeetypes "github.com/CosmosContracts/juno/v16/x/globalfee/types"
	minttypes "github.com/CosmosContracts/juno/v16/x/mint/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v16/x/tokenfactory/types"
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

func CreateV16UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

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

			// POB
			case buildertypes.ModuleName:
				// already SDK v47
				continue

			// juno modules
			case feesharetypes.ModuleName:
				keyTable = feesharetypes.ParamKeyTable() //nolint:staticcheck
			case tokenfactorytypes.ModuleName:
				keyTable = tokenfactorytypes.ParamKeyTable() //nolint:staticcheck
			case minttypes.ModuleName:
				keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
			case globalfeetypes.ModuleName:
				keyTable = globalfeetypes.ParamKeyTable() //nolint:staticcheck
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

		// https://github.com/cosmos/ibc-go/blob/v7.1.0/docs/migrations/v7-to-v7_1.md
		// explicitly update the IBC 02-client params, adding the localhost client type
		params := keepers.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = append(params.AllowedClients, exported.Localhost)
		keepers.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		// Interchain Queries
		icqParams := icqtypes.NewParams(true, nil)
		keepers.ICQKeeper.SetParams(ctx, icqParams)

		// update gov params to use a 20% initial deposit ratio, allowing us to remote the ante handler
		govParams := keepers.GovKeeper.GetParams(ctx)
		govParams.MinInitialDepositRatio = sdk.NewDec(20).Quo(sdk.NewDec(100)).String()
		if err := keepers.GovKeeper.SetParams(ctx, govParams); err != nil {
			return nil, err
		}

		// x/Staking - set minimum commission to 0.050000000000000000
		stakingParams := keepers.StakingKeeper.GetParams(ctx)
		stakingParams.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
		err = keepers.StakingKeeper.SetParams(ctx, stakingParams)
		if err != nil {
			return nil, err
		}

		// Migrate Core-1 vesting account remaining funds -> Core-1, then create a new vesting contract for them (if not wolf).
		if ctx.ChainID() == "juno-1" {
			if err := migrateCore1VestingAccounts(ctx, keepers, nativeDenom); err != nil {
				return nil, err
			}
    }

		// x/POB
		pobAddr := keepers.AccountKeeper.GetModuleAddress(buildertypes.ModuleName)

		builderParams := buildertypes.DefaultGenesisState().GetParams()
		builderParams.EscrowAccountAddress = pobAddr
		builderParams.MaxBundleSize = 4
		builderParams.FrontRunningProtection = false
		builderParams.MinBidIncrement.Denom = nativeDenom
		builderParams.MinBidIncrement.Amount = math.NewInt(1000000)
		builderParams.ReserveFee.Denom = nativeDenom
		builderParams.ReserveFee.Amount = math.NewInt(1000000)
		if err := keepers.BuildKeeper.SetParams(ctx, builderParams); err != nil {
			return nil, err
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
