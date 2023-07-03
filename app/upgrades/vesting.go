package upgrades

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/CosmosContracts/juno/v16/app/keepers"
)

const (
	// TODO: Ensure mainnet codeId is used here
	// Same as Reece, Noah, and Ekez contracts.
	// junod q wasm code 2453 $HOME/Desktop/vesting.wasm --node https://juno-rpc.reece.sh:443
	vestingCodeID = 2453
	// vestingCodeID        = 1 // testing
	junoUnbondingSeconds = 2419200
)

type VestingContract struct {
	Owner                    string `json:"owner"`
	Recipient                string `json:"recipient"`
	Title                    string `json:"title"`
	Description              string `json:"description"`
	Schedule                 string `json:"schedule"`
	UnbondingDurationSeconds uint64 `json:"unbonding_duration_seconds"`
	VestingDurationSeconds   uint64 `json:"vesting_duration_seconds"`
	Total                    string `json:"total"`
	Denom                    Denom  `json:"denom"`
}

type Denom struct {
	Native string `json:"native"`
}

// Stops a vesting account and returns all tokens back to the Core-1 SubDAO.
func MoveVestingCoinFromVestingAccount(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string, name string, accAddr sdk.AccAddress, core1AccAddr sdk.AccAddress) error {
	now := ctx.BlockHeader().Time

	stdAcc := keepers.AccountKeeper.GetAccount(ctx, accAddr)
	vacc, ok := stdAcc.(*authvestingtypes.PeriodicVestingAccount)
	if !ok {
		// For e2e testing
		fmt.Printf("account " + accAddr.String() + " is not a vesting account.\n")
		return nil
	}

	fmt.Printf("\n\n== Vesting Account Address: %s (%s) ==\n", vacc.GetAddress().String(), name)

	// Gets non-vested coins (These get returned back to Core-1 SubDAO / a new vesting contract made from)
	unvestedCoins := vacc.GetVestingCoins(now)
	fmt.Printf("Locked / waiting to vest Coins: %v\n", unvestedCoins)

	// Pre migration balance
	core1BeforeBal := keepers.BankKeeper.GetBalance(ctx, core1AccAddr, bondDenom)

	// Clears the account so all all future vesting periods are removed.
	clearVestingAccount(ctx, vacc, keepers, unvestedCoins)

	// Complete any re-deleations to become standard delegations.
	if err := completeAllRedelegations(ctx, keepers, accAddr, now); err != nil {
		return err
	}

	// Instant unbond all delegations.
	_, err := unbondAllAndFinish(ctx, now, keepers, accAddr)
	if err != nil {
		return err
	}

	// set the vesting account to a base account
	keepers.AccountKeeper.SetAccount(ctx, vacc.BaseAccount)

	// Moves unvested tokens to the Core-1 SubDAO.
	if err := transferUnvestedTokensToCore1SubDao(ctx, keepers, bondDenom, accAddr, core1AccAddr, unvestedCoins); err != nil {
		return err
	}

	// Ensure the post validation checks are met.
	if err := postValidation(ctx, keepers, bondDenom, accAddr, core1AccAddr, unvestedCoins, core1BeforeBal); err != nil {
		return err
	}

	// Create a new vesting contract owned by Core-1 (and Juno Governance by proxy)
	// Wolfs resignation terminates his future vesting.
	if name != "wolf" {
		if err := newVestingContract(ctx, keepers, core1AccAddr, accAddr, name, unvestedCoins); err != nil {
			return err
		}
	}

	return nil
}

func newVestingContract(ctx sdk.Context, keepers *keepers.AppKeepers, core1AccAddr, accAddr sdk.AccAddress, name string, unvestedCoins sdk.Coins) error {
	now := ctx.BlockHeader().Time.Unix()

	// End Vesting Time (Juno Network launch Oct 1st, 2021. Vested 12 years = 2033)
	endVestingEpochDate := time.Date(2033, 10, 1, 0, 0, 0, 0, time.UTC)
	endVestingEpochSeconds := uint64(endVestingEpochDate.Unix())
	vestingDurationSeconds := endVestingEpochSeconds - uint64(now)

	owner := core1AccAddr.String()
	recipient := accAddr.String()

	// https://github.com/DA0-DA0/dao-contracts/blob/main/contracts/external/cw-vesting/src/msg.rs#L11
	msgBz, err := json.Marshal(VestingContract{
		Owner:                    owner,
		Recipient:                recipient,
		Title:                    fmt.Sprintf("%s core-1", name),
		Description:              "Core-1 Vesting contract",
		Schedule:                 "saturating_linear",
		UnbondingDurationSeconds: junoUnbondingSeconds,
		VestingDurationSeconds:   vestingDurationSeconds,
		Total:                    strconv.FormatInt(unvestedCoins[0].Amount.Int64(), 10),
		Denom: Denom{
			Native: "ujuno",
		},
	})
	if err != nil {
		return err
	}

	fmt.Printf("Moving %v from Core1 to new contract\n", unvestedCoins)
	fmt.Println(string(msgBz))

	contractAcc, _, err := keepers.ContractKeeper.Instantiate(
		ctx,
		uint64(vestingCodeID),
		core1AccAddr,
		core1AccAddr,
		msgBz,
		fmt.Sprintf("vest_to_%s_%d", recipient, now),
		unvestedCoins,
	)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such code") {
			fmt.Printf("No such codeId: %d - skipping (e2e testing, not mainnet)\n", vestingCodeID)
			return nil
		}

		return err
	}

	fmt.Printf("Contract Address: %s for %s with amount %d\n", contractAcc.String(), name, unvestedCoins[0].Amount.Int64())
	return nil
}

