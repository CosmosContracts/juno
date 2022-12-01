package types

// feeshare events
const (
	EventTypeRegisterFeeShare      = "register_feeshare"
	EventTypeCancelFeeShare        = "cancel_feeshare"
	EventTypeUpdateFeeShare        = "update_feeshare"
	EventTypeDistributeDevFeeShare = "distribute_dev_feeshare"

	AttributeKeyContract          = "contract"
	AttributeKeyWithdrawerAddress = "withdrawer_address"
)
