package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	dripv1 "github.com/CosmosContracts/juno/v29/api/juno/drip/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: dripv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current drip module parameters",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              dripv1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "DistributeTokens",
					Use:       "distribute-tokens [sender_address] [amount]",
					Short:     "Distribute tokens to all stakers in the next block",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "amount"},
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
