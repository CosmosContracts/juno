package keeper_test

import (
	_ "embed"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v22/x/drip/types"
)

func (s *IntegrationTestSuite) TestDripDistributeTokensMsgs() {
	_, _, allowedSender := testdata.KeyTestPubAddr()
	_, _, notAllowedSender := testdata.KeyTestPubAddr()
	_ = s.FundAccount(s.ctx, allowedSender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))
	_ = s.FundAccount(s.ctx, notAllowedSender, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1_000_000))))

	_ = s.app.AppKeepers.DripKeeper.SetParams(s.ctx, types.Params{
		EnableDrip: true,
		AllowedAddresses: []string{
			allowedSender.String(),
		},
	})

	for _, tc := range []struct {
		desc       string
		senderAddr string
		coins      sdk.Coins
		success    bool
	}{
		{
			desc:       "Success - Allowed sender with proper funds",
			senderAddr: allowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1))),
			success:    true,
		},
		{
			desc:       "Fail - Allowed sender no proper funds",
			senderAddr: allowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin("notarealtoken", sdk.NewInt(1))),
			success:    false,
		},
		{
			desc:       "Fail - Allowed sender no tokens",
			senderAddr: allowedSender.String(),
			coins:      nil,
			success:    false,
		},
		{
			desc:       "Fail - Allowed sender empty tokens",
			senderAddr: allowedSender.String(),
			coins:      sdk.NewCoins(),
			success:    false,
		},
		{
			desc:       "Fail - No sender withproper funds",
			senderAddr: "",
			coins:      sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1))),
			success:    false,
		},
		{
			desc:       "Fail - Non Allowed sender proper funds",
			senderAddr: notAllowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1))),
			success:    false,
		},
		{
			desc:       "Fail - Non Allowed sender improper funds",
			senderAddr: notAllowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin("notarealtoken", sdk.NewInt(1))),
			success:    false,
		},
	} {
		s.Run(tc.desc, func() {
			msg := types.MsgDistributeTokens{
				SenderAddress: tc.senderAddr,
				Amount:        tc.coins,
			}
			_, err := s.app.AppKeepers.DripKeeper.DistributeTokens(s.ctx, &msg)

			if !tc.success {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateDripParams() {
	_, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()

	for _, tc := range []struct {
		desc             string
		isEnabled        bool
		AllowedAddresses []string
		success          bool
	}{
		{
			desc:             "Success - Valid on",
			isEnabled:        true,
			AllowedAddresses: []string{},
			success:          true,
		},
		{
			desc:             "Success - Valid off",
			isEnabled:        false,
			AllowedAddresses: []string{},
			success:          true,
		},
		{
			desc:             "Success - On and 1 allowed address",
			isEnabled:        true,
			AllowedAddresses: []string{addr.String()},
			success:          true,
		},
		{
			desc:             "Fail - On and 2 duplicate addresses",
			isEnabled:        true,
			AllowedAddresses: []string{addr.String(), addr.String()},
			success:          false,
		},
		{
			desc:             "Success - On and 2 unique",
			isEnabled:        true,
			AllowedAddresses: []string{addr.String(), addr2.String()},
			success:          true,
		},
		{
			desc:             "Success - On and 2 duplicate 1 unique",
			isEnabled:        true,
			AllowedAddresses: []string{addr.String(), addr2.String(), addr.String()},
			success:          false,
		},
	} {
		s.Run(tc.desc, func() {
			params := types.Params{
				EnableDrip:       tc.isEnabled,
				AllowedAddresses: tc.AllowedAddresses,
			}
			err := s.app.AppKeepers.DripKeeper.SetParams(s.ctx, params)

			if !tc.success {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
