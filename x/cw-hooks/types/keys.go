package types

var ParamsKey = []byte{0x00}

const (
	ModuleName = "cw-hooks"

	StoreKey = ModuleName

	QuerierRoute = ModuleName

	RouterKey = ModuleName
)

var (
	KeyPrefixStaking = []byte{0x01}
	KeyPrefixGov     = []byte{0x02}
)
