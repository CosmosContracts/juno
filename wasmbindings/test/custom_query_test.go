package bindings_test

import (
	"fmt"

	types "github.com/CosmosContracts/juno/v30/wasmbindings/types"
)

func (s *BindingsTestSuite) TestQueryFullDenom() {
	s.SetupTest()
	actor := s.RandomAccountAddress()
	s.StoreReflectCode(actor)

	reflect := s.instantiateReflectContract(actor)
	s.Require().NotEmpty(reflect)

	// query full denom
	query := types.TokenFactoryQuery{
		FullDenom: &types.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "ustart",
		},
	}
	resp := types.FullDenomResponse{}
	s.queryCustom(reflect, query, &resp)

	expected := fmt.Sprintf("factory/%s/ustart", reflect.String())
	s.Require().EqualValues(expected, resp.Denom)
}
