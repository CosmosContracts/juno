package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces register implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterFeeShare{},
		&MsgCancelFeeShare{},
		&MsgUpdateFeeShare{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/FeeShare interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization and EIP-712 compatibility.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCancelFeeShare{}, "juno/x/feeshare/MsgCancelFeeShare", nil)
	cdc.RegisterConcrete(&MsgRegisterFeeShare{}, "juno/x/feeshare/MsgRegisterFeeShare", nil)
	cdc.RegisterConcrete(&MsgUpdateFeeShare{}, "juno/x/feeshare/MsgUpdateFeeShare", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "juno/x/feeshare/MsgUpdateParams", nil)
}
