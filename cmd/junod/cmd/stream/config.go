package stream

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	errorsmod "cosmossdk.io/errors"
)

const streamingWSConfig = `
[streaming.ws]
# Circuit breaker prevents cascading failures by temporarily blocking failing connections.
circuit_breaker_enabled = true

# Number of consecutive failures before circuit breaker opens.
circuit_breaker_threshold = 5

# Duration to wait before attempting to close the circuit breaker (e.g., "30s", "1m").
circuit_breaker_timeout = "30s"
`

func ensureStreamingWSSection(configPath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return errorsmod.Wrap(err, "failed to read app.toml")
	}

	if bytes.Contains(content, []byte("[streaming.ws]")) {
		return nil
	}

	insert := []byte(streamingWSConfig)
	// Ensure the snippet ends with a newline so subsequent sections keep their alignment.
	if len(insert) == 0 || insert[len(insert)-1] != '\n' {
		insert = append(insert, '\n')
	}

	// Try to place the new section immediately before the mempool section.
	mempoolIdx := bytes.Index(content, []byte("\n[mempool]"))
	var updated bytes.Buffer

	writeBytes := func(data []byte) error {
		if _, err := updated.Write(data); err != nil {
			return err
		}
		return nil
	}

	writeString := func(value string) error {
		if _, err := updated.WriteString(value); err != nil {
			return err
		}
		return nil
	}

	switch {
	case mempoolIdx >= 0:
		if err := writeBytes(content[:mempoolIdx]); err != nil {
			return errorsmod.Wrap(err, "prepare streaming config prefix")
		}
		if !bytes.HasSuffix(updated.Bytes(), []byte("\n\n")) {
			if err := writeString("\n"); err != nil {
				return errorsmod.Wrap(err, "ensure spacing before streaming config")
			}
		}
		if err := writeBytes(insert); err != nil {
			return errorsmod.Wrap(err, "insert streaming config")
		}
		if !bytes.HasSuffix(updated.Bytes(), []byte("\n")) {
			if err := writeString("\n"); err != nil {
				return errorsmod.Wrap(err, "ensure trailing newline after streaming config")
			}
		}
		if err := writeBytes(content[mempoolIdx:]); err != nil {
			return errorsmod.Wrap(err, "append remainder of config")
		}
	default:
		if err := writeBytes(content); err != nil {
			return errorsmod.Wrap(err, "copy existing config")
		}
		if len(content) == 0 || content[len(content)-1] != '\n' {
			if err := writeString("\n"); err != nil {
				return errorsmod.Wrap(err, "ensure newline before streaming config")
			}
		}
		if err := writeBytes(insert); err != nil {
			return errorsmod.Wrap(err, "append streaming config")
		}
	}

	return os.WriteFile(configPath, updated.Bytes(), 0o600)
}

// EnsureStreamConfig ensures the websocket streaming configuration exists in app.toml.
func EnsureStreamConfig(homeDir string) error {
	configPath := filepath.Join(homeDir, "config", "app.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	return ensureStreamingWSSection(configPath)
}

// WrapInitCmd wraps the init command to add stream configuration after initialization.
func WrapInitCmd(initCmd *cobra.Command) *cobra.Command {
	originalRunE := initCmd.RunE

	initCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := originalRunE(cmd, args); err != nil {
			return err
		}

		homeDir, err := cmd.Flags().GetString("home")
		if err != nil {
			return err
		}

		if err := EnsureStreamConfig(homeDir); err != nil {
			cmd.Printf("Warning: failed to add websocket streaming configuration to app.toml: %v\n", err)
		} else {
			cmd.Println("Websocket streaming configuration added to app.toml")
		}

		return nil
	}

	return initCmd
}
