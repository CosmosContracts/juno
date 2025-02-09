package bindings_test

import (
	"fmt"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bindings "github.com/CosmosContracts/juno/v28/wasmbindings"
)

func (s *BindingsTestSuite) TestFullDenom() {
	actor := s.RandomAccountAddress()

	specs := map[string]struct {
		addr         string
		subdenom     string
		expFullDenom string
		expErr       bool
	}{
		"valid address": {
			addr:         actor.String(),
			subdenom:     "subDenom1",
			expFullDenom: fmt.Sprintf("factory/%s/subDenom1", actor.String()),
		},
		"empty address": {
			addr:     "",
			subdenom: "subDenom1",
			expErr:   true,
		},
		"invalid address": {
			addr:     "invalid",
			subdenom: "subDenom1",
			expErr:   true,
		},
		"empty sub-denom": {
			addr:         actor.String(),
			subdenom:     "",
			expFullDenom: fmt.Sprintf("factory/%s/", actor.String()),
		},
		"invalid sub-denom (contains exclamation point)": {
			addr:     actor.String(),
			subdenom: "subdenom!",
			expErr:   true,
		},
	}
	for name, spec := range specs {
		s.Run(name, func() {
			// when
			gotFullDenom, gotErr := bindings.GetFullDenom(spec.addr, spec.subdenom)
			// then
			if spec.expErr {
				s.Require().Error(gotErr)
				return
			}
			s.Require().NoError(gotErr)
			assert.Equal(s.T(), spec.expFullDenom, gotFullDenom, "exp %s but got %s", spec.expFullDenom, gotFullDenom)
		})
	}
}

func (s *BindingsTestSuite) TestDenomAdmin() {
	addr := s.RandomAccountAddress()
	s.StoreReflectCode(addr)

	// set token creation fee to zero to make testing easier
	tfParams := s.App.AppKeepers.TokenFactoryKeeper.GetParams(s.Ctx)
	tfParams.DenomCreationFee = sdk.NewCoins()
	err := s.App.AppKeepers.TokenFactoryKeeper.SetParams(s.Ctx, tfParams)
	s.Require().NoError(err)

	// create a subdenom via the token factory
	admin := sdk.AccAddress([]byte("addr1_______________"))
	tfDenom, err := s.App.AppKeepers.TokenFactoryKeeper.CreateDenom(s.Ctx, admin.String(), "subdenom")
	s.Require().NoError(err)
	s.Require().NotEmpty(tfDenom)

	queryPlugin := bindings.NewQueryPlugin(s.App.AppKeepers.BankKeeper, &s.App.AppKeepers.TokenFactoryKeeper)

	testCases := []struct {
		name        string
		denom       string
		expectErr   bool
		expectAdmin string
	}{
		{
			name:        "valid token factory denom",
			denom:       tfDenom,
			expectAdmin: admin.String(),
		},
		{
			name:        "invalid token factory denom",
			denom:       sdk.DefaultBondDenom,
			expectErr:   false,
			expectAdmin: "",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			resp, err := queryPlugin.GetDenomAdmin(s.Ctx, tc.denom)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Require().Equal(tc.expectAdmin, resp.Admin)
			}
		})
	}
}
