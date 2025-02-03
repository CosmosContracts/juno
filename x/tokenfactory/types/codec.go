package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateDenom{},
		&MsgMint{},
		&MsgBurn{},
		&MsgForceTransfer{},
		&MsgChangeAdmin{},
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateDenom{}, "osmosis/x/tokenfactory/create-denom", nil)
	cdc.RegisterConcrete(&MsgMint{}, "osmosis/x/tokenfactory/mint", nil)
	cdc.RegisterConcrete(&MsgBurn{}, "osmosis/x/tokenfactory/burn", nil)
	cdc.RegisterConcrete(&MsgForceTransfer{}, "osmosis/x/tokenfactory/force-transfer", nil)
	cdc.RegisterConcrete(&MsgChangeAdmin{}, "osmosis/x/tokenfactory/change-admin", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "osmosis/x/tokenfactory/update-params", nil)
}
