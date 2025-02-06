package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CosmosContracts/juno/v27/x/tokenfactory/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	s.SetupTestForInitGenesis()
	genesisState := types.GenesisState{
		FactoryDenoms: []types.GenesisDenom{
			{
				Denom: "factory/juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym/bitcoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym",
				},
			},
			{
				Denom: "factory/juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym/diff-admin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "juno15czt5nhlnvayqq37xun9s9yus0d6y26dsvkcna",
				},
			},
			{
				Denom: "factory/juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym/litecoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "juno1t7egva48prqmzl59x5ngv4zx0dtrwewcmjwfym",
				},
			},
		},
	}

	// Test both with bank denom metadata set, and not set.
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, sets bank metadata to exist if i != 0, to cover both cases.
		if i != 0 {
			s.App.AppKeepers.BankKeeper.SetDenomMetaData(s.Ctx, banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    denom.GetDenom(),
					Exponent: 0,
				}},
				Base:    denom.GetDenom(),
				Display: denom.GetDenom(),
				Name:    denom.GetDenom(),
				Symbol:  denom.GetDenom(),
			})
		}
	}

	// check before initGenesis that the module account is nil
	tokenfactoryModuleAccount := s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().Nil(tokenfactoryModuleAccount)

	err := s.App.AppKeepers.TokenFactoryKeeper.SetParams(s.Ctx, types.Params{DenomCreationFee: sdk.Coins{sdk.NewInt64Coin("ujuno", 100)}})
	s.Require().NoError(err)
	s.App.AppKeepers.TokenFactoryKeeper.InitGenesis(s.Ctx, genesisState)

	// check that the module account is now initialized
	tokenfactoryModuleAccount = s.App.AppKeepers.AccountKeeper.GetAccount(s.Ctx, s.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().NotNil(tokenfactoryModuleAccount)

	exportedGenesis := s.App.AppKeepers.TokenFactoryKeeper.ExportGenesis(s.Ctx)
	s.Require().NotNil(exportedGenesis)
	s.Require().Equal(genesisState, *exportedGenesis)

	// verify that the exported bank genesis is valid
	err = s.App.AppKeepers.BankKeeper.SetParams(s.Ctx, banktypes.DefaultParams())
	s.Require().NoError(err)
	exportedBankGenesis := s.App.AppKeepers.BankKeeper.ExportGenesis(s.Ctx)
	s.Require().NoError(exportedBankGenesis.Validate())

	s.App.AppKeepers.BankKeeper.InitGenesis(s.Ctx, exportedBankGenesis)
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, check whether bank metadata is not replaced if i != 0, to cover both cases.
		if i != 0 {
			metadata, found := s.App.AppKeepers.BankKeeper.GetDenomMetaData(s.Ctx, denom.GetDenom())
			s.Require().True(found)
			s.Require().EqualValues(metadata, banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    denom.GetDenom(),
					Exponent: 0,
				}},
				Base:    denom.GetDenom(),
				Display: denom.GetDenom(),
				Name:    denom.GetDenom(),
				Symbol:  denom.GetDenom(),
			})
		}
	}
}
