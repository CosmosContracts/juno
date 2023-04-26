package params

const (
	// Name defines the application name of the Juno network.
	Name = "ujuno"

	// BondDenom defines the native staking token denomination.
	BondDenom = "ujuno"

	// DisplayDenom defines the name, symbol, and display value of the Juno token.
	DisplayDenom = "JUNO"

	// DefaultGasLimit - set to the same value as cosmos-sdk flags.DefaultGasLimit
	// this value is currently only used in tests.
	DefaultGasLimit = 200000

	DefaultWeightMsgCreateDenom      int = 100
	DefaultWeightMsgMint             int = 100
	DefaultWeightMsgBurn             int = 100
	DefaultWeightMsgChangeAdmin      int = 100
	DefaultWeightMsgSetDenomMetadata int = 100
	DefaultWeightMsgForceTransfer    int = 100
)
