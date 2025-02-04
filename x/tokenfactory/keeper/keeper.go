package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/log"

	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/store/prefix"
	legacystoretypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"

	"github.com/CosmosContracts/juno/v27/x/tokenfactory/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService storetypes.KVStoreService
		permAddrs    map[string]authtypes.PermissionsForAddress
		permAddrMap  map[string]bool

		accountKeeper      authkeeper.AccountKeeper
		bankKeeper         bankkeeper.Keeper
		distributionKeeper distrkeeper.Keeper

		enabledCapabilities []string

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

// NewKeeper returns a new instance of the x/tokenfactory keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	maccPerms map[string][]string,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	distributionKeeper distrkeeper.Keeper,
	enabledCapabilities []string,
	authority string,
) Keeper {
	permAddrs := make(map[string]authtypes.PermissionsForAddress)
	permAddrMap := make(map[string]bool)
	for name, perms := range maccPerms {
		permsForAddr := authtypes.NewPermissionsForAddress(name, perms)
		permAddrs[name] = permsForAddr
		permAddrMap[permsForAddr.GetAddress().String()] = true
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,

		permAddrs:          permAddrs,
		permAddrMap:        permAddrMap,
		accountKeeper:      accountKeeper,
		bankKeeper:         bankKeeper,
		distributionKeeper: distributionKeeper,

		enabledCapabilities: enabledCapabilities,

		authority: authority,
	}
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetDenomPrefixStore returns the substore for a specific denom
func (k Keeper) GetDenomPrefixStore(ctx context.Context, denom string) legacystoretypes.KVStore {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(store, types.GetDenomPrefixStore(denom))
}

// GetCreatorPrefixStore returns the substore for a specific creator address
func (k Keeper) GetCreatorPrefixStore(ctx context.Context, creator string) legacystoretypes.KVStore {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(store, types.GetCreatorPrefix(creator))
}

// GetCreatorsPrefixStore returns the substore that contains a list of creators
func (k Keeper) GetCreatorsPrefixStore(ctx context.Context) legacystoretypes.KVStore {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(store, types.GetCreatorsPrefix())
}

// CreateModuleAccount creates a module account with minting and burning capabilities
// This account isn't intended to store any coins,
// it purely mints and burns them on behalf of the admin of respective denoms,
// and sends to the relevant address.
func (k Keeper) CreateModuleAccount(ctx context.Context) {
	moduleAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter, authtypes.Burner)
	moduleAccI := (k.accountKeeper.NewAccount(ctx, moduleAcc)).(sdk.ModuleAccountI)
	k.accountKeeper.SetModuleAccount(ctx, moduleAccI)
}
