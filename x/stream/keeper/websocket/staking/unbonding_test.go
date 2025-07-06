package staking_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/common"
	"github.com/CosmosContracts/juno/v30/x/stream/keeper/websocket/staking"
	"github.com/CosmosContracts/juno/v30/x/stream/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUnbondingHandler_HandleUnbondingDelegations(t *testing.T) {
	validDelegatorAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	validValidatorAddress := "junovaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqdl2twf"
	
	// Create test unbonding delegation
	unbondingDelegation := stakingtypes.UnbondingDelegation{
		DelegatorAddress: validDelegatorAddress,
		ValidatorAddress: validValidatorAddress,
		Entries: []stakingtypes.UnbondingDelegationEntry{
			{
				CreationHeight: 1000,
				CompletionTime: time.Now().Add(21 * 24 * time.Hour),
				InitialBalance: sdkmath.NewInt(1000),
				Balance:        sdkmath.NewInt(1000),
			},
		},
	}
	
	tests := []struct {
		name               string
		delegatorAddress   string
		setupMocks         func(*MockKeeperInterface, *MockStakingKeeper)
		expectedStatus     int
		expectedError      string
		skipWebSocketCheck bool
	}{
		{
			name:             "valid address with unbonding delegations",
			delegatorAddress: validDelegatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetAllUnbondingDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.UnbondingDelegation{unbondingDelegation}, nil)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:             "valid address with no unbonding delegations",
			delegatorAddress: validDelegatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetAllUnbondingDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.UnbondingDelegation{}, nil)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:               "invalid address format",
			delegatorAddress:   "invalid!address",
			setupMocks:         func(mk *MockKeeperInterface, ms *MockStakingKeeper) {},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "invalid address",
			skipWebSocketCheck: true,
		},
		{
			name:               "empty address",
			delegatorAddress:   "",
			setupMocks:         func(mk *MockKeeperInterface, ms *MockStakingKeeper) {},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "address cannot be empty",
			skipWebSocketCheck: true,
		},
		{
			name:             "error getting unbonding delegations",
			delegatorAddress: validDelegatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetAllUnbondingDelegations", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:             "multiple unbonding delegations",
			delegatorAddress: validDelegatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				
				unbondingDelegation2 := stakingtypes.UnbondingDelegation{
					DelegatorAddress: validDelegatorAddress,
					ValidatorAddress: "junovaloper1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx123456",
					Entries: []stakingtypes.UnbondingDelegationEntry{
						{
							CreationHeight: 2000,
							CompletionTime: time.Now().Add(14 * 24 * time.Hour),
							InitialBalance: sdkmath.NewInt(500),
							Balance:        sdkmath.NewInt(500),
						},
					},
				}
				
				ms.On("GetAllUnbondingDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.UnbondingDelegation{unbondingDelegation, unbondingDelegation2}, nil)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:             "unbonding delegation with multiple entries",
			delegatorAddress: validDelegatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				
				multiEntryUnbonding := stakingtypes.UnbondingDelegation{
					DelegatorAddress: validDelegatorAddress,
					ValidatorAddress: validValidatorAddress,
					Entries: []stakingtypes.UnbondingDelegationEntry{
						{
							CreationHeight: 1000,
							CompletionTime: time.Now().Add(21 * 24 * time.Hour),
							InitialBalance: sdkmath.NewInt(1000),
							Balance:        sdkmath.NewInt(1000),
						},
						{
							CreationHeight: 1500,
							CompletionTime: time.Now().Add(14 * 24 * time.Hour),
							InitialBalance: sdkmath.NewInt(500),
							Balance:        sdkmath.NewInt(500),
						},
						{
							CreationHeight: 2000,
							CompletionTime: time.Now().Add(7 * 24 * time.Hour),
							InitialBalance: sdkmath.NewInt(250),
							Balance:        sdkmath.NewInt(250),
						},
					},
				}
				
				ms.On("GetAllUnbondingDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.UnbondingDelegation{multiEntryUnbonding}, nil)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockKeeper := new(MockKeeperInterface)
			mockStakingKeeper := new(MockStakingKeeper)
			mockConnManager := new(MockConnectionManager)
			mockRegistry := new(MockSubscriptionRegistry)
			mockLogger := new(MockLogger)

			// Setup mocks
			tc.setupMocks(mockKeeper, mockStakingKeeper)

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
			handler := staking.NewUnbondingHandler(mockKeeper, deps)

			// Create test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				router := mux.NewRouter()
				router.HandleFunc("/staking/unbonding_delegations/{delegator}", handler.HandleUnbondingDelegations)
				router.ServeHTTP(w, r)
			}))
			defer ts.Close()

			// Create request
			url := "ws" + ts.URL[4:] + "/staking/unbonding_delegations/" + tc.delegatorAddress
			
			if tc.skipWebSocketCheck {
				// For cases where we expect an HTTP error
				resp, err := http.Get("http" + ts.URL[4:] + "/staking/unbonding_delegations/" + tc.delegatorAddress)
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
			mockStakingKeeper.AssertExpectations(t)
		})
	}
}

