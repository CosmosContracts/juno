package lupercalia_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/app"
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

var (
	genTokens        = sdk.NewIntFromUint64(100000000000)
	bondTokens       = sdk.NewIntFromUint64(80000000000)
	escapeBondTokens = sdk.NewIntFromUint64(25000000000)
	genCoin          = sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin         = sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)
	escapeBondCoin   = sdk.NewCoin(sdk.DefaultBondDenom, escapeBondTokens)
	maxJunoPerAcc    = sdk.NewIntFromUint64(50000000000)
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *junoapp.App
}

/*
	Test site for lupercalia
*/

func (suite *UpgradeTestSuite) TestAdjustFunds() {
	initialBondPool := sdk.ZeroInt()
	initialUnbondPool := sdk.ZeroInt()

	testCases := []struct {
		msg               string
		pre_adjust_funds  func()
		adjust_funds      func()
		post_adjust_funds func()
		expPass           bool
	}{
		{
			"Test adjusting funds for lupercalia",
			func() {
				suite.ctx = suite.app.BaseApp.NewContext(true, tmproto.Header{})

				initialBondPool = suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.StakingKeeper.GetBondedPool(suite.ctx).GetAddress(), "stake").Amount
				initialUnbondPool = suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.StakingKeeper.GetNotBondedPool(suite.ctx).GetAddress(), "stake").Amount

				// 1. check if bond pool has correct acc2 delegation
				delegation := suite.app.StakingKeeper.GetDelegatorDelegations(suite.ctx, addr2, 120)[0]
				require.Equal(suite.T(), delegation.Shares.RoundInt(), initialBondPool.Sub(bondTokens))

				// 2. check if unbond pool has correct acc2 undelegation
				undelegation := suite.app.StakingKeeper.GetAllUnbondingDelegations(suite.ctx, addr2)[0].Entries[0].Balance
				require.Equal(suite.T(), undelegation, initialUnbondPool)
			},
			func() {
				header := tmproto.Header{Height: suite.app.LastBlockHeight() + 1}
				suite.app.BeginBlock(abci.RequestBeginBlock{Header: header})

				// unbond the accAddr delegations, send all the unbonding and unbonded tokens to the community pool
				bankBaseKeeper, _ := suite.app.BankKeeper.(bankkeeper.BaseKeeper)

				// move all juno from acc to community pool (uncluding bonded juno)
				lupercalia.BurnCoinFromAccount(suite.ctx, addr2, &suite.app.StakingKeeper, &bankBaseKeeper)

				suite.app.EndBlock(abci.RequestEndBlock{})
				suite.app.Commit()
			},
			func() {
				//1. check if fund is moved from unbond and bond pool to community pool
				//acc2 is supposed to lose bondTokens amount to community pool
				afterBondPool := suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.StakingKeeper.GetBondedPool(suite.ctx).GetAddress(), "stake").Amount
				afterUnbondPool := suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.StakingKeeper.GetNotBondedPool(suite.ctx).GetAddress(), "stake").Amount

				initial := initialBondPool.Add(initialUnbondPool)
				later := afterBondPool.Add(afterUnbondPool)
				require.Equal(suite.T(), initial.Sub(bondTokens), later)

				//2. check if acc2 has 0 juno
				afterAcc2Amount := suite.app.BankKeeper.GetBalance(suite.ctx, addr2, sdk.DefaultBondDenom).Amount
				require.Equal(suite.T(), sdk.ZeroInt(), afterAcc2Amount)

				//3. check if all unbonding delegations are removed
				unbondDels := suite.app.StakingKeeper.GetAllUnbondingDelegations(suite.ctx, addr2)

				require.Equal(suite.T(), len(unbondDels), 0)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()
			suite.SetupValidatorDelegator()

			tc.pre_adjust_funds()
			tc.adjust_funds()
			tc.post_adjust_funds()

		})
	}
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

// CheckBalance checks the balance of an account.
func checkBalance(t *testing.T, app *junoapp.App, addr sdk.AccAddress, balances sdk.Coins) {
	ctxCheck := app.BaseApp.NewContext(true, tmproto.Header{})
	require.True(t, balances.IsEqual(app.BankKeeper.GetAllBalances(ctxCheck, addr)))
}

