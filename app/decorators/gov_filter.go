package decorators

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var MiniumInitialDeposit = sdk.NewInt64Coin("ujuno", 10000000)

type GovPreventSpamDecorator struct {
	cdc codec.BinaryCodec
}

func NewGovPreventSpamDecorator(cdc codec.BinaryCodec) MinCommissionDecorator {
	return MinCommissionDecorator{cdc}
}

func (gpsd GovPreventSpamDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx,
	simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()

	err = gpsd.checkSpamSubmitProposalMsg(msgs)

	if err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}

func (gpsd GovPreventSpamDecorator) checkSpamSubmitProposalMsg(msgs []sdk.Msg) error {
	validMsg := func(m sdk.Msg) error {
		switch msg := m.(type) {
		case *govtypes.MsgSubmitProposal:
			// prevent spam gov msg
			if msg.InitialDeposit.IsAllLT(sdk.NewCoins(MiniumInitialDeposit)) {
				return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "not enough initial deposit")
			}
		}

		return nil
	}

	validAuthz := func(execMsg *authz.MsgExec) error {
		for _, v := range execMsg.Msgs {
			var innerMsg sdk.Msg
			err := gpsd.cdc.UnpackAny(v, &innerMsg)
			if err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
			}

			err = validMsg(innerMsg)
			if err != nil {
				return err
			}
		}

		return nil
	}

	for _, m := range msgs {
		if msg, ok := m.(*authz.MsgExec); ok {
			if err := validAuthz(msg); err != nil {
				return err
			}
			continue
		}

		// validate normal msgs
		err := validMsg(m)
		if err != nil {
			return err
		}
	}
	return nil
}
