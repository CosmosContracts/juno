package keeper_test

import (
	_ "embed"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v29/x/drip/types"
)

func (s *KeeperTestSuite) TestDripDistributeTokensMsgs() {
	_, _, allowedSender := testdata.KeyTestPubAddr()
	_, _, notAllowedSender := testdata.KeyTestPubAddr()
	s.FundAcc(allowedSender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))
	s.FundAcc(notAllowedSender, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1_000_000))))

	_ = s.App.AppKeepers.DripKeeper.SetParams(s.Ctx, types.Params{
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
			coins:      sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1))),
			success:    true,
		},
		{
			desc:       "Fail - Allowed sender no proper funds",
			senderAddr: allowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin("notarealtoken", sdkmath.NewInt(1))),
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
			coins:      sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1))),
			success:    false,
		},
		{
			desc:       "Fail - Non Allowed sender proper funds",
			senderAddr: notAllowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1))),
			success:    false,
		},
		{
			desc:       "Fail - Non Allowed sender improper funds",
			senderAddr: notAllowedSender.String(),
			coins:      sdk.NewCoins(sdk.NewCoin("notarealtoken", sdkmath.NewInt(1))),
			success:    false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			msg := types.MsgDistributeTokens{
				SenderAddress: tc.senderAddr,
				Amount:        tc.coins,
			}
			_, err := s.msgServer.DistributeTokens(s.Ctx, &msg)

			if !tc.success {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateDripParams() {
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
		tc := tc
		s.Run(tc.desc, func() {
			params := types.Params{
				EnableDrip:       tc.isEnabled,
				AllowedAddresses: tc.AllowedAddresses,
			}
			err := s.App.AppKeepers.DripKeeper.SetParams(s.Ctx, params)

			if !tc.success {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
