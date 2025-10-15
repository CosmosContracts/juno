package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	tmos "github.com/cometbft/cometbft/libs/os"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
)

// ResetCmd creates a main CLI command to reset different parts of application state
func ResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset commands for different parts of application state",
		RunE:  client.ValidateCmd,
	}

	cmd.AddCommand(ResetWasmCmd)
	cmd.AddCommand(ResetAppCmd)

	return cmd
}

// ResetWasmCmd removes the database of the specified Comet core instance.
var ResetWasmCmd = &cobra.Command{
	Use:   "wasm",
	Short: "Reset WASM files",
	RunE: func(cmd *cobra.Command, _ []string) (err error) {
		clientCtx := client.GetClientContextFromCmd(cmd)
		serverCtx := server.GetServerContextFromCmd(cmd)
		config := serverCtx.Config

		config.SetRoot(clientCtx.HomeDir)

		return resetWasm(config.DBDir())
	},
}

// ResetAppCmd removes the database of the specified Comet core instance.
var ResetAppCmd = &cobra.Command{
	Use:   "app",
	Short: "Reset App files",
	RunE: func(cmd *cobra.Command, _ []string) (err error) {
		clientCtx := client.GetClientContextFromCmd(cmd)
		serverCtx := server.GetServerContextFromCmd(cmd)
		config := serverCtx.Config

		config.SetRoot(clientCtx.HomeDir)

		return resetApp(config.DBDir())
	},
}

// resetWasm removes wasm files
func resetWasm(dbDir string) error {
	wasmDir := filepath.Join(dbDir, "wasm")

	if tmos.FileExists(wasmDir) {
		if err := os.RemoveAll(wasmDir); err == nil {
			_, err = fmt.Println("Removed wasm", "dir", wasmDir)
			if err != nil {
				return errorsmod.Wrapf(err, "error removing wasm dir: %s", wasmDir)
			}
		} else {
			return errorsmod.Wrapf(err, "error removing wasm dir: %s", wasmDir)
		}
	}

	if err := tmos.EnsureDir(wasmDir, 0o700); err != nil {
		return errorsmod.Wrap(err, "unable to recreate wasm")
	}
	return nil
}

// resetApp removes application.db files
func resetApp(dbDir string) error {
	appDir := filepath.Join(dbDir, "application.db")

	if tmos.FileExists(appDir) {
		if err := os.RemoveAll(appDir); err == nil {
			_, err = fmt.Println("Removed application.db", "dir", appDir)
			if err != nil {
				return errorsmod.Wrapf(err, "error removing application.db dir: %s", appDir)
			}
		} else {
			return errorsmod.Wrapf(err, "error removing application.db dir: %s", appDir)
		}
	}

	if err := tmos.EnsureDir(appDir, 0o700); err != nil {
		return errorsmod.Wrap(err, "unable to recreate application.db")
	}
	return nil
}
