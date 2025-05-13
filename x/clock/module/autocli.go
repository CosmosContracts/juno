package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	clockv1 "github.com/CosmosContracts/juno/v29/api/juno/clock/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: clockv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "ClockContracts",
					Use:       "contracts",
					Short:     "Show addresses of all current clock contracts",
				},
				{
					RpcMethod: "ClockContract",
					Use:       "contract [contract_address]",
					Short:     "Get contract by address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "contract_address"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Show all module params",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              clockv1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "RegisterClockContract",
					Use:       "register [sender_address] [contract_address]",
					Short:     "Register a clock contract.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_address"},
					},
				},
				{
					RpcMethod: "UnregisterClockContract",
					Use:       "unregister [sender_address] [contract_address]",
					Short:     "Unregister a clock contract.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_address"},
					},
				},
				{
					RpcMethod: "UnjailClockContract",
					Use:       "unjail [sender_address] [contract_address]",
					Short:     "Unjail a clock contract.",
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
