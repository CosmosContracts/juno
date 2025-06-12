package common

import (
	"context"
	"net/http"
	"time"

	"cosmossdk.io/log"
)

// StreamConfig defines the configuration for the stream module
type StreamConfig struct {
	IntakeBufferSize        int
	SubscriptionBufferSize  int
	EnableConnectionUUID    bool
	CircuitBreakerEnabled   bool
	CircuitBreakerThreshold int
	CircuitBreakerTimeout   time.Duration
}

// Logger defines the logger interface
type Logger interface {
	Info(msg string, keyvals ...any)
	Error(msg string, keyvals ...any)
	Debug(msg string, keyvals ...any)
	Warn(msg string, keyvals ...any)
}

// ConnectionManager defines the connection manager interface
type ConnectionManager interface {
	CheckConnectionLimits(w http.ResponseWriter, r *http.Request) bool
	RegisterConnectionWithHeaders(remoteAddr, xForwardedFor string) string
	UnregisterConnection(connectionID string)
	AddSubscription(connectionID string) bool
	RemoveSubscription(connectionID string)
}

// SubscriptionRegistry defines the subscription registry interface
type SubscriptionRegistry interface {
	Subscribe(key SubscriptionKey, ctx context.Context, sendCh chan<- any) Subscriber
	Unsubscribe(subscriber Subscriber)
}

// SubscriptionKey represents the key for a subscription
type SubscriptionKey interface {
	String() string
}

// Subscriber defines the subscriber interface
type Subscriber interface {
	// Add methods as needed
}

// CircuitBreaker defines the circuit breaker interface
type CircuitBreaker interface {
	AllowRequest(connectionID string) (bool, error)
	RecordSuccess(connectionID string)
	RecordFailure(connectionID string)
}

// LoggerAdapter adapts the cosmos SDK logger to our interface
type LoggerAdapter struct {
	logger log.Logger
}

// NewLoggerAdapter creates a new logger adapter
func NewLoggerAdapter(logger log.Logger) Logger {
	return &LoggerAdapter{logger: logger}
}

func (l *LoggerAdapter) Info(msg string, keyvals ...any) {
	l.logger.Info(msg, keyvals...)
}

func (l *LoggerAdapter) Error(msg string, keyvals ...any) {
	l.logger.Error(msg, keyvals...)
}

func (l *LoggerAdapter) Debug(msg string, keyvals ...any) {
	l.logger.Debug(msg, keyvals...)
}

func (l *LoggerAdapter) Warn(msg string, keyvals ...any) {
	l.logger.Warn(msg, keyvals...)
}
