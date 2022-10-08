package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	tmos "github.com/tendermint/tendermint/libs/os"
)

// ResetWasmCmd removes the database of the specified Tendermint core instance.
var ResetWasmCmd = &cobra.Command{
	Use:   "reset-app-wasm",
	Short: "Remove App and WASM files",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		clientCtx := client.GetClientContextFromCmd(cmd)
		serverCtx := server.GetServerContextFromCmd(cmd)
		config := serverCtx.Config

		config.SetRoot(clientCtx.HomeDir)

		return resetWasm(config.DBDir())
	},
}

func init() {
}

// resetWasm removes wasm files
func resetWasm(dbDir string) error {
	application := filepath.Join(dbDir, "application.db")
	wasm := filepath.Join(dbDir, "wasm")

	if tmos.FileExists(application) {
		if err := os.RemoveAll(application); err == nil {
			fmt.Println("Removed all application.db", "dir", application)
		} else {
			fmt.Println("error removing all application.db", "dir", application, "err", err)
		}
	}

	if tmos.FileExists(wasm) {
		if err := os.RemoveAll(wasm); err == nil {
			fmt.Println("Removed all wasm", "dir", wasm)
		} else {
			fmt.Println("error removing all wasm", "dir", wasm, "err", err)
		}
	}

	if err := tmos.EnsureDir(dbDir, 0700); err != nil {
		fmt.Println("unable to recreate dbDir", "err", err)
	}
	return nil
}
