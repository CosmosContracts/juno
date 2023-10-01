package types

var ParamsKey = []byte{0x00}

const (
	ModuleName = "cw-hooks"

	StoreKey = ModuleName

	QuerierRoute = ModuleName

	RouterKey = ModuleName
)

var (
	KeyPrefixStakingRegister = []byte{0x01}
	KeyPrefixGovRegister     = []byte{0x02}
)
