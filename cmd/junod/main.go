package main

import (
	"os"

	"github.com/CosmosContracts/juno/app"
	"github.com/CosmosContracts/juno/cmd/junod/cmd"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
