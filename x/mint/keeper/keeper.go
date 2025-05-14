package keeper

import (
	"context"
	"errors"
	"fmt"

	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v29/x/mint/types"
)

// Keeper of the mint store
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService

	stakingKeeper types.StakingKeeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper

	feeCollectorName string
	authority        string
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	ss storetypes.KVStoreService,
	sk types.StakingKeeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	feeCollectorName string,
	authority string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("the x/%s module account has not been set", types.ModuleName))
	}

	return Keeper{
		cdc:              cdc,
		storeService:     ss,
		stakingKeeper:    sk,
		bankKeeper:       bk,
		accountKeeper:    ak,
		feeCollectorName: feeCollectorName,
		authority:        authority,
	}
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// get the minter
func (k Keeper) GetMinter(ctx context.Context) (minter types.Minter, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.MinterKey)
	if bz == nil {
		return types.Minter{}, err
	}
	if err != nil {
		return types.Minter{}, err
	}

	k.cdc.MustUnmarshal(bz, &minter)
	return minter, nil
}

// set the minter
func (k Keeper) SetMinter(ctx context.Context, minter types.Minter) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&minter)
	err := store.Set(types.MinterKey, bz)
	if err != nil {
		return err
	}

	return nil
}

// ______________________________________________________________________

// SetParams sets the x/mint module parameters.
func (k Keeper) SetParams(ctx context.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&p)
	err := store.Set(types.ParamsKey, bz)
	if err != nil {
		return err
	}

	return nil
}

// GetParams returns the current x/mint module parameters.
func (k Keeper) GetParams(ctx context.Context) (p types.Params, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if bz == nil {
		return p, err
	}
	if err != nil {
		return p, err
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p, err
}

func (k Keeper) GetAccountKeeper() types.AccountKeeper {
	return k.accountKeeper
}

// ______________________________________________________________________

// StakingTokenSupply implements an alias call to the underlying staking keeper's
// StakingTokenSupply to be used in BeginBlocker.
func (k Keeper) StakingTokenSupply(ctx context.Context) sdkmath.Int {
	supply, err := k.stakingKeeper.StakingTokenSupply(ctx)
	if err != nil {
		return sdkmath.ZeroInt()
	}
	return supply
}

// TokenSupply implements an alias call to the underlying bank keeper's
// TokenSupply to be used in BeginBlocker.
func (k Keeper) TokenSupply(ctx context.Context, denom string) sdkmath.Int {
	return k.bankKeeper.GetSupply(ctx, denom).Amount
}

// GetBalance implements an alias call to the underlying bank keeper's
// GetBalance to be used in BeginBlocker.
func (k Keeper) GetBalance(ctx context.Context, address sdk.AccAddress, denom string) sdkmath.Int {
	return k.bankKeeper.GetBalance(ctx, address, denom).Amount
}

// BondedRatio implements an alias call to the underlying staking keeper's
// BondedRatio to be used in BeginBlocker.
func (k Keeper) BondedRatio(ctx context.Context) sdkmath.LegacyDec {
	br, err := k.stakingKeeper.BondedRatio(ctx)
	if err != nil {
		return sdkmath.LegacyZeroDec()
	}
	return br
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx context.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

func (k Keeper) ReduceTargetSupply(ctx context.Context, burnCoin sdk.Coin) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	if burnCoin.Denom != params.MintDenom {
		return errors.New("tried reducing target supply with non staking token")
	}

	minter, err := k.GetMinter(ctx)
	if err != nil {
		return err
	}
	minter.TargetSupply = minter.TargetSupply.Sub(burnCoin.Amount)
	err = k.SetMinter(ctx, minter)
	if err != nil {
		return err
	}

	return nil
}

// AddCollectedFees implements an alias call to the underlying supply keeper's
// AddCollectedFees to be used in BeginBlocker.
func (k Keeper) AddCollectedFees(ctx context.Context, fees sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, fees)
}
