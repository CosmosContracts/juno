package v15

import (
	"fmt"
	"time"

	"github.com/CosmosContracts/juno/v15/app/keepers"

	"github.com/CosmosContracts/juno/v15/app/upgrades"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	tokenfactorytypes "github.com/CosmosTokenFactory/token-factory/x/tokenfactory/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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
		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		// x/Mint
		// Double blocks per year (from 6 seconds to 3 = 2x blocks per year)
		mintParams := keepers.MintKeeper.GetParams(ctx)
		mintParams.BlocksPerYear *= 2
		keepers.MintKeeper.SetParams(ctx, mintParams)
		logger.Info(fmt.Sprintf("updated minted blocks per year logic to %v", mintParams))

		// x/TokenFactory
		// Use denom creation gas consumtion instead of fee for contract developers
		updatedTf := tokenfactorytypes.Params{
			DenomCreationFee:        nil,
			DenomCreationGasConsume: NewDenomCreationGasConsume,
		}
		keepers.TokenFactoryKeeper.SetParams(ctx, updatedTf)
		logger.Info(fmt.Sprintf("updated tokenfactory params to %v", updatedTf))

		// x/Slashing
		// Double slashing window due to double blocks per year
		slashingParams := keepers.SlashingKeeper.GetParams(ctx)
		slashingParams.SignedBlocksWindow *= 2
		keepers.SlashingKeeper.SetParams(ctx, slashingParams)

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

	// iterate through accounts
	// Instantiate on behalf of the core-1 subDAO as the owner, and move all balance, pending rewards, and staked amounts into the new contract
	for address, memberName := range vestingAccounts {
		fmt.Println(address, memberName)

		addr, _ := sdk.AccAddressFromBech32(address)

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

		// start_time is not set as it is Optional, which then sets when it is instantiated. ("start_time": "1677657600000000000")
		// vesting_duration_seconds a time in the future. 12 years. So get current epoch second, time until 12 year end, difference
		// unbonding_duration_seconds:
		// vesting_duration_seconds (94608000 = 3 years)
		msg := fmt.Sprintf(`{"owner":"%s","recipient":"%s","title":"%s","description":"Core-1 Vesting","total":%d,"denom":{"native":"ujuno"}},"schedule":"saturating_linear","unbonding_duration_seconds":%d,"vesting_duration_seconds":9999}`, core1SubDaoAddr, address, memberName, preVestedCoin.Amount.Int64(), junoUnbondingSeconds)
		// set as label vest_to_juno1addr_1682213004408 where the ending is the current epoch time of prev block
		// also pass through funds which must == total.

		// replace with previous blocktime header in the future
		currentEpochTime := time.Now().Unix() / 1000

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
			fmt.Sprintf("vest_to_%s_%d", address, currentEpochTime),
			coins,
		)
		// log contractAddr
		fmt.Println(contractAddr, err)
		if err != nil {
			panic(err)
		}

	}
}
