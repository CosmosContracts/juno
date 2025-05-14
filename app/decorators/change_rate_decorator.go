package decorators

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	MaxChangeRate = "0.05"
)

// MsgChangeRateDecorator defines the AnteHandler that filters & prevents messages
// that create validators and exceed the max change rate of 5%.
type MsgChangeRateDecorator struct {
	sk                      *stakingkeeper.Keeper
	maxCommissionChangeRate sdkmath.LegacyDec
}

// Create new Change Rate Decorator
func NewChangeRateDecorator(sk *stakingkeeper.Keeper) MsgChangeRateDecorator {
	rate, err := sdkmath.LegacyNewDecFromStr(MaxChangeRate)
	if err != nil {
		panic(err)
	}

	return MsgChangeRateDecorator{
		sk:                      sk,
		maxCommissionChangeRate: rate,
	}
}

// The AnteHandle checks for transactions that exceed the max change rate of 5% on the
// creation of a validator.
func (mcr MsgChangeRateDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	err := mcr.hasInvalidCommissionRateMsgs(ctx, tx.GetMsgs())
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// Check if a tx's messages exceed a validator's max change rate
func (mcr MsgChangeRateDecorator) hasInvalidCommissionRateMsgs(ctx sdk.Context, msgs []sdk.Msg) error {
	for _, msg := range msgs {
		// Check if an authz message, loop through all inner messages, and recursively call this function
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			msgs, err := execMsg.GetMessages()
			if err != nil {
				return err
			}

			// Recursively call this function with the inner messages
			err = mcr.hasInvalidCommissionRateMsgs(ctx, msgs)
			if err != nil {
				return err
			}
		}

		// Check for create validator messages
		if msg, ok := msg.(*stakingtypes.MsgCreateValidator); ok && mcr.isInvalidCreateMessage(msg) {
			return fmt.Errorf("max change rate must not exceed %f%%", mcr.maxCommissionChangeRate)
		}

		// Check for edit validator messages
		if msg, ok := msg.(*stakingtypes.MsgEditValidator); ok {
			err := mcr.isInvalidEditMessage(ctx, msg)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Check if the create validator message is invalid
func (mcr MsgChangeRateDecorator) isInvalidCreateMessage(msg *stakingtypes.MsgCreateValidator) bool {
	return msg.Commission.MaxChangeRate.GT(mcr.maxCommissionChangeRate)
}

// Check if the edit validator message is invalid
func (mcr MsgChangeRateDecorator) isInvalidEditMessage(ctx sdk.Context, msg *stakingtypes.MsgEditValidator) error {
	// Skip if the commission rate is not being modified
	if msg.CommissionRate == nil {
		return nil
	}

	bech32Addr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return errors.New("invalid validator address")
	}

	// Get validator info, if exists
	valInfo, err := mcr.sk.GetValidator(ctx, bech32Addr)
	if err != nil {
		return err
	}

	// Check if new commission rate is out of bounds of the max change rate
	if msg.CommissionRate.LT(valInfo.Commission.Rate.Sub(mcr.maxCommissionChangeRate)) || msg.CommissionRate.GT(valInfo.Commission.Rate.Add(mcr.maxCommissionChangeRate)) {
		return fmt.Errorf("commission rate cannot change by more than %f%%", mcr.maxCommissionChangeRate)
	}

	return nil
}
