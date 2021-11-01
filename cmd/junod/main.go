package main

import (
	"os"

	"github.com/CosmosContracts/juno/app"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	tmcmds "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/spm/cosmoscmd"
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
	rootCmd.AddCommand(tmcmds.RollbackStateCmd)
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
