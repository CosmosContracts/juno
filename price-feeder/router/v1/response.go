package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CosmosContracts/juno/price-feeder/oracle/provider"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Response constants
const (
	StatusAvailable = "available"
)

type (
	// HealthZResponse defines the response type for the healthy API handler.
	HealthZResponse struct {
		Status string `json:"status" yaml:"status"`
		Oracle struct {
			LastSync string `json:"last_sync"`
		} `json:"oracle"`
	}

	// PricesResponse defines the response type for getting the latest exchange
	// rates from the oracle.
	PricesResponse struct {
		Prices map[string]sdk.Dec `json:"prices"`
	}

	PricesPerProviderResponse struct {
		Prices map[provider.Name]map[string]sdk.Dec `json:"providers"`
	}
)

// errorResponse defines the attributes of a JSON error response.
type errorResponse struct {
	Code  int    `json:"code,omitempty"`
	Error string `json:"error"`
}

// newErrorResponse creates a new errorResponse instance.
func newErrorResponse(code int, err string) errorResponse {
	return errorResponse{Code: code, Error: err}
}

// writeErrorResponse prepares and writes a HTTP error
// given a status code and an error message.
func writeErrorResponse(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if bz, err := json.Marshal(newErrorResponse(0, err)); err == nil {
		_, _ = w.Write(bz)
	} else {
		_, _ = w.Write([]byte(fmt.Sprintf("failed to marshal error response: %s", err)))
	}
}
