package keeper

import (
	"errors"

	"cosmossdk.io/collections"
)

func isNotFound(err error) bool { return errors.Is(err, collections.ErrNotFound) }
