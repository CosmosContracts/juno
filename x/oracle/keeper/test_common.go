package keeper

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v4/modules/core"
	ibchost "github.com/cosmos/ibc-go/v4/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v4/modules/core/keeper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	dbm "github.com/tendermint/tm-db"

	appparams "github.com/CosmosContracts/juno/v12/app/params"
)

var moduleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	distribution.AppModuleBasic{},
	gov.NewAppModuleBasic(
		paramsclient.ProposalHandler, distrclient.ProposalHandler, upgradeclient.ProposalHandler,
	),
	params.AppModuleBasic{},
	crisis.AppModuleBasic{},
	slashing.AppModuleBasic{},
	ibc.AppModuleBasic{},
	upgrade.AppModuleBasic{},
	evidence.AppModuleBasic{},
	transfer.AppModuleBasic{},
	vesting.AppModuleBasic{},
)

func MakeTestCodec(t testing.TB) codec.Codec {
	return MakeEncodingConfig(t).Marshaler
}

func MakeEncodingConfig(_ testing.TB) appparams.EncodingConfig {
	encodingConfig := appparams.MakeEncodingConfig()
	amino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(amino)

	moduleBasics.RegisterLegacyAminoCodec(amino)
	moduleBasics.RegisterInterfaces(interfaceRegistry)

	types.RegisterInterfaces(interfaceRegistry)
	types.RegisterLegacyAminoCodec(amino)

	return encodingConfig
}

var TestingStakeParams = stakingtypes.Params{
	UnbondingTime:     100,
	MaxValidators:     10,
	MaxEntries:        10,
	HistoricalEntries: 10,
	BondDenom:         "stake",
}

type TestFaucet struct {
	t                testing.TB
	bankKeeper       bankkeeper.Keeper
	sender           sdk.AccAddress
	balance          sdk.Coins
	minterModuleName string
}

func NewTestFaucet(t testing.TB, ctx sdk.Context, bankKeeper bankkeeper.Keeper, minterModuleName string, initialAmount ...sdk.Coin) *TestFaucet {
	require.NotEmpty(t, initialAmount)
	r := &TestFaucet{t: t, bankKeeper: bankKeeper, minterModuleName: minterModuleName}
	_, _, addr := keyPubAddr()
	r.sender = addr
	r.Mint(ctx, addr, initialAmount...)
	r.balance = initialAmount
	return r
}

