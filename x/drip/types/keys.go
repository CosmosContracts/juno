package types

const (
	// module name
	ModuleName = "drip"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// KVStore key prefixes
var (
	ParamsKey = []byte{0x00} // Prefix for params key
)
