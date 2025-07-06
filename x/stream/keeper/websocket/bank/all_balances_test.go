package bank_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/bank"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAllBalancesHandler_Handle(t *testing.T) {
	validAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"

	tests := []struct {
		name               string
		address            string
		setupMocks         func(*MockKeeperInterface, *MockBankKeeper)
		expectedStatus     int
		expectedError      string
		skipWebSocketCheck bool
	}{
		{
			name:    "valid address with balances",
			address: validAddress,
			setupMocks: func(mk *MockKeeperInterface, mb *MockBankKeeper) {
				mk.On("GetBankKeeper").Return(mb)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				balances := sdk.NewCoins(
					sdk.NewInt64Coin("ujuno", 1000),
					sdk.NewInt64Coin("uatom", 500),
				)
				mb.On("GetAllBalances", mock.Anything, mock.Anything).Return(balances)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:    "valid address with no balances",
			address: validAddress,
			setupMocks: func(mk *MockKeeperInterface, mb *MockBankKeeper) {
				mk.On("GetBankKeeper").Return(mb)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				mb.On("GetAllBalances", mock.Anything, mock.Anything).Return(sdk.NewCoins())
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:               "invalid address format",
			address:            "invalid!address",
			setupMocks:         func(mk *MockKeeperInterface, mb *MockBankKeeper) {},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "invalid address",
			skipWebSocketCheck: true,
		},
		{
			name:               "empty address",
			address:            "",
			setupMocks:         func(mk *MockKeeperInterface, mb *MockBankKeeper) {},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "address cannot be empty",
			skipWebSocketCheck: true,
		},
		{
			name:    "valid address with single balance",
			address: validAddress,
			setupMocks: func(mk *MockKeeperInterface, mb *MockBankKeeper) {
				mk.On("GetBankKeeper").Return(mb)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				balances := sdk.NewCoins(sdk.NewInt64Coin("ujuno", 100000))
				mb.On("GetAllBalances", mock.Anything, mock.Anything).Return(balances)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:    "valid address with many balances",
			address: validAddress,
			setupMocks: func(mk *MockKeeperInterface, mb *MockBankKeeper) {
				mk.On("GetBankKeeper").Return(mb)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				balances := sdk.NewCoins(
					sdk.NewInt64Coin("ujuno", 1000),
					sdk.NewInt64Coin("uatom", 500),
					sdk.NewInt64Coin("uosmo", 250),
					sdk.NewInt64Coin("ustars", 100),
					sdk.NewInt64Coin("uion", 50),
				)
				mb.On("GetAllBalances", mock.Anything, mock.Anything).Return(balances)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:    "address with large balance amounts",
			address: validAddress,
			setupMocks: func(mk *MockKeeperInterface, mb *MockBankKeeper) {
				mk.On("GetBankKeeper").Return(mb)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				largeAmount := sdkmath.NewInt(1000000000000000000)
				balances := sdk.NewCoins(
					sdk.NewCoin("ujuno", largeAmount),
					sdk.NewCoin("uatom", largeAmount),
				)
				mb.On("GetAllBalances", mock.Anything, mock.Anything).Return(balances)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockKeeper := new(MockKeeperInterface)
			mockBankKeeper := new(MockBankKeeper)
			mockConnManager := new(MockConnectionManager)
			mockRegistry := new(MockSubscriptionRegistry)
			mockLogger := new(MockLogger)

			// Setup mocks
			tc.setupMocks(mockKeeper, mockBankKeeper)

			// Setup connection manager mocks
			mockConnManager.On("CheckConnectionLimits", mock.Anything, mock.Anything).Return(true)
			mockConnManager.On("RegisterConnectionWithHeaders", mock.Anything, mock.Anything).Return("conn-123")
			mockConnManager.On("UnregisterConnection", mock.Anything).Return()
			mockConnManager.On("AddSubscription", mock.Anything).Return(true)
			mockConnManager.On("RemoveSubscription", mock.Anything).Return()

			// Setup registry mocks
			mockRegistry.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(&MockSubscriber{})
			mockRegistry.On("Unsubscribe", mock.Anything).Return()

			// Setup logger mocks
			mockLogger.On("With", mock.Anything).Return(mockLogger)
			mockLogger.On("Info", mock.Anything, mock.Anything).Return()
			mockLogger.On("Error", mock.Anything, mock.Anything).Return()
			mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

			// Create handler dependencies
			deps := &common.HandlerDependencies{
				Config:         &common.StreamConfig{},
				Logger:         mockLogger,
				ConnManager:    mockConnManager,
				Registry:       mockRegistry,
				CircuitBreaker: nil,
				AppContext:     context.Background(),
			}

			// Create handler
			handler := bank.NewAllBalancesHandler(mockKeeper, deps)

			// Create test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				router := mux.NewRouter()
				router.HandleFunc("/bank/all_balances/{address}", handler.Handle)
				router.ServeHTTP(w, r)
			}))
			defer ts.Close()

			// Create request
			url := "ws" + ts.URL[4:] + "/bank/all_balances/" + tc.address

			if tc.skipWebSocketCheck {
				// For cases where we expect an HTTP error
				resp, err := http.Get("http" + ts.URL[4:] + "/bank/all_balances/" + tc.address)
				require.NoError(t, err)
				defer resp.Body.Close()

				require.Equal(t, tc.expectedStatus, resp.StatusCode)
				if tc.expectedError != "" {
					body := make([]byte, 1024)
					n, _ := resp.Body.Read(body)
					require.Contains(t, string(body[:n]), tc.expectedError)
				}
			} else {
				dialer := websocket.Dialer{
					HandshakeTimeout: 1 * time.Second,
				}

				conn, resp, err := dialer.Dial(url, nil)

				if tc.expectedStatus == http.StatusSwitchingProtocols {
					require.NoError(t, err)
					require.NotNil(t, conn)
					defer conn.Close()
					require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

					// Test that we can receive messages
					conn.SetReadDeadline(time.Now().Add(1 * time.Second))
					_, message, err := conn.ReadMessage()
					if err == nil {
						require.NotEmpty(t, message)
					}
				} else {
					require.Error(t, err)
					if resp != nil {
						require.Equal(t, tc.expectedStatus, resp.StatusCode)
					}
				}
			}

			// Verify mock expectations
			mockKeeper.AssertExpectations(t)
			mockBankKeeper.AssertExpectations(t)
		})
	}
}

