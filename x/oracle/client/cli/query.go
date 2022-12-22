package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/CosmosContracts/juno/v12/x/oracle/types"
	"github.com/CosmosContracts/juno/v12/x/oracle/util"
)

// GetQueryCmd returns the CLI query commands for the x/oracle module.
func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryAggregatePrevote(),
		GetCmdQueryAggregateVote(),
		GetCmdQueryParams(),
		GetCmdQueryExchangeRates(),
		GetCmdQueryExchangeRate(),
		GetCmdQueryFeederDelegation(),
		GetCmdQueryMissCounter(),
		GetCmdQuerySlashWindow(),
		GetCmdQueryAllPriceHistory(),
	)

	return cmd
}

// GetCmdQueryParams implements the query params command.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current Oracle params",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParams{})
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryAggregateVote implements the query aggregate prevote of the
// validator command.
func GetCmdQueryAggregateVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aggregate-votes [validator]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "Query outstanding oracle aggregate votes",
		Long: strings.TrimSpace(`
Query outstanding oracle aggregate vote.

$ junod query oracle aggregate-votes

Or, you can filter with voter address

$ junod query oracle aggregate-votes junovaloper...
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			query := types.QueryAggregateVote{}

			if len(args) > 0 {
				validator, err := sdk.ValAddressFromBech32(args[0])
				if err != nil {
					return err
				}
				query.ValidatorAddr = validator.String()
			}

			res, err := queryClient.AggregateVote(cmd.Context(), &query)
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryAggregatePrevote implements the query aggregate prevote of the
// validator command.
func GetCmdQueryAggregatePrevote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aggregate-prevotes [validator]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "Query outstanding oracle aggregate prevotes",
		Long: strings.TrimSpace(`
Query outstanding oracle aggregate prevotes.

$ junod query oracle aggregate-prevotes

Or, can filter with voter address

$ junod query oracle aggregate-prevotes junovaloper...
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			query := types.QueryAggregatePrevote{}

			if len(args) > 0 {
				validator, err := sdk.ValAddressFromBech32(args[0])
				if err != nil {
					return err
				}
				query.ValidatorAddr = validator.String()
			}

			res, err := queryClient.AggregatePrevote(cmd.Context(), &query)
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryExchangeRates implements the query rate command.
func GetCmdQueryExchangeRates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rates",
		Args:  cobra.NoArgs,
		Short: "Query the exchange rates",
		Long: strings.TrimSpace(`
Query the current exchange rates of assets based on USD.
You can find the current list of active denoms by running

$ junod query oracle exchange-rates
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ExchangeRates(cmd.Context(), &types.QueryExchangeRates{})
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryExchangeRates implements the query rate command.
func GetCmdQueryExchangeRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rate [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the exchange rates",
		Long: strings.TrimSpace(`
Query the current exchange rates of an asset based on USD.

$ junod query oracle exchange-rate ATOM
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ExchangeRates(
				cmd.Context(),
				&types.QueryExchangeRates{
					Denom: args[0],
				},
			)
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryFeederDelegation implements the query feeder delegation command.
func GetCmdQueryFeederDelegation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feeder-delegation [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the current delegate for a given validator address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			if _, err = sdk.ValAddressFromBech32(args[0]); err != nil {
				return err
			}
			res, err := queryClient.FeederDelegation(cmd.Context(), &types.QueryFeederDelegation{
				ValidatorAddr: args[0],
			})
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryMissCounter implements the miss counter query command.
func GetCmdQueryMissCounter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "miss-counter [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the current miss counter for a given validator address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			if _, err = sdk.ValAddressFromBech32(args[0]); err != nil {
				return err
			}
			res, err := queryClient.MissCounter(cmd.Context(), &types.QueryMissCounter{
				ValidatorAddr: args[0],
			})
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQuerySlashWindow implements the slash window query command.
func GetCmdQuerySlashWindow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slash-window",
		Short: "Query the current slash window progress",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.SlashWindow(cmd.Context(), &types.QuerySlashWindow{})
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdQueryPriceTrackingLists() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price-tracking-lists",
		Short: "Query current price tracking list",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.PriceTrackingLists(cmd.Context(), &types.QueryPriceTrackingLists{})
			if err != nil {
				return err
			}
			return util.PrintOrErr(res, err, clientCtx)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdQueryPriceHistoryAt() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price-history-at [symbolDenom] [blockHeight]",
		Short: "Query price history at specific block height",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			blockHeight, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryPriceHistoryAt{
				Denom:       strings.ToUpper(args[0]),
				BlockHeight: blockHeight,
			}
			res, err := queryClient.PriceHistoryAt(cmd.Context(), req)
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdQueryCurrentVotePeriodCount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current-vote-period-count [denom]",
		Short: "Query current vote period count",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryCurrentVotePeriodCount{
				Denom: strings.ToUpper(args[0]),
			}
			res, err := queryClient.CurrentVotePeriodCount(cmd.Context(), req)
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdQueryAllPriceHistory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all-price-history [denom]",
		Short: "Show all price history storage on chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAllPriceHistory{
				Denom:      strings.ToUpper(args[0]),
				Pagination: pageReq,
			}

			res, err := queryClient.AllPriceHistory(cmd.Context(), req)
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
