package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	streamv1 "github.com/CosmosContracts/juno/v30/api/juno/stream/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: streamv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "StreamBalance",
					Use:       "stream-balance",
					Short:     "Establishes a grpc streaming connection to the balance of an address",
				},
				{
					RpcMethod: "StreamAllBalances",
					Use:       "stream-all-balances",
					Short:     "Establishes a grpc streaming connection to all balances of an address",
				},
				{
					RpcMethod: "StreamDelegations",
					Use:       "stream-delegations",
					Short:     "Establishes a grpc streaming connection to the delegations of an address",
				},
				{
					RpcMethod: "StreamDelegation",
					Use:       "stream-delegation",
					Short:     "Establishes a grpc streaming connection to the delegation of an address",
				},
				{
					RpcMethod: "StreamUnbondingDelegations",
					Use:       "stream-unbonding-delegations",
					Short:     "Establishes a grpc streaming connection to the unbonding delegations of an address",
				},
				{
					RpcMethod: "StreamUnbondingDelegation",
					Use:       "stream-unbonding-delegation",
					Short:     "Establishes a grpc streaming connection to the unbonding delegation of an address",
				},
			},
		},
	}
}
