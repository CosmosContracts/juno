package keeper

import (
	"errors"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	types "github.com/CosmosContracts/juno/v17/x/feepay/types"
	revtypes "github.com/CosmosContracts/juno/v17/x/feeshare/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

const (
	StoreKeyContracts = "contracts"
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

// ===

// Check if a contract is associated with a FeePay contract
func (k Keeper) IsValidContract(ctx sdk.Context, contractAddr string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("contracts"))

	hasData := store.Has([]byte(contractAddr))

	if hasData {
		return nil
	} else {
		return errorsmod.Wrapf(errors.New("invalid contract address"), "contract %s not registered", contractAddr)
	}
}

// Get a contract from KV store
func (k Keeper) GetContract(ctx sdk.Context, contractAddress string) (*types.FeePayContract, error) {

	// Return false because the contract was already registered
	if err := k.IsValidContract(ctx, contractAddress); err != nil {
		return nil, err
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

func (k Keeper) UpdateContractBalance(ctx sdk.Context, contractAddress string, newBalance uint64) error {

	// TODO: Do we need torecheck this? (remove if duplicate)
	// Return false because the contract was already registered
	if err := k.IsValidContract(ctx, contractAddress); err != nil {
		return err
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(StoreKeyContracts))

	key := []byte(contractAddress)
	bz := store.Get(key)

	var fpc types.FeePayContract
	if err := k.cdc.Unmarshal(bz, &fpc); err != nil {
		return err
	}

	fpc.Balance = newBalance

	store.Set(key, k.cdc.MustMarshal(&fpc))
	return nil
}

// Get all registered fee pay contracts
func (k Keeper) GetAllContracts(ctx sdk.Context) ([]types.FeePayContract, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(StoreKeyContracts))

	var contracts []types.FeePayContract
	iterator := sdk.KVStorePrefixIterator(store, nil)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var contract types.FeePayContract
		k.cdc.MustUnmarshal(iterator.Value(), &contract)
		contracts = append(contracts, contract)
	}

	return contracts, nil
}

// Register the contract in the module store
func (k Keeper) RegisterContract(ctx sdk.Context, fpc *types.FeePayContract) error {

	// Return false because the contract was already registered
	if err := k.IsValidContract(ctx, fpc.ContractAddress); err == nil {
		return errorsmod.Wrapf(errors.New("contract already registered"), "contract %s already registered", fpc.ContractAddress)
	}

	ctx.Logger().Error("Registering contract", "Contract is not registered!")

	// Register the new fee pay contract in the KV store
	// TODO: change this to be the function GetContract()
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("contracts"))
	key := []byte(fpc.ContractAddress)
	bz := k.cdc.MustMarshal(fpc)

	ctx.Logger().Error("Registering contract", "Key", key, "Value", bz)

	store.Set(key, bz)
	return nil
}

// Fund an existing fee pay contract
func (k Keeper) FundContract(ctx sdk.Context, mfc *types.MsgFundFeePayContract) error {

	ctx.Logger().Error("Funding contract", "Contract", mfc.ContractAddress)

	// Return false because the contract was already registered
	if err := k.IsValidContract(ctx, mfc.ContractAddress); err != nil {
		return err
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
