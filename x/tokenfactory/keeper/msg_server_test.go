package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CosmosContracts/juno/v27/x/tokenfactory/types"
)

// TestMintDenomMsg tests TypeMsgMint message is emitted on a successful mint
func (s *KeeperTestSuite) TestMintDenomMsg() {
	// Create a denom
	s.CreateDefaultDenom()

	for _, tc := range []struct {
		desc                  string
		amount                int64
		mintDenom             string
		admin                 string
		valid                 bool
		expectedMessageEvents int
	}{
		{
			desc:      "denom does not exist",
			amount:    10,
			mintDenom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos",
			admin:     s.TestAccs[0].String(),
			valid:     false,
		},
		{
			desc:                  "success case",
			amount:                10,
			mintDenom:             s.defaultDenom,
			admin:                 s.TestAccs[0].String(),
			valid:                 true,
			expectedMessageEvents: 1,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))
			// Test mint message
			mintMsg := &types.MsgMint{
				Sender:        tc.admin,
				Amount:        sdk.NewInt64Coin(tc.mintDenom, 10),
				MintToAddress: tc.admin,
			}
			_, err := s.msgServer.Mint(ctx, mintMsg)
			if tc.valid {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// TestBurnDenomMsg tests TypeMsgBurn message is emitted on a successful burn
func (s *KeeperTestSuite) TestBurnDenomMsg() {
	// Create a denom.
	s.CreateDefaultDenom()
	// mint 10 default token for testAcc[0]
	mintMsg := &types.MsgMint{
		Sender:        s.TestAccs[0].String(),
		Amount:        sdk.NewInt64Coin(s.defaultDenom, 10),
		MintToAddress: s.TestAccs[0].String(),
	}
	s.msgServer.Mint(s.Ctx, mintMsg)

	for _, tc := range []struct {
		desc                  string
		amount                int64
		burnDenom             string
		admin                 string
		valid                 bool
		expectedMessageEvents int
	}{
		{
			desc:      "denom does not exist",
			burnDenom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos",
			admin:     s.TestAccs[0].String(),
			valid:     false,
		},
		{
			desc:                  "success case",
			burnDenom:             s.defaultDenom,
			admin:                 s.TestAccs[0].String(),
			valid:                 true,
			expectedMessageEvents: 1,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))
			// Test burn message
			burnMsg := &types.MsgBurn{
				Sender:          tc.admin,
				Amount:          sdk.NewInt64Coin(tc.burnDenom, 10),
				BurnFromAddress: tc.admin,
			}
			_, err := s.msgServer.Burn(ctx, burnMsg)
			if tc.valid {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// TestCreateDenomMsg tests TypeMsgCreateDenom message is emitted on a successful denom creation
func (s *KeeperTestSuite) TestCreateDenomMsg() {
	for _, tc := range []struct {
		desc     string
		subdenom string
		valid    bool
	}{
		{
			desc:     "subdenom too long",
			subdenom: "assadsadsadasdasdsadsadsadsadsadsadsklkadaskkkdasdasedskhanhassyeunganassfnlksdflksafjlkasd",
			valid:    false,
		},
		{
			desc:     "success case: defaultDenomCreationFee",
			subdenom: "evmos",
			valid:    true,
		},
	} {
		s.SetupTest()
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))
			// Test create denom message
			createDenomMsg := &types.MsgCreateDenom{
				Sender:   s.TestAccs[0].String(),
				Subdenom: tc.subdenom,
			}
			_, err := s.msgServer.CreateDenom(ctx, createDenomMsg)
			if tc.valid {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// TestChangeAdminDenomMsg tests TypeMsgChangeAdmin message is emitted on a successful admin change
func (s *KeeperTestSuite) TestChangeAdminDenomMsg() {
	for _, tc := range []struct {
		desc                    string
		msgChangeAdmin          func(denom string) *types.MsgChangeAdmin
		expectedChangeAdminPass bool
		expectedAdminIndex      int
		msgMint                 func(denom string) *types.MsgMint
	}{
		{
			desc: "non-admins can't change the existing admin",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return &types.MsgChangeAdmin{
					Sender:   s.TestAccs[1].String(),
					Denom:    denom,
					NewAdmin: s.TestAccs[2].String(),
				}
			},
			expectedChangeAdminPass: false,
			expectedAdminIndex:      0,
		},
		{
			desc: "success change admin",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return &types.MsgChangeAdmin{
					Sender:   s.TestAccs[0].String(),
					Denom:    denom,
					NewAdmin: s.TestAccs[1].String(),
				}
			},
			expectedAdminIndex:      1,
			expectedChangeAdminPass: true,
			msgMint: func(denom string) *types.MsgMint {
				return &types.MsgMint{
					Sender:        s.TestAccs[1].String(),
					Amount:        sdk.NewInt64Coin(denom, 5),
					MintToAddress: s.TestAccs[1].String(),
				}
			},
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// setup test
			s.SetupTest()
			ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))
			// Create a denom and mint
			createDenomMsg := &types.MsgCreateDenom{
				Sender:   s.TestAccs[0].String(),
				Subdenom: "bitcoin",
			}
			res, err := s.msgServer.CreateDenom(ctx, createDenomMsg)
			s.Require().NoError(err)
			testDenom := res.GetNewTokenDenom()
			mintMsg := &types.MsgMint{
				Sender:        s.TestAccs[0].String(),
				Amount:        sdk.NewInt64Coin(testDenom, 10),
				MintToAddress: s.TestAccs[0].String(),
			}
			_, err = s.msgServer.Mint(ctx, mintMsg)
			s.Require().NoError(err)
			// Test change admin message
			_, err = s.msgServer.ChangeAdmin(ctx, tc.msgChangeAdmin(testDenom))
			if tc.expectedChangeAdminPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// TestSetDenomMetaDataMsg tests TypeMsgSetDenomMetadata message is emitted on a successful denom metadata change
func (s *KeeperTestSuite) TestSetDenomMetaDataMsg() {
	// setup test
	s.SetupTest()
	s.CreateDefaultDenom()

	for _, tc := range []struct {
		desc                  string
		msgSetDenomMetadata   types.MsgSetDenomMetadata
		expectedPass          bool
		expectedMessageEvents int
	}{
		{
			desc: "successful set denom metadata",
			msgSetDenomMetadata: types.MsgSetDenomMetadata{
				Sender: s.TestAccs[0].String(),
				Metadata: banktypes.Metadata{
					Description: "yeehaw",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    s.defaultDenom,
							Exponent: 0,
						},
						{
							Denom:    "uosmo",
							Exponent: 6,
						},
					},
					Base:    s.defaultDenom,
					Display: "uosmo",
					Name:    "OSMO",
					Symbol:  "OSMO",
				},
			},
			expectedPass:          true,
			expectedMessageEvents: 1,
		},
		{
			desc: "non existent factory denom name",
			msgSetDenomMetadata: types.MsgSetDenomMetadata{
				Sender: s.TestAccs[0].String(),
				Metadata: banktypes.Metadata{
					Description: "yeehaw",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    fmt.Sprintf("factory/%s/litecoin", s.TestAccs[0].String()),
							Exponent: 0,
						},
						{
							Denom:    "uosmo",
							Exponent: 6,
						},
					},
					Base:    fmt.Sprintf("factory/%s/litecoin", s.TestAccs[0].String()),
					Display: "uosmo",
					Name:    "OSMO",
					Symbol:  "OSMO",
				},
			},
			expectedPass: false,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc := tc
			ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))
			// Test set denom metadata message
			_, err := s.msgServer.SetDenomMetadata(ctx, &tc.msgSetDenomMetadata)
			if tc.expectedPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
