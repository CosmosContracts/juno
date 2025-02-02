package types

const (
	prefixParamsKey = iota + 1
)

const (
	// module name
	ModuleName = "feepay"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName
)

var ParamsKey = []byte{prefixParamsKey}
