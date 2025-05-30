package burn_test

import (
	"fmt"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
)

type BurnTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestBurnTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{e2esuite.DefaultSpec},
		e2esuite.DefaultTxCfg,
	)

	t.Parallel()
	t.Cleanup(func() {
		_ = s.Ic.Close()
	})

	testSuite := &BurnTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

// TestBurnModule ensures the x/burn module register and execute sharing functions work properly on smart contracts.
// This is required due to how x/mint handles minting tokens for the target supply.
// It is purely for developers ::BurnTokens to function as expected.
func (s *BurnTestSuite) TestBurnModule() {
	t := s.T()
	nativeDenom := s.Chain.Config().Denom
	fees := sdk.NewCoins(sdk.NewCoin(nativeDenom, math.NewInt(30000)))

	// Users
	user := s.GetAndFundTestUser("default", int64(10_000_000), s.Chain)

	// Upload & init contract

	_, contractAddr := s.SetupContract(s.Chain, user.KeyName(), "../../contracts/cw_testburn.wasm", `{}`, false, fees)

	// get balance before execute
	balance, err := s.Chain.GetBalance(s.Ctx, user.FormattedAddress(), nativeDenom)
	if err != nil {
		t.Fatal(err)
	}

	// execute burn of tokens
	burnAmt := int64(1_000_000)
	_, err = s.ExecuteMsgWithAmount(s.Chain, user, contractAddr, strconv.Itoa(int(burnAmt))+nativeDenom, `{"burn_token":{}}`, fees)
	if err != nil {
		t.Fatal(err)
	}

	// verify it is down 1_000_000 tokens since the burn
	updatedBal, err := s.Chain.GetBalance(s.Ctx, user.FormattedAddress(), nativeDenom)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the funds were sent, and burned.
	fmt.Println(balance, updatedBal)
	assert.Equal(t, burnAmt, balance.Sub(updatedBal).Sub(fees.AmountOf(nativeDenom).Mul(math.NewInt(2))).Int64(), fmt.Sprintf("balance should be %d less than updated balance", burnAmt))
}
