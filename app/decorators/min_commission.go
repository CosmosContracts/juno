package decorators

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var DefaultIsAppSimulation = false

type MinCommissionDecorator struct {
	cdc codec.BinaryCodec
}

// NewMinCommissionDecorator returns a new MinCommissionDecorator
func NewMinCommissionDecorator(cdc codec.BinaryCodec) MinCommissionDecorator {
	return MinCommissionDecorator{cdc}
}

// AnteHandle calls the next AnteHandler after validating the commission rate
func (min MinCommissionDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx,
	simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if DefaultIsAppSimulation {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()
	minCommissionRate := sdk.NewDecWithPrec(5, 2)

	if err := min.validateCommissionRate(msgs, minCommissionRate); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// validateCommissionRate validates the commission rate of the msgs
func (MinCommissionDecorator) validateCommissionRate(msgs []sdk.Msg, minCommissionRate sdk.Dec) error {
	validMsg := func(m sdk.Msg) error {
		switch msg := m.(type) {
		case *stakingtypes.MsgCreateValidator:
			// prevent new validators joining the set with
			// commission set below 5%
			c := msg.Commission
			if c.Rate.LT(minCommissionRate) {
				return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "commission can't be lower than 5%")
			}
		case *stakingtypes.MsgEditValidator:
			// if commission rate is nil, it means only
			// other fields are affected - skip
			if msg.CommissionRate == nil {
				break
			}
			if msg.CommissionRate.LT(minCommissionRate) {
				return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "commission can't be lower than 5%")
			}
		}

		return nil
	}

	// Check every msg in the tx, if it's a MsgExec, check the inner msgs.
	for _, m := range msgs {
		err := validMsg(m)
		if err != nil {
			return err
		}
	}

	return nil
}
