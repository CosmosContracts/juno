package cmd

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
)

func AddGenesisWasmMsgCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "add-wasm-genesis-message",
		Short:                      "Wasm genesis subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		wasmcli.ClearContractAdminCmd(),
		wasmcli.ExecuteContractCmd(),
		wasmcli.InstantiateContractCmd(),
		wasmcli.MigrateContractCmd(),
		wasmcli.UpdateContractAdminCmd(),
		wasmcli.GetCmdGetContractStateAll(),
	)
	return txCmd
}
