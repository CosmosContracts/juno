package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	cwhooksv1 "github.com/CosmosContracts/juno/v27/api/juno/cwhooks/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: cwhooksv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Show all module params",
				},
				{
					RpcMethod: "StakingContracts",
					Use:       "staking-contracts",
					Short:     "Show all staking contracts",
				},
				{
					RpcMethod: "GovernanceContracts",
					Use:       "governance-contracts",
					Short:     "Show all governance contracts",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: cwhooksv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "RegisterStaking",
					Use:       "register-staking [sender_address] [contract_address]",
					Short:     "Register a staking contract for sudo message updates",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_address"},
					},
				},
				{
					RpcMethod: "RegisterGovernance",
					Use:       "register-governance [sender_address] [contract_address]",
					Short:     "Register a governance contract for sudo message updates",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_address"},
					},
				},
				{
					RpcMethod: "UnregisterStaking",
					Use:       "unregister-staking [sender_address] [contract_address]",
					Short:     "Remove a staking contract from receiving sudo message updates",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_address"},
					},
				},
				{
					RpcMethod: "UnregisterGovernance",
					Use:       "unregister-governance [sender_address] [contract_address]",
					Short:     "Remove a governance contract from receiving sudo message updates",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_address"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
			},
		},
	}
}
