package websocket

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config contains the runtime configuration for the websocket server
type Config struct {
	CORSAllowedOrigins []string
}

// LoadConfig reads the websocket configuration from app.toml located under the node's home directory
func LoadConfig(homePath string) (Config, error) {
	configPath := filepath.Join(homePath, "config", "config.toml")
	serverConfig := viper.New()
	serverConfig.SetConfigFile(configPath)
	serverConfig.SetConfigType("toml")

	if err := serverConfig.ReadInConfig(); err != nil {
		var configErr viper.ConfigFileNotFoundError
		if !errors.As(err, &configErr) && !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
	}

	return Config{CORSAllowedOrigins: serverConfig.GetStringSlice("rpc.cors_allowed_origins")}, nil
}
