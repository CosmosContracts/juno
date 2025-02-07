package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	tokenfactoryv1beta1 "github.com/CosmosContracts/juno/v27/api/osmosis/tokenfactory/v1beta1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: tokenfactoryv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Get the params for the x/tokenfactory module",
				},
				{
					RpcMethod: "DenomAuthorityMetadata",
					Use:       "denom-authority-metadata [denom]",
					Short:     "Get the authority metadata for a specific denom",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
					},
				},
				{
					RpcMethod: "DenomsFromCreator",
					Use:       "denoms-from-creator [creator",
					Short:     "Returns a list of all tokens created by a specific creator address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "creator"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              tokenfactoryv1beta1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CreateDenom",
					Use:       "create-denom [sender] [subdenom]",
					Short:     "Create a new denom from an account",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "subdenom"},
					},
				},
				{
					RpcMethod: "Mint",
					Use:       "mint [sender] [amount] [mint_to_address]",
					Short:     "Mint a denom to your address. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "amount"},
						{ProtoField: "mint_to_address"},
					},
				},
				{
					RpcMethod: "Burn",
					Use:       "burn [sender] [amount] [burn_from_address]",
					Short:     "Burn tokens from an address. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "amount"},
						{ProtoField: "burn_from_address"},
					},
				},
				{
					RpcMethod: "ForceTransfer",
					Use:       "force-transfer [sender] [amount] [transfer-from-address] [transfer-to-address]",
					Short:     "Force transfer tokens from one address to another address. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "amount"},
						{ProtoField: "transfer_from_address"},
						{ProtoField: "transfer_to_address"},
					},
				},
				{
					RpcMethod: "ChangeAdmin",
					Use:       "change-admin [sender] [denom] [new_admin]",
					Short:     "Changes the admin address for a factory-created denom. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "denom"},
						{ProtoField: "new_admin"},
					},
				},
				{
					RpcMethod: "SetDenomMetadata",
					Use:       "modify-metadata [sender] [metadata]",
					Short:     "Changes the base data for frontends to query the data of.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "metadata"},
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