func TestUnbondingHandler_HandleUnbondingDelegation(t *testing.T) {
	validDelegatorAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	validValidatorAddress := "junovaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqdl2twf"
	
	// Create test unbonding delegation
	unbondingDelegation := stakingtypes.UnbondingDelegation{
		DelegatorAddress: validDelegatorAddress,
		ValidatorAddress: validValidatorAddress,
		Entries: []stakingtypes.UnbondingDelegationEntry{
			{
				CreationHeight: 1000,
				CompletionTime: time.Now().Add(21 * 24 * time.Hour),
				InitialBalance: sdkmath.NewInt(1000),
				Balance:        sdkmath.NewInt(1000),
			},
		},
	}
	
	tests := []struct {
		name               string
		delegatorAddress   string
		validatorAddress   string
		setupMocks         func(*MockKeeperInterface, *MockStakingKeeper)
		expectedStatus     int
		expectedError      string
		skipWebSocketCheck bool
	}{
		{
			name:             "valid unbonding delegation found",
			delegatorAddress: validDelegatorAddress,
			validatorAddress: validValidatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetUnbondingDelegation", mock.Anything, mock.Anything, mock.Anything).Return(unbondingDelegation, nil)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:             "unbonding delegation not found",
			delegatorAddress: validDelegatorAddress,
			validatorAddress: validValidatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetUnbondingDelegation", mock.Anything, mock.Anything, mock.Anything).Return(stakingtypes.UnbondingDelegation{}, errors.New("not found"))
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:               "invalid delegator address",
			delegatorAddress:   "invalid!address",
			validatorAddress:   validValidatorAddress,
			setupMocks:         func(mk *MockKeeperInterface, ms *MockStakingKeeper) {},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "invalid address",
			skipWebSocketCheck: true,
		},
		{
			name:               "invalid validator address",
			delegatorAddress:   validDelegatorAddress,
			validatorAddress:   "invalid!validator",
			setupMocks:         func(mk *MockKeeperInterface, ms *MockStakingKeeper) {},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "invalid validator address",
			skipWebSocketCheck: true,
		},
		{
			name:               "both addresses invalid",
			delegatorAddress:   "invalid!address",
			validatorAddress:   "invalid!validator",
			setupMocks:         func(mk *MockKeeperInterface, ms *MockStakingKeeper) {},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "invalid",
			skipWebSocketCheck: true,
		},
		{
			name:               "empty delegator address",
			delegatorAddress:   "",
			validatorAddress:   validValidatorAddress,
			setupMocks:         func(mk *MockKeeperInterface, ms *MockStakingKeeper) {},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "address cannot be empty",
			skipWebSocketCheck: true,
		},
		{
			name:               "empty validator address",
			delegatorAddress:   validDelegatorAddress,
			validatorAddress:   "",
			setupMocks:         func(mk *MockKeeperInterface, ms *MockStakingKeeper) {},
			expectedStatus:     http.StatusBadRequest,
			expectedError:      "validator address cannot be empty",
			skipWebSocketCheck: true,
		},
		{
			name:             "unbonding delegation with zero balance",
			delegatorAddress: validDelegatorAddress,
			validatorAddress: validValidatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				
				zeroBalanceUnbonding := stakingtypes.UnbondingDelegation{
					DelegatorAddress: validDelegatorAddress,
					ValidatorAddress: validValidatorAddress,
					Entries: []stakingtypes.UnbondingDelegationEntry{
						{
							CreationHeight: 1000,
							CompletionTime: time.Now().Add(21 * 24 * time.Hour),
							InitialBalance: sdkmath.NewInt(0),
							Balance:        sdkmath.NewInt(0),
						},
					},
				}
				
				ms.On("GetUnbondingDelegation", mock.Anything, mock.Anything, mock.Anything).Return(zeroBalanceUnbonding, nil)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockKeeper := new(MockKeeperInterface)
			mockStakingKeeper := new(MockStakingKeeper)
			mockConnManager := new(MockConnectionManager)
			mockRegistry := new(MockSubscriptionRegistry)
			mockLogger := new(MockLogger)

			// Setup mocks
			tc.setupMocks(mockKeeper, mockStakingKeeper)

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
			handler := staking.NewUnbondingHandler(mockKeeper, deps)

			// Create test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				router := mux.NewRouter()
				router.HandleFunc("/staking/unbonding_delegation/{delegator}/{validator}", handler.HandleUnbondingDelegation)
				router.ServeHTTP(w, r)
			}))
			defer ts.Close()

			// Create request
			url := "ws" + ts.URL[4:] + "/staking/unbonding_delegation/" + tc.delegatorAddress + "/" + tc.validatorAddress
			
			if tc.skipWebSocketCheck {
				// For cases where we expect an HTTP error
				resp, err := http.Get("http" + ts.URL[4:] + "/staking/unbonding_delegation/" + tc.delegatorAddress + "/" + tc.validatorAddress)
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
			mockStakingKeeper.AssertExpectations(t)
		})
	}
}

