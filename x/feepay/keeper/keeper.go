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
func (k Keeper) RegisterContract(ctx sdk.Context, fpc *types.FeePayContract) error {

	// Return false because the contract was already registered
	if k.IsRegisteredContract(ctx, fpc.ContractAddress) {
		return types.ErrContractAlreadyRegistered
	}

	// Register the new fee pay contract in the KV store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("contracts"))
	key := []byte(fpc.ContractAddress)
	bz := k.cdc.MustMarshal(fpc)

	ctx.Logger().Error("Registering contract", "Key", key, "Value", bz)

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
	ctx.Logger().Error("Funding contract", "Contract", mfc.ContractAddress)

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

	ctx.Logger().Error("Funding contract", "Amount", transferCoin)

	// Confirm the sender has enough funds to fund the contract
	addr, err := sdk.AccAddressFromBech32(mfc.SenderAddress)
	if err != nil {
		return err
	}

	ctx.Logger().Error("Funding contract", "Address", addr)

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, sdk.NewCoins(transferCoin)); err != nil {
		ctx.Logger().Error("Funding contract", "Error", err)
		return err
	}

	ctx.Logger().Error("Funding contract", "Sent Coins", true)

	// Get existing fee pay contract from store
	fpc, err := k.GetContract(ctx, mfc.ContractAddress)
	if err != nil {
		return err
	}

	// Increment the fpc balance
	fpc.Balance += transferCoin.Amount.Uint64()
	k.UpdateContractBalance(ctx, mfc.ContractAddress, fpc.Balance)

	ctx.Logger().Error("Funded contract", "New Details", fpc)
	return nil
}