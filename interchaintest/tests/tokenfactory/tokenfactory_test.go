package interchaintest

import (
	"fmt"
	"testing"

	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
	"github.com/stretchr/testify/suite"

	helpers "github.com/CosmosContracts/juno/tests/interchaintest/helpers"
)

type TokenfactoryTestSuite struct {
	e2esuite.E2ETestSuite
}

func TestTokenfactoryTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		e2esuite.Spec,
		e2esuite.TxCfg,
	)

	suite.Run(t, s)
}

// TestTokenFactory ensures the tokenfactory module & bindings work properly.
func (s *TokenfactoryTestSuite) TestTokenFactory() {
	t := s.T()
	juno := s.Chain
	t.Parallel()

	user := s.GetAndFundTestUser(s.Ctx, "default", 10_000_000, juno)
	uaddr := user.FormattedAddress()

	user2 := s.GetAndFundTestUser(s.Ctx, "default", 10_000_000, juno)
	uaddr2 := user2.FormattedAddress()

	tfDenom := helpers.CreateTokenFactoryDenom(t, s.Ctx, juno, user, "ictestdenom", fmt.Sprintf("0%s", s.Denom))
	t.Log("tfDenom", tfDenom)

	// mint
	helpers.MintTokenFactoryDenom(t, s.Ctx, juno, user, 100, tfDenom)
	t.Log("minted tfDenom to user")
	if balance, err := juno.GetBalance(s.Ctx, uaddr, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 100 {
		t.Fatal("balance not 100")
	}

	// mint-to
	helpers.MintToTokenFactoryDenom(t, s.Ctx, juno, user, user2, 70, tfDenom)
	t.Log("minted tfDenom to user")
	if balance, err := juno.GetBalance(s.Ctx, uaddr2, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 70 {
		t.Fatal("balance not 70")
	}

	// This allows the uaddr here to mint tokens on behalf of the contract. Typically you only allow a contract here, but this is testing.
	coreInitMsg := fmt.Sprintf(`{"allowed_mint_addresses":["%s"],"existing_denoms":["%s"]}`, uaddr, tfDenom)
	_, coreTFContract := helpers.SetupContract(t, s.Ctx, juno, user.KeyName(), "contracts/tokenfactory_core.wasm", coreInitMsg)
	t.Log("coreContract", coreTFContract)

	// change admin to the contract
	helpers.TransferTokenFactoryAdmin(t, s.Ctx, juno, user, coreTFContract, tfDenom)

	// ensure the admin is the contract
	admin := helpers.GetTokenFactoryAdmin(t, s.Ctx, juno, tfDenom)
	t.Log("admin", admin)
	if admin != coreTFContract {
		t.Fatal("admin not coreTFContract. Did not properly transfer.")
	}

	// Mint on the contract for the user to ensure mint bindings work.
	mintMsg := fmt.Sprintf(`{"mint":{"address":"%s","denom":[{"denom":"%s","amount":"31"}]}}`, uaddr2, tfDenom)
	if _, err := juno.ExecuteContract(s.Ctx, user.KeyName(), coreTFContract, mintMsg); err != nil {
		t.Fatal(err)
	}

	// ensure uaddr2 has 31+70 = 101
	if balance, err := juno.GetBalance(s.Ctx, uaddr2, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 101 {
		t.Fatal("balance not 101")
	}

	t.Cleanup(func() {
		s.TearDownSuite()
	})
}
