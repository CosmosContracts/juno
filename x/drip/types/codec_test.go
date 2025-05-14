package types

import (
	"testing"

	s "github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodecTestSuite struct {
	s.Suite
}

func TestCodecSuite(t *testing.T) {
	s.Run(t, new(CodecTestSuite))
}

func (suite *CodecTestSuite) TestRegisterInterfaces() {
	registry := codectypes.NewInterfaceRegistry()
	registry.RegisterInterface(sdk.MsgInterfaceProtoName, (*sdk.Msg)(nil))
	RegisterInterfaces(registry)

	impls := registry.ListImplementations(sdk.MsgInterfaceProtoName)
	suite.Require().Equal(2, len(impls))
	suite.Require().ElementsMatch([]string{
		"/juno.drip.v1.MsgDistributeTokens",
		"/juno.drip.v1.MsgUpdateParams",
	}, impls)
}
