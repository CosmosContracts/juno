package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	errorsmod "cosmossdk.io/errors"
)

// defaultStreamConfig returns the default stream configuration as TOML
func defaultStreamConfig() string {
	return `#######################################################
###            Stream Module Configuration          ###
#######################################################

[stream]
# Buffer size for intake channel that receives state events
# Default: 1000
intake_buffer_size = 1000

# Buffer size for subscription channels per client
# Default: 32
subscription_buffer_size = 32

# Enable UUID-based connection tracking for better connection management
# When enabled, each connection gets a unique ID regardless of IP address
# Default: true
enable_connection_uuid = true

# Connection timeout duration (e.g., "60s", "5m")
# Default: 60s
connection_timeout = "60s"

# Circuit breaker prevents cascading failures by temporarily blocking failing connections
# Default: true
circuit_breaker_enabled = true

# Number of consecutive failures before circuit breaker opens
# Default: 5
circuit_breaker_threshold = 5

# Duration to wait before attempting to close circuit breaker (e.g., "30s", "1m")
# Default: 30s
circuit_breaker_timeout = "30s"
`
}

// AppendStreamConfigToFile appends stream configuration to config.toml if it doesn't exist
func AppendStreamConfigToFile(configPath string) error {
	// Read existing config
	content, err := os.ReadFile(configPath)
	if err != nil {
		return errorsmod.Wrap(err, "failed to read config file")
	}

	// Check if stream config already exists
	if strings.Contains(string(content), "[stream]") {
		return nil
	}

	// Append stream config
	streamConfig := defaultStreamConfig()
	newContent := string(content) + "\n" + streamConfig

	// Write back to file
	if err := os.WriteFile(configPath, []byte(newContent), 0o600); err != nil {
		return errorsmod.Wrap(err, "failed to write config file")
	}

	return nil
}

// EnsureStreamConfig ensures stream configuration exists in config.toml
func EnsureStreamConfig(homeDir string) error {
	configPath := filepath.Join(homeDir, "config", "config.toml")

	// Check if config.toml exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config doesn't exist yet, will be created by init command
		return nil
	}

	// Append stream config if needed
	return AppendStreamConfigToFile(configPath)
}

// WrapInitCmd wraps the init command to add stream configuration after initialization
func WrapInitCmd(initCmd *cobra.Command) *cobra.Command {
	originalRunE := initCmd.RunE

	initCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Run the original init command
		if err := originalRunE(cmd, args); err != nil {
			return err
		}

		// Get home directory
		homeDir, err := cmd.Flags().GetString("home")
		if err != nil {
			return err
		}

		// Add stream configuration to config.toml
		if err := EnsureStreamConfig(homeDir); err != nil {
			// Don't fail init if we can't add stream config
			cmd.Printf("Warning: failed to add stream configuration to config.toml: %v\n", err)
		} else {
			cmd.Println("Stream module configuration added to config.toml")
		}

		return nil
	}

	return initCmd
}