func (f *TestFaucet) Mint(parentCtx sdk.Context, addr sdk.AccAddress, amounts ...sdk.Coin) {
	require.NotEmpty(f.t, amounts)
	ctx := parentCtx.WithEventManager(sdk.NewEventManager()) // discard all faucet related events
	err := f.bankKeeper.MintCoins(ctx, f.minterModuleName, amounts)
	require.NoError(f.t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(ctx, f.minterModuleName, addr, amounts)
	require.NoError(f.t, err)
	f.balance = f.balance.Add(amounts...)
}

func (f *TestFaucet) Fund(parentCtx sdk.Context, receiver sdk.AccAddress, amounts ...sdk.Coin) {
	require.NotEmpty(f.t, amounts)
	// ensure faucet is always filled
	if !f.balance.IsAllGTE(amounts) {
		f.Mint(parentCtx, f.sender, amounts...)
	}
	ctx := parentCtx.WithEventManager(sdk.NewEventManager()) // discard all faucet related events
	err := f.bankKeeper.SendCoins(ctx, f.sender, receiver, amounts)
	require.NoError(f.t, err)
	f.balance = f.balance.Sub(amounts)
}

func (f *TestFaucet) NewFundedRandomAccount(ctx sdk.Context, amounts ...sdk.Coin) sdk.AccAddress {
	_, _, addr := keyPubAddr()
	f.Fund(ctx, addr, amounts...)
	return addr
}

type TestKeepers struct {
	AccountKeeper  authkeeper.AccountKeeper
	StakingKeeper  stakingkeeper.Keeper
	DistKeeper     distributionkeeper.Keeper
	BankKeeper     bankkeeper.Keeper
	GovKeeper      govkeeper.Keeper
	OracleKeeper   Keeper
	IBCKeeper      *ibckeeper.Keeper
	Router         *baseapp.Router
	EncodingConfig appparams.EncodingConfig
	Faucet         *TestFaucet
	MultiStore     sdk.CommitMultiStore
}

// CreateDefaultTestInput common settings for CreateTestInput
func CreateDefaultTestInput(t testing.TB) (sdk.Context, TestKeepers) {
	return CreateTestInput(t, false)
}

// CreateTestInput encoders can be nil to accept the defaults, or set it to override some of the message handlers (like default)
func CreateTestInput(t testing.TB, isCheckTx bool) (sdk.Context, TestKeepers) {
	// Load default wasm config
	return createTestInput(t, isCheckTx, dbm.NewMemDB())
}

func createTestInput(
	t testing.TB,
	isCheckTx bool,
	db dbm.DB,
) (sdk.Context, TestKeepers) {
	tempDir := t.TempDir()

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distributiontypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey,
		capabilitytypes.StoreKey, feegrant.StoreKey, authzkeeper.StoreKey,
		types.StoreKey,
	)
	ms := store.NewCommitMultiStore(db)
	for _, v := range keys {
		ms.MountStoreWithDB(v, sdk.StoreTypeIAVL, db)
	}
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	for _, v := range tkeys {
		ms.MountStoreWithDB(v, sdk.StoreTypeTransient, db)
	}

	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
	for _, v := range memKeys {
		ms.MountStoreWithDB(v, sdk.StoreTypeMemory, db)
	}

	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, isCheckTx, log.NewNopLogger())

	encodingConfig := MakeEncodingConfig(t)
	appCodec, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino

	paramsKeeper := paramskeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)
	for _, m := range []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		minttypes.ModuleName,
		distributiontypes.ModuleName,
		slashingtypes.ModuleName,
		crisistypes.ModuleName,
		ibctransfertypes.ModuleName,
		capabilitytypes.ModuleName,
		ibchost.ModuleName,
		govtypes.ModuleName,
		types.ModuleName,
	} {
		paramsKeeper.Subspace(m)
	}
	subspace := func(m string) paramstypes.Subspace {
		r, ok := paramsKeeper.GetSubspace(m)
		require.True(t, ok)
		return r
	}
	maccPerms := map[string][]string{ // module account permissions
		authtypes.FeeCollectorName:     nil,
		distributiontypes.ModuleName:   nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		types.ModuleName:               nil,
	}
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey], // target store
		subspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, // prototype
		maccPerms,
	)
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		accountKeeper,
		subspace(banktypes.ModuleName),
		blockedAddrs,
	)
	bankKeeper.SetParams(ctx, banktypes.DefaultParams())

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		accountKeeper,
		bankKeeper,
		subspace(stakingtypes.ModuleName),
	)
	stakingKeeper.SetParams(ctx, TestingStakeParams)

	distKeeper := distributionkeeper.NewKeeper(
		appCodec,
		keys[distributiontypes.StoreKey],
		subspace(distributiontypes.ModuleName),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		nil,
	)
	distKeeper.SetParams(ctx, distributiontypes.DefaultParams())
	stakingKeeper.SetHooks(distKeeper.Hooks())

	// set genesis items required for distribution
	distKeeper.SetFeePool(ctx, distributiontypes.InitialFeePool())

	upgradeKeeper := upgradekeeper.NewKeeper(
		map[int64]bool{},
		keys[upgradetypes.StoreKey],
		appCodec,
		tempDir,
		nil,
	)

	faucet := NewTestFaucet(t, ctx, bankKeeper, minttypes.ModuleName, sdk.NewCoin("stake", sdk.NewInt(100_000_000_000)))

	// set some funds ot pay out validatores, based on code from:
	// https://github.com/cosmos/cosmos-sdk/blob/fea231556aee4d549d7551a6190389c4328194eb/x/distribution/keeper/keeper_test.go#L50-L57
	distrAcc := distKeeper.GetDistributionAccount(ctx)
	faucet.Fund(ctx, distrAcc.GetAddress(), sdk.NewCoin("stake", sdk.NewInt(2000000)))
	accountKeeper.SetModuleAccount(ctx, distrAcc)

	capabilityKeeper := capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)
	scopedIBCKeeper := capabilityKeeper.ScopeToModule(ibchost.ModuleName)

	ibcKeeper := ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		subspace(ibchost.ModuleName),
		stakingKeeper,
		upgradeKeeper,
		scopedIBCKeeper,
	)

	router := baseapp.NewRouter()
	bh := bank.NewHandler(bankKeeper)
	router.AddRoute(sdk.NewRoute(banktypes.RouterKey, bh))
	sh := staking.NewHandler(stakingKeeper)
	router.AddRoute(sdk.NewRoute(stakingtypes.RouterKey, sh))
	dh := distribution.NewHandler(distKeeper)
	router.AddRoute(sdk.NewRoute(distributiontypes.RouterKey, dh))

	querier := baseapp.NewGRPCQueryRouter()
	querier.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)

	keeper := NewKeeper(
		appCodec,
		keys[types.StoreKey],
		subspace(types.ModuleName),
		accountKeeper,
		bankKeeper,
		distKeeper,
		stakingKeeper,
		distributiontypes.ModuleName,
	)
	keeper.SetParams(ctx, types.DefaultParams())
	router.AddRoute(sdk.NewRoute(types.RouterKey, TestHandler(keeper)))

	am := module.NewManager( // minimal module set that we use for message/ query tests
		bank.NewAppModule(appCodec, bankKeeper, accountKeeper),
		staking.NewAppModule(appCodec, stakingKeeper, accountKeeper, bankKeeper),
		distribution.NewAppModule(appCodec, distKeeper, accountKeeper, bankKeeper, stakingKeeper),
	)
	am.RegisterServices(module.NewConfigurator(appCodec, msgRouter, querier))
	types.RegisterMsgServer(msgRouter, NewMsgServerImpl(keeper))
	types.RegisterQueryServer(querier, NewQuerier(keeper))

	govRouter := govtypes.NewRouter().
		AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(paramsKeeper)).
		AddRoute(distributiontypes.RouterKey, distribution.NewCommunityPoolSpendProposalHandler(distKeeper)).
		AddRoute(types.RouterKey, NewOracleProposalHandler(keeper))

	govKeeper := govkeeper.NewKeeper(
		appCodec,
		keys[govtypes.StoreKey],
		subspace(govtypes.ModuleName).WithKeyTable(govtypes.ParamKeyTable()),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		govRouter,
	)

	govKeeper.SetProposalID(ctx, govtypes.DefaultStartingProposalID)
	govKeeper.SetDepositParams(ctx, govtypes.DefaultDepositParams())
	govKeeper.SetVotingParams(ctx, govtypes.DefaultVotingParams())
	govKeeper.SetTallyParams(ctx, govtypes.DefaultTallyParams())

	keepers := TestKeepers{
		AccountKeeper:  accountKeeper,
		StakingKeeper:  stakingKeeper,
		DistKeeper:     distKeeper,
		BankKeeper:     bankKeeper,
		GovKeeper:      govKeeper,
		OracleKeeper:   keeper,
		IBCKeeper:      ibcKeeper,
		Router:         router,
		EncodingConfig: encodingConfig,
		Faucet:         faucet,
		MultiStore:     ms,
	}
	return ctx, keepers
}

func TestHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		// ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) { //nolint:gocritic
		default:
			errMsg := fmt.Sprintf("unrecognized wasm message type: %T", msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

var keyCounter uint64

// we need to make this deterministic (same every test run), as encoded address size and thus gas cost,
// depends on the actual bytes (due to ugly CanonicalAddress encoding)
func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) { //nolint:unparam
	keyCounter++
	seed := make([]byte, 8)
	binary.BigEndian.PutUint64(seed, keyCounter)

	key := ed25519.GenPrivKeyFromSecret(seed)
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}
