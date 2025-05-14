package keeper

import (
	corestoretypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// KeeperWrapper is a wrapper around the cosmos sdk gov module keeper
type KeeperWrapper struct {
	*govkeeper.Keeper
	ak govtypes.AccountKeeper
}

// NewKeeper returns a new wrapped gov module keeper instance.
func NewKeeper(
	cdc codec.Codec, storeService corestoretypes.KVStoreService, authKeeper govtypes.AccountKeeper,
	bankKeeper govtypes.BankKeeper, stakingKeeper govtypes.StakingKeeper, distrKeeper govtypes.DistributionKeeper,
	router baseapp.MessageRouter, config govtypes.Config, authority string,
) KeeperWrapper {
	return KeeperWrapper{
		Keeper: govkeeper.NewKeeper(cdc, storeService, authKeeper, bankKeeper, stakingKeeper, distrKeeper, router, config, authority),
		ak:     authKeeper,
	}
}

// SetHooks sets the hooks for governance on the original keeper
// and returns the updated wrapped keeper.
func (k *KeeperWrapper) SetHooks(gh govtypes.GovHooks) *KeeperWrapper {
	keeper := k.Keeper.SetHooks(gh)

	k.Keeper = keeper

	return k
}
