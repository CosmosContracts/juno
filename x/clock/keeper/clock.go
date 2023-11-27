package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v18/x/clock/types"
)

// Store Keys for clock contracts (both jailed and unjailed)
var (
	StoreKeyContracts       = []byte("contracts")
	StoreKeyJailedContracts = []byte("jailed-contracts")
)

// Get the store key for either jailed or unjailed contracts.
func (k Keeper) getStore(ctx sdk.Context, isJailed bool) prefix.Store {
	if isJailed {
		return prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyJailedContracts)
	}

	return prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyContracts)
}

// Set a clock contract address in the KV store.
func (k Keeper) SetClockContract(ctx sdk.Context, contractAddress string, isJailed bool) {
	store := k.getStore(ctx, isJailed)
	store.Set([]byte(contractAddress), []byte(contractAddress))
}

// Check if a clock contract address is in the KV store.
func (k Keeper) IsClockContract(ctx sdk.Context, contractAddress string, isJailed bool) bool {
	store := k.getStore(ctx, isJailed)
	return store.Has([]byte(contractAddress))
}

// Get all clock contract addresses from the KV store.
func (k Keeper) GetAllContracts(ctx sdk.Context, isJailed bool) []string {
	// Get the KV store
	store := k.getStore(ctx, isJailed)

	// Iterate over all contracts
	iterator := sdk.KVStorePrefixIterator(store, []byte(nil))
	defer iterator.Close()

	contracts := []string{}
	for ; iterator.Valid(); iterator.Next() {
		contracts = append(contracts, string(iterator.Value()))
	}

	// Return array of contract addresses
	return contracts
}

// Remove a clock contract address from the KV store.
func (k Keeper) RemoveContract(ctx sdk.Context, contractAddress string, isJailed bool) {
	store := k.getStore(ctx, isJailed)
	key := []byte(contractAddress)

	if store.Has(key) {
		store.Delete(key)
	}
}

// Register a clock contract address in the KV store.
func (k Keeper) RegisterContract(ctx sdk.Context, senderAddress string, contractAddress string) error {
	// Check if the contract is already registered
	if k.IsClockContract(ctx, contractAddress, false) {
		return types.ErrContractAlreadyRegistered
	}

	// Check if the contract is already jailed
	if k.IsClockContract(ctx, contractAddress, true) {
		return types.ErrContractJailed
	}

	// Ensure the sender is the contract admin or creator
	if ok, err := k.IsContractManager(ctx, senderAddress, contractAddress); !ok {
		return err
	}

	// Register contract
	k.SetClockContract(ctx, contractAddress, false)
	return nil
}

// Unregister a clock contract from either the jailed or unjailed KV store.
func (k Keeper) UnregisterContract(ctx sdk.Context, senderAddress string, contractAddress string) error {
	// Check if the contract is registered in either store
	if !k.IsClockContract(ctx, contractAddress, false) && !k.IsClockContract(ctx, contractAddress, true) {
		return types.ErrContractNotRegistered
	}

	// Ensure the sender is the contract admin or creator
	if ok, err := k.IsContractManager(ctx, senderAddress, contractAddress); !ok {
		return err
	}

	// Remove contract from both stores
	k.RemoveContract(ctx, contractAddress, false)
	k.RemoveContract(ctx, contractAddress, true)
	return nil
}

// Jail a clock contract in the jailed KV store.
func (k Keeper) JailContract(ctx sdk.Context, contractAddress string) error {
	// Check if the contract is registered in the unjailed store
	if !k.IsClockContract(ctx, contractAddress, false) {
		return types.ErrContractNotRegistered
	}

	// Remove contract from unjailed store
	k.RemoveContract(ctx, contractAddress, false)

	// Set contract in jailed store
	k.SetClockContract(ctx, contractAddress, true)
	return nil
}

// Unjail a clock contract from the jailed KV store.
func (k Keeper) UnjailContract(ctx sdk.Context, senderAddress string, contractAddress string) error {
	// Check if the contract is registered in the jailed store
	if !k.IsClockContract(ctx, contractAddress, true) {
		return types.ErrContractNotJailed
	}

	// Ensure the sender is the contract admin or creator
	if ok, err := k.IsContractManager(ctx, senderAddress, contractAddress); !ok {
		return err
	}

	// Remove contract from jailed store
	k.RemoveContract(ctx, contractAddress, true)

	// Set contract in unjailed store
	k.SetClockContract(ctx, contractAddress, false)
	return nil
}

// Check if the sender is the designated contract manager for the FeePay contract. If
// an admin is present, they are considered the manager. If there is no admin, the
// contract creator is considered the manager.
func (k Keeper) IsContractManager(ctx sdk.Context, senderAddress string, contractAddress string) (bool, error) {
	contractAddr := sdk.MustAccAddressFromBech32(contractAddress)

	// Ensure the contract is a cosm wasm contract
	if ok := k.wasmKeeper.HasContractInfo(ctx, contractAddr); !ok {
		return false, types.ErrInvalidCWContract
	}

	// Get the contract info
	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddr)

	// Flags for admin existence & sender being admin/creator
	adminExists := len(contractInfo.Admin) > 0
	isSenderAdmin := contractInfo.Admin == senderAddress
	isSenderCreator := contractInfo.Creator == senderAddress

	// Check if the sender is the admin or creator
	if adminExists && !isSenderAdmin {
		return false, types.ErrContractNotAdmin
	} else if !adminExists && !isSenderCreator {
		return false, types.ErrContractNotCreator
	}

	return true, nil
}
