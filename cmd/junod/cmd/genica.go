package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icagenesistypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/genesis/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// AddGenesisAccountCmd returns add-genesis-account cobra Command.
func AddGenesisIcaCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-ica-config",
		Short: "Add ICA config to genesis.json",
		Long:  `Add default ICA configuration to genesis.json`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			controllerGenesisState := icagenesistypes.DefaultControllerGenesis()
			// no params set in upgrade handler, no params set here
			controllerGenesisState.Params = icacontrollertypes.Params{}

			hostGenesisState := icagenesistypes.DefaultHostGenesis()
			// add the messages we want (from old upgrade handler)
			hostGenesisState.Params = icahosttypes.Params{
				HostEnabled:   true,
				AllowMessages: []string{"*"},
			}

			newIcaGenState := icagenesistypes.NewGenesisState(controllerGenesisState, hostGenesisState)

			icaGenStateBz, err := clientCtx.Codec.MarshalJSON(newIcaGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}

			appState[icatypes.ModuleName] = icaGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
