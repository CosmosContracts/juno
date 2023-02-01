package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/CosmosContracts/juno/v13/x/oracle/types"
)

// GetTxCmd returns the CLI transaction commands for the x/oracle module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Transaction commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdDelegateFeedConsent(),
		GetCmdAggregateExchangeRatePrevote(),
		GetCmdAggregateExchangeRateVote(),
	)

	return cmd
}

// GetCmdDelegateFeedConsent creates a Cobra command to generate or
// broadcast a transaction with a MsgDelegateFeedConsent message.
func GetCmdDelegateFeedConsent() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-feeder [feeder] --from [validator_key]",
		Args:  cobra.ExactArgs(1),
		Short: "Delegate the permission to vote for the oracle to an address",
		Long: strings.TrimSpace(`
Delegate the permission to submit exchange rate votes for the oracle to an address.
Delegation can keep your validator operator key offline and use a separate replaceable key online.
$ junod tx oracle set-feeder juno1... --from validator_key
where "juno1..." is the address you want to delegate your voting rights to.
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := cmd.Flags().GetString(flags.FlagFrom); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			feederAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgDelegateFeedConsent(sdk.ValAddress(clientCtx.GetFromAddress()), feederAddr)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdAggregateExchangeRatePrevote creates a Cobra command to generate or
// broadcast a transaction with a MsgAggregateExchangeRatePrevote message.
func GetCmdAggregateExchangeRatePrevote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rate-prevote [hash] [validator-address]",
		Args:  cobra.MinimumNArgs(1),
		Short: "Submit an exchange rate prevote with a hash",
		Long: fmt.Sprintf(`Submit an exchange rate prevote with a hash as a hex string
			representation of a byte array.
			Ex: junod tx oracle exchange-rate-prevote %s --from alice`,
			"19c30cf9ea8aa0e0b03904162cadec0f2024a76d"),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hash, err := types.AggregateVoteHashFromHexString(args[0])
			if err != nil {
				return err
			}

			valAddress := sdk.ValAddress(clientCtx.GetFromAddress())
			if len(args) > 1 {
				valAddress, err = sdk.ValAddressFromBech32(args[1])
				if err != nil {
					return err
				}
			}

			msg := types.NewMsgAggregateExchangeRatePrevote(
				hash,
				clientCtx.GetFromAddress(),
				valAddress,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdAggregateExchangeRateVote creates a Cobra command to generate or
// broadcast a transaction with a NewMsgAggregateExchangeRateVote message.
func GetCmdAggregateExchangeRateVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rate-vote [salt] [exchange-rates] [validator-address]",
		Args:  cobra.MinimumNArgs(2),
		Short: "Submit an exchange rate vote with the salt and exchange rate string",
		Long: fmt.Sprintf(`Submit an exchange rate vote with the salt of the previous hash, and the
			exchange rate string previously used in the hash.
			Ex: junod tx oracle exchange-rate-vote %s %s --from alice`,
			"0cf33fb528b388660c3a42c3f3250e983395290b75fef255050fb5bc48a6025f",
			"foo:1.0,bar:1232.123",
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAddress := sdk.ValAddress(clientCtx.GetFromAddress())
			if len(args) > 2 {
				valAddress, err = sdk.ValAddressFromBech32(args[2])
				if err != nil {
					return err
				}
			}

			msg := types.NewMsgAggregateExchangeRateVote(
				args[0],
				args[1],
				clientCtx.GetFromAddress(),
				valAddress,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
