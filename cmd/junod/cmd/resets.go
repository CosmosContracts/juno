package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	tmos "github.com/cometbft/cometbft/libs/os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
)

// Cmd creates a main CLI command
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

// ResetWasmCmd removes the database of the specified Tendermint core instance.
var ResetWasmCmd = &cobra.Command{
	Use:   "wasm",
	Short: "Reset WASM files",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		clientCtx := client.GetClientContextFromCmd(cmd)
		serverCtx := server.GetServerContextFromCmd(cmd)
		config := serverCtx.Config

		config.SetRoot(clientCtx.HomeDir)

		return resetWasm(config.DBDir())
	},
}

// ResetAppCmd removes the database of the specified Tendermint core instance.
var ResetAppCmd = &cobra.Command{
	Use:   "app",
	Short: "Reset App files",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
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
			fmt.Println("Removed wasm", "dir", wasmDir)
		} else {
			return fmt.Errorf("error removing wasm  dir: %s; err: %w", wasmDir, err)
		}
	}

	if err := tmos.EnsureDir(wasmDir, 0o700); err != nil {
		return fmt.Errorf("unable to recreate wasm %w", err)
	}
	return nil
}

// resetApp removes application.db files
func resetApp(dbDir string) error {
	appDir := filepath.Join(dbDir, "application.db")

	if tmos.FileExists(appDir) {
		if err := os.RemoveAll(appDir); err == nil {
			fmt.Println("Removed application.db", "dir", appDir)
		} else {
			return fmt.Errorf("error removing application.db  dir: %s; err: %w", appDir, err)
		}
	}

	if err := tmos.EnsureDir(appDir, 0o700); err != nil {
		return fmt.Errorf("unable to recreate application.db %w", err)
	}
	return nil
}
