package decorate

import (
	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
)

// ContractGasTXDecorator ante handler to Execute && Instantiate gas limit.
type ContractGasTXDecorator struct {
}

// NewContractGasTXDecorator constructor
func NewContractGasTXDecorator() *ContractGasTXDecorator {
	return &ContractGasTXDecorator{}
}

// Ante handler limit gas used in `MsgInstantiateContract` and `MsgExecuteContract`
func (a ContractGasTXDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if simulate {
		return next(ctx, tx, simulate)
	}

	for _, msg := range tx.GetMsgs() {
		if sdk.MsgTypeURL(msg) == "/cosmwasm.wasm.v1.MsgInstantiateContract" || sdk.MsgTypeURL(msg) == "/cosmwasm.wasm.v1.MsgExecuteContract" {
			gasTx, ok := tx.(authante.GasTx)
			if !ok {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be GasTx")
			}

			if gasTx.GetGas() > 500000 {
				return ctx, sdkerrors.Wrap(types.ErrInvalid, "Gas in Execute or Instantiate contract should be lower than 500000")
			}
		}
	}
	return next(ctx, tx, simulate)
}
