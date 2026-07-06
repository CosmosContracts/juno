package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
)

func isNotFound(err error) bool { return errors.Is(err, collections.ErrNotFound) }

// Sentinel errors for the capped range query. The gRPC layer maps these
// to InvalidArgument / ResourceExhausted; the wasmbinding surfaces them
// verbatim to the calling contract.
var (
	ErrRangeTooWide     = fmt.Errorf("voting power range too wide: max %d blocks", MaxVotingPowerRangeWidth)
	ErrRangeTooManyRows = fmt.Errorf("voting power range matched too many snapshots: max %d rows, narrow the range", MaxVotingPowerRangeRows)
)
