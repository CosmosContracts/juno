package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	feemarketv1 "github.com/CosmosContracts/juno/v30/api/feemarket/feemarket/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: feemarketv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GasPrices",
					Use:       "gas-prices",
					Short:     "Query all current gas prices",
				},
				{
					RpcMethod: "GasPrice",
					Use:       "gas-price [denom]",
					Short:     "Query the current gas price per denom",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
					},
				},
				{
					RpcMethod: "State",
					Use:       "state",
					Short:     "Query the current feemarket state",
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current feemarket module parameters",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              feemarketv1.Msg_ServiceDesc.ServiceName,
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
