package fees_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestFeeShare ensures the x/feeshare module register and execute sharing functions work properly on smart contracts.
func (s *FeesTestSuite) TestFeeShare() {
	t := s.T()
	t.Parallel()

	// Users
	granter := s.GetAndFundTestUser("granter", 10_000_000, s.Chain)
	grantee := s.GetAndFundTestUser("grantee", 10_000_000, s.Chain)
	feeRcvAddr := "juno1v75wlkccpv7le3560zw32v2zjes5n0e7csr4qh"

	// Upload & init contract payment to another address
	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(100000)))
	_, contractAddr := s.SetupContract(s.Chain, granter.KeyName(), "../../contracts/cw_template.wasm", `{"count":0}`, false, fees)

	// register contract to a random address (since we are the creator, though not the admin)
	s.RegisterFeeShare(s.Chain, granter, contractAddr, feeRcvAddr)
	if balance, err := s.Chain.GetBalance(s.Ctx, feeRcvAddr, s.Denom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 0 {
		t.Fatal("balance not 0")
	}

	// execute with a 10000 fee (so 5000 denom should be in the contract now with 50% feeshare default)
	_, err := s.ExecuteMsgWithFeeReturn(s.Chain, granter, contractAddr, "", `{"increment":{}}`, false, fees)
	if err != nil {
		t.Fatal(err)
	}

	// check balance of s.Denom now
	if balance, err := s.Chain.GetBalance(s.Ctx, feeRcvAddr, s.Denom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 5000 {
		t.Fatal("balance not 5,000. it is ", balance, s.Denom)
	}

	// Test authz message execution:
	// Grant contract execute permission to grantee
	s.ExecuteAuthzGrantMsg(s.Chain, granter, grantee, "/cosmos.authz.v1beta1.MsgExec")

	// Execute authz msg as grantee
	s.ExecuteAuthzExecMsgWithFee(s.Chain, grantee, contractAddr, "", "10000"+s.Denom, `{"increment":{}}`)

	// check balance of s.Denom now
	if balance, err := s.Chain.GetBalance(s.Ctx, feeRcvAddr, s.Denom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 10000 {
		t.Fatal("balance not 10,000. it is ", balance, s.Denom)
	}
}