func TestUnbondingHandler_SubscriptionKeys(t *testing.T) {
	validDelegatorAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	validValidatorAddress := "junovaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqdl2twf"
	
	t.Run("unbonding delegations subscription key", func(t *testing.T) {
		// Create expected subscription key
		expectedKey := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegations, validDelegatorAddress, "", "")
		
		// Create mocks
		mockKeeper := new(MockKeeperInterface)
		mockStakingKeeper := new(MockStakingKeeper)
		mockConnManager := new(MockConnectionManager)
		mockRegistry := new(MockSubscriptionRegistry)
		mockLogger := new(MockLogger)

		// Setup mocks
		mockKeeper.On("GetStakingKeeper").Return(mockStakingKeeper)
		mockKeeper.On("GetQueryContext").Return(context.Background(), nil)
		mockStakingKeeper.On("GetAllUnbondingDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.UnbondingDelegation{}, nil)

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
		handler := staking.NewUnbondingHandler(mockKeeper, deps)

		// Create test request
		req := httptest.NewRequest("GET", "/staking/unbonding_delegations/"+validDelegatorAddress, nil)
		req = mux.SetURLVars(req, map[string]string{
			"delegator": validDelegatorAddress,
		})
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Handle request
		handler.HandleUnbondingDelegations(w, req)
		
		// Verify the subscription key matches expected format
		require.Equal(t, expectedKey, capturedKey)
	})

	t.Run("unbonding delegation subscription key", func(t *testing.T) {
		// Create expected subscription key
		expectedKey := types.GenerateSubscriptionKey(types.SubscriptionTypeUnbondingDelegation, validDelegatorAddress, validValidatorAddress, "")
		
		// Create mocks
		mockKeeper := new(MockKeeperInterface)
		mockStakingKeeper := new(MockStakingKeeper)
		mockConnManager := new(MockConnectionManager)
		mockRegistry := new(MockSubscriptionRegistry)
		mockLogger := new(MockLogger)

		// Setup mocks
		mockKeeper.On("GetStakingKeeper").Return(mockStakingKeeper)
		mockKeeper.On("GetQueryContext").Return(context.Background(), nil)
		mockStakingKeeper.On("GetUnbondingDelegation", mock.Anything, mock.Anything, mock.Anything).Return(stakingtypes.UnbondingDelegation{}, errors.New("not found"))

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
		handler := staking.NewUnbondingHandler(mockKeeper, deps)

		// Create test request
		req := httptest.NewRequest("GET", "/staking/unbonding_delegation/"+validDelegatorAddress+"/"+validValidatorAddress, nil)
		req = mux.SetURLVars(req, map[string]string{
			"delegator": validDelegatorAddress,
			"validator": validValidatorAddress,
		})
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Handle request
		handler.HandleUnbondingDelegation(w, req)
		
		// Verify the subscription key matches expected format
		require.Equal(t, expectedKey, capturedKey)
	})
}

