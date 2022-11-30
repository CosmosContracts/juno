package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v12/x/feeshare/types"
)

// NewTxCmd returns a root CLI command handler for certain modules/revenue
// transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "revenue subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewRegisterRevenue(),
		NewCancelRevenue(),
		NewUpdateRevenue(),
	)
	return txCmd
}

// NewRegisterRevenue returns a CLI command handler for registering a
// contract for fee distribution
func NewRegisterRevenue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [contract_bech32] [withdraw_bech32]",
		Short: "Register a contract for fee distribution. Only the contract admin can register a contract.",
		Long:  "Register a contract for feeshare distribution. **NOTE** Please ensure, that the admin of the contract (or the DAO/factory that deployed the contract) is an account that is owned by your project, to avoid that an individual admin who leaves your project becomes malicious.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			contract := args[0]
			if _, err := sdk.AccAddressFromBech32(contract); err != nil {
				return fmt.Errorf("invalid contract hex address %w", err)
			}

			withdrawer := args[1]
			if _, err := sdk.AccAddressFromBech32(withdrawer); err != nil {
				return fmt.Errorf("invalid withdrawer bech32 address %w", err)
			}

			msg := &types.MsgRegisterRevenue{
				ContractAddress:   contract,
				DeployerAddress:   deployer.String(),
				WithdrawerAddress: withdrawer,
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

// NewCancelRevenue returns a CLI command handler for canceling a
// contract for fee distribution
func NewCancelRevenue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [contract_bech32]",
		Short: "Cancel a contract from feeshare distribution",
		Long:  "Cancel a contract from feeshare distribution. The withdraw address will no longer receive fees from users interacting with the contract.\nOnly the contract admin can cancel a contract.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			contract := args[0]
			if _, err := sdk.AccAddressFromBech32(contract); err != nil {
				return fmt.Errorf("invalid contract bech32 address %w", err)
			}

			msg := &types.MsgCancelRevenue{
				ContractAddress: contract,
				DeployerAddress: deployer.String(),
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

// NewUpdateRevenue returns a CLI command handler for updating the withdraw
// address of a contract for fee distribution
func NewUpdateRevenue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [contract_bech32] [",
		Short: "Update withdrawer address for a contract registered for feeshare distribution.",
		Long:  "Update withdrawer address for a contract registered for feeshare distribution. \nOnly the contract admin can update the withdrawer address.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			contract := args[0]
			if _, err := sdk.AccAddressFromBech32(contract); err != nil {
				return fmt.Errorf("invalid contract bech32 address %w", err)
			}

			withdrawer := args[1]
			if _, err := sdk.AccAddressFromBech32(withdrawer); err != nil {
				return fmt.Errorf("invalid withdrawer bech32 address %w", err)
			}

			msg := &types.MsgUpdateRevenue{
				ContractAddress:   contract,
				DeployerAddress:   deployer.String(),
				WithdrawerAddress: withdrawer,
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
