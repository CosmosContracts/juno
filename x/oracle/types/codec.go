package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/oracle module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/staking and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}

// RegisterLegacyAminoCodec registers the necessary x/oracle interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAggregateExchangeRatePrevote{}, "juno/oracle/MsgAggregateExchangeRatePrevote", nil)
	cdc.RegisterConcrete(&MsgAggregateExchangeRateVote{}, "juno/oracle/MsgAggregateExchangeRateVote", nil)
	cdc.RegisterConcrete(&MsgDelegateFeedConsent{}, "juno/oracle/MsgDelegateFeedConsent", nil)

	cdc.RegisterConcrete(&AddTrackingPriceHistoryProposal{}, "juno/oracle/AddTrackingPriceHistoryProposal", nil)
	cdc.RegisterConcrete(&AddTrackingPriceHistoryWithAcceptListProposal{}, "juno/oracle/AddTrackingPriceHistoryWithAcceptListProposal", nil)
	cdc.RegisterConcrete(&RemoveTrackingPriceHistoryProposal{}, "juno/oracle/RemoveTrackingPriceHistoryProposal", nil)
}

// RegisterInterfaces registers the x/oracle interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgDelegateFeedConsent{},
		&MsgAggregateExchangeRatePrevote{},
		&MsgAggregateExchangeRateVote{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&AddTrackingPriceHistoryProposal{},
		&AddTrackingPriceHistoryWithAcceptListProposal{},
		&RemoveTrackingPriceHistoryProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
