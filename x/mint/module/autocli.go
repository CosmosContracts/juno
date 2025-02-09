package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	mintv1 "github.com/CosmosContracts/juno/v28/api/juno/mint/v1"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: mintv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current minting parameters",
				},
				{
					RpcMethod: "Inflation",
					Use:       "inflation",
					Short:     "Query the current minting inflation value",
				},
				{
					RpcMethod: "AnnualProvisions",
					Use:       "annual-provisions",
					Short:     "Query the current minting annual provisions value",
				},
				{
					RpcMethod: "TargetSupply",
					Use:       "target-supply",
					Short:     "Query the current target supply until the next inflation phase",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              mintv1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
			},
		},
	}
}
