package bindings_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/CosmosContracts/juno/v27/wasmbindings/types"
	tftypes "github.com/CosmosContracts/juno/v27/x/tokenfactory/types"
)

func (s *BindingsTestSuite) TestCreateDenomMsg() {
	s.SetupTest()
	creator := s.RandomAccountAddress()
	s.StoreReflectCode(creator)

	lucky := s.RandomAccountAddress()
	reflect := s.instantiateReflectContract(lucky)
	s.Require().NotEmpty(reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(tftypes.DefaultParams().DenomCreationFee[0].Denom, tftypes.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	s.FundAcc(reflect, reflectAmount)

	msg := types.TokenFactoryMsg{CreateDenom: &types.CreateDenom{
		Subdenom: "SUN",
	}}
	err := s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)

	// query the denom and see if it matches
	query := types.TokenFactoryQuery{
		FullDenom: &types.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp := types.FullDenomResponse{}
	s.queryCustom(reflect, query, &resp)

	s.Require().Equal(resp.Denom, fmt.Sprintf("factory/%s/SUN", reflect.String()))
}

func (s *BindingsTestSuite) TestMintMsg() {
	s.SetupTest()
	creator := s.RandomAccountAddress()
	s.StoreReflectCode(creator)
	lucky := s.RandomAccountAddress()
	reflect := s.instantiateReflectContract(lucky)
	s.Require().NotEmpty(reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(tftypes.DefaultParams().DenomCreationFee[0].Denom, tftypes.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	s.FundAcc(reflect, reflectAmount)

	// lucky was broke
	balances := s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, lucky)
	s.Require().Empty(balances)

	// Create denom for minting
	msg := types.TokenFactoryMsg{CreateDenom: &types.CreateDenom{
		Subdenom: "SUN",
	}}
	err := s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)
	sunDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount, ok := sdkmath.NewIntFromString("808010808")
	s.Require().True(ok)
	msg = types.TokenFactoryMsg{MintTokens: &types.MintTokens{
		Denom:         sunDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)

	balances = s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, lucky)
	s.Require().Len(balances, 1)
	coin := balances[0]
	s.Require().Equal(amount, coin.Amount)
	s.Require().Contains(coin.Denom, "factory/")

	// query the denom and see if it matches
	query := types.TokenFactoryQuery{
		FullDenom: &types.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp := types.FullDenomResponse{}
	s.queryCustom(reflect, query, &resp)

	s.Require().Equal(resp.Denom, coin.Denom)

	// mint the same denom again
	err = s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)

	balances = s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, lucky)
	s.Require().Len(balances, 1)
	coin = balances[0]
	s.Require().Equal(amount.MulRaw(2), coin.Amount)
	s.Require().Contains(coin.Denom, "factory/")

	// query the denom and see if it matches
	query = types.TokenFactoryQuery{
		FullDenom: &types.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp = types.FullDenomResponse{}
	s.queryCustom(reflect, query, &resp)

	s.Require().Equal(resp.Denom, coin.Denom)

	// now mint another amount / denom
	// create it first
	msg = types.TokenFactoryMsg{CreateDenom: &types.CreateDenom{
		Subdenom: "MOON",
	}}
	err = s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)
	moonDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount = amount.SubRaw(1)
	msg = types.TokenFactoryMsg{MintTokens: &types.MintTokens{
		Denom:         moonDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)

	balances = s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, lucky)
	s.Require().Len(balances, 2)
	coin = balances[0]
	s.Require().Equal(amount, coin.Amount)
	s.Require().Contains(coin.Denom, "factory/")

	// query the denom and see if it matches
	query = types.TokenFactoryQuery{
		FullDenom: &types.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "MOON",
		},
	}
	resp = types.FullDenomResponse{}
	s.queryCustom(reflect, query, &resp)

	s.Require().Equal(resp.Denom, coin.Denom)

	// and check the first denom is unchanged
	coin = balances[1]
	s.Require().Equal(amount.AddRaw(1).MulRaw(2), coin.Amount)
	s.Require().Contains(coin.Denom, "factory/")

	// query the denom and see if it matches
	query = types.TokenFactoryQuery{
		FullDenom: &types.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp = types.FullDenomResponse{}
	s.queryCustom(reflect, query, &resp)

	s.Require().Equal(resp.Denom, coin.Denom)
}

func (s *BindingsTestSuite) TestForceTransfer() {
	s.SetupTest()
	creator := s.RandomAccountAddress()
	s.StoreReflectCode(creator)
	lucky := s.RandomAccountAddress()
	rcpt := s.RandomAccountAddress()
	reflect := s.instantiateReflectContract(lucky)
	s.Require().NotEmpty(reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(tftypes.DefaultParams().DenomCreationFee[0].Denom, tftypes.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	s.FundAcc(reflect, reflectAmount)

	// lucky was broke
	balances := s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, lucky)
	s.Require().Empty(balances)

	// Create denom for minting
	msg := types.TokenFactoryMsg{CreateDenom: &types.CreateDenom{
		Subdenom: "SUN",
	}}
	err := s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)
	sunDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount, ok := sdkmath.NewIntFromString("808010808")
	s.Require().True(ok)

	// Mint new tokens to lucky
	msg = types.TokenFactoryMsg{MintTokens: &types.MintTokens{
		Denom:         sunDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)

	// Force move 100 tokens from lucky to rcpt
	msg = types.TokenFactoryMsg{ForceTransfer: &types.ForceTransfer{
		Denom:       sunDenom,
		Amount:      sdkmath.NewInt(100),
		FromAddress: lucky.String(),
		ToAddress:   rcpt.String(),
	}}
	err = s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)

	// check the balance of rcpt
	balances = s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, rcpt)
	s.Require().Len(balances, 1)
	coin := balances[0]
	s.Require().Equal(sdkmath.NewInt(100), coin.Amount)
}

