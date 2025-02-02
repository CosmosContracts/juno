package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Params{}, "cwhooks/Params", nil)
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "cwhooks/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgRegisterStaking{}, "cwhooks/MsgRegisterStaking")
	legacy.RegisterAminoMsg(cdc, &MsgRegisterGovernance{}, "cwhooks/MsgRegisterGovernance")
	legacy.RegisterAminoMsg(cdc, &MsgUnregisterGovernance{}, "cwhooks/MsgUnregisterGovernance")
	legacy.RegisterAminoMsg(cdc, &MsgUnregisterStaking{}, "cwhooks/MsgUnregisterStaking")
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgRegisterGovernance{},
		&MsgRegisterStaking{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
