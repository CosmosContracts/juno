package keeper

import (
	"fmt"
	"time"
)

// StreamConfig holds all configuration options for the stream module
type StreamConfig struct {
	// Buffer sizes
	IntakeBufferSize       int `mapstructure:"intake_buffer_size"`
	SubscriptionBufferSize int `mapstructure:"subscription_buffer_size"`

	// Connection settings
	EnableConnectionUUID bool          `mapstructure:"enable_connection_uuid"`
	ConnectionTimeout    time.Duration `mapstructure:"connection_timeout"`

	// Circuit breaker settings
	CircuitBreakerEnabled   bool          `mapstructure:"circuit_breaker_enabled"`
	CircuitBreakerThreshold int           `mapstructure:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `mapstructure:"circuit_breaker_timeout"`

	// CORS settings (from RPC config)
	CORSAllowedOrigins []string
}

// DefaultStreamConfig returns the default configuration
func DefaultStreamConfig() StreamConfig {
	return StreamConfig{
		IntakeBufferSize:        1000,
		SubscriptionBufferSize:  32,
		EnableConnectionUUID:    true,
		ConnectionTimeout:       60 * time.Second,
		CircuitBreakerEnabled:   true,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   30 * time.Second,
		CORSAllowedOrigins:      []string{},
	}
}

// Validate validates the stream configuration
func (c *StreamConfig) Validate() error {
	if c.IntakeBufferSize <= 0 {
		return fmt.Errorf("intake_buffer_size must be positive, got %d", c.IntakeBufferSize)
	}
	if c.IntakeBufferSize > 100000 {
		return fmt.Errorf("intake_buffer_size too large (max 100000), got %d", c.IntakeBufferSize)
	}

	if c.SubscriptionBufferSize <= 0 {
		return fmt.Errorf("subscription_buffer_size must be positive, got %d", c.SubscriptionBufferSize)
	}
	if c.SubscriptionBufferSize > 1000 {
		return fmt.Errorf("subscription_buffer_size too large (max 1000), got %d", c.SubscriptionBufferSize)
	}

	if c.ConnectionTimeout <= 0 {
		return fmt.Errorf("connection_timeout must be positive, got %v", c.ConnectionTimeout)
	}
	if c.ConnectionTimeout > 5*time.Minute {
		return fmt.Errorf("connection_timeout too large (max 5m), got %v", c.ConnectionTimeout)
	}

	if c.CircuitBreakerEnabled {
		if c.CircuitBreakerThreshold <= 0 {
			return fmt.Errorf("circuit_breaker_threshold must be positive when enabled, got %d", c.CircuitBreakerThreshold)
		}
		if c.CircuitBreakerTimeout <= 0 {
			return fmt.Errorf("circuit_breaker_timeout must be positive when enabled, got %v", c.CircuitBreakerTimeout)
		}
	}

	return nil
}

// ApplyDefaults fills in any zero values with defaults
func (c *StreamConfig) ApplyDefaults() {
	defaults := DefaultStreamConfig()

	if c.IntakeBufferSize == 0 {
		c.IntakeBufferSize = defaults.IntakeBufferSize
	}
	if c.SubscriptionBufferSize == 0 {
		c.SubscriptionBufferSize = defaults.SubscriptionBufferSize
	}
	if c.ConnectionTimeout == 0 {
		c.ConnectionTimeout = defaults.ConnectionTimeout
	}
	if c.CircuitBreakerTimeout == 0 {
		c.CircuitBreakerTimeout = defaults.CircuitBreakerTimeout
	}
	if c.CircuitBreakerThreshold == 0 {
		c.CircuitBreakerThreshold = defaults.CircuitBreakerThreshold
	}
}
