package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v17/x/feepay/types"
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
		NewFundFeePayContract(),
	)
	return txCmd
}

// NewRegisterFeeShare returns a CLI command handler for registering a
// contract for fee pay.
func NewRegisterFeePayContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [contract_bech32] [wallet_limit]",
		Short: "Register a contract for fee pay. Only the contract admin can register a contract.",
		Long:  "Register a contract for fee pay. Only the contract admin can register a contract.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer_address := cliCtx.GetFromAddress()
			contract_address := args[0]
			wallet_limit := args[1]
			dec_limit, err := strconv.ParseUint(wallet_limit, 10, 64)

			// todo bech32 validation

			if err != nil {
				return err
			}

			fpc := &types.FeePayContract{
				ContractAddress: contract_address,
				Balance:         uint64(0),
				Limit:           dec_limit,
			}

			msg := &types.MsgRegisterFeePayContract{
				SenderAddress: deployer_address.String(),
				Contract:      fpc,
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

			sender_address := cliCtx.GetFromAddress()
			contract_address := args[0]
			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgFundFeePayContract{
				SenderAddress:   sender_address.String(),
				ContractAddress: contract_address,
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
