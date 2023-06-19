package main

import (
	"os"

	"cosmossdk.io/log"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/CosmosContracts/juno/v16/app"
	"github.com/CosmosContracts/juno/v16/cmd/junod/cmd"
)

func main() {
	app.SetAddressPrefixes()
	rootCmd, _ := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "JUNOD", app.DefaultNodeHome); err != nil {
		log.NewLogger(rootCmd.OutOrStderr()).Error("failure when running app", "err", err)
		os.Exit(1)
	}
}
