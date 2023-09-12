package keeper

import (
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/CosmosContracts/juno/v17/x/feepay/types"
	revtypes "github.com/CosmosContracts/juno/v17/x/feeshare/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/cosmos/cosmos-sdk/types/query"
)

const (
	StoreKeyContracts    = "contracts"
	StoreKeyContractUses = "contract-uses"
)

// Keeper of this module maintains collections of feeshares for contracts
// registered to receive transaction fees.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	bankKeeper    *bankkeeper.BaseKeeper
	wasmKeeper    wasmkeeper.Keeper
	accountKeeper revtypes.AccountKeeper

	feeCollectorName string
	bondDenom        string

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates new instances of the fees Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	bk *bankkeeper.BaseKeeper,
	wk wasmkeeper.Keeper,
	ak revtypes.AccountKeeper,
	feeCollector string,
	bondDenom string,
	authority string,
) Keeper {
	return Keeper{
		storeKey:         storeKey,
		cdc:              cdc,
		bankKeeper:       bk,
		wasmKeeper:       wk,
		accountKeeper:    ak,
		feeCollectorName: feeCollector,
		bondDenom:        bondDenom,
		authority:        authority,
	}
}

// GetAuthority returns the x/feeshare module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", revtypes.ModuleName))
}

// === TODO: MOVE BELOW FUNCTIONS TO NEW FILE ===

// Check if a contract is registered as a fee pay contract
func (k Keeper) IsRegisteredContract(ctx sdk.Context, contractAddr string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(StoreKeyContracts))
	return store.Has([]byte(contractAddr))
}

// Get a contract from KV store
func (k Keeper) GetContract(ctx sdk.Context, contractAddress string) (*types.FeePayContract, error) {

	// Return nil, contract not registered
	if !k.IsRegisteredContract(ctx, contractAddress) {
		return nil, types.ErrContractNotRegistered
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(StoreKeyContracts))

	key := []byte(contractAddress)
	bz := store.Get(key)

	var fpc types.FeePayContract
	if err := k.cdc.Unmarshal(bz, &fpc); err != nil {
		return nil, err
	}

	return &fpc, nil
}

// Get all registered fee pay contracts
func (k Keeper) GetAllContracts(ctx sdk.Context, req *types.QueryFeePayContracts) (*types.QueryFeePayContractsResponse, error) {

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(StoreKeyContracts))

	// Filter and paginate all contracts
	results, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		req.Pagination,
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
		Contracts:  contracts,
		Pagination: pageRes,
	}, nil
}

// Register the contract in the module store
func (k Keeper) RegisterContract(ctx sdk.Context, rfp *types.MsgRegisterFeePayContract) error {

	// Return false because the contract was already registered
	if k.IsRegisteredContract(ctx, rfp.Contract.ContractAddress) {
		return types.ErrContractAlreadyRegistered
	}

	// Check if sender is the owner of the cw contract
	contractAddr, err := sdk.AccAddressFromBech32(rfp.Contract.ContractAddress)
	if err != nil {
		return err
	}

	// Get the contract owner
	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddr)

	// Check if the sender is first the admin & then the creator (if no admin exists)
	adminExists := len(contractInfo.Admin) > 0
	if adminExists && contractInfo.Admin != rfp.SenderAddress {
		return types.ErrContractNotAdmin
	} else if !adminExists && contractInfo.Creator != rfp.SenderAddress {
		return types.ErrContractNotCreator
	}

	// Register the new fee pay contract in the KV store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(StoreKeyContracts))
	key := []byte(rfp.Contract.ContractAddress)
	bz := k.cdc.MustMarshal(rfp.Contract)

	store.Set(key, bz)
	return nil
}

// Update the contract balance in the KV store
func (k Keeper) UpdateContractBalance(ctx sdk.Context, contractAddress string, newBalance uint64) error {

	// Skip if the contract is not registered
	if !k.IsRegisteredContract(ctx, contractAddress) {
		return types.ErrContractNotRegistered
	}

	// Get the existing contract in KV store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(StoreKeyContracts))

	key := []byte(contractAddress)
	bz := store.Get(key)

	var fpc types.FeePayContract
	if err := k.cdc.Unmarshal(bz, &fpc); err != nil {
		return err
	}

	// Set new balance and save to KV store
	fpc.Balance = newBalance

	store.Set(key, k.cdc.MustMarshal(&fpc))
	return nil
}

