package client

import (
	"github.com/CosmosContracts/juno/v13/x/oracle/client/cli"
	"github.com/CosmosContracts/juno/v13/x/oracle/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	ProposalHandlerAddTrackingPriceHistory              = govclient.NewProposalHandler(cli.ProposalAddTrackingPriceHistoryCmd, rest.EmptyRestHandler)
	ProposalHandlerAddTrackingPriceHistoryWithWhitelist = govclient.NewProposalHandler(cli.ProposalAddTrackingPriceHistoryWithWhitelistCmd, rest.EmptyRestHandler)
	ProposalRemoveTrackingPriceHistory                  = govclient.NewProposalHandler(cli.ProposalRemoveTrackingPriceHistoryCmd, rest.EmptyRestHandler)
)
