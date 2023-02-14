package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/CosmosContracts/juno/v13/x/oracle/types"
	"github.com/CosmosContracts/juno/v13/x/oracle/util"
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
		GetCmdQueryPriceHistoryAt(),
		GetCmdQueryTwapTrackingLists(),
		GetCmdQueryTwapPrice(),
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
		Use:   "feeder-address [validator_operator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the current delegated address for a given validator operator address",
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

func GetCmdQueryTwapTrackingLists() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price-tracking-list",
		Short: "Query current price tracking list",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.TwapTrackingLists(cmd.Context(), &types.QueryTwapTrackingLists{})
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
		Use:   "price-history-at [denom] [time_stamp]",
		Short: "Query price history at specific block height",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			time, err := time.Parse(time.RFC3339, args[1])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryPriceHistoryAtTime{
				Denom: strings.ToUpper(args[0]),
				Time:  time,
			}
			res, err := queryClient.PriceHistoryAtTime(cmd.Context(), req)
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

func GetCmdQueryTwapPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "twap [denom] [start_time] [end_time]",
		Short: "Query twap between period time",
		Long: strings.TrimSpace(
			`Query twap for pool. Start and end time must be in RFC3339 format or UNIX format.
Example:
$ junod q oracle twap JUNO 2022-12-25T19:42:07.100Z 2022-12-25T20:42:07.100Z
$ junod q oracle twap JUNO 1675053795 1675093795
`),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)
			var startTime, endTime time.Time
			startTime, endTime, err := twapQueryUnixParseArgs(args)
			if err != nil {
				startTime, endTime, err = twapQueryRFCParseArgs(args)
				if err != nil {
					return err
				}
			}

			req := &types.QueryArithmeticTwapPriceBetweenTime{
				Denom:     strings.ToUpper(args[0]),
				StartTime: startTime,
				EndTime:   endTime,
			}

			res, err := queryClient.ArithmeticTwapPriceBetweenTime(cmd.Context(), req)
			return util.PrintOrErr(res, err, clientCtx)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func twapQueryRFCParseArgs(args []string) (startTime time.Time, endTime time.Time, err error) {
	// RFC3339 TIME PARSE
	startTime, err = time.Parse(time.RFC3339, args[1])
	if err != nil {
		return
	}
	endTime, err = time.Parse(time.RFC3339, args[2])
	if err != nil {
		return
	}

	return startTime, endTime, nil
}

func twapQueryUnixParseArgs(args []string) (startTime time.Time, endTime time.Time, err error) {
	startTime, err = ParseUnixTime(args[1], "start time")
	if err != nil {
		return
	}

	// try parsing in unix time, if failed try parsing in duration
	endTime, err = ParseUnixTime(args[2], "end time")
	if err != nil {
		duration, err2 := time.ParseDuration(args[2])
		if err2 != nil {
			err = err2
			return
		}
		endTime = startTime.Add(duration)
	}
	return startTime, endTime, nil
}

func ParseUnixTime(arg string, fieldName string) (time.Time, error) {
	timeUnix, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse %s as unix time for field %s: %w", arg, fieldName, err)
	}
	startTime := time.Unix(timeUnix, 0)
	return startTime, nil
}
