package decorators

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"

	wasmtype "github.com/CosmWasm/wasmd/x/wasm/types"
)

// prevents a contract from using > some amount of set gas in a single execute

type PreventWasmDecorator struct {
	cdc              codec.BinaryCodec
	contractGasLimit uint64
}

func NewPreventWasmHighGasUsageDecorator(cdc codec.BinaryCodec, gasLimit uint64) PreventWasmDecorator {
	return PreventWasmDecorator{
		cdc:              cdc,
		contractGasLimit: gasLimit,
	}
}

func (pwasmd PreventWasmDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if DefaultIsAppSimulation {
		return next(ctx, tx, simulate)
	}

	if pwasmd.hasAnyWasmMessages(ctx, tx.GetMsgs()) {
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
		}

		gasUsed := feeTx.GetGas()
		if gasUsed > pwasmd.contractGasLimit {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "gas limit of %d exceeded with a wasm execute. You used: %d", pwasmd.contractGasLimit, gasUsed)
		}
	}

	return next(ctx, tx, simulate)
}

func (pwasmd PreventWasmDecorator) hasAnyWasmMessages(ctx sdk.Context, msgs []sdk.Msg) bool {
	isWasmMessage := func(m sdk.Msg) bool {
		switch m.(type) {
		case *wasmtype.MsgInstantiateContract, *wasmtype.MsgInstantiateContract2:
			return true
		case *wasmtype.MsgExecuteContract:
			return true
		case *wasmtype.MsgMigrateContract:
			return true
		default:
			return false
		}
	}

	// Check every msg in the tx, if it's a MsgExec, check the inner msgs.
	for _, m := range msgs {
		if isWasmMessage(m) {
			return true
		}

		if msg, ok := m.(*authz.MsgExec); ok {
			for _, v := range msg.Msgs {
				var innerMsg sdk.Msg
				err := pwasmd.cdc.UnpackAny(v, &innerMsg)
				if err != nil {
					return false
				}

				if isWasmMessage(innerMsg) {
					return true
				}
			}
		}
	}

	return false
}
