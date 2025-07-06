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
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

func (m *MockKeeperInterface) GetStakingKeeper() staking.StakingKeeperInterface {
	args := m.Called()
	return args.Get(0).(staking.StakingKeeperInterface)
}

// MockStakingKeeper is a mock implementation of StakingKeeperInterface
type MockStakingKeeper struct {
	mock.Mock
}

func (m *MockStakingKeeper) GetAllDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.Delegation, error) {
	args := m.Called(ctx, delegator)
	if dels := args.Get(0); dels != nil {
		return dels.([]stakingtypes.Delegation), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStakingKeeper) GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, error) {
	args := m.Called(ctx, delAddr, valAddr)
	if del := args.Get(0); del != nil {
		return del.(stakingtypes.Delegation), args.Error(1)
	}
	return stakingtypes.Delegation{}, args.Error(1)
}

func (m *MockStakingKeeper) GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error) {
	args := m.Called(ctx, addr)
	if val := args.Get(0); val != nil {
		return val.(stakingtypes.Validator), args.Error(1)
	}
	return stakingtypes.Validator{}, args.Error(1)
}

func (m *MockStakingKeeper) BondDenom(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockStakingKeeper) GetAllUnbondingDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.UnbondingDelegation, error) {
	args := m.Called(ctx, delegator)
	if ubds := args.Get(0); ubds != nil {
		return ubds.([]stakingtypes.UnbondingDelegation), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStakingKeeper) GetUnbondingDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.UnbondingDelegation, error) {
	args := m.Called(ctx, delAddr, valAddr)
	if ubd := args.Get(0); ubd != nil {
		return ubd.(stakingtypes.UnbondingDelegation), args.Error(1)
	}
	return stakingtypes.UnbondingDelegation{}, args.Error(1)
}

// Mocks from previous test files
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

type MockSubscriber struct{}

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

