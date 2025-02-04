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
	cdc.RegisterConcrete(Params{}, "juno/x/cwhooks/Params", nil)
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "juno/x/cwhooks/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgRegisterStaking{}, "juno/x/cwhooks/MsgRegisterStaking")
	legacy.RegisterAminoMsg(cdc, &MsgRegisterGovernance{}, "juno/x/cwhooks/MsgRegisterGovernance")
	legacy.RegisterAminoMsg(cdc, &MsgUnregisterGovernance{}, "juno/x/cwhooks/MsgUnregisterGovernance")
	legacy.RegisterAminoMsg(cdc, &MsgUnregisterStaking{}, "juno/x/cwhooks/MsgUnregisterStaking")
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgRegisterGovernance{},
		&MsgRegisterStaking{},
		&MsgUnregisterGovernance{},
		&MsgUnregisterStaking{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
