package types

import (
	"errors"
	"fmt"
	stdmath "math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// static validation errors (err113-friendly; also matchable via errors.Is)
var (
	ErrDuplicateLSTEntry = errors.New("duplicate LST allowlist address")
	ErrParamOutOfRange   = errors.New("parameter exceeds int64 range")
)

// Validate stateless-checks a Params value. Called from MsgUpdateParams
// handling and genesis validation so malformed params can never be
// committed:
//   - every LST allowlist entry must be a valid bech32 account address
//     (a bad entry would make computeTotalPower / IsLST comparisons
//     silently miss the intended contract);
//   - allowlist entries must be unique (duplicates would double-subtract
//     the LST's stake from TotalPower);
//   - RetentionWindowHeights and PruneInterval must fit in an int64, as
//     the prune path casts them to block-height arithmetic.
func (p Params) Validate() error {
	seen := make(map[string]struct{}, len(p.LstAllowlist))
	for _, entry := range p.LstAllowlist {
		if _, err := sdk.AccAddressFromBech32(entry); err != nil {
			return fmt.Errorf("invalid LST allowlist address %q: %w", entry, err)
		}
		if _, dup := seen[entry]; dup {
			return fmt.Errorf("%w: %q", ErrDuplicateLSTEntry, entry)
		}
		seen[entry] = struct{}{}
	}
	if p.RetentionWindowHeights > uint64(stdmath.MaxInt64) {
		return fmt.Errorf("%w: retention_window_heights %d", ErrParamOutOfRange, p.RetentionWindowHeights)
	}
	if p.PruneInterval > uint64(stdmath.MaxInt64) {
		return fmt.Errorf("%w: prune_interval %d", ErrParamOutOfRange, p.PruneInterval)
	}
	return nil
}
