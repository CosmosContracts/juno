package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v23/x/cw-hooks/types"
)

// NewTxCmd returns a root CLI command handler for modules
// transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      types.ModuleName + "subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewRegister(),
		NewUnregister(),
	)
	return txCmd
}

func NewRegister() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [staking|governance] [contract]",
		Short: "Register a contract for sudo message updates",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			registerType := args[0]
			contract := args[1]

			var msg sdk.Msg
			switch registerType {
			case "staking", "stake":
				msg = &types.MsgRegisterStaking{
					ContractAddress: contract,
					RegisterAddress: deployer.String(),
				}
			case "governance", "gov":
				msg = &types.MsgRegisterGovernance{
					ContractAddress: contract,
					RegisterAddress: deployer.String(),
				}
			default:
				return fmt.Errorf("invalid register type: %s", registerType)
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewUnregister() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregister [staking|governance] [contract]",
		Short: "Remove a contract from receiving sudo message updates",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			registerType := args[0]
			contract := args[1]

			var msg sdk.Msg
			switch registerType {
			case "staking", "stake":
				msg = &types.MsgUnregisterStaking{
					ContractAddress: contract,
					RegisterAddress: deployer.String(),
				}
			case "governance", "gov":
				msg = &types.MsgUnregisterGovernance{
					ContractAddress: contract,
					RegisterAddress: deployer.String(),
				}
			default:
				return fmt.Errorf("invalid register type: %s", registerType)
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
