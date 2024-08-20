package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmosContracts/juno/v24/x/drip/types"
)

// NewTxCmd returns a root CLI command handler for certain modules transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Drip subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewDistributeToken(),
	)
	return txCmd
}

// NewDistributeToken returns a CLI command handler for distributing tokens.
func NewDistributeToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "distribute-tokens [amount]",
		Short: "Distribute tokens to all stakers in the next block.",
		Long:  "Distribute tokens to all stakers in the next block **NOTE** ALL the tokens sent will be distributed to stakers in one shot at the next block. If you want to do a gradual airdrop, execute this transaction multiple times splitting the amount. This message can be executed only by authorized addresses.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()
			amount, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := &types.MsgDistributeTokens{
				SenderAddress: sender.String(),
				Amount:        amount,
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
