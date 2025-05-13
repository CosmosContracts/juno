package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	feesharev1 "github.com/CosmosContracts/juno/v29/api/juno/feeshare/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: feesharev1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "FeeShares",
					Use:       "contracts",
					Short:     "Query all FeeShares",
				},
				{
					RpcMethod: "FeeShare",
					Use:       "contract [contract_address]",
					Short:     "Query a registered contract for fee distribution by its bech32 address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "contract_address"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current feeshare module parameters",
				},
				{
					RpcMethod: "DeployerFeeShares",
					Use:       "deployer-contracts [deployer_address]",
					Short:     "Query all contracts that a given deployer has registered for feeshare distribution",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "deployer_address"},
					},
				},
				{
					RpcMethod: "WithdrawerFeeShares",
					Use:       "withdrawer-contracts [withdrawer_address]",
					Short:     "Query all contracts that have been registered for feeshare distribution with a given withdrawer address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "withdrawer_address"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              feesharev1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "RegisterFeeShare",
					Use:       "register [contract_address] [deployer_address] [withdrawer_address]",
					Short:     "Register a contract for fee distribution",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "contract_address"},
						{ProtoField: "deployer_address"},
						{ProtoField: "withdrawer_address"},
					},
				},
				{
					RpcMethod: "CancelFeeShare",
					Use:       "cancel [contract_address] [deployer_address]",
					Short:     "Cancel a contract from feeshare distribution",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "contract_address"},
						{ProtoField: "deployer_address"},
					},
				},
				{
					RpcMethod: "UpdateFeeShare",
					Use:       "update [contract_address] [deployer_address] [withdrawer_address]",
					Short:     "Update withdrawer address for a contract registered for feeshare distribution",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "contract_address"},
						{ProtoField: "deployer_address"},
						{ProtoField: "withdrawer_address"},
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
