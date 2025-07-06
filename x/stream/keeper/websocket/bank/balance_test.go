package bank_test

import (
	"context"
	"errors"
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

func init() {
	// Set the bech32 prefixes for testing
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("juno", "junopub")
	config.SetBech32PrefixForValidator("junovaloper", "junovaloperpub")
	config.SetBech32PrefixForConsensusNode("junovalcons", "junovalconspub")
}

// MockKeeperInterface is a mock implementation of KeeperInterface
type MockKeeperInterface struct {
	mock.Mock
}

func (m *MockKeeperInterface) GetQueryContext() (context.Context, error) {
	args := m.Called()
	return args.Get(0).(context.Context), args.Error(1)
}

func (m *MockKeeperInterface) ValidateDenom(ctx context.Context, denom string) error {
	args := m.Called(ctx, denom)
	return args.Error(0)
}

func (m *MockKeeperInterface) GetBankKeeper() bank.BankKeeperInterface {
	args := m.Called()
	return args.Get(0).(bank.BankKeeperInterface)
}

// MockBankKeeper is a mock implementation of BankKeeperInterface
type MockBankKeeper struct {
	mock.Mock
}

func (m *MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	args := m.Called(ctx, addr, denom)
	return args.Get(0).(sdk.Coin)
}

func (m *MockBankKeeper) GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	args := m.Called(ctx, addr)
	return args.Get(0).(sdk.Coins)
}

// MockConnectionManager for testing
type MockConnectionManager struct {
	mock.Mock
}

func (m *MockConnectionManager) CheckConnectionLimits(w http.ResponseWriter, r *http.Request) bool {
	args := m.Called(w, r)
	return args.Bool(0)
}

func (m *MockConnectionManager) RegisterConnectionWithHeaders(remoteAddr, xForwardedFor string) string {
	args := m.Called(remoteAddr, xForwardedFor)
	return args.String(0)
}

func (m *MockConnectionManager) UnregisterConnection(connectionID string) {
	m.Called(connectionID)
}

func (m *MockConnectionManager) AddSubscription(connectionID string) bool {
	args := m.Called(connectionID)
	return args.Bool(0)
}

func (m *MockConnectionManager) RemoveSubscription(connectionID string) {
	m.Called(connectionID)
}

// MockSubscriptionRegistry for testing
type MockSubscriptionRegistry struct {
	mock.Mock
}

func (m *MockSubscriptionRegistry) Subscribe(key common.SubscriptionKey, ctx context.Context, sendCh chan<- any) common.Subscriber {
	args := m.Called(key, ctx, sendCh)
	if sub := args.Get(0); sub != nil {
		return sub.(common.Subscriber)
	}
	return nil
}

func (m *MockSubscriptionRegistry) Unsubscribe(subscriber common.Subscriber) {
	m.Called(subscriber)
}

// MockSubscriber for testing
type MockSubscriber struct{}

// MockLogger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, keyvals ...any) {
	m.Called(msg, keyvals)
}

func (m *MockLogger) Error(msg string, keyvals ...any) {
	m.Called(msg, keyvals)
}

func (m *MockLogger) Debug(msg string, keyvals ...any) {
	m.Called(msg, keyvals)
}

func (m *MockLogger) Warn(msg string, keyvals ...any) {
	m.Called(msg, keyvals)
}

func (m *MockLogger) With(keyvals ...any) common.Logger {
	args := m.Called(keyvals)
	return args.Get(0).(common.Logger)
}

// MockCircuitBreaker for testing
type MockCircuitBreaker struct {
	mock.Mock
}

