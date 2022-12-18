package keeper_test

import (
	"crypto/sha256"
	_ "embed"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//go:embed testdata/reflect.wasm
var wasmContract []byte

func (s *IntegrationTestSuite) TestStoreCode() {
	_, _, sender := testdata.KeyTestPubAddr()
	msg := wasmtypes.MsgStoreCodeFixture(func(m *wasmtypes.MsgStoreCode) {
		m.WASMByteCode = wasmContract
		m.Sender = sender.String()
	})
	rsp, err := s.app.MsgServiceRouter().Handler(msg)(s.ctx, msg)
	s.Require().NoError(err)
	var result wasmtypes.MsgStoreCodeResponse
	s.Require().NoError(s.app.AppCodec().Unmarshal(rsp.Data, &result))
	s.Require().Equal(uint64(1), result.CodeID)
	expHash := sha256.Sum256(wasmContract)
	s.Require().Equal(expHash[:], result.Checksum)
	// and
	info := s.app.WasmKeeper.GetCodeInfo(s.ctx, 1)
	s.Require().NotNil(info)
	s.Require().Equal(expHash[:], info.CodeHash)
	s.Require().Equal(sender.String(), info.Creator)
	s.Require().Equal(wasmtypes.DefaultParams().InstantiateDefaultPermission.With(sender), info.InstantiateConfig)
}

func (s *IntegrationTestSuite) TestGetContractAdminOrCreatorAddress() {
	_, _, sender := testdata.KeyTestPubAddr()
	_, _, admin := testdata.KeyTestPubAddr()
	s.FundAccount(s.ctx, sender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))
	s.FundAccount(s.ctx, admin, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	msgStoreCode := wasmtypes.MsgStoreCodeFixture(func(m *wasmtypes.MsgStoreCode) {
		m.WASMByteCode = wasmContract
		m.Sender = sender.String()
	})
	_, err := s.app.MsgServiceRouter().Handler(msgStoreCode)(s.ctx, msgStoreCode)
	s.Require().NoError(err)

	msgInstantiate := wasmtypes.MsgInstantiateContractFixture(func(m *wasmtypes.MsgInstantiateContract) {
		m.Sender = sender.String()
		m.Admin = ""
		m.Msg = []byte(`{}`)
	})
	resp, err := s.app.MsgServiceRouter().Handler(msgInstantiate)(s.ctx, msgInstantiate)
	s.Require().NoError(err)
	var resultNoAdmin wasmtypes.MsgInstantiateContractResponse
	s.Require().NoError(s.app.AppCodec().Unmarshal(resp.Data, &resultNoAdmin))
	contractInfo := s.app.WasmKeeper.GetContractInfo(s.ctx, sdk.MustAccAddressFromBech32(resultNoAdmin.Address))
	s.Require().Equal(contractInfo.CodeID, uint64(1))
	s.Require().Equal(contractInfo.Admin, "")
	s.Require().Equal(contractInfo.Creator, sender.String())

	msgInstantiateWithAdmin := wasmtypes.MsgInstantiateContractFixture(func(m *wasmtypes.MsgInstantiateContract) {
		m.Sender = sender.String()
		m.Admin = admin.String()
		m.Msg = []byte(`{}`)
	})
	resp, err = s.app.MsgServiceRouter().Handler(msgInstantiateWithAdmin)(s.ctx, msgInstantiateWithAdmin)
	s.Require().NoError(err)
	var resultWithAdmin wasmtypes.MsgInstantiateContractResponse
	s.Require().NoError(s.app.AppCodec().Unmarshal(resp.Data, &resultWithAdmin))
	contractInfo = s.app.WasmKeeper.GetContractInfo(s.ctx, sdk.MustAccAddressFromBech32(resultWithAdmin.Address))
	s.Require().Equal(contractInfo.CodeID, uint64(1))
	s.Require().Equal(contractInfo.Admin, admin.String())
	s.Require().Equal(contractInfo.Creator, sender.String())

	noAdminContractAddress := resultNoAdmin.Address
	withAdminContractAddress := resultWithAdmin.Address

	for _, tc := range []struct {
		desc            string
		contractAddress string
		deployerAddress string
		shouldErr       bool
	}{
		{
			desc:            "Success - Creator",
			contractAddress: noAdminContractAddress,
			deployerAddress: sender.String(),
			shouldErr:       false,
		},
		{
			desc:            "Success - Admin",
			contractAddress: withAdminContractAddress,
			deployerAddress: admin.String(),
			shouldErr:       false,
		},
		{
			desc:            "Error - Invalid deployer",
			contractAddress: noAdminContractAddress,
			deployerAddress: "Invalid",
			shouldErr:       true,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			if !tc.shouldErr {
				_, err := s.app.FeeShareKeeper.GetContractAdminOrCreatorAddress(s.ctx, sdk.MustAccAddressFromBech32(tc.contractAddress), tc.deployerAddress)
				s.Require().NoError(err)
			} else {
				_, err := s.app.FeeShareKeeper.GetContractAdminOrCreatorAddress(s.ctx, sdk.MustAccAddressFromBech32(tc.contractAddress), tc.deployerAddress)
				s.Require().Error(err)
			}
		})
	}
}
