package keeper

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CosmosContracts/juno/v18/x/feepay/types"
)

// Check if a contract is registered as a fee pay contract
func (k Keeper) IsContractRegistered(ctx sdk.Context, contractAddr string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyContracts)
	return store.Has([]byte(contractAddr))
}

// Get a contract from KV store
func (k Keeper) GetContract(ctx sdk.Context, contractAddress string) (*types.FeePayContract, error) {
	// Return nil, contract not registered
	if !k.IsContractRegistered(ctx, contractAddress) {
		return nil, types.ErrContractNotRegistered
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyContracts)

	key := []byte(contractAddress)
	bz := store.Get(key)

	var fpc types.FeePayContract
	if err := k.cdc.Unmarshal(bz, &fpc); err != nil {
		return nil, err
	}

	return &fpc, nil
}

// Get all registered fee pay contracts
func (k Keeper) GetContracts(ctx sdk.Context, pag *query.PageRequest) (*types.QueryFeePayContractsResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyContracts)

	// Filter and paginate all contracts
	results, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pag,
		func(key []byte, value *types.FeePayContract) (*types.FeePayContract, error) {
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

// GetAllContracts returns all the registered FeePay contracts.
func (k Keeper) GetAllContracts(ctx sdk.Context) []types.FeePayContract {
	contracts := []types.FeePayContract{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, StoreKeyContracts)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var c types.FeePayContract
		k.cdc.MustUnmarshal(iterator.Value(), &c)

		contracts = append(contracts, c)
	}

	return contracts
}

// Register the contract in the module store
func (k Keeper) RegisterContract(ctx sdk.Context, rfp *types.MsgRegisterFeePayContract) error {
	// Return false because the contract was already registered
	if k.IsContractRegistered(ctx, rfp.FeePayContract.ContractAddress) {
		return types.ErrContractAlreadyRegistered
	}

	// Check if sender is the owner of the cw contract
	contractAddr, err := sdk.AccAddressFromBech32(rfp.FeePayContract.ContractAddress)
	if err != nil {
		return err
	}

	if ok := k.wasmKeeper.HasContractInfo(ctx, contractAddr); !ok {
		return types.ErrInvalidCWContract
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

// Set a contract in the KV Store
func (k Keeper) SetFeePayContract(ctx sdk.Context, feepay types.FeePayContract) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyContracts)
	key := []byte(feepay.ContractAddress)
	bz := k.cdc.MustMarshal(&feepay)
	store.Set(key, bz)
}

// Unregister contract (loop through usage store & remove all usage entries for contract)
func (k Keeper) UnregisterContract(ctx sdk.Context, rfp *types.MsgUnregisterFeePayContract) error {
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
		return types.ErrInvalidCWContract
	}

	// Get the contract info
	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddr)

	// Check if sender is the contract manager
	if ok, err := k.IsContractManager(rfp.SenderAddress, contractInfo); !ok {
		return err
	}

	// Remove contract from KV store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyContracts)
	store.Delete([]byte(rfp.ContractAddress))

	// Remove all usage entries for contract
	store = prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyContractUses)
	iterator := sdk.KVStorePrefixIterator(store, []byte(rfp.ContractAddress))

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}

	// Calculate coins to refund
	coins := sdk.NewCoins(sdk.NewCoin(k.bondDenom, math.NewIntFromUint64(contract.Balance)))

	// Default refund address to admin, fallback to creator
	var refundAddr string
	if contractInfo.Admin != "" {
		refundAddr = contractInfo.Admin
	} else {
		refundAddr = contractInfo.Creator
	}

	// Send coins from the FeePay module to the refund address
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(refundAddr), coins)
}

// Set the contract balance in the KV store
func (k Keeper) SetContractBalance(ctx sdk.Context, fpc *types.FeePayContract, newBalance uint64) {
	// Get the existing contract in KV store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyContracts)

	// Set new balance and save to KV store
	fpc.Balance = newBalance
	store.Set([]byte(fpc.ContractAddress), k.cdc.MustMarshal(fpc))
}

// Fund an existing fee pay contract
func (k Keeper) FundContract(ctx sdk.Context, fpc *types.FeePayContract, senderAddr sdk.AccAddress, coins sdk.Coins) error {
	// Only transfer the bond denom
	var transferCoin sdk.Coin
	for _, c := range coins {
		if c.Denom == k.bondDenom {
			transferCoin = c
		}
	}

	// Ensure the transfer coin was set
	if transferCoin == (sdk.Coin{}) {
		return types.ErrInvalidJunoFundAmount.Wrapf("contract must be funded with '%s'", k.bondDenom)
	}

	// Transfer from sender to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, coins); err != nil {
		return err
	}

	// Increment the fpc balance
	k.SetContractBalance(ctx, fpc, fpc.Balance+transferCoin.Amount.Uint64())
	return nil
}

