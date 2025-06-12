package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/bank"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/staking"
)

// BankKeeperAdapter adapts the keeper to implement bank.KeeperInterface
type BankKeeperAdapter struct {
	keeper *Keeper
}

// NewBankKeeperAdapter creates a new bank keeper adapter
func NewBankKeeperAdapter(keeper *Keeper) bank.KeeperInterface {
	return &BankKeeperAdapter{keeper: keeper}
}

// GetQueryContext implements bank.KeeperInterface
func (a *BankKeeperAdapter) GetQueryContext() (context.Context, error) {
	return a.keeper.GetQueryContext()
}

// ValidateDenom implements bank.KeeperInterface
func (a *BankKeeperAdapter) ValidateDenom(ctx context.Context, denom string) error {
	return a.keeper.ValidateDenom(ctx, denom)
}

// GetBankKeeper implements bank.KeeperInterface
func (a *BankKeeperAdapter) GetBankKeeper() bank.BankKeeperInterface {
	return &BankKeeperImpl{bankKeeper: a.keeper.bankKeeper}
}

// BankKeeperImpl implements bank.BankKeeperInterface
type BankKeeperImpl struct {
	bankKeeper interface {
		GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
		GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	}
}

// GetBalance implements bank.BankKeeperInterface
func (b *BankKeeperImpl) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return b.bankKeeper.GetBalance(ctx, addr, denom)
}

// GetAllBalances implements bank.BankKeeperInterface
func (b *BankKeeperImpl) GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return b.bankKeeper.GetAllBalances(ctx, addr)
}

// StakingKeeperAdapter adapts the keeper to implement staking.KeeperInterface
type StakingKeeperAdapter struct {
	keeper *Keeper
}

// NewStakingKeeperAdapter creates a new staking keeper adapter
func NewStakingKeeperAdapter(keeper *Keeper) staking.KeeperInterface {
	return &StakingKeeperAdapter{keeper: keeper}
}

// GetQueryContext implements staking.KeeperInterface
func (a *StakingKeeperAdapter) GetQueryContext() (context.Context, error) {
	return a.keeper.GetQueryContext()
}

// GetStakingKeeper implements staking.KeeperInterface
func (a *StakingKeeperAdapter) GetStakingKeeper() staking.StakingKeeperInterface {
	return &StakingKeeperImpl{stakingKeeper: a.keeper.stakingKeeper}
}

// StakingKeeperImpl implements staking.StakingKeeperInterface
type StakingKeeperImpl struct {
	stakingKeeper interface {
		GetAllDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.Delegation, error)
		GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, error)
		GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
		BondDenom(ctx context.Context) (string, error)
		GetAllUnbondingDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.UnbondingDelegation, error)
		GetUnbondingDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.UnbondingDelegation, error)
	}
}

// GetAllDelegatorDelegations implements staking.StakingKeeperInterface
func (s *StakingKeeperImpl) GetAllDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.Delegation, error) {
	return s.stakingKeeper.GetAllDelegatorDelegations(ctx, delegator)
}

// GetDelegation implements staking.StakingKeeperInterface
func (s *StakingKeeperImpl) GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, error) {
	return s.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
}

// GetValidator implements staking.StakingKeeperInterface
func (s *StakingKeeperImpl) GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error) {
	return s.stakingKeeper.GetValidator(ctx, addr)
}

// BondDenom implements staking.StakingKeeperInterface
func (s *StakingKeeperImpl) BondDenom(ctx context.Context) (string, error) {
	return s.stakingKeeper.BondDenom(ctx)
}

// GetAllUnbondingDelegations implements staking.StakingKeeperInterface
func (s *StakingKeeperImpl) GetAllUnbondingDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.UnbondingDelegation, error) {
	return s.stakingKeeper.GetAllUnbondingDelegations(ctx, delegator)
}

// GetUnbondingDelegation implements staking.StakingKeeperInterface
func (s *StakingKeeperImpl) GetUnbondingDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.UnbondingDelegation, error) {
	return s.stakingKeeper.GetUnbondingDelegation(ctx, delAddr, valAddr)
}