func (m *MockCircuitBreaker) AllowRequest(connectionID string) (bool, error) {
	args := m.Called(connectionID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCircuitBreaker) RecordSuccess(connectionID string) {
	m.Called(connectionID)
}

func (m *MockCircuitBreaker) RecordFailure(connectionID string) {
	m.Called(connectionID)
}

func TestBalanceHandler_Handle(t *testing.T) {
	validAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	
	tests := []struct {
		name               string
		address            string
		denom              string
		setupMocks         func(*MockKeeperInterface, *MockBankKeeper)
		expectedStatus     int
		expectedError      string
		skipWebSocketCheck bool
	}{
		{
			name:    "valid address and denom",
			address: validAddress,
			denom:   "ujuno",
			setupMocks: func(mk *MockKeeperInterface, mb *MockBankKeeper) {
				mk.On("ValidateDenom", mock.Anything, "ujuno").Return(nil)
				mk.On("GetBankKeeper").Return(mb)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				mb.On("GetBalance", mock.Anything, mock.Anything, "ujuno").Return(sdk.NewInt64Coin("ujuno", 1000))
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:               "invalid address format",
			address:            "invalid!address",
			denom:              "ujuno",
			setupMocks:         func(mk *MockKeeperInterface, mb *MockBankKeeper) {},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "invalid address",
			skipWebSocketCheck: true,
		},
		{
			name:               "empty address",
			address:            "",
			denom:              "ujuno",
			setupMocks:         func(mk *MockKeeperInterface, mb *MockBankKeeper) {},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "address cannot be empty",
			skipWebSocketCheck: true,
		},
		{
			name:    "invalid denom",
			address: validAddress,
			denom:   "invalid$denom",
			setupMocks: func(mk *MockKeeperInterface, mb *MockBankKeeper) {
				mk.On("ValidateDenom", mock.Anything, "invalid$denom").Return(errors.New("invalid denom format"))
			},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "invalid denom",
			skipWebSocketCheck: true,
		},
		{
			name:    "empty denom",
			address: validAddress,
			denom:   "",
			setupMocks: func(mk *MockKeeperInterface, mb *MockBankKeeper) {
				mk.On("ValidateDenom", mock.Anything, "").Return(errors.New("denom cannot be empty"))
			},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "invalid denom",
			skipWebSocketCheck: true,
		},
		{
			name:    "query context error",
			address: validAddress,
			denom:   "ujuno",
			setupMocks: func(mk *MockKeeperInterface, mb *MockBankKeeper) {
				mk.On("ValidateDenom", mock.Anything, "ujuno").Return(nil)
				mk.On("GetBankKeeper").Return(mb)
				mk.On("GetQueryContext").Return(context.Background(), errors.New("context error"))
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:    "zero balance",
			address: validAddress,
			denom:   "uatom",
			setupMocks: func(mk *MockKeeperInterface, mb *MockBankKeeper) {
				mk.On("ValidateDenom", mock.Anything, "uatom").Return(nil)
				mk.On("GetBankKeeper").Return(mb)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				mb.On("GetBalance", mock.Anything, mock.Anything, "uatom").Return(sdk.NewInt64Coin("uatom", 0))
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:    "large balance",
			address: validAddress,
			denom:   "ujuno",
			setupMocks: func(mk *MockKeeperInterface, mb *MockBankKeeper) {
				mk.On("ValidateDenom", mock.Anything, "ujuno").Return(nil)
				mk.On("GetBankKeeper").Return(mb)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				// Large balance
				largeAmount := sdkmath.NewInt(1000000000000000000)
				mb.On("GetBalance", mock.Anything, mock.Anything, "ujuno").Return(sdk.NewCoin("ujuno", largeAmount))
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
				CircuitBreaker: nil, // Not testing circuit breaker in this test
				AppContext:     context.Background(),
			}

			// Create handler
			handler := bank.NewBalanceHandler(mockKeeper, deps)

			// Create test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Set up router to parse path variables
				router := mux.NewRouter()
				router.HandleFunc("/bank/balance/{address}/{denom}", handler.Handle)
				router.ServeHTTP(w, r)
			}))
			defer ts.Close()

			// Create request
			url := "ws" + ts.URL[4:] + "/bank/balance/" + tc.address + "/" + tc.denom
			
			if tc.skipWebSocketCheck {
				// For cases where we expect an HTTP error, use regular HTTP client
				resp, err := http.Get("http" + ts.URL[4:] + "/bank/balance/" + tc.address + "/" + tc.denom)
				require.NoError(t, err)
				defer resp.Body.Close()
				
				require.Equal(t, tc.expectedStatus, resp.StatusCode)
				if tc.expectedError != "" {
					body := make([]byte, 1024)
					n, _ := resp.Body.Read(body)
					require.Contains(t, string(body[:n]), tc.expectedError)
				}
			} else {
				// For WebSocket connections
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
						// Successfully read a message
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

func TestBalanceHandler_ConcurrentConnections(t *testing.T) {
	validAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	
	// Create mocks
	mockKeeper := new(MockKeeperInterface)
	mockBankKeeper := new(MockBankKeeper)
	mockConnManager := new(MockConnectionManager)
	mockRegistry := new(MockSubscriptionRegistry)
	mockLogger := new(MockLogger)

	// Setup mocks
	mockKeeper.On("ValidateDenom", mock.Anything, "ujuno").Return(nil)
	mockKeeper.On("GetBankKeeper").Return(mockBankKeeper)
	mockKeeper.On("GetQueryContext").Return(context.Background(), nil)
	mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, "ujuno").Return(sdk.NewInt64Coin("ujuno", 1000))

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
	handler := bank.NewBalanceHandler(mockKeeper, deps)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router := mux.NewRouter()
		router.HandleFunc("/bank/balance/{address}/{denom}", handler.Handle)
		router.ServeHTTP(w, r)
	}))
	defer ts.Close()

	// Test concurrent connections
	numConnections := 5
	done := make(chan bool, numConnections)
	
	for i := 0; i < numConnections; i++ {
		go func(id int) {
			url := "ws" + ts.URL[4:] + "/bank/balance/" + validAddress + "/ujuno"
			dialer := websocket.Dialer{
				HandshakeTimeout: 1 * time.Second,
			}
			
			conn, _, err := dialer.Dial(url, nil)
			if err == nil && conn != nil {
				conn.Close()
			}
			done <- (err == nil)
		}(i)
	}

	// Wait for all connections
	successCount := 0
	for i := 0; i < numConnections; i++ {
		if <-done {
			successCount++
		}
	}

	// At least some connections should succeed
	require.Greater(t, successCount, 0)
}

func TestBalanceHandler_SubscriptionKey(t *testing.T) {
	validAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	denom := "ujuno"
	
	// Create expected subscription key
	expectedKey := types.GenerateSubscriptionKey(types.SubscriptionTypeBalance, validAddress, "", denom)
	
	// Create mocks
	mockKeeper := new(MockKeeperInterface)
	mockBankKeeper := new(MockBankKeeper)
	mockConnManager := new(MockConnectionManager)
	mockRegistry := new(MockSubscriptionRegistry)
	mockLogger := new(MockLogger)

	// Setup mocks
	mockKeeper.On("ValidateDenom", mock.Anything, denom).Return(nil)
	mockKeeper.On("GetBankKeeper").Return(mockBankKeeper)
	mockKeeper.On("GetQueryContext").Return(context.Background(), nil)
	mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, denom).Return(sdk.NewInt64Coin(denom, 1000))

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
	handler := bank.NewBalanceHandler(mockKeeper, deps)

	// Create test request
	req := httptest.NewRequest("GET", "/bank/balance/"+validAddress+"/"+denom, nil)
	req = mux.SetURLVars(req, map[string]string{
		"address": validAddress,
		"denom":   denom,
	})
	
	// Create response recorder
	w := httptest.NewRecorder()
	
	// Handle request
	handler.Handle(w, req)
	
	// Verify the subscription key matches expected format
	require.Equal(t, expectedKey, capturedKey)
}

