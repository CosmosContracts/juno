package keeper_test

import (
	_ "embed"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

//go:embed testdata/reflect.wasm
var wasmContract []byte

func (s *IntegrationTestSuite) TestStoreContract() {
	msg := types.MsgStoreCodeFixture(func(m *wasmtypes.MsgStoreCode) {
		m.WASMByteCode = wasmContract
		m.Sender = sender.String()
	})
}

func (s *IntegrationTestSuite) TestGetContractAdminOrCreatorAddress() {

}