func (suite *UpgradeTestSuite) SetupTest() {
	// acc1 is to create validator
	acc1 := &authtypes.BaseAccount{Address: addr1.String()}
	// acc2 is to delegate funds to acc1 validator
	acc2 := &authtypes.BaseAccount{Address: addr2.String()}

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

	suite.app = app.SetupWithGenesisAccounts(accs, balances...)
	suite.ctx = suite.app.BaseApp.NewContext(true, tmproto.Header{})
	checkBalance(suite.T(), suite.app, addr1, sdk.Coins{genCoin})
	checkBalance(suite.T(), suite.app, addr2, sdk.Coins{genCoin})
}

func (suite *UpgradeTestSuite) SetupValidatorDelegator() {
	// create validator
	description := types.NewDescription("acc1", "", "", "", "")
	createValidatorMsg, err := types.NewMsgCreateValidator(
		sdk.ValAddress(addr1), valKey.PubKey(), bondCoin, description, commissionRates, sdk.OneInt(),
	)
	require.NoError(suite.T(), err)

	header := tmproto.Header{Height: suite.app.LastBlockHeight() + 1}
	txGen := cosmoscmd.MakeEncodingConfig(junoapp.ModuleBasics).TxConfig
	_, _, err = cosmossimapp.SignCheckDeliver(suite.T(), txGen, suite.app.BaseApp, header, []sdk.Msg{createValidatorMsg}, "", []uint64{0}, []uint64{0}, true, true, priv1)
	require.NoError(suite.T(), err)
	checkBalance(suite.T(), suite.app, addr1, sdk.Coins{genCoin.Sub(bondCoin)})

	header = tmproto.Header{Height: suite.app.LastBlockHeight() + 1}
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: header})

	validator := checkValidator(suite.T(), suite.app, sdk.ValAddress(addr1), true)
	require.Equal(suite.T(), sdk.ValAddress(addr1).String(), validator.OperatorAddress)
	require.Equal(suite.T(), types.Bonded, validator.Status)
	require.True(sdk.IntEq(suite.T(), bondTokens, validator.BondedTokens()))

	header = tmproto.Header{Height: suite.app.LastBlockHeight() + 1}
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// delegate
	checkBalance(suite.T(), suite.app, addr2, sdk.Coins{genCoin})
	delegateMsg := types.NewMsgDelegate(addr2, sdk.ValAddress(addr1), bondCoin)

	header = tmproto.Header{Height: suite.app.LastBlockHeight() + 1}
	_, _, err = cosmossimapp.SignCheckDeliver(suite.T(), txGen, suite.app.BaseApp, header, []sdk.Msg{delegateMsg}, "", []uint64{1}, []uint64{0}, true, true, priv2)
	require.NoError(suite.T(), err)

	checkBalance(suite.T(), suite.app, addr2, sdk.Coins{genCoin.Sub(bondCoin)})
	checkDelegation(suite.T(), suite.app, addr2, sdk.ValAddress(addr1), true, bondTokens.ToDec())

	// begin unbonding half
	beginUnbondingMsg := types.NewMsgUndelegate(addr2, sdk.ValAddress(addr1), escapeBondCoin)
	header = tmproto.Header{Height: suite.app.LastBlockHeight() + 1}
	_, _, err = cosmossimapp.SignCheckDeliver(suite.T(), txGen, suite.app.BaseApp, header, []sdk.Msg{beginUnbondingMsg}, "", []uint64{1}, []uint64{1}, true, true, priv2)
	require.NoError(suite.T(), err)

	// delegation should be halved through unbonding cheat to avoid lupercalia hunt
	checkDelegation(suite.T(), suite.app, addr2, sdk.ValAddress(addr1), true, bondTokens.Sub(escapeBondTokens).ToDec())

	// balance should be the same because bonding not yet complete
	checkBalance(suite.T(), suite.app, addr2, sdk.Coins{genCoin.Sub(bondCoin)})
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}
