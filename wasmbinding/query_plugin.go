package wasmbinding

import (
	"encoding/json"
	"fmt"

	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmosContracts/juno/v13/wasmbinding/gov"
	"github.com/CosmosContracts/juno/v13/wasmbinding/types"
)

// QueryDispatcher dispatches custom WASM messages.
type QueryDispatcher struct {
	govHandler gov.QueryHandler
}

// NewQueryDispatcher returns a new QueryDispatcher instance.
func NewQueryDispatcher(govHandler gov.QueryHandler) QueryDispatcher {
	return QueryDispatcher{
		govHandler: govHandler,
	}
}

// DispatchQuery validates and executes a custom WASM query.
func (d QueryDispatcher) DispatchQuery(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	// Parse and validate the input
	var req types.Query
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, sdkErrors.Wrap(sdkErrors.ErrInvalidRequest, fmt.Sprintf("custom query JSON unmarshal: %v", err))
	}
	if err := req.Validate(); err != nil {
		return nil, sdkErrors.Wrap(sdkErrors.ErrInvalidRequest, fmt.Sprintf("custom query validation: %v", err))
	}

	// Execute custom sub-query (one of)
	var resData interface{}
	var resErr error

	switch {
	case req.GovVote != nil:
		resData, resErr = d.govHandler.GetVote(ctx, *req.GovVote)
	default:
		// That should never happen, since we validate the input above
		return nil, wasmVmTypes.UnsupportedRequest{Kind: "no custom querier found"}
	}
	if resErr != nil {
		return nil, resErr
	}

	// Encode the response
	res, err := json.Marshal(resData)
	if err != nil {
		return nil, fmt.Errorf("custom query response JSON marshal: %w", err)
	}

	return res, nil
}
