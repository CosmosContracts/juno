package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	globalfeev1beta1 "github.com/CosmosContracts/juno/v27/api/gaia/globalfee/v1beta1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: globalfeev1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "MinimumGasPrices",
					Use:       "minimum-gas-prices",
					Short:     "Show minimum gas prices",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: globalfeev1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
			},
		},
	}
}
