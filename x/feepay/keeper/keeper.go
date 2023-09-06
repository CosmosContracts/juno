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
)

// Keeper of this module maintains collections of feeshares for contracts
// registered to receive transaction fees.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	bankKeeper    revtypes.BankKeeper
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
	bk revtypes.BankKeeper,
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

// Register the contract in the module store
func (k Keeper) RegisterContract(ctx sdk.Context, fpc *types.FeePayContract) error {

	// Return false because the contract was already registered
	if err := k.IsValidContract(ctx, fpc.ContractAddress); err != nil {
		return err
	}

	// Register the new fee pay contract in the KV store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("contracts"))
	key := []byte(fpc.ContractAddress)
	bz := k.cdc.MustMarshal(fpc)

	store.Set(key, bz)
	return nil
}

// Fund an existing fee pay contract
func (k Keeper) FundContract(ctx sdk.Context, mfc *types.MsgFundFeePayContract) error {

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

	// Confirm the sender has enough funds to fund the contract
	addr, err := sdk.AccAddressFromBech32(mfc.SenderAddress)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, sdk.NewCoins(transferCoin))
	if err != nil {
		return err
	}

	// Get existing fee pay contract from store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("contracts"))
	key := []byte(mfc.ContractAddress)
	bz := store.Get(key)

	var fpc types.FeePayContract
	k.cdc.MustUnmarshal(bz, &fpc)

	// Increment the fpc balance
	fpc.Balance += transferCoin.Amount.Uint64()

	// Update the balance in KV store, return success
	store.Set(key, k.cdc.MustMarshal(&fpc))
	return nil
}
