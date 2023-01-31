package cmd

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
)

func AddGenesisWasmMsgCmd(defaultNodeHome string) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "add-wasm-genesis-message",
		Short:                      "Wasm genesis subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	genesisIO := wasmcli.NewDefaultGenesisIO()
	txCmd.AddCommand(
		wasmcli.StoreCodeCmd(defaultNodeHome, genesisIO),
		wasmcli.InstantiateContractCmd(defaultNodeHome, genesisIO),
		wasmcli.ExecuteContractCmd(defaultNodeHome, genesisIO),
		wasmcli.GenesisListContractsCmd(defaultNodeHome, genesisIO),
		wasmcli.ListCodesCmd(defaultNodeHome, genesisIO),
	)
	return txCmd
}
