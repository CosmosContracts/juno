package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	feepayv1 "github.com/CosmosContracts/juno/v27/api/juno/feepay/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: feepayv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "FeePayContract",
					Use:       "contract [contract_address]",
					Short:     "Query a FeePay contract by address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "contract_address"},
					},
				},
				{
					RpcMethod: "FeePayContracts",
					Use:       "contracts",
					Short:     "Query all FeePay contracts",
				},
				{
					RpcMethod: "FeePayContractUses",
					Use:       "uses [contract_address] [wallet_address]",
					Short:     "Query wallet usage on FeePay contract",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "contract_address"},
						{ProtoField: "wallet_address"},
					},
				},
				{
					RpcMethod: "FeePayWalletIsEligible",
					Use:       "is-eligible [contract_address] [wallet_address]",
					Short:     "Query if a wallet is eligible to interact with a FeePay contract",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "contract_address"},
						{ProtoField: "wallet_address"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current feepay module parameters",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: feepayv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "RegisterFeePayContract",
					Use:       "register [sender_address] [contract_address] [wallet_limit]",
					Short:     "Register a contract for fee pay",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_address"},
						{ProtoField: "wallet_limit"},
					},
				},
				{
					RpcMethod: "UnregisterFeePayContract",
					Use:       "unregister [sender_address] [contract_address]",
					Short:     "Unregister a contract for fee pay",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_address"},
					},
				},
				{
					RpcMethod: "FundFeePayContract",
					Use:       "fund [sender_address] [contract_address] [amount]",
					Short:     "Send funds to a registered fee pay contract",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_address"},
						{ProtoField: "amount"},
					},
				},
				{
					RpcMethod: "UpdateFeePayContractWalletLimit",
					Use:       "update-wallet-limit [sender_address] [contract_address] [wallet_limit]",
					Short:     "Update the wallet limit of a fee pay contract",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender_address"},
						{ProtoField: "contract_address"},
						{ProtoField: "wallet_limit"},
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
