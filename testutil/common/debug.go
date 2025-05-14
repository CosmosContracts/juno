package common

import (
	"github.com/cosmos/cosmos-sdk/server"
)

// DebugAppOptions is a stub implementing AppOptions
type DebugAppOptions struct{}

// Get implements AppOptions
func (DebugAppOptions) Get(o string) any {
	if o == server.FlagTrace {
		return true
	}
	return nil
}

func IsDebugLogEnabled() bool {
	return true
}