func (s *BindingsTestSuite) TestBurnMsg() {
	s.SetupTest()
	creator := s.RandomAccountAddress()
	s.StoreReflectCode(creator)

	lucky := s.RandomAccountAddress()
	reflect := s.instantiateReflectContract(lucky)
	s.Require().NotEmpty(reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(tftypes.DefaultParams().DenomCreationFee[0].Denom, tftypes.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	s.FundAcc(reflect, reflectAmount)

	// lucky was broke
	balances := s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, lucky)
	s.Require().Empty(balances)

	// Create denom for minting
	msg := types.TokenFactoryMsg{CreateDenom: &types.CreateDenom{
		Subdenom: "SUN",
	}}
	err := s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)
	sunDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount, ok := sdkmath.NewIntFromString("808010809")
	s.Require().True(ok)

	msg = types.TokenFactoryMsg{MintTokens: &types.MintTokens{
		Denom:         sunDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)

	// can burn from different address with burnFrom
	amt, ok := sdkmath.NewIntFromString("1")
	s.Require().True(ok)
	msg = types.TokenFactoryMsg{BurnTokens: &types.BurnTokens{
		Denom:           sunDenom,
		Amount:          amt,
		BurnFromAddress: lucky.String(),
	}}
	err = s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)

	// lucky needs to send balance to reflect contract to burn it
	luckyBalance := s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, lucky)
	err = s.App.AppKeepers.BankKeeper.SendCoins(s.Ctx, lucky, reflect, luckyBalance)
	s.Require().NoError(err)

	msg = types.TokenFactoryMsg{BurnTokens: &types.BurnTokens{
		Denom:           sunDenom,
		Amount:          amount.Abs().Sub(sdkmath.NewInt(1)),
		BurnFromAddress: reflect.String(),
	}}
	err = s.executeCustom(reflect, lucky, msg, sdk.Coin{})
	s.Require().NoError(err)
}
