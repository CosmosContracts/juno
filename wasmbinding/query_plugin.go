package wasmbinding

import (
	"encoding/json"

	"github.com/CosmosContracts/juno/v12/wasmbinding/bindings"
	oracle "github.com/CosmosContracts/juno/v12/x/oracle/wasm"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CustomQuerier dispatches custom CosmWasm bindings queries.
func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindings.CosmosQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, sdkerrors.Wrap(err, "kujira query")
		}

		if contractQuery.Oracle != nil {
			res, err := oracle.Handle(qp.oraclekeeper, ctx, contractQuery.Oracle)
			if err != nil {
				return nil, err
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
			}

			return bz, nil
		} else {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown Custom variant"}
		}
	}
}