// Check if a fee pay contract has a balance greater than or equal to the fee
func (k Keeper) CanContractCoverFee(fpc *types.FeePayContract, fee uint64) bool {
	return fpc.Balance >= fee
}

// Get the number of times a wallet has interacted with a fee pay contract (err only if contract not registered)
func (k Keeper) GetContractUses(ctx sdk.Context, fpc *types.FeePayContract, walletAddress string) (uint64, error) {
	// Get usage from store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyContractUses)
	key := []byte(fpc.ContractAddress + "-" + walletAddress)
	bz := store.Get(key)

	var walletUsage types.FeePayWalletUsage
	if err := k.cdc.Unmarshal(bz, &walletUsage); err != nil {
		return 0, err
	}

	return walletUsage.Uses, nil
}

// Set the number of times a wallet has interacted with a fee pay contract
func (k Keeper) IncrementContractUses(ctx sdk.Context, fpc *types.FeePayContract, walletAddress string, increment uint64) error {
	uses, err := k.GetContractUses(ctx, fpc, walletAddress)
	if err != nil {
		return err
	}

	// Get store, key, & value for setting usage
	store := prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyContractUses)
	key := []byte(fpc.ContractAddress + "-" + walletAddress)
	bz, err := k.cdc.Marshal(&types.FeePayWalletUsage{
		ContractAddress: fpc.ContractAddress,
		WalletAddress:   walletAddress,
		Uses:            uses + increment,
	})
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// Check if a wallet exceeded usage limit (defaults to true if contract not registered)
func (k Keeper) HasWalletExceededUsageLimit(ctx sdk.Context, fpc *types.FeePayContract, walletAddress string) bool {
	// Get account uses
	uses, err := k.GetContractUses(ctx, fpc, walletAddress)
	if err != nil {
		return true
	}

	// Return if the wallet has used the contract too many times
	return uses >= fpc.WalletLimit
}

// Update the wallet limit of an existing fee pay contract
func (k Keeper) UpdateContractWalletLimit(ctx sdk.Context, fpc *types.FeePayContract, senderAddress string, walletLimit uint64) error {
	// Check if a cw contract
	contractAddr, err := sdk.AccAddressFromBech32(fpc.ContractAddress)
	if err != nil {
		return err
	}

	if ok := k.wasmKeeper.HasContractInfo(ctx, contractAddr); !ok {
		return types.ErrInvalidCWContract
	}

	// Get the contract info & ensure sender is the manager
	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddr)

	if ok, err := k.IsContractManager(senderAddress, contractInfo); !ok {
		return err
	}

	// Update the store with the new limit
	store := prefix.NewStore(ctx.KVStore(k.storeKey), StoreKeyContracts)
	fpc.WalletLimit = walletLimit
	store.Set([]byte(fpc.ContractAddress), k.cdc.MustMarshal(fpc))

	return nil
}

// Check if a wallet is eligible to interact with a contract
func (k Keeper) IsWalletEligible(ctx sdk.Context, fpc *types.FeePayContract, walletAddress string) (bool, error) {
	// Check if wallet has exceeded usage limit
	if k.HasWalletExceededUsageLimit(ctx, fpc, walletAddress) {
		return false, types.ErrWalletExceededUsageLimit
	}

	// Check if contract has enough funds
	if fpc.Balance == 0 {
		return true, types.ErrContractNotEnoughFunds
	}

	return true, nil
}

// Check if the sender is the designated contract manager for the FeePay contract. If
// an admin is present, they are considered the manager. If there is no admin, the
// contract creator is considered the manager.
func (k Keeper) IsContractManager(senderAddress string, contractInfo *wasmtypes.ContractInfo) (bool, error) {
	// Flags for admin existence & sender being admin/creator
	adminExists := len(contractInfo.Admin) > 0
	isSenderAdmin := contractInfo.Admin == senderAddress
	isSenderCreator := contractInfo.Creator == senderAddress

	if adminExists && !isSenderAdmin {
		return false, types.ErrContractNotAdmin
	} else if !adminExists && !isSenderCreator {
		return false, types.ErrContractNotCreator
	}

	return true, nil
}
