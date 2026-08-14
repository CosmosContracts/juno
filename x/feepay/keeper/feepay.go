package keeper

import (
	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CosmosContracts/juno/v31/app/utils"
	"github.com/CosmosContracts/juno/v31/x/feepay/types"
)

// IsContractRegistered checks if a contract is registered as a feepay contract
func (k Keeper) IsContractRegistered(ctx context.Context, contractAddr string) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	contractsPrefix := prefix.NewStore(store, StoreKeyContracts)
	return contractsPrefix.Has([]byte(contractAddr))
}

// GetContract gets a feepay contract from KV store
func (k Keeper) GetContract(ctx context.Context, contractAddress string) (*types.FeePayContract, error) {
	// Return nil, contract not registered
	if !k.IsContractRegistered(ctx, contractAddress) {
		return nil, utils.ErrContractNotRegistered
	}

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	contractsPrefix := prefix.NewStore(store, StoreKeyContracts)

	key := []byte(contractAddress)
	bz := contractsPrefix.Get(key)

	var fpc types.FeePayContract
	if err := k.cdc.Unmarshal(bz, &fpc); err != nil {
		return nil, err
	}

	return &fpc, nil
}

// GetContracts gets registered feepay contracts, paginated
func (k Keeper) GetContracts(ctx context.Context, pag *query.PageRequest) (*types.QueryFeePayContractsResponse, error) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	contractsPrefix := prefix.NewStore(store, StoreKeyContracts)

	// Filter and paginate all contracts
	results, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		contractsPrefix,
		pag,
		func(_ []byte, value *types.FeePayContract) (*types.FeePayContract, error) {
			return value, nil
		},
		func() *types.FeePayContract {
			return &types.FeePayContract{}
		},
	)
	if err != nil {
		return nil, err
	}

	// Dereference pointer array of contracts
	var contracts []types.FeePayContract
	for _, contract := range results {
		contracts = append(contracts, *contract)
	}

	return &types.QueryFeePayContractsResponse{
		FeePayContracts: contracts,
		Pagination:      pageRes,
	}, nil
}

// GetAllContracts returns ALL registered FeePay contracts.
func (k Keeper) GetAllContracts(ctx context.Context) []types.FeePayContract {
	contracts := []types.FeePayContract{}

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, StoreKeyContracts)
	defer iterator.Close() //nolint:errcheck

	for ; iterator.Valid(); iterator.Next() {
		var c types.FeePayContract
		k.cdc.MustUnmarshal(iterator.Value(), &c)

		contracts = append(contracts, c)
	}

	return contracts
}

// HasOutstandingBalances reports whether changing the configured fee denom
// would reinterpret any existing denomination-less FeePay liability.
func (k Keeper) HasOutstandingBalances(ctx sdk.Context) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, StoreKeyContracts)
	defer iterator.Close() //nolint:errcheck

	for ; iterator.Valid(); iterator.Next() {
		var contract types.FeePayContract
		k.cdc.MustUnmarshal(iterator.Value(), &contract)
		if contract.Balance != 0 {
			return true
		}
	}

	return false
}

// RegisterContract registers a contract in the KV store
func (k Keeper) RegisterContract(ctx context.Context, rfp *types.MsgRegisterFeePayContract) error {
	_, err := sdk.AccAddressFromBech32(rfp.SenderAddress)
	if err != nil {
		return err
	}

	if rfp.FeePayContract.WalletLimit > 1_000_000 {
		return types.ErrInvalidWalletLimit
	}

	// Return false because the contract was already registered
	if k.IsContractRegistered(ctx, rfp.FeePayContract.ContractAddress) {
		return utils.ErrContractAlreadyRegistered
	}

	// Check if sender is the owner of the cw contract
	contractAddr, err := sdk.AccAddressFromBech32(rfp.FeePayContract.ContractAddress)
	if err != nil {
		return err
	}

	if ok := k.wasmKeeper.HasContractInfo(ctx, contractAddr); !ok {
		return utils.ErrInvalidCWContract
	}

	// Get the contract owner
	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddr)

	// Check if sender is contract manager
	if ok, err := k.IsContractManager(rfp.SenderAddress, contractInfo); !ok {
		return err
	}

	k.SetFeePayContract(ctx, *rfp.FeePayContract)
	return nil
}