func TestBalanceHandler_ConnectionLimit(t *testing.T) {
	validAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	
	// Create mocks
	mockKeeper := new(MockKeeperInterface)
	mockBankKeeper := new(MockBankKeeper)
	mockConnManager := new(MockConnectionManager)
	mockRegistry := new(MockSubscriptionRegistry)
	mockLogger := new(MockLogger)

	// Setup mocks
	mockKeeper.On("ValidateDenom", mock.Anything, "ujuno").Return(nil)
	mockKeeper.On("GetBankKeeper").Return(mockBankKeeper)
	
	// Connection limit reached
	mockConnManager.On("CheckConnectionLimits", mock.Anything, mock.Anything).Return(false)
	
	// Setup logger mocks
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

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
	handler := bank.NewBalanceHandler(mockKeeper, deps)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router := mux.NewRouter()
		router.HandleFunc("/bank/balance/{address}/{denom}", handler.Handle)
		router.ServeHTTP(w, r)
	}))
	defer ts.Close()

	// Try to connect
	url := "ws" + ts.URL[4:] + "/bank/balance/" + validAddress + "/ujuno"
	dialer := websocket.Dialer{
		HandshakeTimeout: 1 * time.Second,
	}
	
	_, resp, err := dialer.Dial(url, nil)
	require.Error(t, err)
	require.NotNil(t, resp)
	// When connection limit is reached, we expect a non-switching protocols status
	require.NotEqual(t, http.StatusSwitchingProtocols, resp.StatusCode)
}