func TestDelegationsHandler_Handle(t *testing.T) {
	validDelegatorAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	validValidatorAddress := "junovaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqdl2twf"
	
	// Create test validator and delegation
	validator := stakingtypes.Validator{
		OperatorAddress: validValidatorAddress,
		Status:          stakingtypes.Bonded,
		Tokens:          sdkmath.NewInt(1000000),
		DelegatorShares: sdkmath.LegacyNewDec(1000000),
	}
	
	delegation := stakingtypes.Delegation{
		DelegatorAddress: validDelegatorAddress,
		ValidatorAddress: validValidatorAddress,
		Shares:           sdkmath.LegacyNewDec(1000),
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
			name:             "valid address with delegations",
			delegatorAddress: validDelegatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetAllDelegatorDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.Delegation{delegation}, nil)
				ms.On("BondDenom", mock.Anything).Return("ujuno", nil)
				ms.On("GetValidator", mock.Anything, mock.Anything).Return(validator, nil)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:             "valid address with no delegations",
			delegatorAddress: validDelegatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetAllDelegatorDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.Delegation{}, nil)
				ms.On("BondDenom", mock.Anything).Return("ujuno", nil)
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
			name:             "error getting delegations",
			delegatorAddress: validDelegatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetAllDelegatorDelegations", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
				ms.On("BondDenom", mock.Anything).Return("ujuno", nil)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:             "error getting bond denom",
			delegatorAddress: validDelegatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetAllDelegatorDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.Delegation{delegation}, nil)
				ms.On("BondDenom", mock.Anything).Return("", errors.New("bond denom error"))
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:             "delegation with non-existent validator",
			delegatorAddress: validDelegatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetAllDelegatorDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.Delegation{delegation}, nil)
				ms.On("BondDenom", mock.Anything).Return("ujuno", nil)
				ms.On("GetValidator", mock.Anything, mock.Anything).Return(stakingtypes.Validator{}, errors.New("validator not found"))
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:             "multiple delegations",
			delegatorAddress: validDelegatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				
				delegation2 := stakingtypes.Delegation{
					DelegatorAddress: validDelegatorAddress,
					ValidatorAddress: "junovaloper1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx123456",
					Shares:           sdkmath.LegacyNewDec(500),
				}
				
				validator2 := stakingtypes.Validator{
					OperatorAddress: "junovaloper1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx123456",
					Status:          stakingtypes.Bonded,
					Tokens:          sdkmath.NewInt(500000),
					DelegatorShares: sdkmath.LegacyNewDec(500000),
				}
				
				ms.On("GetAllDelegatorDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.Delegation{delegation, delegation2}, nil)
				ms.On("BondDenom", mock.Anything).Return("ujuno", nil)
				ms.On("GetValidator", mock.Anything, mock.MatchedBy(func(addr sdk.ValAddress) bool {
					return addr.String() == validValidatorAddress
				})).Return(validator, nil)
				ms.On("GetValidator", mock.Anything, mock.MatchedBy(func(addr sdk.ValAddress) bool {
					return addr.String() == "junovaloper1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx123456"
				})).Return(validator2, nil)
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
			handler := staking.NewDelegationsHandler(mockKeeper, deps)

			// Create test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				router := mux.NewRouter()
				router.HandleFunc("/staking/delegations/{delegator}", handler.Handle)
				router.ServeHTTP(w, r)
			}))
			defer ts.Close()

			// Create request
			url := "ws" + ts.URL[4:] + "/staking/delegations/" + tc.delegatorAddress
			
			if tc.skipWebSocketCheck {
				// For cases where we expect an HTTP error
				resp, err := http.Get("http" + ts.URL[4:] + "/staking/delegations/" + tc.delegatorAddress)
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

func TestDelegationsHandler_HandleDelegation(t *testing.T) {
	validDelegatorAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	validValidatorAddress := "junovaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqdl2twf"
	
	// Create test validator and delegation
	validator := stakingtypes.Validator{
		OperatorAddress: validValidatorAddress,
		Status:          stakingtypes.Bonded,
		Tokens:          sdkmath.NewInt(1000000),
		DelegatorShares: sdkmath.LegacyNewDec(1000000),
	}
	
	delegation := stakingtypes.Delegation{
		DelegatorAddress: validDelegatorAddress,
		ValidatorAddress: validValidatorAddress,
		Shares:           sdkmath.LegacyNewDec(1000),
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
			name:             "valid delegation found",
			delegatorAddress: validDelegatorAddress,
			validatorAddress: validValidatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetDelegation", mock.Anything, mock.Anything, mock.Anything).Return(delegation, nil)
				ms.On("BondDenom", mock.Anything).Return("ujuno", nil)
				ms.On("GetValidator", mock.Anything, mock.Anything).Return(validator, nil)
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:             "delegation not found",
			delegatorAddress: validDelegatorAddress,
			validatorAddress: validValidatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetDelegation", mock.Anything, mock.Anything, mock.Anything).Return(stakingtypes.Delegation{}, errors.New("not found"))
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
			name:             "delegation found but validator not found",
			delegatorAddress: validDelegatorAddress,
			validatorAddress: validValidatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetDelegation", mock.Anything, mock.Anything, mock.Anything).Return(delegation, nil)
				ms.On("BondDenom", mock.Anything).Return("ujuno", nil)
				ms.On("GetValidator", mock.Anything, mock.Anything).Return(stakingtypes.Validator{}, errors.New("validator not found"))
			},
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:             "delegation found but bond denom error",
			delegatorAddress: validDelegatorAddress,
			validatorAddress: validValidatorAddress,
			setupMocks: func(mk *MockKeeperInterface, ms *MockStakingKeeper) {
				mk.On("GetStakingKeeper").Return(ms)
				mk.On("GetQueryContext").Return(context.Background(), nil)
				ms.On("GetDelegation", mock.Anything, mock.Anything, mock.Anything).Return(delegation, nil)
				ms.On("BondDenom", mock.Anything).Return("", errors.New("bond denom error"))
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
			handler := staking.NewDelegationsHandler(mockKeeper, deps)

			// Create test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				router := mux.NewRouter()
				router.HandleFunc("/staking/delegation/{delegator}/{validator}", handler.HandleDelegation)
				router.ServeHTTP(w, r)
			}))
			defer ts.Close()

			// Create request
			url := "ws" + ts.URL[4:] + "/staking/delegation/" + tc.delegatorAddress + "/" + tc.validatorAddress
			
			if tc.skipWebSocketCheck {
				// For cases where we expect an HTTP error
				resp, err := http.Get("http" + ts.URL[4:] + "/staking/delegation/" + tc.delegatorAddress + "/" + tc.validatorAddress)
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

func TestDelegationsHandler_SubscriptionKeys(t *testing.T) {
	validDelegatorAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	validValidatorAddress := "junovaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqdl2twf"
	
	t.Run("delegations subscription key", func(t *testing.T) {
		// Create expected subscription key
		expectedKey := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegations, validDelegatorAddress, "", "")
		
		// Create mocks
		mockKeeper := new(MockKeeperInterface)
		mockStakingKeeper := new(MockStakingKeeper)
		mockConnManager := new(MockConnectionManager)
		mockRegistry := new(MockSubscriptionRegistry)
		mockLogger := new(MockLogger)

		// Setup mocks
		mockKeeper.On("GetStakingKeeper").Return(mockStakingKeeper)
		mockKeeper.On("GetQueryContext").Return(context.Background(), nil)
		mockStakingKeeper.On("GetAllDelegatorDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.Delegation{}, nil)
		mockStakingKeeper.On("BondDenom", mock.Anything).Return("ujuno", nil)

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
		handler := staking.NewDelegationsHandler(mockKeeper, deps)

		// Create test request
		req := httptest.NewRequest("GET", "/staking/delegations/"+validDelegatorAddress, nil)
		req = mux.SetURLVars(req, map[string]string{
			"delegator": validDelegatorAddress,
		})
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Handle request
		handler.Handle(w, req)
		
		// Verify the subscription key matches expected format
		require.Equal(t, expectedKey, capturedKey)
	})

	t.Run("delegation subscription key", func(t *testing.T) {
		// Create expected subscription key
		expectedKey := types.GenerateSubscriptionKey(types.SubscriptionTypeDelegation, validDelegatorAddress, validValidatorAddress, "")
		
		// Create mocks
		mockKeeper := new(MockKeeperInterface)
		mockStakingKeeper := new(MockStakingKeeper)
		mockConnManager := new(MockConnectionManager)
		mockRegistry := new(MockSubscriptionRegistry)
		mockLogger := new(MockLogger)

		// Setup mocks
		mockKeeper.On("GetStakingKeeper").Return(mockStakingKeeper)
		mockKeeper.On("GetQueryContext").Return(context.Background(), nil)
		mockStakingKeeper.On("GetDelegation", mock.Anything, mock.Anything, mock.Anything).Return(stakingtypes.Delegation{}, errors.New("not found"))

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
		handler := staking.NewDelegationsHandler(mockKeeper, deps)

		// Create test request
		req := httptest.NewRequest("GET", "/staking/delegation/"+validDelegatorAddress+"/"+validValidatorAddress, nil)
		req = mux.SetURLVars(req, map[string]string{
			"delegator": validDelegatorAddress,
			"validator": validValidatorAddress,
		})
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Handle request
		handler.HandleDelegation(w, req)
		
		// Verify the subscription key matches expected format
		require.Equal(t, expectedKey, capturedKey)
	})
}

func TestDelegationsHandler_EdgeCases(t *testing.T) {
	validDelegatorAddress := "juno1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq89mgve"
	
	t.Run("delegation with invalid validator address in response", func(t *testing.T) {
		// Create delegation with invalid validator address
		delegationWithInvalidValidator := stakingtypes.Delegation{
			DelegatorAddress: validDelegatorAddress,
			ValidatorAddress: "invalid!validator!address",
			Shares:           sdkmath.LegacyNewDec(1000),
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
		mockStakingKeeper.On("GetAllDelegatorDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.Delegation{delegationWithInvalidValidator}, nil)
		mockStakingKeeper.On("BondDenom", mock.Anything).Return("ujuno", nil)

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
		handler := staking.NewDelegationsHandler(mockKeeper, deps)

		// Create test server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			router := mux.NewRouter()
			router.HandleFunc("/staking/delegations/{delegator}", handler.Handle)
			router.ServeHTTP(w, r)
		}))
		defer ts.Close()

		// Connect via WebSocket
		url := "ws" + ts.URL[4:] + "/staking/delegations/" + validDelegatorAddress
		dialer := websocket.Dialer{
			HandshakeTimeout: 1 * time.Second,
		}
		
		conn, _, err := dialer.Dial(url, nil)
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer conn.Close()

		// Should handle gracefully and return empty response
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, message, err := conn.ReadMessage()
		if err == nil {
			// Should receive a response with empty array
			require.NotEmpty(t, message)
		}
	})

	t.Run("validator with zero tokens", func(t *testing.T) {
		validValidatorAddress := "junovaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqdl2twf"
		
		// Create validator with zero tokens
		zeroTokenValidator := stakingtypes.Validator{
			OperatorAddress: validValidatorAddress,
			Status:          stakingtypes.Bonded,
			Tokens:          sdkmath.NewInt(0),
			DelegatorShares: sdkmath.LegacyNewDec(1000),
		}
		
		delegation := stakingtypes.Delegation{
			DelegatorAddress: validDelegatorAddress,
			ValidatorAddress: validValidatorAddress,
			Shares:           sdkmath.LegacyNewDec(1000),
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
		mockStakingKeeper.On("GetAllDelegatorDelegations", mock.Anything, mock.Anything).Return([]stakingtypes.Delegation{delegation}, nil)
		mockStakingKeeper.On("BondDenom", mock.Anything).Return("ujuno", nil)
		mockStakingKeeper.On("GetValidator", mock.Anything, mock.Anything).Return(zeroTokenValidator, nil)

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
		handler := staking.NewDelegationsHandler(mockKeeper, deps)

		// Create test server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			router := mux.NewRouter()
			router.HandleFunc("/staking/delegations/{delegator}", handler.Handle)
			router.ServeHTTP(w, r)
		}))
		defer ts.Close()

		// Connect via WebSocket
		url := "ws" + ts.URL[4:] + "/staking/delegations/" + validDelegatorAddress
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