// SetFeePayContract sets a contract in the KV Store without validation -> ONLY USED IN FEEPAY GENESIS
func (k Keeper) SetFeePayContract(ctx context.Context, feepay types.FeePayContract) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	contractsPrefix := prefix.NewStore(store, StoreKeyContracts)
	key := []byte(feepay.ContractAddress)
	bz := k.cdc.MustMarshal(&feepay)
	contractsPrefix.Set(key, bz)
}

// UnregisterContract UnregisterS contract (loop through usage store & remove all usage entries for contract)
func (k Keeper) UnregisterContract(ctx context.Context, rfp *types.MsgUnregisterFeePayContract) error {
	if _, err := sdk.AccAddressFromBech32(rfp.SenderAddress); err != nil {
		return err
	}

	// Get fee pay contract
	contract, err := k.GetContract(ctx, rfp.ContractAddress)
	if err != nil {
		return err
	}

	// Get contract address
	contractAddr, err := sdk.AccAddressFromBech32(rfp.ContractAddress)
	if err != nil {
		return err
	}

	// Ensure CW contract is valid
	if ok := k.wasmKeeper.HasContractInfo(ctx, contractAddr); !ok {
		return utils.ErrInvalidCWContract
	}

	// Get the contract info
	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddr)

	// Check if sender is the contract manager
	if ok, err := k.IsContractManager(rfp.SenderAddress, contractInfo); !ok {
		return err
	}

	feeDenom, err := k.feeDenom(ctx)
	if err != nil {
		return err
	}

	// Calculate coins to refund
	coins := sdk.NewCoins(sdk.NewCoin(feeDenom, math.NewIntFromUint64(contract.Balance)))

	// Default refund address to admin, fallback to creator
	var refundAddr string
	if contractInfo.Admin != "" {
		refundAddr = contractInfo.Admin
	} else {
		refundAddr = contractInfo.Creator
	}
	refundAccount, err := sdk.AccAddressFromBech32(refundAddr)
	if err != nil {
		return err
	}

	// Complete every fallible operation before deleting the contract ledger and
	// usage state. This keeps direct keeper calls atomic on refund failure, not
	// only calls wrapped by BaseApp's transaction cache.
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, refundAccount, coins); err != nil {
		return err
	}

	// Remove contract from KV store
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, StoreKeyContracts)
	prefixStore.Delete([]byte(rfp.ContractAddress))

	// Remove all usage entries for contract
	prefixStore = prefix.NewStore(store, StoreKeyContractUses)
	iterator := storetypes.KVStorePrefixIterator(prefixStore, []byte(rfp.ContractAddress))
	defer iterator.Close() //nolint:errcheck

	for ; iterator.Valid(); iterator.Next() {
		prefixStore.Delete(iterator.Key())
	}

	return nil
}

// SetContractBalance sets the contract's balance in the KV store
func (k Keeper) SetContractBalance(ctx context.Context, fpc *types.FeePayContract, newBalance uint64) {
	// Get the existing contract in KV store
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	contractsPrefix := prefix.NewStore(store, StoreKeyContracts)

	// Set new balance and save to KV store
	fpc.Balance = newBalance
	contractsPrefix.Set([]byte(fpc.ContractAddress), k.cdc.MustMarshal(fpc))
}

// FundContract funds an existing feepay contract with tokens
func (k Keeper) FundContract(ctx context.Context, fpc *types.FeePayContract, senderAddr sdk.AccAddress, coins sdk.Coins) error {
	feeDenom, err := k.feeDenom(ctx)
	if err != nil {
		return err
	}

	// Only transfer the configured fee denom.
	var transferCoin sdk.Coin
	for _, c := range coins {
		if c.Denom == feeDenom {
			transferCoin = c
		}
	}

	// Ensure the transfer coin was set
	if transferCoin == (sdk.Coin{}) {
		return types.ErrInvalidJunoFundAmount.Wrapf("contract must be funded with '%s'", feeDenom)
	}

	newBalance, err := types.ContractBalanceAfterAddition(fpc.Balance, transferCoin.Amount)
	if err != nil {
		return err
	}

	// Complete all validation before transferring bank funds. A rejected
	// amount must not move coins without a matching uint64 ledger credit.
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(transferCoin)); err != nil {
		return err
	}

	k.SetContractBalance(ctx, fpc, newBalance)
	return nil
}

