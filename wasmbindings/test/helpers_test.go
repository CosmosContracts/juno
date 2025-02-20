package bindings_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/stretchr/testify/suite"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v28/testutil"
	types "github.com/CosmosContracts/juno/v28/wasmbindings/types"
)

type ReflectExec struct {
	ReflectMsg    *ReflectMsgs    `json:"reflect_msg,omitempty"`
	ReflectSubMsg *ReflectSubMsgs `json:"reflect_sub_msg,omitempty"`
}

type ReflectMsgs struct {
	Msgs []wasmvmtypes.CosmosMsg `json:"msgs"`
}

type ReflectSubMsgs struct {
	Msgs []wasmvmtypes.SubMsg `json:"msgs"`
}

type ReflectQuery struct {
	Chain *ChainRequest `json:"chain,omitempty"`
}

type ChainRequest struct {
	Request wasmvmtypes.QueryRequest `json:"request"`
}

type ChainResponse struct {
	Data []byte `json:"data"`
}

type BindingsTestSuite struct {
	testutil.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(BindingsTestSuite))
}

func (s *BindingsTestSuite) SetupTest() {
	s.Setup()
}

// we need to make this deterministic (same every test run), as content might affect gas costs
func (s *BindingsTestSuite) keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func (s *BindingsTestSuite) RandomAccountAddress() sdk.AccAddress {
	_, _, addr := s.keyPubAddr()
	return addr
}

func (s *BindingsTestSuite) RandomBech32AccountAddress() string {
	return s.RandomAccountAddress().String()
}

func (s *BindingsTestSuite) storeReflectCode(addr sdk.AccAddress) uint64 {
	wasmCode, err := os.ReadFile("./testdata/token_reflect.wasm")
	s.Require().NoError(err)

	contractKeeper := s.App.AppKeepers.ContractKeeper
	sdkCtx := sdk.UnwrapSDKContext(s.Ctx)
	codeID, _, err := contractKeeper.Create(sdkCtx, addr, wasmCode, nil)
	s.Require().NoError(err)

	return codeID
}

func (s *BindingsTestSuite) instantiateReflectContract(funder sdk.AccAddress) sdk.AccAddress {
	s.T().Helper()

	initMsgBz := []byte("{}")
	contractKeeper := s.App.AppKeepers.ContractKeeper
	codeID := uint64(1)
	sdkCtx := sdk.UnwrapSDKContext(s.Ctx)
	addr, _, err := contractKeeper.Instantiate(sdkCtx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	s.Require().NoError(err)

	return addr
}

func (s *BindingsTestSuite) StoreReflectCode(addr sdk.AccAddress) {
	s.storeReflectCode(addr)
	cInfo := s.App.AppKeepers.WasmKeeper.GetCodeInfo(s.Ctx, 1)
	s.Require().NotNil(cInfo)
}

func (s *BindingsTestSuite) executeCustom(contract sdk.AccAddress, sender sdk.AccAddress, msg types.TokenFactoryMsg, funds sdk.Coin) error { //nolint:unparam // funds is always nil but could change in the future.
	customBz, err := json.Marshal(msg)
	s.Require().NoError(err)

	reflectMsg := ReflectExec{
		ReflectMsg: &ReflectMsgs{
			Msgs: []wasmvmtypes.CosmosMsg{{
				Custom: customBz,
			}},
		},
	}
	reflectBz, err := json.Marshal(reflectMsg)
	s.Require().NoError(err)

	// no funds sent if amount is 0
	var coins sdk.Coins
	if !funds.Amount.IsNil() {
		coins = sdk.Coins{funds}
	}

	contractKeeper := keeper.NewDefaultPermissionKeeper(s.App.AppKeepers.WasmKeeper)
	_, err = contractKeeper.Execute(s.Ctx, contract, sender, reflectBz, coins)
	return err
}

func (s *BindingsTestSuite) queryCustom(contract sdk.AccAddress, request types.TokenFactoryQuery, response interface{}) {
	msgBz, err := json.Marshal(request)
	s.Require().NoError(err)
	fmt.Println("queryCustom1", string(msgBz))

	query := ReflectQuery{
		Chain: &ChainRequest{
			Request: wasmvmtypes.QueryRequest{Custom: msgBz},
		},
	}
	queryBz, err := json.Marshal(query)
	s.Require().NoError(err)
	fmt.Println("queryCustom2", string(queryBz))

	resBz, err := s.App.AppKeepers.WasmKeeper.QuerySmart(s.Ctx, contract, queryBz)
	s.Require().NoError(err)
	var resp ChainResponse
	err = json.Unmarshal(resBz, &resp)
	s.Require().NoError(err)
	err = json.Unmarshal(resp.Data, response)
	s.Require().NoError(err)
}
