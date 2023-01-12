package v1_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"github.com/CosmosContracts/juno/price-feeder/config"
	"github.com/CosmosContracts/juno/price-feeder/oracle"
	"github.com/CosmosContracts/juno/price-feeder/oracle/provider"
	v1 "github.com/CosmosContracts/juno/price-feeder/router/v1"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

var (
	_ v1.Oracle = (*mockOracle)(nil)

	mockPrices = map[string]sdk.Dec{
		"ATOM": sdk.MustNewDecFromStr("34.84"),
		"UMEE": sdk.MustNewDecFromStr("4.21"),
	}

	mockComputedPrices = map[provider.Name]map[string]sdk.Dec{
		provider.ProviderBinance: {
			"ATOM": sdk.MustNewDecFromStr("28.21000000"),
			"UMEE": sdk.MustNewDecFromStr("1.13000000"),
		},
		provider.ProviderKraken: {
			"ATOM": sdk.MustNewDecFromStr("28.268700"),
			"UMEE": sdk.MustNewDecFromStr("1.13000000"),
		},
	}
)

type mockOracle struct{}

func (m mockOracle) GetLastPriceSyncTimestamp() time.Time {
	return time.Now()
}

func (m mockOracle) GetPrices() map[string]sdk.Dec {
	return mockPrices
}

func (m mockOracle) GetTvwapPrices() oracle.PricesByProvider {
	return mockComputedPrices
}

func (m mockOracle) GetVwapPrices() oracle.PricesByProvider {
	return mockComputedPrices
}

type mockMetrics struct{}

func (mockMetrics) Gather(format string) (telemetry.GatherResponse, error) {
	return telemetry.GatherResponse{}, nil
}

type RouterTestSuite struct {
	suite.Suite

	mux    *mux.Router
	router *v1.Router
}

// SetupSuite executes once before the suite's tests are executed.
func (rts *RouterTestSuite) SetupSuite() {
	mux := mux.NewRouter()
	cfg := config.Config{
		Server: config.Server{
			AllowedOrigins: []string{},
			VerboseCORS:    false,
		},
	}

	r := v1.New(zerolog.Nop(), cfg, mockOracle{}, mockMetrics{})
	r.RegisterRoutes(mux, v1.APIPathPrefix)

	rts.mux = mux
	rts.router = r
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}

func (rts *RouterTestSuite) executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	rts.mux.ServeHTTP(rr, req)

	return rr
}

func (rts *RouterTestSuite) TestHealthz() {
	req, err := http.NewRequest("GET", "/api/v1/healthz", nil)
	rts.Require().NoError(err)

	response := rts.executeRequest(req)
	rts.Require().Equal(http.StatusOK, response.Code)

	var respBody map[string]interface{}
	rts.Require().NoError(json.Unmarshal(response.Body.Bytes(), &respBody))
	rts.Require().Equal(respBody["status"], v1.StatusAvailable)
}

func (rts *RouterTestSuite) TestPrices() {
	req, err := http.NewRequest("GET", "/api/v1/prices", nil)
	rts.Require().NoError(err)

	response := rts.executeRequest(req)
	rts.Require().Equal(http.StatusOK, response.Code)

	var respBody v1.PricesResponse
	rts.Require().NoError(json.Unmarshal(response.Body.Bytes(), &respBody))
	rts.Require().Equal(respBody.Prices["ATOM"], mockPrices["ATOM"])
	rts.Require().Equal(respBody.Prices["UMEE"], mockPrices["UMEE"])
	rts.Require().Equal(respBody.Prices["FOO"], sdk.Dec{})
}

func (rts *RouterTestSuite) TestTvwap() {
	req, err := http.NewRequest("GET", "/api/v1/prices/providers/tvwap", nil)
	rts.Require().NoError(err)
	response := rts.executeRequest(req)
	rts.Require().Equal(http.StatusOK, response.Code)

	var respBody v1.PricesPerProviderResponse
	rts.Require().NoError(json.Unmarshal(response.Body.Bytes(), &respBody))
	rts.Require().Equal(
		respBody.Prices[provider.ProviderBinance]["ATOM"],
		mockComputedPrices[provider.ProviderBinance]["ATOM"],
	)
}

func (rts *RouterTestSuite) TestVwap() {
	req, err := http.NewRequest("GET", "/api/v1/prices/providers/vwap", nil)
	rts.Require().NoError(err)
	response := rts.executeRequest(req)
	rts.Require().Equal(http.StatusOK, response.Code)

	var respBody v1.PricesPerProviderResponse
	rts.Require().NoError(json.Unmarshal(response.Body.Bytes(), &respBody))
	rts.Require().Equal(
		respBody.Prices[provider.ProviderBinance]["ATOM"],
		mockComputedPrices[provider.ProviderBinance]["ATOM"],
	)
}
