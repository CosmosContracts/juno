package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

const (
	// Amino names
	registerFeePayContract   = "juno/x/feepay/MsgRegisterFeePayContract"
	unregisterFeePayContract = "juno/x/feepay/MsgUnregisterFeePayContract"
	fundFeePayContract       = "juno/x/feepay/MsgFundFeePayContract"
	updateFeeShareParams     = "juno/x/feepay/MsgUpdateParams"
)

// RegisterInterfaces register implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterFeePayContract{},
		&MsgUnregisterFeePayContract{},
		&MsgFundFeePayContract{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/FeeShare interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization and EIP-712 compatibility.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterFeePayContract{}, registerFeePayContract, nil)
	cdc.RegisterConcrete(&MsgUnregisterFeePayContract{}, unregisterFeePayContract, nil)
	cdc.RegisterConcrete(&MsgFundFeePayContract{}, fundFeePayContract, nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, updateFeeShareParams, nil)
}
