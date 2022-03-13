package lupercalia_test

import (
	"fmt"
	"testing"

	junoapp "github.com/CosmosContracts/juno/app"
	lupercalia "github.com/CosmosContracts/juno/app/upgrade"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	cosmossimapp "github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/starport/starport/pkg/cosmoscmd"
)

var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = secp256k1.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())

	valKey = ed25519.GenPrivKey()

	commissionRates = types.NewCommissionRates(sdk.NewDecWithPrec(5, 2), sdk.NewDecWithPrec(5, 2), sdk.NewDecWithPrec(5, 2))
)

/*
	Test site for lupercalia
*/
func lupercaliaHunt(
	app *junoapp.App,
) {
	ctxCheck := app.BaseApp.NewContext(true, tmproto.Header{})

	newBonded := app.BankKeeper.GetBalance(ctxCheck, app.StakingKeeper.GetBondedPool(ctxCheck).GetAddress(), "stake").Amount
	newNotBonded := app.BankKeeper.GetBalance(ctxCheck, app.StakingKeeper.GetNotBondedPool(ctxCheck).GetAddress(), "stake").Amount

	fmt.Printf("before bonded = %v, before notBonded = %v \n", newBonded, newNotBonded)

	completionTime := ctxCheck.BlockHeader().Time.Add(app.StakingKeeper.UnbondingTime(ctxCheck))
	dvPair := app.StakingKeeper.GetUBDQueueTimeSlice(ctxCheck, completionTime)

	for i, pair := range dvPair {
		fmt.Printf("before UBD queue item %d = \n%v \n", i, pair)
	}

	unbondDels := app.StakingKeeper.GetAllUnbondingDelegations(ctxCheck, addr2)

	for i, unbond := range unbondDels {
		fmt.Printf("before unbond item %d = \n%v \n", i, unbond)
	}

	coin := app.BankKeeper.GetAllBalances(ctxCheck, addr2)
	fmt.Printf("before addr 2 amount = %v \n", coin)

	distAmount := app.BankKeeper.GetBalance(ctxCheck, app.AccountKeeper.GetModuleAccount(ctxCheck, distrtypes.ModuleName).GetAddress(), "stake")
	fmt.Printf("before Distribution module amount = %v \n", distAmount)

	fmt.Printf("===ADJUSTING DELEGATION=== \n")

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// unbond the accAddr delegations, send all the unbonding and unbonded tokens to the community pool
	bankBaseKeeper, _ := app.BankKeeper.(bankkeeper.BaseKeeper)

	lupercalia.MoveDelegatorDelegationsToCommunityPool(ctxCheck, addr2, &app.StakingKeeper, &bankBaseKeeper, &app.DistrKeeper)
	// send 50k juno from the community pool to the accAddr if the master account has less than 50k juno
	accAddrAmount := bankBaseKeeper.GetBalance(ctxCheck, addr2, app.StakingKeeper.BondDenom(ctxCheck)).Amount
	if sdk.NewIntFromUint64(50000000000).GT(accAddrAmount) {
		bankBaseKeeper.SendCoinsFromModuleToAccount(ctxCheck, distrtypes.ModuleName, addr2, sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctxCheck), sdk.NewIntFromUint64(50000000000).Sub(accAddrAmount))))
	}

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	fmt.Printf("===AFTER ADJUSTING DELEGATION=== \n")

	newBonded = app.BankKeeper.GetBalance(ctxCheck, app.StakingKeeper.GetBondedPool(ctxCheck).GetAddress(), "stake").Amount
	newNotBonded = app.BankKeeper.GetBalance(ctxCheck, app.StakingKeeper.GetNotBondedPool(ctxCheck).GetAddress(), "stake").Amount

	fmt.Printf("bonded = %v, notBonded = %v \n", newBonded, newNotBonded)

	dvPair = app.StakingKeeper.GetUBDQueueTimeSlice(ctxCheck, completionTime)

	for i, pair := range dvPair {
		fmt.Printf("UBD queue item %d = \n%v \n", i, pair)
	}

	unbondDels = app.StakingKeeper.GetAllUnbondingDelegations(ctxCheck, addr2)

	for i, unbond := range unbondDels {
		fmt.Printf("unbond item %d = \n%v \n", i, unbond)
	}

	coin = app.BankKeeper.GetAllBalances(ctxCheck, addr2)
	fmt.Printf("addr 2 amount = %v \n", coin)

	distAmount = app.BankKeeper.GetBalance(ctxCheck, app.AccountKeeper.GetModuleAccount(ctxCheck, distrtypes.ModuleName).GetAddress(), "stake")
	fmt.Printf("Distribution module amount = %v \n", distAmount)
}

