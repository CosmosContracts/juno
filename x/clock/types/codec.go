package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterClockContract{}, "clock/MsgRegisterClockContract", nil)
	cdc.RegisterConcrete(&MsgUnregisterClockContract{}, "clock/MsgUnregisterClockContract", nil)
	cdc.RegisterConcrete(&MsgUnjailClockContract{}, "clock/MsgUnjailClockContract", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "clock/MsgUpdateParams", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterClockContract{},
		&MsgUnregisterClockContract{},
		&MsgUnjailClockContract{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
