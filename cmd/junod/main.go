package main

import (
	"os"

	"github.com/CosmosContracts/juno/v9/app"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"
)

func main() {
	cmdOptions := GetWasmCmdOptions()
	rootCmd, _ := cosmoscmd.NewRootCmd(
		app.Name,
		app.AccountAddressPrefix,
		app.DefaultNodeHome,
		app.Name,
		app.ModuleBasics,
		app.New,
		// this line is used by starport scaffolding # root/arguments
		cmdOptions...,
	)

	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
