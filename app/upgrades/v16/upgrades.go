package v15

import (
	"fmt"
	"time"

	"github.com/CosmosContracts/juno/v16/app/keepers"
	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/CosmosContracts/juno/v16/app/upgrades"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	tokenfactorytypes "github.com/CosmosContracts/juno/v16/x/tokenfactory/types"
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
	exported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	// Juno modules
	feesharetypes "github.com/CosmosContracts/juno/v16/x/feeshare/types"
)

// We now charge 2 million gas * gas price to create a denom.
const NewDenomCreationGasConsume uint64 = 2_000_000

func CreateV16UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		nativeDenom := upgrades.GetChainsDenomToken(ctx.ChainID())
		logger.Info(fmt.Sprintf("With native denom %s", nativeDenom))

		// https://github.com/cosmos/ibc-go/blob/v7.1.0/docs/migrations/v7-to-v7_1.md
		// explicitly update the IBC 02-client params, adding the localhost client type
		params := keepers.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = append(params.AllowedClients, exported.Localhost)
		keepers.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		// TODO: Our mint, feeshare, globalfee, and tokenfactory module needs to be migrated to v47 for minttypes.ModuleName
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
			// juno modules
			case feesharetypes.ModuleName:
				keyTable = feesharetypes.ParamKeyTable() //nolint:staticcheck
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

		// Migrate Core-1 vesting accounts
		if ctx.ChainID() == "juno-1" {
			migrateCore1VestingAccountsToVestingContract(ctx, keepers)
		}

		return versionMap, err
	}
}

func migrateCore1VestingAccountsToVestingContract(ctx sdk.Context, keepers *keepers.AppKeepers) {
	// TODO: Easier solution - return all funds back to core1SubDao, then propose new vesting contracts from there.
	// Can print out values here for each account on the nodes to ensure each gets what they need + remove the accounts & upgrade time

	// https://github.com/DA0-DA0/dao-contracts/tree/main/contracts/external/cw-vesting

	// This uses the same CodeId as Reece, Noah, and Ekez's vesting contracts are on for core-1
	// https://indexer.daodao.zone/juno-1/contract/juno1axkh35fx7vdtga3s3tj6hadzwdkq8meeq3gen5sardn69lgnmxgslauw4p/cwPayrollFactory/listVestingContracts?
	// junod q wasm contract juno1d232p6f2rn4s66j5mp8fqt8h00rh2z6g84vs4ww58zgtulm2fheqs3u3c4 --output json | jq -r .contract_info.code_id

	vestingCodeID := 2453
	junoUnbondingSeconds := 2419200
	core1SubDaoAddr := "juno1j6glql3xmrcnga0gytecsucq3kd88jexxamxg3yn2xnqhunyvflqr7lxx3"

	vestingAccounts := map[string]string{
		"juno1a...": "Dimi",
		"juno1b...": "JackZ",
		"juno1c...": "JakeH",
		"juno1d...": "Wolf",
		"juno1e...": "Block",
		// max & alex?
	}

	// CurrentTime. Used for label in seconds
	currentUnixSeconds := getBlockTime(ctx)

	// End Vesting Time (Juno Network launch Oct 1st, 2021. Vested 12 years = 2033)
	endVestingEpochDate := time.Date(2033, 10, 1, 0, 0, 0, 0, time.UTC)
	endVestingEpochSeconds := uint64(endVestingEpochDate.Unix())
	vestingDurationSeconds := endVestingEpochSeconds - currentUnixSeconds

	// iterate through accounts
	// Instantiate on behalf of the core-1 subDAO as the owner, and move all balance, pending rewards, and staked amounts into the new contract
	for address, memberName := range vestingAccounts {
		fmt.Println(address, memberName)

		addr, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			panic(err)
		}

		acc := keepers.AccountKeeper.GetAccount(ctx, addr)
		if acc == nil {
			panic("account not found")
		}
		// ensure this is a vesting account.

		// Does this work for vesting accounts as well under the hood?
		preVestedCoin := MoveVestingCoinFromVestingAccount(ctx, addr, keepers, core1SubDaoAddr)
		fmt.Printf("moved %d ujuno from %s to %s\n", preVestedCoin.Amount.Int64(), address, core1SubDaoAddr)

		// delete the old vesting base account
		keepers.AccountKeeper.RemoveAccount(ctx, acc)

		// Now funds are in Core-1 Subdao Control, and we can instantiate a vesting contract on behalf of the subdao for the amount stated

		// start_time is NOT set as it is Optional. Sets when it is instantiated in nano seconds.
		msg := fmt.Sprintf(`{"owner":"%s","recipient":"%s","title":"%s","description":"Core-1 Vesting","total":%d,"denom":{"native":"ujuno"}},"schedule":"saturating_linear","unbonding_duration_seconds":%d,"vesting_duration_seconds":%d}`,
			core1SubDaoAddr,
			address,
			memberName,
			preVestedCoin.Amount.Int64(),
			junoUnbondingSeconds,
			vestingDurationSeconds,
		)

		// set as label vest_to_juno1addr_1682213004408 where the ending is the current epoch time of prev block
		// also pass through funds which must == total.

		coins := []sdk.Coin{
			sdk.NewCoin("ujuno", sdk.NewInt(preVestedCoin.Amount.Int64())),
		}

		// use wasmtypes.ContractOpsKeeper here instead of permissioned keeper? or does it matter since we are permissionless anyways
		contractAddr, _, err := keepers.ContractKeeper.Instantiate(
			ctx,
			uint64(vestingCodeID),
			sdk.MustAccAddressFromBech32(core1SubDaoAddr),
			sdk.MustAccAddressFromBech32(core1SubDaoAddr),
			[]byte(msg),
			fmt.Sprintf("vest_to_%s_%d", address, currentUnixSeconds),
			coins,
		)
		fmt.Println("Contract Created for:", contractAddr, address, memberName, "With ujuno Amount:", preVestedCoin.Amount.Int64())
		if err != nil {
			panic(err)
		}

	}
}

func getBlockTime(ctx sdk.Context) uint64 {
	now := ctx.BlockHeader().Time
	// get the block time in seconds
	return uint64(now.Unix())
}
