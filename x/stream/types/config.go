package types

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DefaultSubscriptionBufferSize    = 32
	DefaultWsMaxConnections          = 200
	DefaultGrpcMaxConnections        = 200
	DefaultMaxSubscriptionsPerClient = 100

	// ModuleName defines the module name
	ModuleName = "stream"
)

type StreamConfig struct {
	SubscriptionBufferSize    uint32
	WsMaxConnections          uint32
	GrpcMaxConnections        uint32
	MaxSubscriptionsPerClient uint32
}

func LoadStreamConfig(homePath string) (StreamConfig, error) {
	cfg := StreamConfig{
		SubscriptionBufferSize:    DefaultSubscriptionBufferSize,
		WsMaxConnections:          DefaultWsMaxConnections,
		GrpcMaxConnections:        DefaultGrpcMaxConnections,
		MaxSubscriptionsPerClient: DefaultMaxSubscriptionsPerClient,
	}

	configPath := filepath.Join(homePath, "config", "config.toml")
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		var configErr viper.ConfigFileNotFoundError
		if errors.As(err, &configErr) || errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	if val := v.GetUint32("rpc.experimental_subscription_buffer_size"); v.IsSet("rpc.experimental_subscription_buffer_size") {
		cfg.SubscriptionBufferSize = val
	}
	if val := v.GetUint32("rpc.max_open_connections"); v.IsSet("rpc.max_open_connections") {
		cfg.WsMaxConnections = val
	}
	if val := v.GetUint32("rpc.grpc_max_open_connections"); v.IsSet("rpc.grpc_max_open_connections") {
		cfg.GrpcMaxConnections = val
	}
	if val := v.GetUint32("rpc.max_subscriptions_per_client"); v.IsSet("rpc.max_subscriptions_per_client") {
		cfg.MaxSubscriptionsPerClient = val
	}

	return cfg, nil
}