func checkValidator(t *testing.T, app *junoapp.App, addr sdk.ValAddress, expFound bool) types.Validator {
	ctxCheck := app.BaseApp.NewContext(true, tmproto.Header{})
	validator, found := app.StakingKeeper.GetValidator(ctxCheck, addr)

	require.Equal(t, expFound, found)
	return validator
}

func checkDelegation(
	t *testing.T, app *junoapp.App, delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress, expFound bool, expShares sdk.Dec,
) {

	ctxCheck := app.BaseApp.NewContext(true, tmproto.Header{})
	delegation, found := app.StakingKeeper.GetDelegation(ctxCheck, delegatorAddr, validatorAddr)
	if expFound {
		require.True(t, found)
		require.True(sdk.DecEq(t, expShares, delegation.Shares))

		return
	}

	require.False(t, found)
}

func TestUndelegate(t *testing.T) {
	genTokens := sdk.NewIntFromUint64(100000000000)
	bondTokens := sdk.NewIntFromUint64(80000000000)
	escapeBondTokens := sdk.NewIntFromUint64(25000000000)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)
	escapeBondCoin := sdk.NewCoin(sdk.DefaultBondDenom, escapeBondTokens)

	// acc1 is to create validator
	acc1 := &authtypes.BaseAccount{Address: addr1.String()}
	// acc2 is to delegate funds to acc1 validator
	acc2 := &authtypes.BaseAccount{Address: addr2.String()}

	fmt.Printf("acc1 val address = %s, acc2 address = %s \n", sdk.ValAddress(addr1).String(), addr2.String())

	accs := authtypes.GenesisAccounts{acc1, acc2}
	balances := []banktypes.Balance{
		{
			Address: addr1.String(),
			Coins:   sdk.Coins{genCoin},
		},
		{
			Address: addr2.String(),
			Coins:   sdk.Coins{genCoin},
		},
	}

	app := setupWithGenesisAccounts(accs, balances...)
	checkBalance(t, app, addr1, sdk.Coins{genCoin})
	checkBalance(t, app, addr2, sdk.Coins{genCoin})

	// create validator
	description := types.NewDescription("acc1", "", "", "", "")
	createValidatorMsg, err := types.NewMsgCreateValidator(
		sdk.ValAddress(addr1), valKey.PubKey(), bondCoin, description, commissionRates, sdk.OneInt(),
	)
	require.NoError(t, err)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	txGen := cosmoscmd.MakeEncodingConfig(junoapp.ModuleBasics).TxConfig
	_, _, err = cosmossimapp.SignCheckDeliver(t, txGen, app.BaseApp, header, []sdk.Msg{createValidatorMsg}, "", []uint64{0}, []uint64{0}, true, true, priv1)
	require.NoError(t, err)
	checkBalance(t, app, addr1, sdk.Coins{genCoin.Sub(bondCoin)})

	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	validator := checkValidator(t, app, sdk.ValAddress(addr1), true)
	require.Equal(t, sdk.ValAddress(addr1).String(), validator.OperatorAddress)
	require.Equal(t, types.Bonded, validator.Status)
	require.True(sdk.IntEq(t, bondTokens, validator.BondedTokens()))

	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// delegate
	checkBalance(t, app, addr2, sdk.Coins{genCoin})
	delegateMsg := types.NewMsgDelegate(addr2, sdk.ValAddress(addr1), bondCoin)

	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = cosmossimapp.SignCheckDeliver(t, txGen, app.BaseApp, header, []sdk.Msg{delegateMsg}, "", []uint64{1}, []uint64{0}, true, true, priv2)
	require.NoError(t, err)

	checkBalance(t, app, addr2, sdk.Coins{genCoin.Sub(bondCoin)})
	checkDelegation(t, app, addr2, sdk.ValAddress(addr1), true, bondTokens.ToDec())

	// begin unbonding half
	beginUnbondingMsg := types.NewMsgUndelegate(addr2, sdk.ValAddress(addr1), escapeBondCoin)
	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = cosmossimapp.SignCheckDeliver(t, txGen, app.BaseApp, header, []sdk.Msg{beginUnbondingMsg}, "", []uint64{1}, []uint64{1}, true, true, priv2)
	require.NoError(t, err)

	// delegation should be halved through unbonding cheat to avoid lupercalia hunt
	bondTokens.Sub(escapeBondTokens)
	checkDelegation(t, app, addr2, sdk.ValAddress(addr1), true, bondTokens.Sub(escapeBondTokens).ToDec())

	// balance should be the same because bonding not yet complete
	checkBalance(t, app, addr2, sdk.Coins{genCoin.Sub(bondCoin)})

	lupercaliaHunt(app)
}
