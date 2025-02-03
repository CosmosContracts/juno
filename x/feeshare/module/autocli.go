package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	feesharev1 "github.com/CosmosContracts/juno/v27/api/juno/feeshare/v1"
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
			Service: feesharev1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "RegisterFeeShare",
					Use:       "register [sender_address] [contract_bech32] [withdraw_bech32]",
					Short:     "Register a contract for fee distribution",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_bech32"},
						{ProtoField: "withdraw_bech32"},
					},
				},
				{
					RpcMethod: "CancelFeeShare",
					Use:       "cancel [sender_address] [contract_bech32]",
					Short:     "Cancel a contract from feeshare distribution",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_bech32"},
					},
				},
				{
					RpcMethod: "UpdateFeeShare",
					Use:       "update [sender_address] [contract_bech32] [new_withdraw_bech32]",
					Short:     "Update withdrawer address for a contract registered for feeshare distribution",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_bech32"},
						{ProtoField: "new_withdraw_bech32"},
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
