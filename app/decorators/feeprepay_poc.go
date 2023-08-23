package decorators

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgFeePrepayDecorator struct{}

func NewMsgFeePrepayDecorator() MsgFeePrepayDecorator {
	return MsgFeePrepayDecorator{}
}

func (mfd MsgFeePrepayDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if hasInvalidMsgs(tx.GetMsgs()) {
		currHeight := ctx.BlockHeight()
		return ctx, fmt.Errorf("tx contains unsupported message types at height %d", currHeight)
	}

	return next(ctx, tx, simulate)
}

// func hasInvalidMsgs(msgs []sdk.Msg) bool {
// 	for _, msg := range msgs {
// 		if _, ok := msg.(*ibcchanneltypes.MsgTimeoutOnClose); ok {
// 			return true
// 		}
// 	}

// 	return false
// }
