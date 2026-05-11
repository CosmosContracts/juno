package fees_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestFeeShare ensures the x/feeshare module register and execute sharing functions work properly on smart contracts.
func (s *FeesTestSuite) TestFeeShare() {
	t := s.T()
	// Do not call t.Parallel here: this runs as a subtest of TestFeesTestSuite which
	// registers a t.Cleanup to close the interchain. With t.Parallel the cleanup
	// fires when the parent returns (after the synchronous subtest finishes) but
	// before this parallel subtest resumes — so DNS lookups against the validator
	// container fail with "no such host" when the test body actually runs.

	// Users
	granter := s.GetAndFundTestUser("granter", 10_000_000, s.Chain)
	grantee := s.GetAndFundTestUser("grantee", 10_000_000, s.Chain)
	feeRcvAddr := "juno1v75wlkccpv7le3560zw32v2zjes5n0e7csr4qh"

	// Upload & init contract payment to another address.
	// setupFees covers wasm-store at v30 minBaseGasPrice 0.075 (≥225k); execFees covers a wasm-execute (~33k req).
	setupFees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(1_000_000)))
	execFees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(50_000)))
	_, contractAddr := s.SetupContract(s.Chain, granter.KeyName(), "../../contracts/cw_template.wasm", `{"count":0}`, false, setupFees)

	// register contract to a random address (since we are the creator, though not the admin)
	s.RegisterFeeShare(s.Chain, granter, contractAddr, feeRcvAddr)
	if balance, err := s.Chain.GetBalance(s.Ctx, feeRcvAddr, s.Denom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 0 {
		t.Fatal("balance not 0")
	}

	// 50% default DeveloperShares of the full --fees amount goes to the registered address.
	expectedShare := execFees.AmountOf(s.Denom).QuoRaw(2).Int64()
	_, err := s.ExecuteMsgWithFeeReturn(s.Chain, granter, contractAddr, "", `{"increment":{}}`, false, execFees)
	if err != nil {
		t.Fatal(err)
	}

	if balance, err := s.Chain.GetBalance(s.Ctx, feeRcvAddr, s.Denom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != expectedShare {
		t.Fatalf("balance not %d. it is %s%s", expectedShare, balance, s.Denom)
	}

	// Test authz message execution:
	// Grant contract execute permission to grantee
	s.ExecuteAuthzGrantMsg(s.Chain, granter, grantee, "/cosmos.authz.v1beta1.MsgExec")

	// Execute authz msg as grantee using the same fee amount so we expect another expectedShare added.
	s.ExecuteAuthzExecMsgWithFee(s.Chain, grantee, contractAddr, "", execFees.String(), `{"increment":{}}`)

	if balance, err := s.Chain.GetBalance(s.Ctx, feeRcvAddr, s.Denom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 2*expectedShare {
		t.Fatalf("balance not %d. it is %s%s", 2*expectedShare, balance, s.Denom)
	}
}
