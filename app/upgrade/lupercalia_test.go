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
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

	fmt.Printf("Acc2 balances before adjust = %v \n", app.BankKeeper.GetAllBalances(ctxCheck, addr2))

	lupercalia.AdjustDelegation(ctxCheck, &app.StakingKeeper, addr2)

	fmt.Printf("Acc2 balances after adjust = %v \n", app.BankKeeper.GetAllBalances(ctxCheck, addr2))
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
	genTokens := sdk.NewIntFromUint64(1000000000000)
	bondTokens := sdk.NewIntFromUint64(500000000000)
	escapeBondTokens := sdk.NewIntFromUint64(250000000000)
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

	// begin unbonding half to escape lupercalia hunt
	beginUnbondingMsg := types.NewMsgUndelegate(addr2, sdk.ValAddress(addr1), escapeBondCoin)
	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = cosmossimapp.SignCheckDeliver(t, txGen, app.BaseApp, header, []sdk.Msg{beginUnbondingMsg}, "", []uint64{1}, []uint64{1}, true, true, priv2)
	require.NoError(t, err)

	// delegation should be halved through unbonding cheat to avoid lupercalia hunt
	checkDelegation(t, app, addr2, sdk.ValAddress(addr1), true, escapeBondTokens.ToDec())

	// balance should be the same because bonding not yet complete
	checkBalance(t, app, addr2, sdk.Coins{genCoin.Sub(bondCoin)})

	lupercaliaHunt(app)
}