func TestUnbondingHandler_EdgeCases(t *testing.T) {
	validDelegatorAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	validValidatorAddress := "junovaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqdl2twf"
	
	t.Run("unbonding delegation with completed entries", func(t *testing.T) {
		// Create unbonding delegation with mixed completion times
		unbondingWithMixedEntries := stakingtypes.UnbondingDelegation{
			DelegatorAddress: validDelegatorAddress,
			ValidatorAddress: validValidatorAddress,
			Entries: []stakingtypes.UnbondingDelegationEntry{
				{
					CreationHeight: 500,
					CompletionTime: time.Now().Add(-24 * time.Hour), // Already completed
					InitialBalance: sdkmath.NewInt(100),
					Balance:        sdkmath.NewInt(100),
				},
				{
					CreationHeight: 1000,
					CompletionTime: time.Now().Add(21 * 24 * time.Hour), // Still unbonding
					InitialBalance: sdkmath.NewInt(1000),
					Balance:        sdkmath.NewInt(1000),
				},
			},
		}
		
		// Create mocks
		mockKeeper := new(MockKeeperInterface)
		mockStakingKeeper := new(MockStakingKeeper)
		mockConnManager := new(MockConnectionManager)
		mockRegistry := new(MockSubscriptionRegistry)
		mockLogger := new(MockLogger)

		// Setup mocks
		mockKeeper.On("GetStakingKeeper").Return(mockStakingKeeper)
		mockKeeper.On("GetQueryContext").Return(context.Background(), nil)
		mockStakingKeeper.On("GetUnbondingDelegation", mock.Anything, mock.Anything, mock.Anything).Return(unbondingWithMixedEntries, nil)

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
		handler := staking.NewUnbondingHandler(mockKeeper, deps)

		// Create test server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			router := mux.NewRouter()
			router.HandleFunc("/staking/unbonding_delegation/{delegator}/{validator}", handler.HandleUnbondingDelegation)
			router.ServeHTTP(w, r)
		}))
		defer ts.Close()

		// Connect via WebSocket
		url := "ws" + ts.URL[4:] + "/staking/unbonding_delegation/" + validDelegatorAddress + "/" + validValidatorAddress
		dialer := websocket.Dialer{
			HandshakeTimeout: 1 * time.Second,
		}
		
		conn, _, err := dialer.Dial(url, nil)
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer conn.Close()

		// Should handle gracefully
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, message, err := conn.ReadMessage()
		if err == nil {
			require.NotEmpty(t, message)
			// Should contain the unbonding delegation data
			require.Contains(t, string(message), "unbonding_delegation")
		}
	})

	t.Run("empty unbonding delegation entries", func(t *testing.T) {
		// Create unbonding delegation with no entries
		emptyEntriesUnbonding := stakingtypes.UnbondingDelegation{
			DelegatorAddress: validDelegatorAddress,
			ValidatorAddress: validValidatorAddress,
			Entries:          []stakingtypes.UnbondingDelegationEntry{}, // Empty entries
		}
		
		// Create mocks
		mockKeeper := new(MockKeeperInterface)
		mockStakingKeeper := new(MockStakingKeeper)
		mockConnManager := new(MockConnectionManager)
		mockRegistry := new(MockSubscriptionRegistry)
		mockLogger := new(MockLogger)

		// Setup mocks
		mockKeeper.On("GetStakingKeeper").Return(mockStakingKeeper)
		mockKeeper.On("GetQueryContext").Return(context.Background(), nil)
		mockStakingKeeper.On("GetAllUnbondingDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.UnbondingDelegation{emptyEntriesUnbonding}, nil)

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
		handler := staking.NewUnbondingHandler(mockKeeper, deps)

		// Create test server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			router := mux.NewRouter()
			router.HandleFunc("/staking/unbonding_delegations/{delegator}", handler.HandleUnbondingDelegations)
			router.ServeHTTP(w, r)
		}))
		defer ts.Close()

		// Connect via WebSocket
		url := "ws" + ts.URL[4:] + "/staking/unbonding_delegations/" + validDelegatorAddress
		dialer := websocket.Dialer{
			HandshakeTimeout: 1 * time.Second,
		}
		
		conn, _, err := dialer.Dial(url, nil)
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer conn.Close()

		// Should handle gracefully
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, message, err := conn.ReadMessage()
		if err == nil {
			require.NotEmpty(t, message)
		}
	})
}