func TestAllBalancesHandler_SubscriptionKey(t *testing.T) {
	validAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"

	// Create expected subscription key
	expectedKey := types.GenerateSubscriptionKey(types.SubscriptionTypeAllBalances, validAddress, "", "")

	// Create mocks
	mockKeeper := new(MockKeeperInterface)
	mockBankKeeper := new(MockBankKeeper)
	mockConnManager := new(MockConnectionManager)
	mockRegistry := new(MockSubscriptionRegistry)
	mockLogger := new(MockLogger)

	// Setup mocks
	mockKeeper.On("GetBankKeeper").Return(mockBankKeeper)
	mockKeeper.On("GetQueryContext").Return(context.Background(), nil)
	mockBankKeeper.On("GetAllBalances", mock.Anything, mock.Anything).Return(sdk.NewCoins())

	// Setup connection manager mocks
	mockConnManager.On("CheckConnectionLimits", mock.Anything, mock.Anything).Return(true)
	mockConnManager.On("RegisterConnectionWithHeaders", mock.Anything, mock.Anything).Return("conn-123")
	mockConnManager.On("UnregisterConnection", mock.Anything).Return()
	mockConnManager.On("AddSubscription", mock.Anything).Return(true)
	mockConnManager.On("RemoveSubscription", mock.Anything).Return()

	// Capture the subscription key used
	var capturedKey string
	mockRegistry.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		if key, ok := args.Get(0).(common.SubscriptionKey); ok {
			capturedKey = key.String()
		}
	}).Return(&MockSubscriber{})
	mockRegistry.On("Unsubscribe", mock.Anything).Return()

	// Setup logger mocks
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	// Create handler dependencies
	deps := &common.HandlerDependencies{
		Config:         &common.StreamConfig{},
		Logger:         mockLogger,
		ConnManager:    mockConnManager,
		Registry:       mockRegistry,
		CircuitBreaker: nil,
		AppContext:     context.Background(),
	}

	// Create handler
	handler := bank.NewAllBalancesHandler(mockKeeper, deps)

	// Create test request
	req := httptest.NewRequest("GET", "/bank/all_balances/"+validAddress, nil)
	req = mux.SetURLVars(req, map[string]string{
		"address": validAddress,
	})

	// Create response recorder
	w := httptest.NewRecorder()

	// Handle request
	handler.Handle(w, req)

	// Verify the subscription key matches expected format
	require.Equal(t, expectedKey, capturedKey)
}

