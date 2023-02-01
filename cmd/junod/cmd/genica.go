package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	icacontrollertypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tokenfactorytypes "github.com/CosmWasm/token-factory/x/tokenfactory/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	feesharetypes "github.com/CosmosContracts/juno/v13/x/feeshare/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

			controllerGenesisState := icatypes.DefaultControllerGenesis()
			// no params set in upgrade handler, no params set here
			controllerGenesisState.Params = icacontrollertypes.Params{}

			hostGenesisState := icatypes.DefaultHostGenesis()
			// add the messages we want (from old upgrade handler)
			hostGenesisState.Params = icahosttypes.Params{
				HostEnabled: true,
				AllowMessages: []string{
					// bank
					sdk.MsgTypeURL(&banktypes.MsgSend{}),
					sdk.MsgTypeURL(&banktypes.MsgMultiSend{}),
					// staking
					sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
					sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
					sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
					sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
					sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
					// distribution
					sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
					sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
					sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
					sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
					// gov
					sdk.MsgTypeURL(&govtypes.MsgVote{}),
					sdk.MsgTypeURL(&govtypes.MsgVoteWeighted{}),
					sdk.MsgTypeURL(&govtypes.MsgSubmitProposal{}),
					sdk.MsgTypeURL(&govtypes.MsgDeposit{}),
					// authz
					sdk.MsgTypeURL(&authz.MsgExec{}),
					sdk.MsgTypeURL(&authz.MsgGrant{}),
					sdk.MsgTypeURL(&authz.MsgRevoke{}),
					// wasm
					sdk.MsgTypeURL(&wasmtypes.MsgStoreCode{}),
					sdk.MsgTypeURL(&wasmtypes.MsgInstantiateContract{}),
					sdk.MsgTypeURL(&wasmtypes.MsgInstantiateContract2{}),
					sdk.MsgTypeURL(&wasmtypes.MsgExecuteContract{}),
					sdk.MsgTypeURL(&wasmtypes.MsgMigrateContract{}),
					sdk.MsgTypeURL(&wasmtypes.MsgUpdateAdmin{}),
					sdk.MsgTypeURL(&wasmtypes.MsgClearAdmin{}),
					sdk.MsgTypeURL(&wasmtypes.MsgIBCSend{}),
					sdk.MsgTypeURL(&wasmtypes.MsgIBCCloseChannel{}),
					// tokenfactory
					sdk.MsgTypeURL(&tokenfactorytypes.MsgCreateDenom{}),
					sdk.MsgTypeURL(&tokenfactorytypes.MsgMint{}),
					sdk.MsgTypeURL(&tokenfactorytypes.MsgBurn{}),
					sdk.MsgTypeURL(&tokenfactorytypes.MsgChangeAdmin{}),
					sdk.MsgTypeURL(&tokenfactorytypes.MsgSetDenomMetadata{}),
					// feeshare
					sdk.MsgTypeURL(&feesharetypes.MsgRegisterFeeShare{}),
					sdk.MsgTypeURL(&feesharetypes.MsgUpdateFeeShare{}),
					sdk.MsgTypeURL(&feesharetypes.MsgUpdateFeeShare{}),
					sdk.MsgTypeURL(&feesharetypes.MsgCancelFeeShare{}),
				},
			}

			newIcaGenState := icatypes.NewGenesisState(controllerGenesisState, hostGenesisState)

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
