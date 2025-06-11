package tokenfactory_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	e2esuite "github.com/CosmosContracts/juno/tests/interchaintest/suite"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tftypes "github.com/CosmosContracts/juno/v30/x/tokenfactory/types"
)

type TokenfactoryTestSuite struct {
	*e2esuite.E2ETestSuite
}

func TestTokenfactoryTestSuite(t *testing.T) {
	s := e2esuite.NewE2ETestSuite(
		[]*interchaintest.ChainSpec{e2esuite.DefaultSpec},
		e2esuite.DefaultTxCfg,
	)

	t.Parallel()
	t.Cleanup(func() {
		_ = s.Ic.Close()
	})

	testSuite := &TokenfactoryTestSuite{E2ETestSuite: s}
	suite.Run(t, testSuite)
}

// TestTokenFactory ensures the x/tokenfactory module & bindings work properly.
func (s *TokenfactoryTestSuite) TestTokenfactoryModule() {
	t := s.T()

	user := s.GetAndFundTestUser("default", 100_000_000, s.Chain)
	uaddr := user.FormattedAddress()

	user2 := s.GetAndFundTestUser("default", 100_000_000, s.Chain)
	uaddr2 := user2.FormattedAddress()

	fees := sdk.NewCoins(sdk.NewCoin(s.Denom, math.NewInt(50_000)))

	tfDenom := s.CreateTokenFactoryDenom(s.Chain, user, "ictestdenom", fees)
	t.Log("tfDenom", tfDenom)

	// mint
	s.MintTokenFactoryDenom(s.Chain, user, 100, tfDenom, fees)
	t.Log("minted tfDenom to user")
	if balance, err := s.Chain.GetBalance(s.Ctx, uaddr, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 100 {
		t.Fatal("balance not 100")
	}

	// mint-to
	s.MintToTokenFactoryDenom(s.Chain, user, user2, 70, tfDenom, fees)
	t.Log("minted tfDenom to user")
	if balance, err := s.Chain.GetBalance(s.Ctx, uaddr2, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 70 {
		t.Fatal("balance not 70")
	}

	// This allows the uaddr here to mint tokens on behalf of the contract. Typically you only allow a contract here, but this is testing.
	coreInitMsg := fmt.Sprintf(`{"allowed_mint_addresses":["%s"],"existing_denoms":["%s"]}`, uaddr, tfDenom)
	_, coreTFContract := s.SetupContract(s.Chain, user.KeyName(), "../../contracts/tokenfactory_core.wasm", coreInitMsg, false, fees)
	t.Log("coreContract", coreTFContract)

	// change admin to the contract
	s.ChangeTokenFactoryAdmin(s.Chain, user, coreTFContract, tfDenom, fees)

	// ensure the admin is the contract
	admin, err := s.QueryClients.TokenfactoryClient.DenomAuthorityMetadata(s.Ctx,
		&tftypes.QueryDenomAuthorityMetadataRequest{
			Denom: tfDenom,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("admin", admin)
	if admin.AuthorityMetadata.Admin != coreTFContract {
		t.Fatal("admin not coreTFContract. Did not properly transfer.")
	}

	// Mint on the contract for the user to ensure mint bindings work.
	mintMsg := fmt.Sprintf(`{"mint":{"address":"%s","denom":[{"denom":"%s","amount":"31"}]}}`, uaddr2, tfDenom)
	if _, err := s.Chain.ExecuteContract(s.Ctx, user.KeyName(), coreTFContract, mintMsg,
		"--gas", "auto", "--fees", fees.String()); err != nil {
		t.Fatal(err)
	}

	// ensure uaddr2 has 31+70 = 101
	if balance, err := s.Chain.GetBalance(s.Ctx, uaddr2, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 101 {
		t.Fatal("balance not 101")
	}
}
