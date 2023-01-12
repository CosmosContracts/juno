package client

import (
	"github.com/CosmosContracts/juno/v12/x/oracle/client/cli"
	"github.com/CosmosContracts/juno/v12/x/oracle/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	ProposalHandlerAddTrackingPriceHistory               = govclient.NewProposalHandler(cli.ProposalAddTrackingPriceHistoryCmd, rest.EmptyRestHandler)
	ProposalHandlerAddTrackingPriceHistoryWithAcceptList = govclient.NewProposalHandler(cli.ProposalAddTrackingPriceHistoryWithAcceptListCmd, rest.EmptyRestHandler)
	ProposalRemoveTrackingPriceHistory                   = govclient.NewProposalHandler(cli.ProposalRemoveTrackingPriceHistoryCmd, rest.EmptyRestHandler)
)