// Fund an existing fee pay contract
func (k Keeper) FundContract(ctx sdk.Context, mfc *types.MsgFundFeePayContract) error {
	// Check if the contract is registered
	if !k.IsRegisteredContract(ctx, mfc.ContractAddress) {
		return types.ErrContractNotRegistered
	}

	// Only transfer the bond denom
	var transferCoin sdk.Coin
	for _, c := range mfc.Amount {
		if c.Denom == k.bondDenom {
			transferCoin = c
		}
	}

	// Confirm the sender has enough funds to fund the contract
	addr, err := sdk.AccAddressFromBech32(mfc.SenderAddress)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, sdk.NewCoins(transferCoin)); err != nil {
		return err
	}

	// Get existing fee pay contract from store
	fpc, err := k.GetContract(ctx, mfc.ContractAddress)
	if err != nil {
		return err
	}

	// Increment the fpc balance
	fpc.Balance += transferCoin.Amount.Uint64()
	k.UpdateContractBalance(ctx, mfc.ContractAddress, fpc.Balance)
	return nil
}

// Get the funds associated with a contract
func (k Keeper) GetContractFunds(ctx sdk.Context, contractAddress string) (uint64, error) {
	contract, err := k.GetContract(ctx, contractAddress)

	if err != nil {
		return 0, err
	}

	return contract.Balance, nil
}

// Check if a contract can cover the fees of a transaction
func (k Keeper) CanContractCoverFee(ctx sdk.Context, contractAddress string, fee uint64) bool {

	funds, err := k.GetContractFunds(ctx, contractAddress)

	// Check if contract exists in KV store
	if err != nil {
		return false
	}

	// Check for enough funds
	if funds < fee {
		return false
	}

	return true
}

// Get the number of times a wallet has interacted with a fee pay contract (err only if contract not registered)
func (k Keeper) GetContractUses(ctx sdk.Context, contractAddress string, walletAddress string) (uint64, error) {

	if !k.IsRegisteredContract(ctx, contractAddress) {
		return 0, types.ErrContractNotRegistered
	}

	// Get usage from store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(StoreKeyContractUses))
	key := []byte(contractAddress + "-" + walletAddress)
	bz := store.Get(key)

	var walletUsage types.FeePayWalletUsage
	if err := k.cdc.Unmarshal(bz, &walletUsage); err != nil {
		return 0, err
	}

	return walletUsage.Uses, nil
}

// Set the number of times a wallet has interacted with a fee pay contract
func (k Keeper) SetContractUses(ctx sdk.Context, contractAddress string, walletAddress string, uses uint64) error {

	if !k.IsRegisteredContract(ctx, contractAddress) {
		return types.ErrContractNotRegistered
	}

	// Get store for updating usage
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(StoreKeyContractUses))
	key := []byte(contractAddress + "-" + walletAddress)
	bz, err := k.cdc.Marshal(&types.FeePayWalletUsage{
		ContractAddress: contractAddress,
		WalletAddress:   walletAddress,
		Uses:            uses,
	})

	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// Check if a wallet exceeded usage limit (defaults to true if contract not registered)
func (k Keeper) HasWalletExceededUsageLimit(ctx sdk.Context, contractAddress string, walletAddress string) bool {

	contract, err := k.GetContract(ctx, contractAddress)

	// Check if contract exists in KV store
	if err != nil {
		return true
	}

	// Get account uses
	uses, err := k.GetContractUses(ctx, contractAddress, walletAddress)

	if err != nil {
		return true
	}

	// Return if the wallet has used the contract too many times
	return uses >= contract.WalletLimit
}

// Check if a wallet is eligible to interact with a contract
func (k Keeper) IsWalletEligible(ctx sdk.Context, contractAddress string, walletAddress string) (bool, string) {

	// Check if contract is registered
	if !k.IsRegisteredContract(ctx, contractAddress) {
		return false, types.ErrContractNotRegistered.Error()
	}

	// Check if wallet has exceeded usage limit
	if k.HasWalletExceededUsageLimit(ctx, contractAddress, walletAddress) {
		return false, types.ErrWalletExceededUsageLimit.Error()
	}

	// Check if contract has enough funds
	funds, err := k.GetContractFunds(ctx, contractAddress)

	if err != nil || funds <= 0 {
		return false, types.ErrContractNotEnoughFunds.Error()
	}

	return true, ""
}
