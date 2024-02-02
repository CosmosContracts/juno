package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v20/x/feepay/types"
)

// NewTxCmd returns a root CLI command handler for certain modules/FeeShare
// transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "FeePay subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewRegisterFeePayContract(),
		NewUnregisterFeePayContract(),
		NewFundFeePayContract(),
		NewUpdateFeePayContractWalletLimit(),
	)
	return txCmd
}

// NewRegisterFeeShare returns a CLI command handler for registering a
// contract for fee pay.
func NewRegisterFeePayContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [contract_bech32] [wallet_limit]",
		Short: "Register a contract for fee pay. Only the contract admin can register a contract.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployerAddress := cliCtx.GetFromAddress()
			contractAddress := args[0]
			walletLimit := args[1]
			decLimit, err := strconv.ParseUint(walletLimit, 10, 64)
			if err != nil {
				return err
			}

			fpc := &types.FeePayContract{
				ContractAddress: contractAddress,
				Balance:         uint64(0),
				WalletLimit:     decLimit,
			}

			msg := &types.MsgRegisterFeePayContract{
				SenderAddress:  deployerAddress.String(),
				FeePayContract: fpc,
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

// NewUnregisterFeePayContract returns a CLI command handler for
// unregistering a fee pay contract.
func NewUnregisterFeePayContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregister [contract_bech32]",
		Short: "Unregister a contract for fee pay.",
		Long:  "Unregister a contract for fee pay. All remaining funds will return to the contract admin or the creator (as a fallback).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			senderAddress := cliCtx.GetFromAddress()
			contractAddress := args[0]

			msg := &types.MsgUnregisterFeePayContract{
				SenderAddress:   senderAddress.String(),
				ContractAddress: contractAddress,
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

// NewRegisterFeeShare returns a CLI command handler for
// funding a fee pay contract.
func NewFundFeePayContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund [contract_bech32] [amount]",
		Short: "Send funds to a registered fee pay contract.",
		Long:  "Send funds to a registered fee pay contract.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			senderAddress := cliCtx.GetFromAddress()
			contractAddress := args[0]
			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgFundFeePayContract{
				SenderAddress:   senderAddress.String(),
				ContractAddress: contractAddress,
				Amount:          amount,
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

// NewUpdateFeePayContractWalletLimit returns a CLI command handler for
// updating the wallet limit of a fee pay contract.
func NewUpdateFeePayContractWalletLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-wallet-limit [contract_bech32] [wallet_limit]",
		Short: "Update the wallet limit of a fee pay contract.",
		Long:  "Update the wallet limit of a fee pay contract.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			senderAddress := cliCtx.GetFromAddress()
			contractAddress := args[0]
			walletLimit := args[1]
			decLimit, err := strconv.ParseUint(walletLimit, 10, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateFeePayContractWalletLimit{
				SenderAddress:   senderAddress.String(),
				ContractAddress: contractAddress,
				WalletLimit:     decLimit,
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
