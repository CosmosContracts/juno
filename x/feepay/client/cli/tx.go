package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

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