func postValidation(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string, accAddr sdk.AccAddress, core1AccAddr sdk.AccAddress, unvestedCoins sdk.Coins, core1BeforeBal sdk.Coin) error {
	// Core1 balance should only increase by exactly the core1Bal + unvestedCoins
	core1BalAfter := keepers.BankKeeper.GetBalance(ctx, core1AccAddr, bondDenom)
	if !core1BeforeBal.Add(unvestedCoins[0]).IsEqual(core1BalAfter) {
		return fmt.Errorf("ERROR: core1BeforeBal (%v) + unvestedCoins (%v) != core1BalAfter (%v)", core1BeforeBal, unvestedCoins, core1BalAfter)
	}

	// vesting account should have no future vesting periods
	newVacc := keepers.AccountKeeper.GetAccount(ctx, accAddr)
	if _, ok := newVacc.(*authvestingtypes.PeriodicVestingAccount); ok {
		return fmt.Errorf("ERROR: account %s still is a vesting account", accAddr.String())
	}

	// ensure the account has 0 delegations, redelegations, or unbonding delegations
	delegations := keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr)
	if len(delegations) != 0 {
		return fmt.Errorf("ERROR: account %s still has delegations", accAddr.String())
	}

	redelegations := keepers.StakingKeeper.GetRedelegations(ctx, accAddr, 65535)
	if len(redelegations) != 0 {
		return fmt.Errorf("ERROR: account %s still has redelegations", accAddr.String())
	}

	unbondingDelegations := keepers.StakingKeeper.GetAllUnbondingDelegations(ctx, accAddr)
	if len(unbondingDelegations) != 0 {
		return fmt.Errorf("ERROR: account %s still has unbonding delegations", accAddr.String())
	}

	return nil
}

func transferUnvestedTokensToCore1SubDao(ctx sdk.Context, keepers *keepers.AppKeepers, bondDenom string, accAddr, core1AccAddr sdk.AccAddress, unvestedCoins sdk.Coins) error {
	fmt.Printf("Sending Unvested Coins back to Core-1: %v\n", unvestedCoins)
	if err := keepers.BankKeeper.SendCoins(ctx, accAddr, core1AccAddr, unvestedCoins); err != nil {
		return err
	}

	core1BalC := keepers.BankKeeper.GetBalance(ctx, core1AccAddr, bondDenom)
	fmt.Printf("Updated Core1 SubDAO Balance: %v\n", core1BalC)

	return nil
}

func completeAllRedelegations(ctx sdk.Context, keepers *keepers.AppKeepers, accAddr sdk.AccAddress, now time.Time) error {
	for _, activeRedelegation := range keepers.StakingKeeper.GetRedelegations(ctx, accAddr, 65535) {
		redelegationSrc, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorSrcAddress)
		redelegationDst, _ := sdk.ValAddressFromBech32(activeRedelegation.ValidatorDstAddress)

		// set all entry completionTime to now so we can complete re-delegation
		for i := range activeRedelegation.Entries {
			activeRedelegation.Entries[i].CompletionTime = now
		}

		keepers.StakingKeeper.SetRedelegation(ctx, activeRedelegation)
		_, err := keepers.StakingKeeper.CompleteRedelegation(ctx, accAddr, redelegationSrc, redelegationDst)
		if err != nil {
			return err
		}
	}

	return nil
}

// Returns the amount of tokens which were unbonded (not rewards)
func unbondAllAndFinish(ctx sdk.Context, now time.Time, keepers *keepers.AppKeepers, accAddr sdk.AccAddress) (math.Int, error) {
	unbondedAmt := math.ZeroInt()

	// Unbond all delegations from the account
	for _, delegation := range keepers.StakingKeeper.GetAllDelegatorDelegations(ctx, accAddr) {
		validatorValAddr := delegation.GetValidatorAddr()
		_, found := keepers.StakingKeeper.GetValidator(ctx, validatorValAddr)
		if !found {
			continue
		}

		_, err := keepers.StakingKeeper.Undelegate(ctx, accAddr, validatorValAddr, delegation.GetShares())
		if err != nil {
			return math.ZeroInt(), err
		}
	}

	// Take all unbonding and complete them.
	for _, unbondingDelegation := range keepers.StakingKeeper.GetAllUnbondingDelegations(ctx, accAddr) {
		validatorStringAddr := unbondingDelegation.ValidatorAddress
		validatorValAddr, _ := sdk.ValAddressFromBech32(validatorStringAddr)

		// Complete unbonding delegation
		for i := range unbondingDelegation.Entries {
			unbondingDelegation.Entries[i].CompletionTime = now
			unbondedAmt = unbondedAmt.Add(unbondingDelegation.Entries[i].Balance)
		}

		keepers.StakingKeeper.SetUnbondingDelegation(ctx, unbondingDelegation)
		_, err := keepers.StakingKeeper.CompleteUnbonding(ctx, accAddr, validatorValAddr)
		if err != nil {
			return math.ZeroInt(), err
		}
	}

	return unbondedAmt, nil
}

func clearVestingAccount(ctx sdk.Context, vacc *authvestingtypes.PeriodicVestingAccount, keepers *keepers.AppKeepers, unvestedCoins sdk.Coins) {
	vacc.BaseVestingAccount.EndTime = ctx.BlockTime().Unix()

	for i := range vacc.VestingPeriods {
		vacc.VestingPeriods[i].Length = 0
	}

	vacc.DelegatedFree = unvestedCoins
	vacc.DelegatedVesting = sdk.Coins{}

	vacc.BaseVestingAccount.DelegatedFree = unvestedCoins
	vacc.BaseVestingAccount.DelegatedVesting = sdk.Coins{}

	keepers.AccountKeeper.SetAccount(ctx, vacc)
}
