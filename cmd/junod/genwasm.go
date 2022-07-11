package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
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
		wasmcli.GenesisStoreCodeCmd(defaultNodeHome, genesisIO),
		wasmcli.GenesisInstantiateContractCmd(defaultNodeHome, genesisIO),
		wasmcli.GenesisExecuteContractCmd(defaultNodeHome, genesisIO),
		wasmcli.GenesisListContractsCmd(defaultNodeHome, genesisIO),
		wasmcli.GenesisListCodesCmd(defaultNodeHome, genesisIO),
	)
	return txCmd
}

// Generate contents for `app.toml`. Take the default template and config, append custom parameters
func initAppConfig() (string, interface{}) {
	template := serverconfig.DefaultConfigTemplate
	cfg := serverconfig.DefaultConfig()

	// The SDK's default minimum gas price is set to "" (empty value) inside app.toml. If left empty
	// by validators, the node will halt on startup. However, the chain developer can set a default
	// app.toml value for their validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their own app.toml to override,
	// or use this default value.
	cfg.MinGasPrices = "0.0025ujuno"

	return template, cfg
}
