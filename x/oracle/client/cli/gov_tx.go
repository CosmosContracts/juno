package cli

import (
	"os"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/spf13/cobra"
)

func ProposalAddTrackingPriceHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-tracking-price-history [json-proposal]",
		Short: "Add tracking price history list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, _, _, deposit, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			proposal, err := ParseAirdopInitialProposal(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}
			if err = proposal.ValidateBasic(); err != nil {
				return err
			}

			msg, err := govtypes.NewMsgSubmitProposal(&proposal, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			if err = msg.ValidateBasic(); err != nil {
				return nil
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

func getProposalInfo(cmd *cobra.Command) (client.Context, string, string, sdk.Coins, error) {
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return client.Context{}, "", "", nil, err
	}

	proposalTitle, err := cmd.Flags().GetString(cli.FlagTitle) //nolint:staticcheck
	if err != nil {
		return clientCtx, proposalTitle, "", nil, err
	}

	proposalDescr, err := cmd.Flags().GetString(cli.FlagDescription) //nolint:staticcheck
	if err != nil {
		return client.Context{}, proposalTitle, proposalDescr, nil, err
	}

	depositArg, err := cmd.Flags().GetString(cli.FlagDeposit)
	if err != nil {
		return client.Context{}, proposalTitle, proposalDescr, nil, err
	}

	deposit, err := sdk.ParseCoinsNormalized(depositArg)
	if err != nil {
		return client.Context{}, proposalTitle, proposalDescr, deposit, err
	}

	return clientCtx, proposalTitle, proposalDescr, deposit, nil
}

func ParseAirdopInitialProposal(cdc codec.JSONCodec, proposalFile string) (types.AddTrackingPriceHistoryProposal, error) {
	var proposal types.AddTrackingPriceHistoryProposal

	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	err = cdc.UnmarshalJSON(contents, &proposal)
	if err != nil {
		return proposal, err
	}

	return proposal, nil
}
