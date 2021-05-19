package main

import (
	"os"

	"github.com/CosmosContracts/Juno/app"
	"github.com/CosmosContracts/Juno/cmd/Junod/cmd"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
