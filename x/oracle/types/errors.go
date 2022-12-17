package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// Oracle sentinel errors
var (
	ErrInvalidExchangeRate   = sdkerrors.Register(ModuleName, 1, "invalid exchange rate")
	ErrNoPrevote             = sdkerrors.Register(ModuleName, 2, "no prevote")
	ErrNoVote                = sdkerrors.Register(ModuleName, 3, "no vote")
	ErrNoVotingPermission    = sdkerrors.Register(ModuleName, 4, "unauthorized voter")
	ErrInvalidHash           = sdkerrors.Register(ModuleName, 5, "invalid hash")
	ErrInvalidHashLength     = sdkerrors.Register(ModuleName, 6, fmt.Sprintf("invalid hash length; should equal %d", tmhash.TruncatedSize)) //nolint: lll
	ErrVerificationFailed    = sdkerrors.Register(ModuleName, 7, "hash verification failed")
	ErrRevealPeriodMissMatch = sdkerrors.Register(ModuleName, 8, "reveal period of submitted vote does not match with registered prevote") //nolint: lll
	ErrInvalidSaltLength     = sdkerrors.Register(ModuleName, 9, "invalid salt length; must be 64")
	ErrInvalidSaltFormat     = sdkerrors.Register(ModuleName, 10, "invalid salt format")
	ErrNoAggregatePrevote    = sdkerrors.Register(ModuleName, 11, "no aggregate prevote")
	ErrNoAggregateVote       = sdkerrors.Register(ModuleName, 12, "no aggregate vote")
	ErrUnknownDenom          = sdkerrors.Register(ModuleName, 13, "unknown denom")
	ErrNegativeOrZeroRate    = sdkerrors.Register(ModuleName, 14, "invalid exchange rate; should be positive")
	ErrExistingPrevote       = sdkerrors.Register(ModuleName, 15, "prevote already submitted for this voting period")
	ErrBallotNotSorted       = sdkerrors.Register(ModuleName, 16, "ballot must be sorted before this operation")
	ErrInvalidVotePeriod     = sdkerrors.Register(ModuleName, 17, "invalid voting period")
	ErrEmpty                 = sdkerrors.Register(ModuleName, 18, "empty")

	// 4XX = Price Sensitive
	ErrInvalidOraclePrice = sdkerrors.Register(ModuleName, 401, "invalid oracle price")
)
