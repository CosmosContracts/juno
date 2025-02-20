package types

const (
	// module name
	ModuleName = "drip"
	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName
)

// KVStore key prefixes
var (
	ParamsKey = []byte{0x00} // Prefix for params key
)
