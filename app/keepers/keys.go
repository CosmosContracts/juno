package keepers

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/types"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v10/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	storetypes "cosmossdk.io/store/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	clocktypes "github.com/CosmosContracts/juno/v30/x/clock/types"
	cwhookstypes "github.com/CosmosContracts/juno/v30/x/cw-hooks/types"
	driptypes "github.com/CosmosContracts/juno/v30/x/drip/types"
	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	feepaytypes "github.com/CosmosContracts/juno/v30/x/feepay/types"
	feesharetypes "github.com/CosmosContracts/juno/v30/x/feeshare/types"
	minttypes "github.com/CosmosContracts/juno/v30/x/mint/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v30/x/tokenfactory/types"
	votingsnapshottypes "github.com/CosmosContracts/juno/v30/x/voting-snapshot/types"
)

func (appKeepers *AppKeepers) GenerateKeys() {
	appKeepers.keys = storetypes.NewKVStoreKeys(
		// cosmos sdk store keys
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		consensusparamtypes.StoreKey,
		upgradetypes.StoreKey,
		feegrant.StoreKey,
		evidencetypes.StoreKey,
		authzkeeper.StoreKey,

		// ibc store keys
		capabilitytypes.StoreKey,
		ibcexported.StoreKey,
		ibctransfertypes.StoreKey,
		icahosttypes.StoreKey,
		icacontrollertypes.StoreKey,
		packetforwardtypes.StoreKey,
		ibchookstypes.StoreKey,

		// wasm store keys
		wasmtypes.StoreKey,

		// juno store keys
		tokenfactorytypes.StoreKey,
		feemarkettypes.StoreKey,
		feepaytypes.StoreKey,
		feesharetypes.StoreKey,
		driptypes.StoreKey,
		clocktypes.StoreKey,
		cwhookstypes.StoreKey,
		votingsnapshottypes.StoreKey,
	)

	appKeepers.memKeys = storetypes.NewMemoryStoreKeys(
		capabilitytypes.MemStoreKey,
	)
}

func (appKeepers *AppKeepers) GetKVStoreKeys() map[string]*storetypes.KVStoreKey {
	return appKeepers.keys
}

func (appKeepers *AppKeepers) GetMemoryStoreKeys() map[string]*storetypes.MemoryStoreKey {
	return appKeepers.memKeys
}

func (appKeepers *AppKeepers) GetKey(storeKey string) *storetypes.KVStoreKey {
	return appKeepers.keys[storeKey]
}

func (appKeepers *AppKeepers) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return appKeepers.memKeys[storeKey]
}
