package bindings_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bindings "github.com/CosmosContracts/juno/v28/wasmbindings"
	types "github.com/CosmosContracts/juno/v28/wasmbindings/types"
	tftypes "github.com/CosmosContracts/juno/v28/x/tokenfactory/types"
)

func (s *BindingsTestSuite) TestCreateDenom() {
	actor := s.RandomAccountAddress()
	s.StoreReflectCode(actor)

	// Fund actor with 100 base denom creation fees
	actorAmount := sdk.NewCoins(sdk.NewCoin(tftypes.DefaultParams().DenomCreationFee[0].Denom, tftypes.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	s.FundAcc(actor, actorAmount)

	specs := map[string]struct {
		createDenom *types.CreateDenom
		expErr      bool
	}{
		"valid sub-denom": {
			createDenom: &types.CreateDenom{
				Subdenom: "MOON",
			},
		},
		"empty sub-denom": {
			createDenom: &types.CreateDenom{
				Subdenom: "",
			},
			expErr: false,
		},
		"invalid sub-denom": {
			createDenom: &types.CreateDenom{
				Subdenom: "sub-denom_2",
			},
			expErr: false,
		},
		"null create denom": {
			createDenom: nil,
			expErr:      true,
		},
	}
	for name, spec := range specs {
		s.Run(name, func() {
			// when
			_, gotErr := bindings.PerformCreateDenom(
				s.Ctx,
				&s.App.AppKeepers.TokenFactoryKeeper,
				s.App.AppKeepers.BankKeeper,
				actor,
				spec.createDenom,
			)
			// then
			if spec.expErr {
				s.T().Logf("validate_msg_test got error: %v", gotErr)
				s.Require().Error(gotErr)
				return
			}
			s.Require().NoError(gotErr)
		})
	}
}

func (s *BindingsTestSuite) TestChangeAdmin() {
	const validDenom = "validdenom"

	tokenCreator := s.RandomAccountAddress()

	specs := map[string]struct {
		actor       sdk.AccAddress
		changeAdmin *types.ChangeAdmin

		expErrMsg string
	}{
		"valid": {
			changeAdmin: &types.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: s.RandomBech32AccountAddress(),
			},
			actor: tokenCreator,
		},
		"typo in factory in denom name": {
			changeAdmin: &types.ChangeAdmin{
				Denom:           fmt.Sprintf("facory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: s.RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "denom prefix is incorrect. Is: facory.  Should be: factory: invalid denom",
		},
		"invalid address in denom": {
			changeAdmin: &types.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", s.RandomBech32AccountAddress(), validDenom),
				NewAdminAddress: s.RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "failed changing admin from message: unauthorized account",
		},
		"other denom name in 3 part name": {
			changeAdmin: &types.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), "invalid denom"),
				NewAdminAddress: s.RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: fmt.Sprintf("invalid denom: factory/%s/invalid denom", tokenCreator.String()),
		},
		"empty denom": {
			changeAdmin: &types.ChangeAdmin{
				Denom:           "",
				NewAdminAddress: s.RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "invalid denom: ",
		},
		"empty address": {
			changeAdmin: &types.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: "",
			},
			actor:     tokenCreator,
			expErrMsg: "address from bech32: empty address string is not allowed",
		},
		"creator is a different address": {
			changeAdmin: &types.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: s.RandomBech32AccountAddress(),
			},
			actor:     s.RandomAccountAddress(),
			expErrMsg: "failed changing admin from message: unauthorized account",
		},
		"change to the same address": {
			changeAdmin: &types.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: tokenCreator.String(),
			},
			actor: tokenCreator,
		},
		"nil binding": {
			actor:     tokenCreator,
			expErrMsg: "invalid request: changeAdmin is nil - original request: ",
		},
	}
	for name, spec := range specs {
		s.Run(name, func() {
			s.Reset()

			// Fund actor with 100 base denom creation fees
			actorAmount := sdk.NewCoins(sdk.NewCoin(tftypes.DefaultParams().DenomCreationFee[0].Denom, tftypes.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
			s.FundAcc(tokenCreator, actorAmount)

			_, err := bindings.PerformCreateDenom(s.Ctx, &s.App.AppKeepers.TokenFactoryKeeper, s.App.AppKeepers.BankKeeper, tokenCreator, &types.CreateDenom{
				Subdenom: validDenom,
			})
			s.Require().NoError(err)

			err = bindings.ChangeAdmin(s.Ctx, &s.App.AppKeepers.TokenFactoryKeeper, spec.actor, spec.changeAdmin)
			if len(spec.expErrMsg) > 0 {
				s.Require().Error(err)
				actualErrMsg := err.Error()
				s.Require().Equal(spec.expErrMsg, actualErrMsg)
				return
			}
			s.Require().NoError(err)
		})
	}
}

func (s *BindingsTestSuite) TestMint() {
	creator := s.RandomAccountAddress()
	s.StoreReflectCode(creator)

	// Fund actor with 100 base denom creation fees
	tokenCreationFeeAmt := sdk.NewCoins(sdk.NewCoin(tftypes.DefaultParams().DenomCreationFee[0].Denom, tftypes.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	s.FundAcc(creator, tokenCreationFeeAmt)

	// Create denoms for valid mint tests
	validDenom := types.CreateDenom{
		Subdenom: "MOON",
	}
	_, err := bindings.PerformCreateDenom(s.Ctx, &s.App.AppKeepers.TokenFactoryKeeper, s.App.AppKeepers.BankKeeper, creator, &validDenom)
	s.Require().NoError(err)

	emptyDenom := types.CreateDenom{
		Subdenom: "",
	}
	_, err = bindings.PerformCreateDenom(s.Ctx, &s.App.AppKeepers.TokenFactoryKeeper, s.App.AppKeepers.BankKeeper, creator, &emptyDenom)
	s.Require().NoError(err)

	validDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), validDenom.Subdenom)
	emptyDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), emptyDenom.Subdenom)

	lucky := s.RandomAccountAddress()

	// lucky was broke
	balances := s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, lucky)
	s.Require().Empty(balances)

	amount, ok := sdkmath.NewIntFromString("8080")
	s.Require().True(ok)

	specs := map[string]struct {
		mint   *types.MintTokens
		expErr bool
	}{
		"valid mint": {
			mint: &types.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
		},
		"empty sub-denom": {
			mint: &types.MintTokens{
				Denom:         emptyDenomStr,
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
			expErr: false,
		},
		"nonexistent sub-denom": {
			mint: &types.MintTokens{
				Denom:         fmt.Sprintf("factory/%s/%s", creator.String(), "SUN"),
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"invalid sub-denom": {
			mint: &types.MintTokens{
				Denom:         "sub-denom_2",
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"zero amount": {
			mint: &types.MintTokens{
				Denom:         validDenomStr,
				Amount:        sdkmath.ZeroInt(),
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"negative amount": {
			mint: &types.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount.Neg(),
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"empty recipient": {
			mint: &types.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: "",
			},
			expErr: true,
		},
		"invalid recipient": {
			mint: &types.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: "invalid",
			},
			expErr: true,
		},
		"null mint": {
			mint:   nil,
			expErr: true,
		},
	}
	for name, spec := range specs {
		s.Run(name, func() {
			// when
			gotErr := bindings.PerformMint(s.Ctx, &s.App.AppKeepers.TokenFactoryKeeper, s.App.AppKeepers.BankKeeper, creator, spec.mint)
			// then
			if spec.expErr {
				s.Require().Error(gotErr)
				return
			}
			s.Require().NoError(gotErr)
		})
	}
}

func (s *BindingsTestSuite) TestBurn() {
	creator := s.RandomAccountAddress()
	s.StoreReflectCode(creator)

	// Fund actor with 100 base denom creation fees
	tokenCreationFeeAmt := sdk.NewCoins(sdk.NewCoin(tftypes.DefaultParams().DenomCreationFee[0].Denom, tftypes.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	s.FundAcc(creator, tokenCreationFeeAmt)

	// Create denoms for valid burn tests
	validDenom := types.CreateDenom{
		Subdenom: "MOON",
	}
	_, err := bindings.PerformCreateDenom(s.Ctx, &s.App.AppKeepers.TokenFactoryKeeper, s.App.AppKeepers.BankKeeper, creator, &validDenom)
	s.Require().NoError(err)

	emptyDenom := types.CreateDenom{
		Subdenom: "",
	}
	_, err = bindings.PerformCreateDenom(s.Ctx, &s.App.AppKeepers.TokenFactoryKeeper, s.App.AppKeepers.BankKeeper, creator, &emptyDenom)
	s.Require().NoError(err)

	lucky := s.RandomAccountAddress()

	// lucky was broke
	balances := s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, lucky)
	s.Require().Empty(balances)

	validDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), validDenom.Subdenom)
	emptyDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), emptyDenom.Subdenom)
	mintAmount, ok := sdkmath.NewIntFromString("8080")
	s.Require().True(ok)

	specs := map[string]struct {
		burn   *types.BurnTokens
		expErr bool
	}{
		"valid burn": {
			burn: &types.BurnTokens{
				Denom:           validDenomStr,
				Amount:          mintAmount,
				BurnFromAddress: creator.String(),
			},
			expErr: false,
		},
		"non admin address": {
			burn: &types.BurnTokens{
				Denom:           validDenomStr,
				Amount:          mintAmount,
				BurnFromAddress: lucky.String(),
			},
			expErr: true,
		},
		"empty sub-denom": {
			burn: &types.BurnTokens{
				Denom:           emptyDenomStr,
				Amount:          mintAmount,
				BurnFromAddress: creator.String(),
			},
			expErr: false,
		},
		"invalid sub-denom": {
			burn: &types.BurnTokens{
				Denom:           "sub-denom_2",
				Amount:          mintAmount,
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
		"non-minted denom": {
			burn: &types.BurnTokens{
				Denom:           fmt.Sprintf("factory/%s/%s", creator.String(), "SUN"),
				Amount:          mintAmount,
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
		"zero amount": {
			burn: &types.BurnTokens{
				Denom:           validDenomStr,
				Amount:          sdkmath.ZeroInt(),
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
		"negative amount": {
			burn:   nil,
			expErr: true,
		},
		"null burn": {
			burn: &types.BurnTokens{
				Denom:           validDenomStr,
				Amount:          mintAmount.Neg(),
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
	}

	for name, spec := range specs {
		s.Run(name, func() {
			// Mint valid denom str and empty denom string for burn test
			mintBinding := &types.MintTokens{
				Denom:         validDenomStr,
				Amount:        mintAmount,
				MintToAddress: creator.String(),
			}
			err := bindings.PerformMint(s.Ctx, &s.App.AppKeepers.TokenFactoryKeeper, s.App.AppKeepers.BankKeeper, creator, mintBinding)
			s.Require().NoError(err)

			emptyDenomMintBinding := &types.MintTokens{
				Denom:         emptyDenomStr,
				Amount:        mintAmount,
				MintToAddress: creator.String(),
			}
			err = bindings.PerformMint(s.Ctx, &s.App.AppKeepers.TokenFactoryKeeper, s.App.AppKeepers.BankKeeper, creator, emptyDenomMintBinding)
			s.Require().NoError(err)

			// when
			gotErr := bindings.PerformBurn(s.Ctx, &s.App.AppKeepers.TokenFactoryKeeper, creator, spec.burn)
			// then
			if spec.expErr {
				s.Require().Error(gotErr)
				return
			}
			s.Require().NoError(gotErr)
		})
	}
}
