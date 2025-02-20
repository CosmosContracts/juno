package common

import (
	"github.com/cosmos/cosmos-sdk/server"
)

// DebugAppOptions is a stub implementing AppOptions
type DebugAppOptions struct{}

// Get implements AppOptions
func (ao DebugAppOptions) Get(o string) interface{} {
	if o == server.FlagTrace {
		return true
	}
	return nil
}

func IsDebugLogEnabled() bool {
	return true
}