func TestAllBalancesHandler_ResponseFormat(t *testing.T) {
	validAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"

	// Create mocks
	mockKeeper := new(MockKeeperInterface)
	mockBankKeeper := new(MockBankKeeper)
	mockConnManager := new(MockConnectionManager)
	mockRegistry := new(MockSubscriptionRegistry)
	mockLogger := new(MockLogger)

	// Setup expected balances
	expectedBalances := sdk.NewCoins(
		sdk.NewInt64Coin("ujuno", 1000),
		sdk.NewInt64Coin("uatom", 500),
		sdk.NewInt64Coin("uosmo", 250),
	)

	// Setup mocks
	mockKeeper.On("GetBankKeeper").Return(mockBankKeeper)
	mockKeeper.On("GetQueryContext").Return(context.Background(), nil)
	mockBankKeeper.On("GetAllBalances", mock.Anything, mock.Anything).Return(expectedBalances)

	// Setup connection manager mocks
	mockConnManager.On("CheckConnectionLimits", mock.Anything, mock.Anything).Return(true)
	mockConnManager.On("RegisterConnectionWithHeaders", mock.Anything, mock.Anything).Return("conn-123")
	mockConnManager.On("UnregisterConnection", mock.Anything).Return()
	mockConnManager.On("AddSubscription", mock.Anything).Return(true)
	mockConnManager.On("RemoveSubscription", mock.Anything).Return()

	// Setup registry mocks
	mockRegistry.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(&MockSubscriber{})
	mockRegistry.On("Unsubscribe", mock.Anything).Return()

	// Setup logger mocks
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	// Create handler dependencies
	deps := &common.HandlerDependencies{
		Config:         &common.StreamConfig{},
		Logger:         mockLogger,
		ConnManager:    mockConnManager,
		Registry:       mockRegistry,
		CircuitBreaker: nil,
		AppContext:     context.Background(),
		Upgrader:       common.GetUpgrader(true),
	}

	// Create handler
	handler := bank.NewAllBalancesHandler(mockKeeper, deps)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router := mux.NewRouter()
		router.HandleFunc("/bank/all_balances/{address}", handler.Handle)
		router.ServeHTTP(w, r)
	}))
	defer ts.Close()

	// Connect via WebSocket
	url := "ws" + ts.URL[4:] + "/bank/all_balances/" + validAddress
	dialer := websocket.Dialer{
		HandshakeTimeout: 1 * time.Second,
	}

	conn, _, err := dialer.Dial(url, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Read the initial message
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, message, err := conn.ReadMessage()
	if err == nil {
		// Verify the response format includes balances
		require.Contains(t, string(message), "balances")
	}

	// Verify mock expectations
	mockKeeper.AssertExpectations(t)
	mockBankKeeper.AssertExpectations(t)
}

func TestAllBalancesHandler_CircuitBreaker(t *testing.T) {
	validAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"

	// Create mocks
	mockKeeper := new(MockKeeperInterface)
	mockBankKeeper := new(MockBankKeeper)
	mockConnManager := new(MockConnectionManager)
	mockRegistry := new(MockSubscriptionRegistry)
	mockLogger := new(MockLogger)
	mockCircuitBreaker := new(MockCircuitBreaker)

	// Setup mocks
	mockKeeper.On("GetBankKeeper").Return(mockBankKeeper)

	// Circuit breaker is open
	mockCircuitBreaker.On("AllowRequest", mock.Anything).Return(false, nil)

	// Setup connection manager mocks
	mockConnManager.On("CheckConnectionLimits", mock.Anything, mock.Anything).Return(true)
	mockConnManager.On("RegisterConnectionWithHeaders", mock.Anything, mock.Anything).Return("conn-123")
	mockConnManager.On("UnregisterConnection", mock.Anything).Return()

	// Setup logger mocks
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	// Create handler dependencies with circuit breaker
	deps := &common.HandlerDependencies{
		Config:         &common.StreamConfig{CircuitBreakerEnabled: true},
		Logger:         mockLogger,
		ConnManager:    mockConnManager,
		Registry:       mockRegistry,
		CircuitBreaker: mockCircuitBreaker,
		AppContext:     context.Background(),
		Upgrader:       common.GetUpgrader(true),
	}

	// Create handler
	handler := bank.NewAllBalancesHandler(mockKeeper, deps)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router := mux.NewRouter()
		router.HandleFunc("/bank/all_balances/{address}", handler.Handle)
		router.ServeHTTP(w, r)
	}))
	defer ts.Close()

	// Try to connect
	url := "ws" + ts.URL[4:] + "/bank/all_balances/" + validAddress
	dialer := websocket.Dialer{
		HandshakeTimeout: 1 * time.Second,
	}

	conn, _, err := dialer.Dial(url, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Should receive an error message about circuit breaker
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, message, err := conn.ReadMessage()
	if err == nil {
		// Circuit breaker should send an error
		require.Contains(t, string(message), "error")
	}
}
