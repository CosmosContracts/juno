package decorate_test

import (
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/CosmosContracts/juno/v10/app/decorate"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestContractGasTXDecorator(t *testing.T) {
	specs := map[string]struct {
		maxBlockGas uint64
		msg         sdk.Msg
		gasUsed     uint64
		expErr      bool
	}{
		"valid gas": {
			msg:     &types.MsgExecuteContract{},
			gasUsed: 400000,
			expErr:  true,
		},
		"unvalid gas": {
			msg:     &types.MsgExecuteContract{},
			gasUsed: 600000,
			expErr:  false,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			maxBlockGas := uint64(1000000)
			nextAnte := consumeGasAnteHandler(maxBlockGas)
			ctx := sdk.Context{}.WithGasMeter(sdk.NewInfiniteGasMeter())
			ante := decorate.NewContractGasTXDecorator()
			msgs := []sdk.Msg{spec.msg}

			encodingConfig := simappparams.MakeTestEncodingConfig()
			txBuilder := encodingConfig.TxConfig.NewTxBuilder()
			txBuilder.SetMsgs(msgs...)
			txBuilder.SetGasLimit(spec.gasUsed)

			_, err := ante.AnteHandle(ctx, txBuilder.GetTx(), false, nextAnte)
			if spec.expErr == false {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func consumeGasAnteHandler(gasToConsume sdk.Gas) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		ctx.GasMeter().ConsumeGas(gasToConsume, "testing")
		return ctx, nil
	}
}
