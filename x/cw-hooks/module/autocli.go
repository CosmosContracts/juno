package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/CosmosContracts/juno/v30/api/juno/cwhooks/v2"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: cwhooksv2.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Show all module params",
				},
				{
					RpcMethod: "Contracts",
					Use:       "contracts",
					Short:     "Show all registered contracts for a given module",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "module"},
					},
				},
				{
					RpcMethod: "ContractInfo",
					Use:       "contract",
					Short:     "Show information about a registered contract",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "module"},
						{ProtoField: "contract_address"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              cwhooksv2.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "RegisterContract",
					Use:       "register",
					Short:     "Register a contract for a given module",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "module"},
						{ProtoField: "contract_address"},
					},
				},
				{
					RpcMethod: "UnregisterContract",
					Use:       "unregister",
					Short:     "Unregister a contract from a given module",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "module"},
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
