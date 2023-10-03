package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/CosmosContracts/juno/v18/x/cw-hooks/types"
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
		NewRegisterStaking(),
		NewUnregisterStaking(),
		NewRegisterGovernance(),
		NewUnregisterGovernance(),
	)
	return txCmd
}

func NewRegisterStaking() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-staking [contract_bech32]",
		Short: "Register a contract for staking sudo message updates",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			contract := args[0]

			msg := &types.MsgRegisterStaking{
				ContractAddress: contract,
				RegisterAddress: deployer.String(),
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

func NewRegisterGovernance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-governance [contract_bech32]",
		Short: "Register a contract for governance sudo message updates",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			contract := args[0]

			msg := &types.MsgRegisterGovernance{
				ContractAddress: contract,
				RegisterAddress: deployer.String(),
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

func NewUnregisterStaking() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregister-staking [contract_bech32]",
		Short: "Remove a contract for receiving staking sudo message updates",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			contract := args[0]

			msg := &types.MsgUnregisterStaking{
				ContractAddress: contract,
				RegisterAddress: deployer.String(),
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

func NewUnregisterGovernance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregister-governance [contract_bech32]",
		Short: "Remove a contract for receiving governance sudo message updates",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			contract := args[0]

			msg := &types.MsgUnregisterGovernance{
				ContractAddress: contract,
				RegisterAddress: deployer.String(),
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