// CanContractCoverFee checks if a feepay contract has a balance greater than or equal to the fee to pay
func (Keeper) CanContractCoverFee(fpc *types.FeePayContract, fee uint64) bool {
	return fpc.Balance >= fee
}

// GetContractUses gets the number of times a wallet has interacted with a feepay contract (err only if contract not registered)
func (k Keeper) GetContractUses(ctx context.Context, fpc *types.FeePayContract, walletAddress string) (uint64, error) {
	// Get usage from store
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	contractUsesPrefix := prefix.NewStore(store, StoreKeyContractUses)
	key := []byte(fpc.ContractAddress + "-" + walletAddress)
	bz := contractUsesPrefix.Get(key)

	var walletUsage types.FeePayWalletUsage
	if err := k.cdc.Unmarshal(bz, &walletUsage); err != nil {
		return 0, err
	}

	return walletUsage.Uses, nil
}

// IncrementContractUses sets the number of times a wallet has interacted with a feepay contract
func (k Keeper) IncrementContractUses(ctx context.Context, fpc *types.FeePayContract, walletAddress string, increment uint64) error {
	uses, err := k.GetContractUses(ctx, fpc, walletAddress)
	if err != nil {
		return err
	}

	// Get store, key, & value for setting usage
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	contractUsesPrefix := prefix.NewStore(store, StoreKeyContractUses)
	key := []byte(fpc.ContractAddress + "-" + walletAddress)
	bz, err := k.cdc.Marshal(&types.FeePayWalletUsage{
		ContractAddress: fpc.ContractAddress,
		WalletAddress:   walletAddress,
		Uses:            uses + increment,
	})
	if err != nil {
		return err
	}

	contractUsesPrefix.Set(key, bz)
	return nil
}

// HasWalletExceededUsageLimit checks if a wallet exceeded usage limit (defaults to true if contract not registered)
func (k Keeper) HasWalletExceededUsageLimit(ctx context.Context, fpc *types.FeePayContract, walletAddress string) bool {
	// Get account uses
	uses, err := k.GetContractUses(ctx, fpc, walletAddress)
	if err != nil {
		return true
	}

	// Return if the wallet has used the contract too many times
	return uses >= fpc.WalletLimit
}

// UpdateContractWalletLimit updates the wallet limit of an existing feepay contract
func (k Keeper) UpdateContractWalletLimit(ctx context.Context, fpc *types.FeePayContract, senderAddress string, walletLimit uint64) error {
	// Check if a cw contract
	contractAddr, err := sdk.AccAddressFromBech32(fpc.ContractAddress)
	if err != nil {
		return err
	}

	if ok := k.wasmKeeper.HasContractInfo(ctx, contractAddr); !ok {
		return utils.ErrInvalidCWContract
	}

	// Get the contract info & ensure sender is the manager
	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddr)

	if ok, err := k.IsContractManager(senderAddress, contractInfo); !ok {
		return err
	}

	// Update the store with the new limit
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	contractsPrefix := prefix.NewStore(store, StoreKeyContracts)
	fpc.WalletLimit = walletLimit
	contractsPrefix.Set([]byte(fpc.ContractAddress), k.cdc.MustMarshal(fpc))

	return nil
}

// IsWalletEligible checks if a wallet is eligible to interact with a contract
func (k Keeper) IsWalletEligible(ctx context.Context, fpc *types.FeePayContract, walletAddress string) (bool, error) {
	// Check if wallet has exceeded usage limit
	if k.HasWalletExceededUsageLimit(ctx, fpc, walletAddress) {
		return false, types.ErrWalletExceededUsageLimit
	}

	return true, nil
}

// IsContractManager checks if the sender is the designated contract manager for the feepay contract. If
// an admin is present, they are considered the manager. If there is no admin, the
// contract creator is considered the manager.
func (Keeper) IsContractManager(senderAddress string, contractInfo *wasmtypes.ContractInfo) (bool, error) {
	// Flags for admin existence & sender being admin/creator
	adminExists := len(contractInfo.Admin) > 0
	isSenderAdmin := contractInfo.Admin == senderAddress
	isSenderCreator := contractInfo.Creator == senderAddress

	if adminExists && !isSenderAdmin {
		return false, utils.ErrContractNotAdmin
	} else if !adminExists && !isSenderCreator {
		return false, utils.ErrContractNotCreator
	}

	return true, nil
}
