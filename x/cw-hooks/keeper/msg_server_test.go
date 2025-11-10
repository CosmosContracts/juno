//nolint:dupl // intentional duplication across table-driven tests
package keeper_test

import (
	_ "embed"

	"github.com/CosmosContracts/juno/v30/x/cw-hooks/types"
)

func (s *KeeperTestSuite) TestRegisterContracts() {
	type registerCase struct {
		desc         string
		module       string
		contractAddr func(contractTestContext) string
		senderAddr   func(contractTestContext) string
		pre          func(contractTestContext)
		shouldErr    bool
	}

	cases := []registerCase{
		{
			desc:   "invalid contract address",
			module: "staking",
			contractAddr: func(contractTestContext) string {
				return "Invalid"
			},
			senderAddr: func(ctx contractTestContext) string {
				return ctx.sender.String()
			},
			shouldErr: true,
		},
		{
			desc:   "invalid register address",
			module: "staking",
			contractAddr: func(ctx contractTestContext) string {
				return ctx.contract
			},
			senderAddr: func(contractTestContext) string {
				return "Invalid"
			},
			shouldErr: true,
		},
		{
			desc:   "unauthorized creator",
			module: "staking",
			contractAddr: func(ctx contractTestContext) string {
				return ctx.contract
			},
			senderAddr: func(ctx contractTestContext) string {
				return ctx.notAuthorized.String()
			},
			shouldErr: true,
		},
		{
			desc:   "unauthorized admin",
			module: "staking",
			contractAddr: func(ctx contractTestContext) string {
				return ctx.contractWithAdmin
			},
			senderAddr: func(ctx contractTestContext) string {
				return ctx.notAuthorized.String()
			},
			shouldErr: true,
		},
		{
			desc:   "success staking",
			module: "staking",
			contractAddr: func(ctx contractTestContext) string {
				return ctx.contract
			},
			senderAddr: func(ctx contractTestContext) string {
				return ctx.sender.String()
			},
		},
		{
			desc:   "duplicate staking registration",
			module: "staking",
			contractAddr: func(ctx contractTestContext) string {
				return ctx.contract
			},
			senderAddr: func(ctx contractTestContext) string {
				return ctx.sender.String()
			},
			pre: func(ctx contractTestContext) {
				_, err := s.msgServer.RegisterContract(s.Ctx, &types.MsgRegisterContract{
					Module:          "staking",
					SenderAddress:   ctx.sender.String(),
					ContractAddress: ctx.contract,
				})
				s.Require().NoError(err)
			},
			shouldErr: true,
		},
		{
			desc:   "success register gov after staking",
			module: "gov",
			contractAddr: func(ctx contractTestContext) string {
				return ctx.contract
			},
			senderAddr: func(ctx contractTestContext) string {
				return ctx.sender.String()
			},
			pre: func(ctx contractTestContext) {
				_, err := s.msgServer.RegisterContract(s.Ctx, &types.MsgRegisterContract{
					Module:          "staking",
					SenderAddress:   ctx.sender.String(),
					ContractAddress: ctx.contract,
				})
				s.Require().NoError(err)
			},
		},
		{
			desc:   "duplicate governance registration",
			module: "gov",
			contractAddr: func(ctx contractTestContext) string {
				return ctx.contract
			},
			senderAddr: func(ctx contractTestContext) string {
				return ctx.sender.String()
			},
			pre: func(ctx contractTestContext) {
				_, err := s.msgServer.RegisterContract(s.Ctx, &types.MsgRegisterContract{
					Module:          "gov",
					SenderAddress:   ctx.sender.String(),
					ContractAddress: ctx.contract,
				})
				s.Require().NoError(err)
			},
			shouldErr: true,
		},
		{
			desc:   "register DAODAO factory child",
			module: "staking",
			contractAddr: func(ctx contractTestContext) string {
				return ctx.daoChild
			},
			senderAddr: func(ctx contractTestContext) string {
				return ctx.dao
			},
		},
		{
			desc:   "unsupported module",
			module: "unknown",
			contractAddr: func(ctx contractTestContext) string {
				return ctx.contract
			},
			senderAddr: func(ctx contractTestContext) string {
				return ctx.sender.String()
			},
			shouldErr: true,
		},
	}

	for _, tc := range cases {
		s.Run(tc.desc, func() {
			s.SetupTest()
			ctxData := s.buildContractTestContext()

			if tc.pre != nil {
				tc.pre(ctxData)
			}

			resp, err := s.msgServer.RegisterContract(s.Ctx, &types.MsgRegisterContract{
				Module:          tc.module,
				SenderAddress:   tc.senderAddr(ctxData),
				ContractAddress: tc.contractAddr(ctxData),
			})
			if !tc.shouldErr {
				s.Require().NoError(err)
				s.Require().Equal(&types.MsgRegisterContractResponse{}, resp)
			} else {
				s.Require().Error(err)
				s.Require().Nil(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUnRegisterContracts() {
	type unregisterCase struct {
		desc      string
		module    string
		pre       func(contractTestContext)
		shouldErr bool
	}

	cases := []unregisterCase{
		{
			desc:      "invalid contract address",
			module:    "staking",
			shouldErr: true,
		},
		{
			desc:   "invalid register address",
			module: "staking",
			pre: func(ctx contractTestContext) {
				s.Require().NoError(s.registerContract("staking", ctx.sender.String(), ctx.contract))
			},
			shouldErr: true,
		},
		{
			desc:   "unauthorized sender",
			module: "staking",
			pre: func(ctx contractTestContext) {
				s.Require().NoError(s.registerContract("staking", ctx.sender.String(), ctx.contract))
			},
			shouldErr: true,
		},
		{
			desc:   "success staking unregister",
			module: "staking",
			pre: func(ctx contractTestContext) {
				s.Require().NoError(s.registerContract("staking", ctx.sender.String(), ctx.contract))
			},
		},
		{
			desc:   "duplicate staking unregister",
			module: "staking",
			pre: func(ctx contractTestContext) {
				s.Require().NoError(s.registerContract("staking", ctx.sender.String(), ctx.contract))
				s.Require().NoError(s.unregisterContract("staking", ctx.sender.String(), ctx.contract))
			},
			shouldErr: true,
		},
		{
			desc:   "success governance unregister",
			module: "gov",
			pre: func(ctx contractTestContext) {
				s.Require().NoError(s.registerContract("gov", ctx.sender.String(), ctx.contract))
			},
		},
		{
			desc:   "duplicate governance unregister",
			module: "gov",
			pre: func(ctx contractTestContext) {
				s.Require().NoError(s.registerContract("gov", ctx.sender.String(), ctx.contract))
				s.Require().NoError(s.unregisterContract("gov", ctx.sender.String(), ctx.contract))
			},
			shouldErr: true,
		},
		{
			desc:      "unsupported module",
			module:    "unknown",
			shouldErr: true,
		},
	}

	for _, tc := range cases {
		s.Run(tc.desc, func() {
			s.SetupTest()
			ctxData := s.buildContractTestContext()

			if tc.pre != nil {
				tc.pre(ctxData)
			}

			resp, err := s.msgServer.UnregisterContract(s.Ctx, &types.MsgUnregisterContract{
				Module:          tc.module,
				SenderAddress:   ctxData.sender.String(),
				ContractAddress: ctxData.contract,
			})
			if !tc.shouldErr {
				s.Require().NoError(err)
				s.Require().Equal(&types.MsgUnregisterContractResponse{}, resp)
			} else {
				s.Require().Error(err)
				s.Require().Nil(resp)
			}
		})
	}
}
