package ante

import (
	ibcchanneltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
)

// MsgFilterDecorator defines an AnteHandler decorator for the v9 upgrade that
// provide height-gated message filtering acceptance.
type MsgFilterDecorator struct{}

// AnteHandle performs an AnteHandler check that returns an error if the tx contains a message
// that is blocked.
// Right now, we block MsgTimeoutOnClose due to incorrect behavior that could occur if a packet is re-enabled.
func (MsgFilterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if hasInvalidMsgs(tx.GetMsgs()) {
		currHeight := ctx.BlockHeight()
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "tx contains unsupported message types at height %d", currHeight)
	}

	return next(ctx, tx, simulate)
}

// hasInvalidMsgs walks top-level messages and recurses into authz.MsgExec
// payloads so that wrapping a blocked message in MsgExec does not bypass
// the filter. Mirrors the recursion pattern used by x/feeshare/ante.
func hasInvalidMsgs(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		if _, ok := msg.(*ibcchanneltypes.MsgTimeoutOnClose); ok {
			return true
		}
		if exec, ok := msg.(*authz.MsgExec); ok {
			inner, err := exec.GetMessages()
			if err != nil {
				return true
			}
			if hasInvalidMsgs(inner) {
				return true
			}
		}
	}

	return false
}
