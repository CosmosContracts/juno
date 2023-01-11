package decorators

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var MiniumInitialDepositRate = sdk.NewDecWithPrec(20, 2)

type GovPreventSpamDecorator struct {
	govKeeper govkeeper.Keeper
	cdc       codec.BinaryCodec
}

func NewGovPreventSpamDecorator(cdc codec.BinaryCodec, govKeeper govkeeper.Keeper) GovPreventSpamDecorator {
	return GovPreventSpamDecorator{
		govKeeper: govKeeper,
		cdc:       cdc,
	}
}

func (gpsd GovPreventSpamDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx,
	simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if DefaultIsAppSimulation {
		return next(ctx, tx, simulate)
	}
	msgs := tx.GetMsgs()

	err = gpsd.checkSpamSubmitProposalMsg(ctx, msgs)

	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (gpsd GovPreventSpamDecorator) checkSpamSubmitProposalMsg(ctx sdk.Context, msgs []sdk.Msg) error {
	validMsg := func(m sdk.Msg) error {
		if msg, ok := m.(*govtypes.MsgSubmitProposal); ok {
			// prevent spam gov msg
			depositParams := gpsd.govKeeper.GetDepositParams(ctx)
			miniumInitialDeposit := gpsd.calcMiniumInitialDeposit(depositParams.MinDeposit)
			if msg.InitialDeposit.IsAllLT(miniumInitialDeposit) {
				return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "not enough initial deposit. required: %v", miniumInitialDeposit)
			}
		}
		return nil
	}

	// Check every msg in the tx, if it's a MsgExec, check the inner msgs.
	// If it's a MsgSubmitProposal, check the initial deposit is enough.
	for _, m := range msgs {
		var innerMsg sdk.Msg
		if msg, ok := m.(*authz.MsgExec); ok {
			for _, v := range msg.Msgs {
				err := gpsd.cdc.UnpackAny(v, &innerMsg)
				if err != nil {
					return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
				}

				err = validMsg(innerMsg)
				if err != nil {
					return err
				}
			}
		} else {
			err := validMsg(m)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (gpsd GovPreventSpamDecorator) calcMiniumInitialDeposit(minDeposit sdk.Coins) (miniumInitialDeposit sdk.Coins) {
	for _, coin := range minDeposit {
		miniumInitialCoin := MiniumInitialDepositRate.MulInt(coin.Amount).RoundInt()
		miniumInitialDeposit = miniumInitialDeposit.Add(sdk.NewCoin(coin.Denom, miniumInitialCoin))
	}

	return
}
