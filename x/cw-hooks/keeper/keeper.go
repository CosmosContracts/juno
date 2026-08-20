package keeper

import (
	"context"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/CosmosContracts/juno/v31/x/cw-hooks/types"
)

type contractIndexes struct {
	ByAddress *indexes.Multi[sdk.AccAddress, collections.Pair[[]byte, sdk.AccAddress], types.ContractInfo]
}

func (i contractIndexes) IndexesList() []collections.Index[collections.Pair[[]byte, sdk.AccAddress], types.ContractInfo] {
	return []collections.Index[collections.Pair[[]byte, sdk.AccAddress], types.ContractInfo]{i.ByAddress}
}

func newContractIndexes(sb *collections.SchemaBuilder) contractIndexes {
	return contractIndexes{
		ByAddress: indexes.NewMulti(
			sb,
			types.ContractsByAddressKey,
			"contracts_by_address",
			sdk.AccAddressKey,
			collections.PairKeyCodec(
				collections.BytesKey,
				sdk.AccAddressKey,
			),
			func(pk collections.Pair[[]byte, sdk.AccAddress], _ types.ContractInfo) (sdk.AccAddress, error) {
				return pk.K2(), nil
			},
		),
	}
}

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService

	stakingKeeper  stakingkeeper.Keeper
	govKeeper      govkeeper.Keeper
	wk             wasmkeeper.Keeper
	contractKeeper wasmtypes.ContractOpsKeeper

	Schema collections.Schema
	Params collections.Item[types.Params]

	Contracts *collections.IndexedMap[collections.Pair[[]byte, sdk.AccAddress], types.ContractInfo, contractIndexes]

	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	ss storetypes.KVStoreService,
	sk stakingkeeper.Keeper,
	gk govkeeper.Keeper,
	wk wasmkeeper.Keeper,
	ck wasmtypes.ContractOpsKeeper,
	authority string,
) Keeper {
	schemaBuilder := collections.NewSchemaBuilder(ss)

	k := Keeper{
		cdc:            cdc,
		storeService:   ss,
		stakingKeeper:  sk,
		govKeeper:      gk,
		contractKeeper: ck,
		authority:      authority,
		wk:             wk,
		Params: collections.NewItem(
			schemaBuilder,
			types.ParamsKey,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		Contracts: collections.NewIndexedMap(
			schemaBuilder,
			types.ContractsKey,
			"contracts",
			collections.PairKeyCodec(
				collections.BytesKey,
				sdk.AccAddressKey,
			),
			codec.CollValue[types.ContractInfo](cdc),
			newContractIndexes(schemaBuilder),
		),
	}

	schema, err := schemaBuilder.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema
	return k
}

// GetAuthority returns the x/cw-hooks module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetContractKeeper returns the x/wasm module's contract keeper.
func (k Keeper) GetContractKeeper() wasmtypes.ContractOpsKeeper {
	return k.contractKeeper
}

func (k Keeper) GetWasmKeeper() wasmkeeper.Keeper {
	return k.wk
}

func (k Keeper) GetStakingKeeper() stakingkeeper.Keeper {
	return k.stakingKeeper
}

func (k Keeper) GetStoreService() storetypes.KVStoreService {
	return k.storeService
}

func (k Keeper) GetCdc() codec.BinaryCodec {
	return k.cdc
}

// Logger returns a module-specific logger.
func (Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}
